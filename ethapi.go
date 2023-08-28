// Copyright 2021 TikBridge
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/abi"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/ThinkiumGroup/go-common/math"
	"github.com/ThinkiumGroup/go-tkmrpc/client"
	"github.com/ethereum/go-ethereum"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	reqTimeOut = 5 * time.Second // in seconds

	DistributedLockKeyInContext = "distributed_lock"
)

var (
	retryInterval time.Duration = 5
)

type EthClient struct {
	Client          *ethclient.Client
	IsTKMChain      bool
	ChainId         *big.Int
	SuggestGasPrice *Expirable[*big.Int]
}

func NewEthClient(ctx context.Context, addr string, chainid *big.Int, gpttlseconds int64, isTKMChain ...bool) (ec *EthClient, err error) {
	var c *ethclient.Client
	c, err = ethclient.Dial(addr)
	if err != nil || c == nil {
		return nil, fmt.Errorf("connect TARGET@%s failed: %w", addr, err)
	}
	defer func() {
		if err != nil {
			if c != nil {
				c.Close()
			}
		}
	}()

	var id *big.Int
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut*6)
	defer cancel()
	id, err = c.ChainID(cctx)
	if err != nil {
		return nil, err
	}
	if chainid != nil && math.CompareBigInt(chainid, id) != 0 {
		return nil, fmt.Errorf("chain id not match, want:%s got:%s", math.BigIntForPrint(chainid), math.BigIntForPrint(id))
	}
	log.Infof("TARGET@EthClient(%s) ChainID:%s connected", addr, id)
	ec = &EthClient{
		Client:          c,
		ChainId:         id,
		SuggestGasPrice: NewExpirable(big.NewInt(0), 1000*gpttlseconds, 0),
	}
	if len(isTKMChain) > 0 && isTKMChain[0] {
		log.Infof("TARGET is an TKM chain")
		ec.IsTKMChain = true
	}
	return ec, nil
}

func (c *EthClient) Close() {
	if c.Client != nil {
		log.Warnf("%s closing", c)
		c.Client.Close()
		c.Client = nil
	}
}

func (c *EthClient) String() string {
	if c == nil {
		return "EthClient<nil>"
	}
	return fmt.Sprintf("EthClient{ChainID:%s}", c.ChainId)
}

func (c *EthClient) _check() error {
	if c.Client == nil || c.ChainId == nil {
		return errors.New("invalid client")
	}
	return nil
}

func (c *EthClient) nonceWithBalanceMoreThan(ctx context.Context, addr common.Address, checkBalance bool, levels ...*big.Int) (uint64, error) {
	if err := c._check(); err != nil {
		return 0, err
	}
	acc := T2E.Address(addr)
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut*2)
	defer cancel()
	if checkBalance {
		balance, err := c.Client.PendingBalanceAt(cctx, acc)
		if err != nil {
			return 0, err
		}
		var level *big.Int
		if len(levels) > 0 {
			level = levels[0]
		}
		bl := (*math.BigInt)(level).MustInt()
		if (*math.BigInt)(balance).CompareInt(bl) <= 0 {
			return 0, fmt.Errorf("balance of %x is less than %s", addr[:], math.BigIntForPrint(bl))
		}
	}
	return c.Client.PendingNonceAt(cctx, acc)
}

func (c *EthClient) nonceAndBalance(ctx context.Context, addr common.Address) (uint64, *big.Int, error) {
	if err := c._check(); err != nil {
		return 0, nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut*2)
	defer cancel()
	acc := T2E.Address(addr)
	nonce, err := c.Client.PendingNonceAt(cctx, acc)
	if err != nil {
		return 0, nil, err
	}
	balance, err := c.Client.PendingBalanceAt(cctx, acc)
	if err != nil {
		return nonce, nil, err
	}
	return nonce, balance, nil
}

func (c *EthClient) getBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	if err := c._check(); err != nil {
		return nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	return c.Client.PendingBalanceAt(cctx, T2E.Address(addr))
}

