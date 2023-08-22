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
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/ThinkiumGroup/go-tkmrpc"
	"github.com/ThinkiumGroup/go-tkmrpc/client"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stephenfire/go-rtl"
	"github.com/urfave/cli/v2"
)

const (
	redisTimeout   = 1 * time.Second
	runningLockTTL = 30 * time.Second
	sendingLockTTL = 30 * time.Second

	senderLockPrefix = "targetSender"
)

type basicHandler interface {
	Name() string
	prepareConfig(ctx *cli.Context) error
	confirmConfig(ctx *cli.Context) error
	doWork(ctx *cli.Context) error
}

const (
	NeedSource = iota
	NeedTarget
	NeedRedis
	NeedRunningLock
)

type BitFlags big.Int

func (f *BitFlags) Bool(bit int) bool {
	if f == nil {
		return false
	}
	return (*big.Int)(f).Bit(bit) == 1
}

func (f *BitFlags) Set(bit int) *BitFlags {
	var i *big.Int
	if f != nil {
		i = (*big.Int)(f)
	} else {
		i = big.NewInt(0)
	}
	i.SetBit(i, bit, 1)
	return (*BitFlags)(i)
}

func (f *BitFlags) Clear(bit int) *BitFlags {
	if f == nil {
		return f
	}
	i := (*big.Int)(f)
	i.SetBit(i, bit, 0)
	return f
}

type runner struct {
	conf     *Config
	needs    *BitFlags
	src      *client.Client
	target   *EthClient
	redis    *redis.Client
	bHandler basicHandler

	keys        redisKeys
	runningLock *redisLock
	sendingLock *redisLock

	// local value
	targetPriv common.Identifier
	once       atomic.Bool
}

func (a *runner) Name() string {
	return fmt.Sprintf("RUNNER_%s", a.conf.TargetName)
}

func (a *runner) String() string {
	if a.bHandler == nil {
		return fmt.Sprintf("%s{%s}", a.Name(), a.keys.runnerLockValue)
	}
	return fmt.Sprintf("%s{%s}", a.bHandler.Name(), a.keys.runnerLockValue)
}

func (a *runner) connectionCheck() error {
	// TODO: try to find out that the link is broken and reconnect
	return nil
}

func (a *runner) getFetchInterval() time.Duration {
	return time.Duration(a.conf.SrcFetchInterval)
}

func (a *runner) _targetKey(ctx *cli.Context) (common.Identifier, error) {
	str := ctx.String(_targetPrivFlag.Name)
	if str == "" {
		return nil, nil
	}
	priv, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	sender, err := models.NewIdentifier(priv)
	if err != nil {
		return nil, fmt.Errorf("identifer failed: %w", err)
	}
	return sender, nil
}

func (a *runner) _targetPEM(ctx *cli.Context) (common.Identifier, error) {
	path := ctx.String(_targetPEMFlag.Name)
	if path == "" {
		return nil, nil
	}
	filebytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read PEM failed: %w", err)
	}
	pwdstr := ctx.String(_targetPEMPwdFlag.Name)
	needPwd, pkcs8Bytes, err := ValidPEM(filebytes)
	if err != nil {
		return nil, fmt.Errorf("validate PEM failed: %w", err)
	}
	pwd := []byte(pwdstr)
	if needPwd && pwdstr == "" {
		pwd, err = readPwd("please input the password of PEM: ", "", "")
		if err != nil {
			return nil, fmt.Errorf("read password failed: %w", err)
		}
	}
	sk, err := ParsePKCS8PrivateKey(pkcs8Bytes, pwd)
	if err != nil {
		log.Debugf("parse pem failed: %v", err)
		return nil, errors.New("PEM error")
	}
	priv := ETHSigner.PrivToBytes(sk)
	sender, err := models.NewIdentifier(priv)
	if err != nil {
		return nil, fmt.Errorf("identifer failed: %w", err)
	}
	return sender, nil
}

func (a *runner) _targetSender(ctx *cli.Context) (common.Identifier, error) {
	sender, err := a._targetKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid target.senderkey: %w", err)
	}
	if sender != nil {
		return sender, nil
	}
	sender, err = a._targetPEM(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid target.senderpem: %w", err)
	}
	if sender == nil {
		return nil, errors.New("sender is missing")
	}
	return sender, nil
}

