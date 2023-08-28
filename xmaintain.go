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
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/stephenfire/go-rtl"
	"github.com/urfave/cli/v2"
)

// Update the consensus committee of the x-relay chain from the x-relay chain to the x-relay
// chain light node deployed on the target chain.
//
//	ATTENTION: The xsyncer process that matches this xmaintainer must use the same src.chainid
//	and target.name, otherwise there will be unexpected problems when updating the committee
//	because of the wrong sycnEpoch value
type xmaintainer struct {
	looper
	syncStartHeightKey string // key for
}

func (a *xmaintainer) Name() string {
	return fmt.Sprintf("XMAINTAIN_%s", a.conf.TargetName)
}

func (a *xmaintainer) prepareConfig(ctx *cli.Context) error {
	initRelayLNAbi()
	if err := a.looper.prepareConfig(ctx); err != nil {
		return err
	}
	a.keys.startHeightKey = fmt.Sprintf("%s_start_%d", strings.ToLower(a.Name()), a.conf.SrcChainId)
	a.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(a.Name()), a.conf.SrcChainId)
	a.syncStartHeightKey = fmt.Sprintf("xsync_%s_start_%d", a.conf.TargetName, a.conf.SrcChainId)
	log.Infof("%s, SyncStartHeightKey: %s", a.keys, a.syncStartHeightKey)

	if _, exist := XLightNodeAbi.Events[xUpdateCommEvent]; !exist {
		return fmt.Errorf("event %s must be exist", xUpdateCommEvent)
	}

	addr, err := stringToAddress(ctx, _xmaintainTargetLCFlag.Name)
	if err != nil {
		return err
	}
	a.conf.XMaintainer.TargetLCAddr = addr
	if err := a.conf.XMaintainer.validate(); err != nil {
		return err
	}
	models.SysContractLogger.Register(a.conf.Maintainer.TargetLCAddr, LightNodeABI)
	return nil
}

func (a *xmaintainer) confirmConfig(ctx *cli.Context) error {
	if err := a.looper.confirmConfig(ctx); err != nil {
		return err
	}

	// check availability of light-node and get lastHeight
	lastHeight := common.NilHeight
	input, err := XLightNodeAbi.Pack(xLastHeightName)
	if err != nil {
		return fmt.Errorf("encode xlightnode.lastHeight() failed: %w", err)
	}
	output, err := a.target.callContract(ctx.Context, a.targetPriv.Address(), &a.conf.XMaintainer.TargetLCAddr, defaultGas, nil, nil, input)
	if err != nil {
		return fmt.Errorf("xlightnode.lastHeight() failed: %w", err)
	}
	outobj := new(struct {
		Height uint64
	})
	if err = XLightNodeAbi.UnpackReturns(outobj, xLastHeightName, output); err != nil {
		return fmt.Errorf("parse lastHeight failed: %w", err)
	}
	lastHeight = common.Height(outobj.Height)
	log.Infof("lastHeight of X-light-node: %s", &lastHeight)

	// update lastHeight
	newHeight := lastHeight + 1
	startHeight := a.getStartHeight(ctx)
	diff, cmp := startHeight.Diff(newHeight)
	if cmp < 0 {
		log.Warnf("replace start height from:%s to X-light-node.lastHeight+1: %s", &startHeight, &newHeight)
		if err := a.updateStartHeight(ctx, newHeight); err != nil {
			return fmt.Errorf("update start height (%d) failed: %w", newHeight, err)
		}
	} else if cmp > 0 {
		if diff >= common.BlocksInEpoch {
			return fmt.Errorf("startHeight:%s but X-light-node.lastHeight:%s", &startHeight, &lastHeight)
		}
	}
	log.Infof("start height: %d", a.getStartHeight(ctx))

	if ctx.Bool(_checkLNCommFlag.Name) {
		if err := a._checkLatestComm(ctx); err != nil {
			return cli.Exit(err, ExitLCErr)
		}
		return cli.Exit(errors.New("comms checked"), 0)
	}

	return nil
}

