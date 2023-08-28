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
	"strings"
	"time"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/ThinkiumGroup/go-common/math"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/redis/go-redis/v9"
	"github.com/stephenfire/go-rtl"
	"github.com/urfave/cli/v2"
)

// let CanonicalTime = NowUnixTime - (NowUnixTime % Interval);
// if CanonicalTime > LastUpdateTIme, then update start
// update LastUpdateTime to now() when update is complete
type updater struct {
	runner
	lastUpdateTimeKey string
}

func (u *updater) Name() string {
	return fmt.Sprintf("UPDATE_%s", u.conf.TargetName)
}

func (u *updater) _lastEpochInLC(cctx context.Context) (common.EpochNum, error) {
	outobj := new(struct {
		Epoch uint64
	})
	if err := u.target.getter(cctx, u.targetPriv.Address(), &u.conf.Updater.TargetLCAddr,
		UpdatableLightNodeAbi.Methods[uLastEpochName], outobj); err != nil {
		return common.NilEpoch, fmt.Errorf("updatable lightnode.%s failed: %w", uLastEpochName, err)
	}
	return common.EpochNum(outobj.Epoch), nil
}

func (u *updater) _lastCommitteeInLC(cctx context.Context) ([]common.Address, error) {
	outobj := new(struct {
		Comm []common.Address
	})
	if err := u.target.getter(cctx, u.targetPriv.Address(), &u.conf.Updater.TargetLCAddr,
		UpdatableLightNodeAbi.Methods[uCheckEpochCommName], outobj); err != nil {
		return nil, fmt.Errorf("updatable lightnode.%s failed: %w", uCheckEpochCommName, err)
	}
	return outobj.Comm, nil
}

// return 0 means not found
func (u *updater) _lastUpdateTimeInCache(ctx context.Context) (int64, error) {
	cctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	l, err := u.redis.Get(cctx, u.lastUpdateTimeKey).Int64()
	switch {
	case err == redis.Nil:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return l, nil
	}
}

func (u *updater) _updateToLastUpdateTimeInCache(ctx context.Context, newValue int64) (int64, error) {
	cctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	if err := u.redis.Set(cctx, u.lastUpdateTimeKey, fmt.Sprintf("%d", newValue), 0).Err(); err != nil {
		return 0, err
	}
	return newValue, nil
}

func (u *updater) _updateLastUpdateTimeInCache(ctx context.Context) (int64, error) {
	now := time.Now().Unix() + 10 // 10 more seconds for ensurance
	return u._updateToLastUpdateTimeInCache(ctx, now)
}

func (u *updater) prepareConfig(ctx *cli.Context) error {
	initUpdatableLNAbi()
	if err := u.runner.prepareConfig(ctx); err != nil {
		return err
	}
	if ctx.Uint64(_updaterPostponeFlag.Name) > 0 {
		u.needs = u.needs.Clear(NeedSource)
		u.needs = u.needs.Clear(NeedTarget)
		u.needs = u.needs.Clear(NeedRunningLock)
	}
	u.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(u.Name()), u.conf.SrcChainId)
	u.lastUpdateTimeKey = fmt.Sprintf("%s_lastTimeStamp_%d", strings.ToLower(u.Name()), u.conf.SrcChainId)
	log.Infof("%s, lastUpdateKey: %s", u.keys, u.lastUpdateTimeKey)

	if _, exist := UpdatableLightNodeAbi.Events[uUpdateCommEvent]; !exist {
		return fmt.Errorf("event %s must be exist", uUpdateCommEvent)
	}

	u.conf.Updater.Interval = ctx.Uint64(_updaterIntervalFlag.Name)
	if u.conf.Updater.Interval > math.MaxInt64 {
		return errors.New("update.interval too big")
	}
	addr, err := stringToAddress(ctx, _updaterTargetLCFlag.Name)
	if err != nil {
		return err
	}
	u.conf.Updater.TargetLCAddr = addr
	if err := u.conf.Updater.validate(); err != nil {
		return err
	}
	u.conf.Updater.ForceEpoch = ctx.Uint64(_updaterForceEpochFlag.Name)
	models.SysContractLogger.Register(u.conf.Updater.TargetLCAddr, UpdatableLightNodeAbi)
	return nil
}