func (a *runner) prepareConfig(ctx *cli.Context) error {
	sender, err := a._targetSender(ctx)
	if err != nil {
		return fmt.Errorf("invalid sender key: %w", err)
	}
	a.targetPriv = sender
	log.Infof("target.sender: 0x%x", sender.Address().Bytes())
	conf := &Config{
		RedisAddr:           ctx.String(_redisFlag.Name),
		SrcFetchInterval:    ctx.Int64(_intervalFlag.Name),
		SrcRpcAddr:          ctx.String(_srcRpcFlag.Name),
		SrcChainId:          common.ChainID(ctx.Uint64(_srcChainFlag.Name)),
		TargetName:          strings.ToUpper(ctx.String(_targetNameFlag.Name)),
		TargetApiAddr:       ctx.String(_targetApiFlag.Name),
		TargetRetryInterval: ctx.Int64(_retryIntervalFlag.Name),
		TargetIsTKM:         ctx.Bool(_targetIsTKM.Name),
		TargetGPTTL:         int64(ctx.Uint64(_targetGPTTL.Name)),
		SrcStartHeight:      ctx.Uint64(_startHeightFlag.Name),
	}
	if cid := ctx.Uint64(_targetChainIDFlag.Name); cid > 0 {
		conf.TargetChainID = new(big.Int).SetUint64(cid)
	}
	if conf.TargetName == "" {
		return cli.Exit(errors.New("target.name required"), ExitByConfig)
	}
	a.conf = conf
	retryInterval = time.Duration(conf.TargetRetryInterval)
	a.keys.startHeightKey = fmt.Sprintf("%s_start_%d", strings.ToLower(a.Name()), conf.SrcChainId)
	a.keys.runnerLockKey = fmt.Sprintf("%s_lock_%d", strings.ToLower(a.Name()), conf.SrcChainId)
	a.keys.runnerLockValue = a._runnerLockValue()
	a.keys.senderLockKey = fmt.Sprintf("%s_%d_0x%x", senderLockPrefix, a.conf.TargetChainID, a.targetPriv.Address().Bytes())
	a.needs = a.needs.Set(NeedSource)
	a.needs = a.needs.Set(NeedTarget)
	a.needs = a.needs.Set(NeedRedis)
	a.needs = a.needs.Set(NeedRunningLock)

	return nil
}

func (a *runner) _ipAndPid() (string, int) {
	ip, err := localip()
	if err != nil {
		panic(err)
	}
	return ip, os.Getpid()
}
func (a *runner) _runnerLockValue() string {
	ip, pid := a._ipAndPid()
	return fmt.Sprintf("%s@%d", ip, pid)
}

func (a *runner) confirmConfig(cctx *cli.Context) error {
	if a.needs.Bool(NeedRedis) {
		ctx, cancel := context.WithTimeout(cctx.Context, redisTimeout*2)
		defer cancel()
		status := a.redis.Ping(ctx)
		if status.Val() != "PONG" {
			return errors.New("ping redis server failed")
		}
		info := a.redis.Info(ctx, "server")
		r := bufio.NewReader(bytes.NewBufferString(info.Val()))
		version := ""
		for i := 0; i < 20; i++ {
			line, _, err := r.ReadLine()
			if len(line) > 0 {
				if strings.Contains(string(line), "redis_version") {
					parts := strings.Split(string(line), ":")
					if len(parts) > 1 {
						version = strings.TrimSpace(parts[1])
					}
					break
				}
			}
			if err != nil {
				break
			}
		}
		log.Infof("redis server version: %s", version)
	}

	return nil
}

func (a *runner) _connectSource(ctx *cli.Context) (c *client.Client, errr error) {
	if !a.needs.Bool(NeedSource) {
		return nil, nil
	}
	conn := &client.Client{
		Server:       a.conf.SrcRpcAddr,
		CurrentChain: a.conf.SrcChainId,
	}
	if err := conn.NewClient(); err != nil {
		return nil, fmt.Errorf("connect TKM @%s failed: %w", a.conf.SrcRpcAddr, err)
	}

	defer func() {
		if errr != nil {
			c = nil
			_ = conn.Close()
		}
	}()

	stats, err := conn.ChainStats(ctx.Context)
	if err != nil {
		return nil, fmt.Errorf("tkm stats failed: %w", err)
	}
	if stats == nil || stats.ChainID != a.conf.SrcChainId {
		return nil, fmt.Errorf("TKM@%s ChainID:%d required, but %s", a.conf.SrcRpcAddr, a.conf.SrcChainId, stats.String())
	}
	log.Infof("TKM@%s connected: %s", a.conf.SrcRpcAddr, stats)
	return conn, nil
}

