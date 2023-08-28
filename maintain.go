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

type maintainer struct {
	looper
}

func (a *maintainer) Name() string {
	return fmt.Sprintf("MAINTAIN_%s", a.conf.TargetName)
}

func (a *maintainer) prepareConfig(ctx *cli.Context) error {
	initTKMLNAbi()
	if err := a.looper.prepareConfig(ctx); err != nil {
		return err
	}
	a.keys.startHeightKey = fmt.Sprintf("%s_start_%d", strings.ToLower(a.Name()), a.conf.SrcChainId)
	a.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(a.Name()), a.conf.SrcChainId)
	log.Infof("%s", a.keys)

	if _, exist := LightNodeABI.Events[updateCommEvent]; !exist {
		return fmt.Errorf("event %s must be exist", updateCommEvent)
	}

	addr, err := stringToAddress(ctx, _maintainTargetLCFlag.Name)
	if err != nil {
		return err
	}
	a.conf.Maintainer.TargetLCAddr = addr
	if err := a.conf.Maintainer.validate(); err != nil {
		return err
	}
	models.SysContractLogger.Register(a.conf.Maintainer.TargetLCAddr, LightNodeABI)
	return nil
}

func (a *maintainer) confirmConfig(ctx *cli.Context) error {
	if err := a.looper.confirmConfig(ctx); err != nil {
		return err
	}

	// check availability of light-node and get lastHeight
	lastHeight := common.NilHeight
	input, err := LightNodeABI.Pack(lastHeightName)
	if err != nil {
		return fmt.Errorf("encode lightnode.lastHeight() failed: %w", err)
	}
	output, err := a.target.callContract(ctx.Context, a.targetPriv.Address(), &a.conf.Maintainer.TargetLCAddr, defaultGas, nil, nil, input)
	if err != nil {
		return fmt.Errorf("lightnode.lastHeight() failed: %w", err)
	}
	outobj := new(struct {
		Height uint64
	})
	if err = LightNodeABI.UnpackReturns(outobj, lastHeightName, output); err != nil {
		return fmt.Errorf("parse lastHeight failed: %w", err)
	}
	lastHeight = common.Height(outobj.Height)
	log.Infof("lastHeight of light-node: %s", &lastHeight)

	// update lastHeight
	newHeight := lastHeight + 1
	startHeight := a.getStartHeight(ctx)
	diff, cmp := startHeight.Diff(newHeight)
	if cmp < 0 {
		log.Warnf("replace start height from:%s to light-node.lastHeight+1: %s", &startHeight, &newHeight)
		if err := a.updateStartHeight(ctx, newHeight); err != nil {
			return fmt.Errorf("update start height (%d) failed: %w", newHeight, err)
		}
	} else if cmp > 0 {
		if diff >= common.BlocksInEpoch {
			return fmt.Errorf("startHeight:%s but light-node.lastHeight:%s", &startHeight, &lastHeight)
		}
	}
	log.Infof("start height: %d", a.getStartHeight(ctx))

	if ctx.Bool(_checkLNCommFlag.Name) {
		if err := a._checkComms(ctx); err != nil {
			return cli.Exit(err, ExitLCErr)
		}
		return cli.Exit(errors.New("comms checked"), 0)
	}

	return nil
}

func (a *maintainer) processBlock(cctx *cli.Context, block *models.BlockEMessage) (fatal, warning error) {
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

func (a *maintainer) _targetUpdateComm(cctx *cli.Context, comm *CommitteeProof) error {
	if err := comm.Verify(true); err != nil {
		return err
	}

	lockingValue, err := a.sendingLock.Fetch(cctx.Context)
	if err != nil {
		return fmt.Errorf("[%s] is sending, fetch %s failed: %w", lockingValue, a.sendingLock, err)
	}
	defer func() {
		_ = a.sendingLock.Release()
	}()

	proof := comm.ForABI()
	data, err := LightNodeABI.Methods[updateNextCommName].Inputs.Pack(proof)
	if err != nil {
		return fmt.Errorf("packdata failed: %w", err)
	}
	input, err := LightNodeABI.Pack(updateCommName, data)
	if err != nil {
		return fmt.Errorf("packinput failed: %w", err)
	}
	to := a.conf.Maintainer.TargetLCAddr

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
	updateEvent := LightNodeABI.Events[updateCommEvent]
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

func (a *maintainer) _parseLNUpdateEvent(l *models.Log) (*updateEvent, error) {
	if l == nil || len(l.Topics) < 3 {
		return nil, errors.New("invalid log")
	}
	epoch := rtl.Numeric.BytesToUint64(l.Topics[1][24:])
	return &updateEvent{
		Epoch:    common.EpochNum(epoch),
		CommHash: l.Topics[2],
	}, nil
}

func (a *maintainer) _checkOneLatestEpoch(ctx *cli.Context, index int64) error {
	from := a.targetPriv.Address()
	to := a.conf.Maintainer.TargetLCAddr

	latestObj := new(struct{ Epoch uint64 })
	latestObj.Epoch = uint64(common.NilEpoch)
	if err := a.target.getter(ctx.Context, from, &to,
		LightNodeABI.Methods[latest2EpochName], latestObj, big.NewInt(index)); err != nil {
		return fmt.Errorf("%s[index:%d] failed: %w", latest2EpochName, index, err)
	}
	if common.EpochNum(latestObj.Epoch).IsNil() {
		log.Warnf("target.%s[%d] not set", latest2EpochName, index)
		return nil
	}

	commObj := new(struct{ Comms []common.Address })
	if err := a.target.getter(ctx.Context, from, &to,
		LightNodeABI.Methods[checkEpochCommName], commObj, latestObj.Epoch); err != nil {
		return fmt.Errorf("%s[index:%d] -> %s(epoch:%d) failed: %w", latest2EpochName, index,
			checkEpochCommName, latestObj.Epoch, err)
	}
	if len(commObj.Comms) == 0 {
		return fmt.Errorf("%s(index:%d) -> %s(epoch:%d) got nothing", latest2EpochName, index,
			checkEpochCommName, latestObj.Epoch)
	}

	srcComm, err := getSourceCommOfEpoch(ctx.Context, a.src, common.EpochNum(latestObj.Epoch))
	if err != nil {
		return fmt.Errorf("src.Committee(epoch:%d) failed: %w", latestObj.Epoch, err)
	}

	if !committeeEquals(srcComm, commObj.Comms) {
		return fmt.Errorf("%s[index:%d]=epoch:%d addrs(%s) not match with %s",
			latest2EpochName, index, latestObj.Epoch, commObj.Comms, srcComm)
	}

	log.Infof("%s[index:%d]=epoch:%d\naddrs:%s matchs:%s",
		latest2EpochName, index, latestObj.Epoch, common.IndentLevel(0).InfoString(commObj.Comms), srcComm.InfoString(0))
	return nil
}

func (a *maintainer) _checkComms(ctx *cli.Context) error {
	if err := a._checkOneLatestEpoch(ctx, 0); err != nil {
		return err
	}
	if err := a._checkOneLatestEpoch(ctx, 1); err != nil {
		return err
	}
	return nil
}