func (c *EthClient) getNonce(ctx context.Context, addr common.Address) (uint64, error) {
	if err := c._check(); err != nil {
		return 0, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	nonce, err := c.Client.PendingNonceAt(cctx, T2E.Address(addr))
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

func (c *EthClient) getCode(ctx context.Context, addr common.Address) ([]byte, error) {
	if err := c._check(); err != nil {
		return nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	return c.Client.PendingCodeAt(cctx, T2E.Address(addr))
}

func (c *EthClient) estimateGas(ctx context.Context, from common.Address, to *common.Address, gas uint64,
	gasPrice *big.Int, value *big.Int, data []byte) (uint64, error) {
	if err := c._check(); err != nil {
		return 0, err
	}
	msg := ethereum.CallMsg{
		From:     T2E.Address(from),
		To:       T2E.AddressP(to),
		Gas:      gas,
		GasPrice: gasPrice,
		Value:    value,
		Data:     data,
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	return c.Client.EstimateGas(cctx, msg)
}

func (c *EthClient) getter(ctx context.Context, from common.Address, to *common.Address,
	method abi.Method, outObj interface{}, args ...interface{}) error {
	data, err := method.Inputs.Pack(args...)
	if err != nil {
		return fmt.Errorf("pack input failed: %w", err)
	}
	input := make([]byte, 0, 4+len(data))
	input = append(input, method.ID...)
	input = append(input, data...)
	output, err := c.callContract(ctx, from, to, defaultGas, nil, nil, input)
	if err != nil {
		return fmt.Errorf("call failed: %w", err)
	}
	if err = method.Outputs.UnpackIntoInterface(outObj, output); err != nil {
		return fmt.Errorf("parse returns failed: %w", err)
	}
	return nil
}

func (c *EthClient) callContract(ctx context.Context, from common.Address, to *common.Address, gas uint64,
	gasPrice *big.Int, value *big.Int, data []byte) ([]byte, error) {
	if err := c._check(); err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{
		From:     T2E.Address(from),
		To:       T2E.AddressP(to),
		Gas:      gas,
		GasPrice: gasPrice,
		Value:    value,
		Data:     data,
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	return c.Client.PendingCallContract(cctx, msg)
}

func (c *EthClient) suggestGasPrice(ctx context.Context) (*big.Int, error) {
	gp, exist := c.SuggestGasPrice.Get()
	if exist {
		return gp, nil
	}
	var err error
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	gp, err = c.Client.SuggestGasPrice(cctx)
	if err != nil || gp == nil {
		return gp, err
	}
	c.SuggestGasPrice.Update(gp)
	log.Debugf("suggest GasPrice=%s get and cached", math.BigForPrint(gp))
	return gp, nil
}

func (c *EthClient) sendLegacyTx(ctx context.Context, priv []byte, to *common.Address, nonce uint64, gas uint64,
	gasPrice *big.Int, value *big.Int, input []byte) (*types.Transaction, *common.Hash, error) {
	if err := c._check(); err != nil {
		return nil, nil, err
	}
	pk, err := crypto.ToECDSA(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("private key error: %w", err)
	}
	if gasPrice == nil {
		gp, err := c.suggestGasPrice(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("suggest gas price failed: %w", err)
		}
		gasPrice = new(big.Int).Set(gp)
	}
	if value == nil {
		value = big.NewInt(0)
	}
	log.Debugf("trying to send: {Nonce:%d GasPrice:%s Gas:%d To:%x Val:%s len(Data):%d}",
		nonce, math.BigIntForPrint(gasPrice), gas, common.ForPrint(to, 0), math.BigIntForPrint(value), len(input))
	signer := types.LatestSignerForChainID(c.ChainId)
	tx, err := types.SignNewTx(pk, signer, &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gas,
		To:       T2E.AddressP(to),
		Value:    value,
		Data:     input,
	})
	if err != nil {
		return nil, nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	err = c.Client.SendTransaction(cctx, tx)
	if err != nil {
		return tx, nil, err
	}
	ethHash := tx.Hash()
	return tx, common.BytesToHashP(ethHash[:]), nil
}

func (c *EthClient) getReceipt(ctx context.Context, txHash common2.Hash) (*types.Receipt, error) {
	if err := c._check(); err != nil {
		return nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()
	log.Debugf("try get receipt of txHash: %x", txHash[:])
	ethrecept, err := c.Client.TransactionReceipt(cctx, txHash)
	return ethrecept, err
}

func putDistributedLock(ctx context.Context, lock DistributedLock) context.Context {
	if lock == nil {
		return ctx
	}
	return context.WithValue(ctx, DistributedLockKeyInContext, lock)
}

func getDistributedLock(ctx context.Context) DistributedLock {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(DistributedLockKeyInContext)
	if v == nil {
		return nil
	}
	l, ok := v.(DistributedLock)
	if !ok {
		return nil
	}
	return l
}

func (c *EthClient) checkReceipt(ctx context.Context, tx *types.Transaction) (*client.ReceiptWithFwds, error) {
	lock := getDistributedLock(ctx)
	txHash := tx.Hash()

	for i := 0; i < 5; i++ {
		if lock != nil {
			if err := lock.Refresh(ctx); err != nil {
				log.Warnf("refresh %s failed: %v", lock, err)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(retryInterval * time.Second)
			rec, err := c.getReceipt(ctx, txHash)
			if err != nil || rec == nil {
				continue
			}

			txrpt, err := E2T.TxReceipt(tx, rec, c.IsTKMChain)
			if err != nil {
				return nil, err
			}
			return client.NewReceiptWithForwards(txrpt)
		}
	}
	return nil, client.ErrNoReceipt
}

func (c *EthClient) checkReceipts(ctx context.Context, ethtxs ...*types.Transaction) ([]*client.ReceiptWithFwds, error) {
	if len(ethtxs) == 0 {
		return nil, nil
	}
	txHashList := make([]common.Hash, 0, len(ethtxs))
	txMap := make(map[common.Hash]*types.Transaction)
	rptMap := make(map[common.Hash]*client.ReceiptWithFwds)
	for _, ethtx := range ethtxs {
		if ethtx == nil {
			continue
		}
		txhash := E2T.Hash(ethtx.Hash())
		txHashList = append(txHashList, txhash)
		txMap[txhash] = ethtx
	}

	lock := getDistributedLock(ctx)
	for i := 0; i < 12; i++ {
		if lock != nil {
			if err := lock.Refresh(ctx); err != nil {
				log.Warnf("refresh %s failed: %v", lock, err)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(retryInterval * time.Second)
			for _, txhash := range txHashList {
				if _, exist := rptMap[txhash]; exist {
					continue
				}
				rec, err := c.getReceipt(ctx, T2E.Hash(txhash))
				if err != nil || rec == nil {
					continue
				}
				// log.Debugf("receipt of 0x%x: %+v", txhash[:], rec)
				ethtx := txMap[txhash]
				txrpt, err := E2T.TxReceipt(ethtx, rec, c.IsTKMChain)
				if err != nil {
					continue
				}
				rptwf := &client.ReceiptWithFwds{TransactionReceipt: *txrpt}
				log.Debugf("receipt: %s", rptwf.InfoString(0))
				rptMap[txhash] = rptwf
			}
		}
		if len(rptMap) == len(txMap) {
			break
		}
	}

	ret := make([]*client.ReceiptWithFwds, len(ethtxs))
	for i, txhash := range txHashList {
		ret[i] = rptMap[txhash]
	}
	return ret, nil
}