func (a *runner) _connectTarget(ctx *cli.Context) (c *EthClient, errr error) {
	if !a.needs.Bool(NeedTarget) {
		return nil, nil
	}
	var chainid *big.Int
	if a.conf.TargetChainID != nil {
		chainid = new(big.Int).Set(a.conf.TargetChainID)
	}
	cl, err := NewEthClient(ctx.Context, a.conf.TargetApiAddr, chainid, a.conf.TargetGPTTL)
	if err != nil || cl == nil {
		return nil, fmt.Errorf("connect TARGET@%s failed: %w", a.conf.TargetApiAddr, err)
	}

	if a.conf.TargetChainID == nil {
		a.conf.TargetChainID = new(big.Int).Set(cl.ChainId)
		a.keys.senderLockKey = fmt.Sprintf("%s_%d_0x%x", senderLockPrefix, a.conf.TargetChainID, a.targetPriv.Address().Bytes())
		log.Infof("SENDER_LOCK_KEY_UPDATED: %s", a.keys)
	}
	return cl, nil
}

func (a *runner) _targetSuggestBalance(ctx context.Context) (gas uint64, mustHave *big.Int) {
	gas = defaultGas
	gasprice, err := a.target.suggestGasPrice(ctx)
	if err != nil || gasprice == nil {
		return gas, nil
	}
	mustHave = new(big.Int).Mul(new(big.Int).SetUint64(gas), gasprice)
	return gas, mustHave
}

func (a *runner) start(ctx *cli.Context) (errr error) {
	if a.once.CompareAndSwap(false, true) {
		ip, pid := a._ipAndPid()
		log.SetFields(logrus.Fields{ip: pid})
		defer func() {
			if errr != nil {
				_ = a.close(ctx)
			}
		}()
		conn, err := a._connectSource(ctx)
		if err != nil {
			return err
		}
		a.src = conn

		cl, err := a._connectTarget(ctx)
		if err != nil {
			return err
		}
		a.target = cl

		if a.needs.Bool(NeedRedis) {
			opts, err := redis.ParseURL(a.conf.RedisAddr)
			if err != nil {
				return cli.Exit(fmt.Errorf("redis URL parse failed: %w", err), ExitByConfig)
			}
			log.Debugf("connecting redis at ADDRESS: %s, DB: %d", opts.Addr, opts.DB)
			rds := redis.NewClient(opts)
			a.redis = rds
			locker := redislock.New(a.redis)
			a.runningLock = newRedisLock(a.redis, locker, a.keys.runnerLockKey, a.keys.runnerLockValue, runningLockTTL)
			a.sendingLock = newRedisLock(a.redis, locker, a.keys.senderLockKey, a.keys.runnerLockValue, sendingLockTTL)
		} else {
			a.redis = nil
			a.runningLock = nil
			a.sendingLock = nil
		}

		if a.bHandler != nil {
			if err := a.bHandler.confirmConfig(ctx); err != nil {
				return err
			}
		} else {
			if err := a.confirmConfig(ctx); err != nil {
				return err
			}
		}

		log.Infof("config: %+v", a.conf)
		log.Infof("%s STARTED", a.String())
		return nil
	} else {
		return errors.New("restart error")
	}
}

func (a *runner) close(_ *cli.Context) error {
	if a.once.CompareAndSwap(true, false) {
		if a.src != nil {
			_ = a.src.Close()
			a.src = nil
		}
		if a.target != nil {
			a.target.Close()
			a.target = nil
		}
		if a.sendingLock != nil {
			// use an available context for making sure to release
			_ = a.sendingLock.Release()
		}
		if a.runningLock != nil {
			_ = a.runningLock.Release()
		}
		if a.redis != nil {
			_ = a.redis.Close()
			a.redis = nil
		}
		log.Warnf("%s CLOSED", a.String())
		return nil
	} else {
		return errors.New("closing an not started runner")
	}
}

