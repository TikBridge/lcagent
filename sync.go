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
	"github.com/ThinkiumGroup/go-tkmrpc"
	"github.com/ThinkiumGroup/go-tkmrpc/client"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stephenfire/go-rtl"
	"github.com/urfave/cli/v2"
)

type provableHeights struct {
	main common.Height
	sub  common.Height
}

func (p *provableHeights) String() string {
	if p == nil {
		return "ProvableHeights<nil>"
	}
	return fmt.Sprintf("ProvableHeights{Main:%s Sub:%s}", &p.main, &p.sub)
}

type syncer struct {
	looper
	watchTopicId       common.Hash
	maxProvableHeights *Expirable[*provableHeights]
}

func (n *syncer) Name() string {
	return fmt.Sprintf("SYNC_%s", n.conf.TargetName)
}

func (n *syncer) prepareConfig(ctx *cli.Context) error {
	initTKMLNAbi()
	initMCSAbis()
	initUpdatableLNAbi()
	if err := n.looper.prepareConfig(ctx); err != nil {
		return err
	}
	n.keys.startHeightKey = fmt.Sprintf("%s_start_%d", strings.ToLower(n.Name()), n.conf.SrcChainId)
	n.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(n.Name()), n.conf.SrcChainId)
	log.Infof("%s", n.keys)

	tkmMcs, err := stringToAddress(ctx, _syncTkmMCSFlag.Name)
	if err != nil {
		return err
	}
	targetMcs, err := stringToAddress(ctx, _syncTargetMCSFlag.Name)
	if err != nil {
		return err
	}
	targetlc, err := stringToAddress(ctx, _syncTargetLCFlag.Name)
	if err != nil {
		return err
	}
	n.conf.Synchronizer.TargetMSCAddr = targetMcs
	n.conf.Synchronizer.TkmChainId = big.NewInt(0).SetUint64(ctx.Uint64(_syncTkmChainIDFlag.Name))
	n.conf.Synchronizer.TkmMCSAddress = tkmMcs
	n.conf.Synchronizer.TargetLCAddr = targetlc
	n.conf.Synchronizer.UpdatableLC = ctx.Bool(_syncUpdatableLCFlag.Name)
	n.conf.Synchronizer.MaxHeightTTL = int64(ctx.Uint64(_syncMaxHeightTTLFlag.Name))

	if err := n.conf.Synchronizer.validate(); err != nil {
		return err
	}
	n.maxProvableHeights = NewExpirable[*provableHeights](
		(*provableHeights)(nil),
		n.conf.Synchronizer.MaxHeightTTL*1000,
		0,
	)
	models.SysContractLogger.Register(n.conf.Synchronizer.TkmMCSAddress, MCSRelayAbi)
	models.SysContractLogger.Register(n.conf.Synchronizer.TargetMSCAddr, MCSAbi)
	if n.conf.Synchronizer.UpdatableLC {
		models.SysContractLogger.Register(n.conf.Synchronizer.TargetLCAddr, UpdatableLightNodeAbi)
	} else {
		models.SysContractLogger.Register(n.conf.Synchronizer.TargetLCAddr, LightNodeABI)
	}
	event, ok := MCSRelayAbi.Events[transferOutEvent]
	if !ok {
		return fmt.Errorf("%s event signature not found in MCSRelayABI", transferOutEvent)
	}
	n.watchTopicId = event.ID
	log.Infof("watching: Address:%x EventTopic:%x", n.conf.Synchronizer.TkmMCSAddress[:], n.watchTopicId[:])
	return nil
}

func (n *syncer) confirmConfig(ctx *cli.Context) error {
	if err := n.looper.confirmConfig(ctx); err != nil {
		return err
	}
	acc, err := n.src.Account(ctx.Context, n.conf.Synchronizer.TkmMCSAddress)
	if err != nil {
		return fmt.Errorf("get account failed: %w", err)
	}
	if acc == nil || len(acc.Code) == 0 {
		return fmt.Errorf("TKM MCS contract at 0x%x not found", n.conf.Synchronizer.TkmMCSAddress[:])
	}

	if !n._targetMustContract(ctx, n.conf.Synchronizer.TargetMSCAddr) {
		return fmt.Errorf("target MSC address %x not a contract", n.conf.Synchronizer.TargetMSCAddr[:])
	}
	return nil
}

