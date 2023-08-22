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
	"strings"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/ThinkiumGroup/go-common/math"
	"github.com/ThinkiumGroup/go-tkmrpc"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stephenfire/go-rtl"
	"github.com/urfave/cli/v2"
)

type xsyncer struct {
	looper
	watchTopicId      common.Hash
	maxProvableHeight *Expirable[*common.Height]
}

func (n *xsyncer) Name() string {
	return fmt.Sprintf("XSYNC_%s", n.conf.TargetName)
}

func (n *xsyncer) prepareConfig(ctx *cli.Context) error {
	initRelayLNAbi()
	initMCSAbis()
	if err := n.looper.prepareConfig(ctx); err != nil {
		return err
	}
	n.keys.startHeightKey = fmt.Sprintf("%s_start_%d", strings.ToLower(n.Name()), n.conf.SrcChainId)
	n.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(n.Name()), n.conf.SrcChainId)
	log.Infof("%s", n.keys)

	xMcs, err := stringToAddress(ctx, _xSyncMCSFlag.Name)
	if err != nil {
		return err
	}
	targetMcs, err := stringToAddress(ctx, _xSyncTargetMCSFlag.Name)
	if err != nil {
		return err
	}
	targetlc, err := stringToAddress(ctx, _xSyncTargetLCFlag.Name)
	if err != nil {
		return err
	}
	n.conf.XSynchronizer.TargetMSCAddr = targetMcs
	n.conf.XSynchronizer.XChainId = big.NewInt(0).SetUint64(ctx.Uint64(_xSyncChainIDFlag.Name))
	n.conf.XSynchronizer.XMCSAddress = xMcs
	n.conf.XSynchronizer.TargetLCAddr = targetlc
	n.conf.XSynchronizer.MaxHeightTTL = int64(ctx.Uint64(_xSyncMaxHeightTTLFlag.Name))

	if err := n.conf.XSynchronizer.validate(); err != nil {
		return err
	}
	n.maxProvableHeight = NewExpirable[*common.Height](
		(*common.Height)(nil),
		n.conf.XSynchronizer.MaxHeightTTL*1000,
		0,
	)
	models.SysContractLogger.Register(n.conf.XSynchronizer.XMCSAddress, MCSRelayAbi)
	models.SysContractLogger.Register(n.conf.XSynchronizer.TargetMSCAddr, MCSAbi)
	models.SysContractLogger.Register(n.conf.XSynchronizer.TargetLCAddr, XLightNodeAbi)
	event, ok := MCSRelayAbi.Events[transferOutEvent]
	if !ok {
		return fmt.Errorf("%s event signature not found in MCSRelayABI", transferOutEvent)
	}
	n.watchTopicId = event.ID
	log.Infof("watching: Address:%x EventTopic:%x", n.conf.XSynchronizer.XMCSAddress[:], n.watchTopicId[:])
	return nil
}

func (n *xsyncer) confirmConfig(ctx *cli.Context) error {
	if err := n.looper.confirmConfig(ctx); err != nil {
		return err
	}
	acc, err := n.src.Account(ctx.Context, n.conf.XSynchronizer.XMCSAddress)
	if err != nil {
		return fmt.Errorf("get account failed: %w", err)
	}
	if acc == nil || len(acc.Code) == 0 {
		return fmt.Errorf("X-Relay MCS contract at 0x%x not found", n.conf.XSynchronizer.XMCSAddress[:])
	}

	if !n._targetMustContract(ctx, n.conf.XSynchronizer.TargetMSCAddr) {
		return fmt.Errorf("target MSC address %x not a contract", n.conf.XSynchronizer.TargetMSCAddr[:])
	}
	return nil
}

func (n *xsyncer) _targetMustContract(cctx *cli.Context, addr common.Address) bool {
	ctx, cancel := context.WithCancel(cctx.Context)
	defer cancel()
	code, err := n.target.Client.CodeAt(ctx, T2E.Address(addr), nil)
	if err != nil {
		return false
	}
	if len(code) == 0 {
		return false
	}
	return true
}

func (n *xsyncer) _maxValidatableHeightFromLC(ctx context.Context) (common.Height, error) {
	outobj := new(struct{ LastEpoch uint64 })
	if err := n.target.getter(ctx, n.targetPriv.Address(), &n.conf.XSynchronizer.TargetLCAddr,
		XLightNodeAbi.Methods[xEndsOfEpochName], outobj, big.NewInt(1)); err != nil {
		return common.NilHeight, fmt.Errorf("getter XLC.%s failed: %w", xEndsOfEpochName, err)
	}
	lastEpoch := common.EpochNum(outobj.LastEpoch)
	if lastEpoch.IsNil() {
		return common.NilHeight, cli.Exit(errors.New("unavailable last epoch in LC"), ExitLCErr)
	}
	return lastEpoch.LastHeight(), nil
}