func (a *runner) run(cctx *cli.Context) error {
	if logpath := cctx.String(_logFileFlag.Name); len(logpath) > 0 {
		pid := strconv.Itoa(os.Getpid())
		log.InitLogWithSuffix(logpath, pid)
	}
	// set common.BlocksInEpoch
	if blocksInEpoch := cctx.Uint64(_srcBlocksInEpochFlag.Name); blocksInEpoch > 0 {
		common.BlocksInEpoch = blocksInEpoch
		log.Warnf("common.BlocksInEpoch = %d", common.BlocksInEpoch)
		log.SetFields(logrus.Fields{"Blocks": common.BlocksInEpoch})
	}
	// set common.BigChainIDBase if needed
	if baseChainID := cctx.Uint64(_srcBaseChainIDFlag.Name); baseChainID > 0 {
		common.BigChainIDBase = baseChainID
		log.Warnf("common.BigChainIDBase = %d", baseChainID)
		log.SetFields(logrus.Fields{"Base": common.BigChainIDBase})
	}

	if a.bHandler != nil {
		if err := a.bHandler.prepareConfig(cctx); err != nil {
			return err
		}
	} else {
		if err := a.prepareConfig(cctx); err != nil {
			return err
		}
	}
	log.SetFields(logrus.Fields{a.bHandler.Name(): cctx.Uint(_srcChainFlag.Name)})
	defer func() {
		_ = a.close(cctx)
	}()
	if err := a.start(cctx); err != nil {
		return err
	}
	if a.bHandler != nil {
		if err := a.bHandler.doWork(cctx); err != nil {
			return err
		}
		return nil
	} else {
		return cli.Exit(errors.New("basic handler is missing"), ExitBasicHandlerErr)
	}
}

type looperHandler interface {
	prepareToGet(cctx *cli.Context, start common.Height) error
	processBlocks(cctx *cli.Context, blocks *client.RpcBlocks) (next common.Height, errr error)
	prepareBlocks(cctx *cli.Context, blocks *client.RpcBlocks) (goon bool, err error)
	processBlock(cctx *cli.Context, block *models.BlockEMessage) (fatal, warning error)
}

type looper struct {
	runner
	lHander looperHandler
}

func (a *looper) Name() string {
	return fmt.Sprintf("LOOPER_%s", a.conf.TargetName)
}

//
// func (a *looper) prepareConfig(ctx *cli.Context) error {
// 	if err := a.runner.prepareConfig(ctx); err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// func (a *looper) confirmConfig(ctx *cli.Context) error {
// 	if err := a.runner.confirmConfig(ctx); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (a *looper) doWork(ctx *cli.Context) error {
	interval := time.Second * a.getFetchInterval()
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			if err := a.connectionCheck(); err != nil {
				return err
			}
			value, err := a.runningLock.FetchOrRefresh(ctx.Context)
			if err != nil {
				log.Debugf("[%s] is running, fetch-refresh %s failed: %v", value, a.runningLock, err)
			} else {
				if err := a.iterateBlocks(ctx); err != nil {
					var exitErr cli.ExitCoder
					if errors.As(err, &exitErr) {
						return err
					} else {
						var unlockErr LockError
						if errors.As(err, &unlockErr) && !unlockErr.Unlock() {
							log.Warnf("iterate blocks failed and not release locks: %v", err)
						} else {
							log.Errorf("iterate blocks failed and release locks: %v", err)
							// error occurred, release the locks so that other processes can take over
							// or waiting for connection reconnect
							_ = a.runningLock.Release()
							_ = a.sendingLock.Release()
						}
					}
				}
			}
			timer.Reset(interval)
		case <-ctx.Done():
			return cli.Exit(ctx.Err(), ExitByContext)
		}
	}
}

func (a *looper) getStartHeight(cctx *cli.Context) common.Height {
	ctx, cancel := context.WithTimeout(cctx.Context, redisTimeout)
	defer cancel()
	h, err := a.redis.Get(ctx, a.keys.startHeightKey).Uint64()
	switch {
	case err != nil:
		return common.Height(a.conf.SrcStartHeight)
	default:
		return common.Height(h)
	}
}