func (n *syncer) _targetMustContract(cctx *cli.Context, addr common.Address) bool {
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

func (n *syncer) prepareToGet(cctx *cli.Context, start common.Height) error {
	maxMain, maxSub, err := n._maxProvableHeights(cctx.Context)
	if err != nil {
		return err
	}
	if start.Compare(maxSub) > 0 {
		return NotUnlockError(fmt.Errorf("max provable height exceeded: Main:%s, Sub:%s, but start:%s",
			&maxMain, &maxSub, &start))
	}
	return nil
}

func (n *syncer) processBlock(cctx *cli.Context, block *models.BlockEMessage) (fatal, warning error) {
	maxMain, maxSub, err := n._maxProvableHeights(cctx.Context)
	if err != nil {
		return err, nil
	}
	if block != nil && block.BlockHeader != nil && block.BlockBody != nil {
		if block.BlockHeader.Height.Compare(maxSub) > 0 {
			return NotUnlockError(fmt.Errorf("max provable height exceeded: Main:%s, Sub:%s, but Block.Height:%s",
				&maxMain, &maxSub, &block.BlockHeader.Height)), nil
		}
		var txproofs []*models.TxFinalProof
		for _, tx := range block.BlockBody.Txs {
			if tx.To != nil && len(tx.Input) > 0 {
				_ = n.runningLock.Refresh(cctx.Context)
				txHash := tx.Hash()
				proof, err := n._txFinalProof(cctx.Context, n.conf.SrcChainId, txHash, maxMain)
				if err != nil || proof == nil {
					return fmt.Errorf("get final proof of TxHash:%x failed: %w", txHash[:], err), nil
				}
				if !proof.Receipt.Success() {
					log.Debugf("%s failed", tx)
					continue
				}
				if err := proof.FinalVerify(); err != nil {
					return fmt.Errorf("final proof %s verify failed: %w", proof, err), nil
				}
				if proof.Receipt == nil {
					return fmt.Errorf("get receipt of TxHash:%x failed", txHash[:]), nil
				}
				if i, rlog := locateLog(proof.Receipt.Logs, n.conf.Synchronizer.TkmMCSAddress, n.watchTopicId); i >= 0 {
					out := new(MapTransferOutLog)
					if err := MCSRelayAbi.UnpackEvent(out, rlog.Topics, rlog.Data); err != nil {
						return fmt.Errorf("unpack log %s failed: %w", rlog, err), nil
					} else {
						log.Infof("%s found", out)
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

		if err := n._mcsProofs(cctx, txproofs); err != nil {
			return fmt.Errorf("MCS proof failed: %w", err), nil
		}
	}
	return nil, nil
}

func (n *syncer) _txFinalProof(baseCtx context.Context, chainid common.ChainID,
	txHash common.Hash, anchorHeight common.Height) (*models.TxFinalProof, error) {
	ctx, cancel := context.WithTimeout(baseCtx, reqTimeOut)
	defer cancel()
	resp, err := n.src.NodeClient.GetTxFinalProof(ctx,
		&tkmrpc.RpcTxProofReq{
			Chainid:           uint32(chainid),
			Hash:              txHash[:],
			ProofedMainHeight: uint64(anchorHeight),
		})
	if err != nil {
		return nil, fmt.Errorf("TxFinalProof: ChainID:%d TxHash:%x Anchor:%s failed: %w",
			chainid, txHash[:], &anchorHeight, err)
	}
	finalProof := new(models.TxFinalProof)
	if err = rtl.Unmarshal(resp.Stream, finalProof); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	return finalProof, nil
}

func (n *syncer) _lnProof(cctx *cli.Context, txProof *models.TxFinalProof) error {
	proof, err := T2LN.ReceiptProof(txProof, n.conf.Synchronizer.TkmMCSAddress, n.watchTopicId)
	if err != nil {
		return err
	}
	log.Infof("proofs: %s", common.IndentLevel(0).InfoString(proof.Proofs))
	data, err := LightNodeABI.Methods[verifyReceiptStruct].Inputs.Pack(proof)
	if err != nil {
		return fmt.Errorf("packdata failed: %w", err)
	}
	input, err := LightNodeABI.Pack(verifyReceiptData, data)
	if err != nil {
		return fmt.Errorf("packinput failed: %w", err)
	}
	log.Infof("input of verify:\n%x", input)
	to := n.conf.Synchronizer.TargetMSCAddr
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
	if err := LightNodeABI.UnpackReturns(retObj, verifyReceiptData, output); err != nil {
		return fmt.Errorf("return parse failed: %w", err)
	}
	if retObj.Success {
		return nil
	}
	return errors.New("verify failed")
}

func (n *syncer) _mcsProofs(cctx *cli.Context, txProofs []*models.TxFinalProof) error {
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
	to := n.conf.Synchronizer.TargetMSCAddr
	gas, mustHave := n._targetSuggestBalance(cctx.Context)
	nonce, err := n.target.nonceWithBalanceMoreThan(cctx.Context, n.targetPriv.Address(), n.conf.TargetCheckBalance, mustHave)
	if err != nil {
		return err
	}

	// send txs
	var ethtxs []*types.Transaction
	for i, txProof := range txProofs {
		proof, err := T2LN.ReceiptProof(txProof, n.conf.Synchronizer.TkmMCSAddress, n.watchTopicId)
		if err != nil {
			return err
		}
		log.Infof("proofs: %s", proof.String())
		data, err := LightNodeABI.Methods[verifyReceiptStruct].Inputs.Pack(proof)
		if err != nil {
			return fmt.Errorf("packdata failed: %w", err)
		}
		input, err := MCSAbi.Pack(transferInName, n.conf.Synchronizer.TkmChainId, data)
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

func (n *syncer) _checkOrderId(ctx *cli.Context, orderId common.Hash) (alreadyTransferred bool, err error) {
	input, err := MCSAbi.Pack(orderListName, orderId)
	if err != nil {
		return false, fmt.Errorf("pack %s failed: %w", orderListName, err)
	}
	to := n.conf.Synchronizer.TargetMSCAddr
	output, err := n.target.callContract(ctx.Context, n.targetPriv.Address(), &to, defaultGas, nil, nil, input)
	if err != nil {
		return false, fmt.Errorf("call %s failed: %w", orderListName, err)
	}
	outObj := new(struct {
		Exist bool
	})
	if err := MCSAbi.UnpackReturns(outObj, orderListName, output); err != nil {
		return false, fmt.Errorf("parse output failed: %w", err)
	}
	return outObj.Exist, nil
}

func (n *syncer) _maxValidatableHeightFromLC(ctx context.Context) (common.Height, error) {
	if n.conf.Synchronizer.UpdatableLC {
		outobj := new(struct{ Epoch uint64 })
		if err := n.target.getter(ctx, n.targetPriv.Address(), &n.conf.Synchronizer.TargetLCAddr,
			UpdatableLightNodeAbi.Methods[uLastEpochName], outobj); err != nil {
			return common.NilHeight, fmt.Errorf("UpdatableLC.%s failed: %w", uLastEpochName, err)
		}
		epoch := common.EpochNum(outobj.Epoch)
		if epoch.IsNil() {
			return common.NilHeight, cli.Exit(errors.New("unavailable last epoch in LC"), ExitLCErr)
		}
		return epoch.LastHeight(), nil
	} else {
		outobj := new(struct{ LastHeight uint64 })
		if err := n.target.getter(ctx, n.targetPriv.Address(), &n.conf.Synchronizer.TargetLCAddr,
			LightNodeABI.Methods[lastHeightName], outobj); err != nil {
			return common.NilHeight, fmt.Errorf("LC.%s failed: %w", lastHeightName, err)
		}
		height := common.Height(outobj.LastHeight)
		if height.IsNil() {
			return common.NilHeight, cli.Exit(errors.New("unavailable last height in LC"), ExitLCErr)
		}
		return (height.EpochNum() + 1).LastHeight(), nil
	}
}

func (n *syncer) _lastConfirmedByMain(ctx context.Context, maxValidatableMainHeight common.Height) (
	maxMainHeight, confirmedHeight common.Height, err error) {
	// get current confirmeds
	confirmeds, err := n.src.LastConfirmedsAt(ctx, common.MainChainID, common.NilHeight)
	if err != nil {
		return common.NilHeight, common.NilHeight, fmt.Errorf("get confirmeds of main-chain failed: %w", err)
	}

	check := func(cs *client.Confirmeds) (common.Height, common.Height, error) {
		for _, info := range cs.Data {
			if info.ChainID == n.conf.SrcChainId {
				if info.Info == nil {
					// not confirmed yet
					return cs.At, common.NilHeight, errors.New("no confirmed block yet")
				} else {
					return cs.At, info.Info.Height, nil
				}
			}
		}
		return cs.At, common.NilHeight, errors.New("no confirmed info yet")
	}

	mheight := maxValidatableMainHeight
	if confirmeds.At.Compare(mheight) <= 0 {
		// maxValidatableHeight >= CurrentMainHeight
		if n.conf.SrcChainId == common.MainChainID {
			// main-chain current height
			return confirmeds.At, confirmeds.At, nil
		}
		return check(confirmeds)
	} else {
		// maxValidatableHeight < CurrentMainHeight
		if n.conf.SrcChainId == common.MainChainID {
			return maxValidatableMainHeight, maxValidatableMainHeight, nil
		}
		confirmeds, err := n.src.LastConfirmedsAt(ctx, common.MainChainID, maxValidatableMainHeight)
		if err != nil {
			return common.NilHeight, common.NilHeight,
				fmt.Errorf("get confirmeds of main-chain MaxValidatable:%s failed: %w", &maxValidatableMainHeight, err)
		}
		return check(confirmeds)
	}
}

func (n *syncer) _maxProvableHeights(ctx context.Context) (main, sub common.Height, err error) {
	max, exist := n.maxProvableHeights.Get()
	if exist && max != nil {
		return max.main, max.sub, nil
	}
	log.Debugf("provable height cache missed, try get")

	maxValidatableMainHeight, err := n._maxValidatableHeightFromLC(ctx)
	if err != nil {
		return common.NilHeight, common.NilHeight, err
	}

	maxMain, maxSub, err := n._lastConfirmedByMain(ctx, maxValidatableMainHeight)
	if err != nil {
		return common.NilHeight, common.NilHeight,
			fmt.Errorf("_lastConfirmedByMain(MMH:%s) failed: %w", &maxValidatableMainHeight, err)
	}

	max = &provableHeights{
		main: maxMain,
		sub:  maxSub,
	}
	n.maxProvableHeights.Update(max)
	log.Debugf("provable height cache put: %s", max)
	return maxMain, maxSub, nil
}