func (n *xsyncer) _maxProvableHeight(ctx context.Context) (common.Height, error) {
	max, exist := n.maxProvableHeight.Get()
	if exist && max != nil {
		return *max, nil
	}
	log.Debugf("provable height cache missed, try get")

	maxValidatableHeight, err := n._maxValidatableHeightFromLC(ctx)
	if err != nil {
		return common.NilHeight, err
	}
	n.maxProvableHeight.Update(&maxValidatableHeight)
	log.Warnf("provable height cache put: %s", &maxValidatableHeight)
	return maxValidatableHeight, nil
}

func (n *xsyncer) prepareToGet(cctx *cli.Context, start common.Height) error {
	max, err := n._maxProvableHeight(cctx.Context)
	if err != nil {
		return err
	}
	if start.Compare(max) > 0 {
		return NotUnlockError(fmt.Errorf("max provable height exceeded: Max:%s, but start:%s", &max, &start))
	}
	return nil
}

func (n *xsyncer) processBlock(cctx *cli.Context, block *models.BlockEMessage) (fatal, warning error) {
	max, err := n._maxProvableHeight(cctx.Context)
	if err != nil {
		return err, nil
	}
	if block != nil && block.BlockHeader != nil && block.BlockBody != nil {
		if block.BlockHeader.Height.Compare(max) > 0 {
			return NotUnlockError(fmt.Errorf("max provable height exceeded: Max:%s, but Block.Height:%s",
				&max, &block.BlockHeader.Height)), nil
		}
		var txproofs []*models.TxFinalProof
		for _, tx := range block.BlockBody.Txs {
			if tx.To != nil && len(tx.Input) > 0 {
				_ = n.runningLock.Refresh(cctx.Context)
				txHash := tx.Hash()
				proof, err := n._txLocalProof(cctx.Context, n.conf.SrcChainId, txHash)
				if err != nil || proof == nil {
					return fmt.Errorf("get local proof of TxHash:%x failed: %w", txHash[:], err), nil
				}
				if !proof.Receipt.Success() {
					log.Debugf("%s failed", tx)
					continue
				}
				if err := proof.LocalVerify(); err != nil {
					return fmt.Errorf("local proof %s verify failed: %w", proof, err), nil
				}
				if proof.Receipt == nil {
					return fmt.Errorf("get receipt of TxHash:%x failed", txHash[:]), nil
				}
				if i, rlog := locateLog(proof.Receipt.Logs, n.conf.XSynchronizer.XMCSAddress, n.watchTopicId); i >= 0 {
					out := new(MapTransferOutLog)
					if err := MCSRelayAbi.UnpackEvent(out, rlog.Topics, rlog.Data); err != nil {
						return fmt.Errorf("unpack log %s failed: %w", rlog, err), nil
					} else {
						if math.CompareBigInt(out.ToChain, n.conf.TargetChainID) == 0 {
							log.Infof("%s found", out)
						} else {
							log.Warnf("%s found, but TargetChainID:%s not match", out, n.conf.TargetChainID)
							continue
						}
					}
					if exist, err := n._checkOrderId(cctx, out.OrderId); err != nil {
						return fmt.Errorf("check orderid %x failed: %w", out.OrderId[:], err), nil
					} else if exist {
						log.Warnf("%s already in order list", out)
						continue
					}
					txproofs = append(txproofs, proof)
					log.Debugf("try to send %d: %s", len(txproofs), proof.InfoString(0))
				}
			}
		}

		// if err := n._lnProofs(cctx, txproofs...); err != nil {
		if err := n._mcsProofs(cctx, txproofs); err != nil {
			return fmt.Errorf("MCS proof failed: %w", err), nil
		}
	}
	return nil, nil
}

func (n *xsyncer) _txLocalProof(baseCtx context.Context, chainid common.ChainID, txHash common.Hash) (*models.TxFinalProof, error) {
	ctx, cancel := context.WithTimeout(baseCtx, reqTimeOut)
	defer cancel()
	resp, err := n.src.NodeClient.GetTxLocalProof(ctx, &tkmrpc.RpcTXHash{
		Chainid: uint32(chainid),
		Hash:    txHash[:],
	})
	if err != nil {
		return nil, fmt.Errorf("TxLocalProof: ChainID:%d TxHash:%x failed: %w",
			chainid, txHash[:], err)
	}
	localProof := new(models.TxFinalProof)
	if err = rtl.Unmarshal(resp.Stream, localProof); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	return localProof, nil
}

func (n *xsyncer) _checkOrderId(ctx *cli.Context, orderId common.Hash) (alreadyTransferred bool, err error) {
	outobj := new(struct{ Exist bool })
	if err := n.target.getter(ctx.Context, n.targetPriv.Address(), &n.conf.XSynchronizer.TargetMSCAddr,
		MCSAbi.Methods[orderListName], outobj, orderId); err != nil {
		return false, fmt.Errorf("getter target.MCS.%s failed: %w", orderListName, err)
	}
	return outobj.Exist, nil
}