func (u *updater) confirmConfig(ctx *cli.Context) error {
	if err := u.runner.confirmConfig(ctx); err != nil {
		return err
	}

	if u.needs.Bool(NeedTarget) {
		// try lastEpoch
		if lastEpoch, err := u._lastEpochInLC(ctx.Context); err != nil {
			return fmt.Errorf("check last epoch in target:%x failed: %w", u.conf.Updater.TargetLCAddr[:], err)
		} else {
			log.Infof("lastEpoch of updatable light-node: %s", lastEpoch)
		}
	}

	if u.needs.Bool(NeedRedis) {
		// try lastUpdate
		if lastTime, err := u._lastUpdateTimeInCache(ctx.Context); err != nil {
			return fmt.Errorf("check last update time in cache failed: %w", err)
		} else {
			log.Infof("last update time: %s", unixSecondsString(lastTime))
		}
	}

	return nil
}

func (u *updater) doWork(ctx *cli.Context) error {
	postpone := ctx.Uint64(_updaterPostponeFlag.Name)
	if postpone > 0 {
		if postpone >= math.MaxInt64 {
			return cli.Exit(errors.New("too big for unix time"), ExitByInput)
		}
		lastUpdate, err := u._lastUpdateTimeInCache(ctx.Context)
		if err != nil {
			return cli.Exit(fmt.Errorf("get last time failed: %w", err), ExitRedisErr)
		}
		if postpone > uint64(math.MaxInt64-lastUpdate) {
			return cli.Exit(errors.New("unix time overflow"), ExitByInput)
		}
		if updated, err := u._updateToLastUpdateTimeInCache(ctx.Context, lastUpdate+int64(postpone)); err != nil {
			return cli.Exit(fmt.Errorf("update redis failed: %w", err), ExitRedisErr)
		} else {
			log.Infof("last update time from: %s update to: %s", unixSecondsString(lastUpdate), unixSecondsString(updated))
		}
		return nil
	}
	switch {
	case u.conf.Updater.ForceEpoch > 0:
		value, err := u.runningLock.Fetch(ctx.Context)
		if err != nil {
			return cli.Exit(fmt.Errorf("[%s] is running, fetch %s failed: %w", value, u.runningLock, err), ExitRunningLockErr)
		}
		defer func() {
			_ = u.runningLock.Release()
		}()
		epoch := common.EpochNum(u.conf.Updater.ForceEpoch)
		if epoch.IsNil() {
			return cli.Exit(errors.New("nil epoch"), ExitByInput)
		}
		comm, err := u._getCommitteeOfEpoch(ctx.Context, epoch)
		if err != nil {
			return cli.Exit(err, ExitSourceErr)
		}
		if err = u._forceUpdate(ctx.Context, epoch, comm); err != nil {
			return cli.Exit(err, ExitTargetErr)
		}
		return nil
	case u.conf.Updater.Interval <= 1:
		value, err := u.runningLock.Fetch(ctx.Context)
		if err != nil {
			return cli.Exit(fmt.Errorf("[%s] is running, fetch %s failed: %w", value, u.runningLock, err), ExitRunningLockErr)
		}
		defer func() {
			_ = u.runningLock.Release()
		}()
		return u._once(ctx)
	default:
		awake := time.Second * u.getFetchInterval()
		timer := time.NewTimer(awake)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				if err := u.connectionCheck(); err != nil {
					return err
				}
				value, err := u.runningLock.FetchOrRefresh(ctx.Context)
				if err != nil {
					log.Debugf("[%s] is running, fetch-refresh %s failed: %v", value, u.runningLock, err)
				} else {
					if err := u._once(ctx); err != nil {
						switch err.(type) {
						case cli.ExitCoder:
							return err
						default:
							var unlockErr LockError
							if errors.As(err, &unlockErr) && !unlockErr.Unlock() {
								log.Warnf("update failed and not release locks: %v", err)
							} else {
								log.Errorf("update failed and release locks: %v", err)
								// error occurred, release the locks so that other processes can take over
								// or waiting for connection reconnect
								_ = u.runningLock.Release()
								_ = u.sendingLock.Release()
							}
						}
					}
				}
				timer.Reset(awake)
			case <-ctx.Done():
				return cli.Exit(ctx.Err(), ExitByContext)
			}
		}
	}
}