func (a *looper) updateStartHeight(cctx *cli.Context, newHeight common.Height) error {
	ctx, cancel := context.WithTimeout(cctx.Context, redisTimeout)
	defer cancel()
	return a.redis.Set(ctx, a.keys.startHeightKey, fmt.Sprintf("%d", newHeight), 0).Err()
}

func (a *looper) prepareToGet(_ *cli.Context, _ common.Height) error {
	return nil
}

func (a *looper) _tkmBlocks(cctx *cli.Context, start common.Height) (*client.RpcBlocks, error) {
	if err := a.lHander.prepareToGet(cctx, start); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(cctx.Context, reqTimeOut)
	defer cancel()
	resp, err := a.src.NodeClient.GetBlocks(ctx, &tkmrpc.RpcBlockHeight{
		Chainid: uint32(a.conf.SrcChainId),
		Height:  uint64(start),
	})
	if err != nil {
		return nil, fmt.Errorf("get blocks starting at %d failed: %w", start, err)
	}
	if !resp.Success() {
		log.Warnf("get blocks starting at %d failed, wait another fetch", start)
		return nil, nil
	}
	blocks := new(client.RpcBlocks)
	if err = rtl.Unmarshal(resp.Stream, blocks); err != nil {
		return nil, fmt.Errorf("unmarshal blocks failed: %w", err)
	}
	log.Infof("get %s starting at %d", blocks, start)
	if blocks == nil || len(blocks.Blocks) == 0 {
		return nil, nil
	}
	return blocks, nil
}

func (a *looper) iterateBlocks(cctx *cli.Context) error {
	if a.lHander == nil {
		return cli.Exit(errors.New("looper handler is missing"), ExitLooperHandlerErr)
	}
	start := a.getStartHeight(cctx)
	if start.IsNil() {
		start = 0
	}
	for {
		select {
		case <-cctx.Done():
			return cli.Exit(cctx.Err(), ExitByContext)
		default:
			_ = a.runningLock.Refresh(cctx.Context)
			blocks, err := a._tkmBlocks(cctx, start)
			if err != nil {
				return err
			}
			if blocks == nil || len(blocks.Blocks) == 0 {
				return nil
			}
			next, err := a.lHander.processBlocks(cctx, blocks)
			if err != nil {
				return err
			}
			if next.Compare(start) >= 0 {
				start = next
				if err = a.updateStartHeight(cctx, start); err != nil {
					log.Warnf("update start height to %s failed: %v", &start, err)
				}
			} else {
				log.Warnf("looperHandler next height (%s) less than start (%s)", &next, &start)
			}
			if start.Compare(blocks.Current) > 0 {
				return nil
			}
			time.Sleep(time.Second)
		}
	}
}

func (a *looper) processBlocks(cctx *cli.Context, blocks *client.RpcBlocks) (next common.Height, errr error) {
	if goon, err := a.lHander.prepareBlocks(cctx, blocks); err != nil {
		return common.NilHeight, err
	} else if !goon {
		return common.NilHeight, nil
	}
	if blocks == nil || len(blocks.Blocks) == 0 {
		return common.NilHeight, nil
	}
	start := blocks.Blocks[0].GetHeight()
	for i, block := range blocks.Blocks {
		select {
		case <-cctx.Done():
			return start, cli.Exit(cctx.Err(), ExitByContext)
		default:
			if block != nil && block.BlockHeader != nil && block.BlockBody != nil {
				if fatal, warning := a.lHander.processBlock(cctx, block); fatal != nil {
					return start, fmt.Errorf("processing %d/%d %s fatal: %w", i, len(blocks.Blocks), block.String(), fatal)
				} else if warning != nil {
					log.Warnf("processing %d/%d %s warned: %v", i, len(blocks.Blocks), block.String(), warning)
				}
				start = block.GetHeight() + 1
				if err := a.updateStartHeight(cctx, start); err != nil {
					log.Warnf("%d/%d: update start height to %s failed: %v", i, len(blocks.Blocks), &start, err)
				}
			}
		}
	}
	return start, nil
}

func (a *looper) prepareBlocks(_ *cli.Context, _ *client.RpcBlocks) (goon bool, err error) {
	return true, nil
}

func (a *looper) processBlock(_ *cli.Context, _ *models.BlockEMessage) (fatal, warning error) {
	return common.ErrUnsupported, nil
}