func (n *xsyncer) _mcsProofs(cctx *cli.Context, txProofs []*models.TxFinalProof) error {
	if len(txProofs) == 0 {
		return nil
	}
	lockingValue, err := n.sendingLock.Fetch(cctx.Context)
	if err != nil {
		return fmt.Errorf("[%s] is sending, fetch %s failed: %w", lockingValue, n.sendingLock, err)
	}
	defer func() {
		_ = n.sendingLock.Release()
	}()

	dlocks := redisLocks{n.runningLock, n.sendingLock}

	_ = dlocks.Refresh(cctx.Context)
	to := n.conf.XSynchronizer.TargetMSCAddr
	gas, mustHave := n._targetSuggestBalance(cctx.Context)
	nonce, err := n.target.nonceWithBalanceMoreThan(cctx.Context, n.targetPriv.Address(), mustHave)
	if err != nil {
		return err
	}

	// send txs
	var ethtxs []*types.Transaction
	for i, txProof := range txProofs {
		// proof, err := T2LN.ReceiptProof(txProof)
		proof, err := T2LN.ReceiptData(txProof, n.conf.XSynchronizer.XMCSAddress, n.watchTopicId)
		if err != nil {
			return err
		}
		log.Infof("proofs: %s", proof.String())
		data, err := XLightNodeAbi.Methods[xVerifyReceiptStruct].Inputs.Pack(proof)
		if err != nil {
			return fmt.Errorf("packdata failed: %w", err)
		}
		input, err := MCSAbi.Pack(transferInName, n.conf.XSynchronizer.XChainId, data)
		if err != nil {
			return fmt.Errorf("packinput failed: %w", err)
		}
		ethtx, _, err := n.target.sendLegacyTx(cctx.Context, n.targetPriv.Priv(), &to, nonce, gas, nil, nil, input)
		if err != nil {
			return fmt.Errorf("send tx failed: %w", err)
		}
		ethtxs = append(ethtxs, ethtx)
		nonce++
		if i > 0 && i%10 == 0 {
			_ = dlocks.Refresh(cctx.Context)
		}
	}

	// get receipts
	rcpts, err := n.target.checkReceipts(putDistributedLock(cctx.Context, dlocks), ethtxs...)
	if err != nil {
		return fmt.Errorf("get receipt failed: %w", err)
	}
	var successes, faileds []common.Hash
	for i, rpt := range rcpts {
		if rpt != nil && rpt.Success() {
			successes = append(successes, rpt.TxHash)
		} else {
			if rpt != nil {
				faileds = append(faileds, rpt.TxHash)
			} else {
				if i < len(ethtxs) && ethtxs[i] != nil {
					faileds = append(faileds, E2T.Hash(ethtxs[i].Hash()))
				} else {
					faileds = append(faileds, common.EmptyHash)
				}
			}
		}
	}
	if len(successes) > 0 {
		log.Infof("MCS Success: %s", successes)
	}
	if len(faileds) > 0 {
		log.Errorf("MCS failed: %s", faileds)
		return fmt.Errorf("transfer failed occurs: %d successed, %d failed", len(successes), len(faileds))
	}
	return nil
}

func (n *xsyncer) _lnProofs(cctx *cli.Context, txProofs ...*models.TxFinalProof) error {
	if len(txProofs) == 0 {
		return nil
	}
	proof, err := T2LN.ReceiptData(txProofs[0], n.conf.XSynchronizer.XMCSAddress, n.watchTopicId)
	if err != nil {
		return err
	}
	log.Infof("proofs: %s", common.IndentLevel(0).InfoString(proof.Proofs))
	data, err := XLightNodeAbi.Methods[xVerifyReceiptStruct].Inputs.Pack(proof)
	if err != nil {
		return fmt.Errorf("packdata failed: %w", err)
	}
	input, err := XLightNodeAbi.Pack(xVerifyReceiptData, data)
	if err != nil {
		return fmt.Errorf("packinput failed: %w", err)
	}
	log.Infof("input of verify:\n%x", input)
	to := n.conf.XSynchronizer.TargetLCAddr
	output, err := n.target.callContract(cctx.Context, n.targetPriv.Address(), &to, 30000000, nil, nil, input)
	if err != nil {
		return fmt.Errorf("call failed: %w", err)
	}
	log.Infof("return: %x", output)
	retObj := new(struct {
		Success  bool   `abi:"success"`
		Mesage   string `abi:"message"`
		LogBytes []byte `abi:"logBytes"`
	})
	if err := XLightNodeAbi.UnpackReturns(retObj, xVerifyReceiptData, output); err != nil {
		return fmt.Errorf("return parse failed: %w", err)
	}
	if retObj.Success {
		return nil
	}
	return errors.New("verify failed")
}