func (u *updater) _canonicalTime() int64 {
	now := time.Now().Unix()
	interval := int64(u.conf.Updater.Interval)
	if interval <= 1 {
		return now
	}
	mod := now % interval
	now = now - mod
	return now
}

func (u *updater) _getCommitteeOfEpoch(ctx context.Context, epoch common.EpochNum) (*models.Committee, error) {
	return getSourceCommOfEpoch(ctx, u.src, epoch)
}

func (u *updater) _currentCommittee(ctx context.Context) (common.EpochNum, *models.Committee, error) {
	cctx, cancel := context.WithTimeout(ctx, reqTimeOut)
	defer cancel()

	stats, err := u.src.ChainStats(cctx)
	if err != nil {
		return common.NilEpoch, nil, fmt.Errorf("tkm stats failed: %w", err)
	}
	epochNum := common.Height(stats.CurrentHeight).EpochNum()
	comm := models.NewCommittee().SetMembers(stats.CurrentComm)
	if epochNum.IsNil() {
		return common.NilEpoch, nil, errors.New("invalid epoch")
	}
	if !comm.IsAvailable() {
		return common.NilEpoch, nil, fmt.Errorf("invalid committee: %s", comm)
	}
	return epochNum, comm, nil
}

func (u *updater) _parseULNUpdateEvent(l *models.Log) (*updateEvent, error) {
	if l == nil || len(l.Topics) < 3 {
		return nil, errors.New("invalid log")
	}
	epoch := rtl.Numeric.BytesToUint64(l.Topics[1][24:])
	return &updateEvent{
		Epoch:    common.EpochNum(epoch),
		CommHash: l.Topics[2],
	}, nil
}

func (u *updater) _updateCommittee(ctx context.Context, epoch common.EpochNum, comm *models.Committee) error {
	lockingValue, err := u.sendingLock.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("[%s] is sending, fetch %s failed: %w", lockingValue, u.sendingLock, err)
	}
	defer func() {
		_ = u.sendingLock.Release()
	}()
	gas, mustHave := u._targetSuggestBalance(ctx)
	nonce, err := u.target.nonceWithBalanceMoreThan(ctx, u.targetPriv.Address(), u.conf.TargetCheckBalance, mustHave)
	if err != nil {
		return fmt.Errorf("get nonce of %x failed: %w", u.targetPriv.Address().Bytes(), err)
	}
	input, err := UpdatableLightNodeAbi.Pack(uUpdateCommName, uint64(epoch), common.NodeIDs(comm.Members).ToBytesSlice())
	if err != nil {
		return fmt.Errorf("packinput failed: %w", err)
	}
	to := u.conf.Updater.TargetLCAddr

	ethtx, txhash, err := u.target.sendLegacyTx(ctx, u.targetPriv.Priv(), &to, nonce, gas, nil, nil, input)
	if err != nil {
		return fmt.Errorf("send tx failed: %w", err)
	}

	log.Infof("update comm TxHash: %x", common.ForPrint(txhash, 0))

	rcpt, err := u.target.checkReceipt(putDistributedLock(ctx, redisLocks{u.runningLock, u.sendingLock}), ethtx)
	if err != nil {
		return fmt.Errorf("get receipt failed: %w", err)
	}
	log.Debugf("%s", rcpt.InfoString(0))
	if !rcpt.Success() {
		return fmt.Errorf("tx failed: %w", rcpt.Err())
	}
	updateEvent := UpdatableLightNodeAbi.Events[uUpdateCommEvent]
	for _, l := range rcpt.Logs {
		if l != nil && len(l.Topics) > 0 && l.Topics[0] == updateEvent.ID {
			event, err := u._parseULNUpdateEvent(l)
			if err != nil {
				return fmt.Errorf("event parse failed: %w", err)
			}
			if epoch != event.Epoch {
				return fmt.Errorf("updates are not performing as expected: want Epoch:%d, got:%s", epoch, event)
			}
			commhash := comm.Hash()
			if event.CommHash != commhash {
				return fmt.Errorf("updates are not performing as expected: want Comm:%x, got:%s", commhash[:], event)
			}
			log.Infof("{Epoch:%s %s} updated %s", epoch, comm, event)
			return nil
		}
	}

	return errors.New("no update committee event found")
}