func (a *xmaintainer) _getSyncingEpoch(cctx *cli.Context) (common.EpochNum, error) {
	ctx, cancel := context.WithTimeout(cctx.Context, redisTimeout)
	defer cancel()
	h, err := a.redis.Get(ctx, a.syncStartHeightKey).Uint64()
	switch {
	case err != nil:
		return 0, err
	default:
		return common.Height(h).EpochNum(), nil
	}
}

func (a *xmaintainer) processBlock(cctx *cli.Context, block *models.BlockEMessage) (fatal, warning error) {
	if block.BlockBody.NextCommittee.Size() > 0 || block.BlockBody.NextRealCommittee.Size() > 0 {
		// update
		cp := &CommitteeProof{
			Header: block.BlockHeader,
			Comm:   nil,
			PaSs:   block.BlockPass,
		}
		if block.BlockBody.NextCommittee.IsAvailable() {
			cp.Comm = block.BlockBody.NextCommittee
		} else if block.BlockBody.NextRealCommittee.IsAvailable() {
			cp.Comm = block.BlockBody.NextRealCommittee
		} else {
			return nil, fmt.Errorf("no available committee found: Next:%s Real:%s",
				block.BlockBody.NextCommittee, block.BlockBody.NextRealCommittee)
		}
		syncEpoch, err := a._getSyncingEpoch(cctx)
		if err != nil {
			log.Warnf("get Syncing Epoch of XSYNC:%s failed: %v", a.syncStartHeightKey, err)
			cp.SyncingEpoch = 0
		} else {
			cp.SyncingEpoch = syncEpoch
		}
		// cp = cp.MalicousTest()
		if cp.Comm != nil {
			log.Infof("found: %s", cp)
			if err := a._targetUpdateComm(cctx, cp); err != nil {
				return fmt.Errorf("update %s failed: %w", cp, err), nil
			}
		}
	}
	return nil, nil
}

func (a *xmaintainer) _targetUpdateComm(cctx *cli.Context, comm *CommitteeProof) error {
	if err := comm.Verify(false); err != nil {
		return err
	}

	lockingValue, err := a.sendingLock.Fetch(cctx.Context)
	if err != nil {
		return fmt.Errorf("[%s] is sending, fetch %s failed: %w", lockingValue, a.sendingLock, err)
	}
	defer func() {
		_ = a.sendingLock.Release()
	}()

	// proof := comm.ForXABI()
	proof, err := comm.ForXDataABI()
	if err != nil {
		return fmt.Errorf("failed at ForXDataABI: %w", err)
	}
	log.Infof("\n%s", proof.InfoString(1))
	input, err := XLightNodeAbi.Pack(xUpdateCommName, proof)
	if err != nil {
		return fmt.Errorf("packinput failed: %w", err)
	}
	to := a.conf.XMaintainer.TargetLCAddr

	// gas, err := a.target.estimateGas(id.Address(), &to, 30000000, nil, nil, input)
	// if err != nil {
	// 	return fmt.Errorf("estimate failed: %w", err)
	// }
	gas, mustHave := a._targetSuggestBalance(cctx.Context)
	nonce, err := a.target.nonceWithBalanceMoreThan(cctx.Context, a.targetPriv.Address(), a.conf.TargetCheckBalance, mustHave)
	if err != nil {
		return err
	}

	ethtx, txhash, err := a.target.sendLegacyTx(cctx.Context, a.targetPriv.Priv(), &to, nonce, gas, nil, nil, input)
	if err != nil {
		return fmt.Errorf("send tx failed: %w", err)
	}

	log.Infof("update comm TxHash: %x", common.ForPrint(txhash, 0))
	rcpt, err := a.target.checkReceipt(putDistributedLock(cctx.Context, redisLocks{a.runningLock, a.sendingLock}), ethtx)
	if err != nil {
		return fmt.Errorf("get receipt failed: %w", err)
	}
	log.Debugf("%s", rcpt.InfoString(0))
	if !rcpt.Success() {
		return fmt.Errorf("tx failed: %w", rcpt.Err())
	}
	updateEvent := XLightNodeAbi.Events[xUpdateCommEvent]
	for _, l := range rcpt.Logs {
		if l != nil && len(l.Topics) > 0 && l.Topics[0] == updateEvent.ID {
			event, err := a._parseLNUpdateEvent(l)
			if err != nil {
				return fmt.Errorf("event parse failed: %w", err)
			}
			headerEpoch := comm.Header.Height.EpochNum()
			if diff, cmp := headerEpoch.Diff(event.Epoch); cmp >= 0 || diff != 1 {
				return fmt.Errorf("updates are not performing as expected: want Epoch:%d, got:%s", headerEpoch, event)
			}
			if !event.CommHash.Equal(comm.Header.ElectedNextRoot) {
				return fmt.Errorf("updates are not performing as expected: want Comm:%x, got:%s", comm.Header.ElectedNextRoot.Slice(), event)
			}
			log.Infof("%s updated %s", comm.Comm, event)
			return nil
		}
	}

	return errors.New("no update committee event found")
}