func (u *updater) _once(ctx *cli.Context) error {
	lastTime, err := u._lastUpdateTimeInCache(ctx.Context)
	if err != nil {
		return cli.Exit(fmt.Errorf("get last time failed: %w", err), ExitRedisErr)
	}
	now := u._canonicalTime()
	if now > lastTime {
		// do work one time
		epoch, comm, err := u._currentCommittee(ctx.Context)
		if err != nil {
			return fmt.Errorf("get current comm from TKM failed: %w", err)
		}
		return u._forceUpdate(ctx.Context, epoch, comm)
	} else {
		log.Debugf("lastUpdate: %s, canonical now: %s next: %s, ignoring update",
			unixSecondsString(lastTime), unixSecondsString(now), unixSecondsString(now+int64(u.conf.Updater.Interval)))
		return nil
	}
}

func (u *updater) _forceUpdate(ctx context.Context, epoch common.EpochNum, comm *models.Committee) error {
	if epoch.IsNil() || !comm.IsAvailable() {
		return fmt.Errorf("{Epoch:%s %s} not available", epoch, comm)
	} else {
		log.Infof("about to update {Epoch:%s %s} to target", epoch, comm)
	}
	if lastEpoch, err := u._lastEpochInLC(ctx); err != nil {
		log.Warnf("check target last epoch failed: %v", err)
	} else {
		if cmp := lastEpoch.Compare(epoch); cmp > 0 {
			log.Warnf("updating an older data {Epoch:%s %s} to LC.lastEpoch:%d", epoch, comm, lastEpoch)
		} else if cmp == 0 {
			// check committee equals
			if lastComm, err := u._lastCommitteeInLC(ctx); err != nil {
				log.Warnf("check target last committee on Epoch:%d failed: %v", lastComm, err)
			} else {
				if committeeEquals(comm, lastComm) {
					log.Warnf("{Epoch:%s %s} equals data in LC, ignoring update", epoch, comm)
					updatedTime, err := u._updateLastUpdateTimeInCache(ctx)
					if err != nil {
						return cli.Exit(fmt.Errorf("last update time failed in setting: %w", err), ExitRedisErr)
					}
					log.Infof("last update time set to: %s", unixSecondsString(updatedTime))
					return nil
				}
				log.Warnf("epoch same but Comm not match: %s <> %s", comm, lastComm)
			}
		}
	}

	if err := u._updateCommittee(ctx, epoch, comm); err != nil {
		return fmt.Errorf("update current comm failed: %w", err)
	}
	updatedTime, err := u._updateLastUpdateTimeInCache(ctx)
	if err != nil {
		return cli.Exit(fmt.Errorf("last update time failed in setting: %w", err), ExitRedisErr)
	}
	log.Infof("last update time set to: %s", unixSecondsString(updatedTime))
	return nil
}