func (a *xmaintainer) _parseLNUpdateEvent(l *models.Log) (*updateEvent, error) {
	if l == nil || len(l.Topics) < 3 {
		return nil, errors.New("invalid log")
	}
	epoch := rtl.Numeric.BytesToUint64(l.Topics[1][24:])
	return &updateEvent{
		Epoch:    common.EpochNum(epoch),
		CommHash: l.Topics[2],
	}, nil
}

func (a *xmaintainer) _checkLatestComm(ctx *cli.Context) error {
	from := a.targetPriv.Address()
	to := a.conf.XMaintainer.TargetLCAddr

	endsEpochObj := new(struct{ Epoch uint64 })
	endsEpochObj.Epoch = uint64(common.NilEpoch)
	if err := a.target.getter(ctx.Context, from, &to,
		XLightNodeAbi.Methods[xEndsOfEpochName], endsEpochObj, big.NewInt(1)); err != nil {
		return fmt.Errorf("target.0x%x.%s[1] failed: %w", to[:], xEndsOfEpochName, err)
	}
	if common.EpochNum(endsEpochObj.Epoch).IsNil() {
		return fmt.Errorf("target.0x%x last epoch not found", to[:])
	}

	commObj := new(struct{ Comms []common.Address })
	if err := a.target.getter(ctx.Context, from, &to,
		XLightNodeAbi.Methods[xCheckEpochCommName], commObj, endsEpochObj.Epoch); err != nil {
		return fmt.Errorf("target.0x%x.%s[1] -> %s(epoch:%d) failed: %w",
			to[:], xEndsOfEpochName, xCheckEpochCommName, endsEpochObj.Epoch, err)
	}
	if len(commObj.Comms) == 0 {
		return fmt.Errorf("\"target.0x%x.%s[1] -> %s(epoch:%d) got nothing",
			to[:], xEndsOfEpochName, xCheckEpochCommName, endsEpochObj.Epoch)
	}

	srcComm, err := getSourceCommOfEpoch(ctx.Context, a.src, common.EpochNum(endsEpochObj.Epoch))
	if err != nil {
		return fmt.Errorf("src.Committee(epoch:%d) failed: %w", endsEpochObj.Epoch, err)
	}

	if !committeeEquals(srcComm, commObj.Comms) {
		return fmt.Errorf("target.0x%x.%s[1]=epoch:%d addrs(%s) not match with %s",
			to[:], xEndsOfEpochName, endsEpochObj.Epoch, commObj.Comms, srcComm)
	}

	log.Infof("target.0x%x.%s[1]=epoch:%d\naddrs:%s matchs:%s",
		to[:], xEndsOfEpochName, endsEpochObj.Epoch, common.IndentLevel(0).InfoString(commObj.Comms), srcComm.InfoString(0))
	return nil
}
