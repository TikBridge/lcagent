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
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ThinkiumGroup/go-common"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

type (
	Config struct {
		RedisAddr           string         // default redis://@127.0.0.1:6379/0
		RunningLockTTL      int64          // TTL for running lock key in redis (seconds)
		SendingLockTTL      int64          // TTL for sending lock key in redis (seconds)
		SrcFetchInterval    int64          // in seconds
		SrcRpcAddr          string         // no default (ip:port)
		SrcChainId          common.ChainID // 0 for maintainer
		SrcStartHeight      uint64         // start height
		SrcIgnoreBlocks     bool           // ignore blocks where its BlockNum<(EpochLength-100) in maintaining
		TargetName          string         // the unique name
		TargetApiAddr       string         // target chain eth_api address, no default (ip:addr)
		TargetChainID       *big.Int       // target chain id
		TargetRetryInterval int64          // retry to fetch receipt from target chain, in seconds
		TargetIsTKM         bool           // target chain is a Thinkium chain (for testing)
		TargetGPTTL         int64          // TTL of gasprice cache for target chain in seconds
		TargetCheckBalance  bool           // whether to check the balance in target
		Maintainer          Maintain
		Synchronizer        Synchronize
		Updater             Update
		XMaintainer         XMaintain
		XSynchronizer       XSynchronize
	}

	Maintain struct {
		TargetLCAddr common.Address // address of TKM light-client contract in target chain
	}

	XMaintain struct {
		TargetLCAddr        common.Address // address of X-Relay light-client contract in target chain
		XsyncStartHeightKey string
	}

	Synchronize struct {
		TkmChainId    *big.Int       // chain id of source chain of cross-chain tx
		TkmMCSAddress common.Address // address of contract map-cross-chain-service in source tkm chain
		TargetMSCAddr common.Address // address of contract map-cross-chain-service in target chain
		TargetLCAddr  common.Address // address of contract TKM Light-Node in target chain
		UpdatableLC   bool           // if the TKM Light-Node is updatable
		MaxHeightTTL  int64          // TTL for cache of max validatable sub-chain height in TKM-Light-Node
	}

	XSynchronize struct {
		XChainId      *big.Int       // chain id of source chain of cross-chain tx
		XMCSAddress   common.Address // address of contract map-cross-chain-service in source X-Relay chain
		TargetMSCAddr common.Address // address of contract map-cross-chain-service in target chain
		TargetLCAddr  common.Address // address of contract X-Light-Node in target chain
		MaxHeightTTL  int64          // TTL for cache of max validatable X-Relay height in X-Light-Node
	}

	Update struct {
		Interval     uint64         // in seconds
		TargetLCAddr common.Address // address of updatable light-client in target chain
		ForceEpoch   uint64         // force update values of the specific epoch, which >0
	}
)

func (m Maintain) validate() error {
	if m.TargetLCAddr == common.EmptyAddress {
		return errors.New("target light-client contract address is missing")
	}
	return nil
}

func (m XMaintain) validate() error {
	if m.TargetLCAddr == common.EmptyAddress {
		return errors.New("target light-client contract address is missing")
	}
	if m.XsyncStartHeightKey == "" {
		return errors.New("xsync start height key is missing")
	}
	return nil
}

func (s Synchronize) validate() error {
	if s.TkmChainId == nil || s.TkmChainId.Sign() < 0 {
		return errors.New("invalid TKM ChainID")
	}
	if s.TkmMCSAddress == common.EmptyAddress {
		return errors.New("TKM MapCrossChainService contract address missing")
	}
	if s.TargetMSCAddr == common.EmptyAddress {
		return errors.New("target MapCrossChainService contract address missing")
	}
	return nil
}

func (s XSynchronize) validate() error {
	if s.XChainId == nil || s.XChainId.Sign() < 0 {
		return errors.New("invalid X-Relay ChainID")
	}
	if s.XMCSAddress == common.EmptyAddress {
		return errors.New("X-Relay MapCrossChainServiceRelay contract address missing")
	}
	if s.TargetMSCAddr == common.EmptyAddress {
		return errors.New("target MapCrossChainService contract address missing")
	}
	return nil
}

func (u Update) validate() error {
	if u.TargetLCAddr == common.EmptyAddress {
		return errors.New("target updatable light-client contract address missing")
	}
	return nil
}

const (
	MaintainFlagCategory  = "MAINTAIN FLAGS"
	SyncFlagCategory      = "SYNCHRONIZE FLAGS"
	UpdaterFlagCategory   = "UDPATE FLAGS"
	XMaintainFlagCategory = "X-RELAY MAINTAIN FLAGS"
	XSyncFlagCategory     = "X-RELAY SYNCHRONIZE FLAGS"
	BasicCategory         = "BASIC"
	SourceCategory        = "SOURCE"
	TargetCategory        = "TARGET"
)

var (
	_confFileFlag = &cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"f"},
		Category: BasicCategory,
		Usage:    "configuation `FILE` in YAML format",
	}

	_redisFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "redis",
		Category: BasicCategory,
		Usage:    "redis server address",
		Value:    "redis://@127.0.0.1:6379/0",
	})

	_ttlRunningLcokFlag = altsrc.NewInt64Flag(&cli.Int64Flag{
		Name:     "runningLockTTL",
		Category: BasicCategory,
		Usage:    "TTL of agent running lock key in redis (seconds)",
		Value:    30,
	})

	_ttlSendingLockFlag = altsrc.NewInt64Flag(&cli.Int64Flag{
		Name:     "sendingLockTTL",
		Category: BasicCategory,
		Usage:    "TTL of agent sending lock key in redis (seconds)",
		Value:    30,
	})

	_intervalFlag = altsrc.NewIntFlag(&cli.IntFlag{
		Name:     "interval",
		Category: BasicCategory,
		Usage:    "The interval for reading block information on the Thinkium source chain (in second)",
		Value:    10,
	})

	_retryIntervalFlag = altsrc.NewInt64Flag(&cli.Int64Flag{
		Name:     "retryInterval",
		Category: BasicCategory,
		Usage:    "The interval for repeatedly reading the receipt on the target chain (in second)",
		Value:    5,
	})

	_logFileFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "log",
		Category: BasicCategory,
		Usage:    "log to `PATH`",
		Aliases:  []string{"l"},
	})

	_checkLNCommFlag = &cli.BoolFlag{
		Name:     "checkLNComm",
		Aliases:  []string{"cc"},
		Category: BasicCategory,
		Usage:    "check if the committee informations in target Light-Node contract match the infos in the source chain, only works for maintain/xmaintain",
		Value:    false,
	}

	_srcRpcFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "src.rpc",
		Category: SourceCategory,
		Usage:    "rpc address of source chain node",
	})

	_srcChainFlag = altsrc.NewUintFlag(&cli.UintFlag{
		Name:     "src.chainid",
		Category: SourceCategory,
		Usage:    "TKM-ChainID of source chain",
	})

	_srcBaseChainIDFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "src.basechainid",
		Category: SourceCategory,
		Usage:    "set source `BigChainIDBase` if the source TKM chain's ETH-Base-ChainID is not default",
	})

	_srcBlocksInEpochFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "src.blocksinepoch",
		Category: SourceCategory,
		Usage:    "set common.BlocksInEpoch if `BlocksInEpoch`>0",
	})

	_startHeightFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "src.start",
		Category: SourceCategory,
		Usage:    "starting height of source chain",
	})

	_srcIgnoreBlocks = altsrc.NewBoolFlag(&cli.BoolFlag{
		Name:     "src.ignoreblocks",
		Category: SourceCategory,
		Usage:    "whether ignoring blocks where their BlockNum<(EpochLength-100) in maintaining",
		Value:    false,
	})

	_targetNameFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "target.name",
		Category: TargetCategory,
		Usage:    "`UNIQUE_NAME` used to distinguish processes of the same type with different targets, same target same type MUST use same name",
	})

	_targetApiFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "target.api",
		Category: TargetCategory,
		Usage:    "Ethereum-like API address of target chain",
	})

	_targetChainIDFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "target.chainid",
		Category: TargetCategory,
		Usage:    "target Ethereum-like chain id",
	})

	_targetPrivFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "target.senderkey",
		Category: TargetCategory,
		Usage:    "account private key used for send transaction on target chain, override the value in target.senderpem",
	})

	_targetPEMFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "target.senderpem",
		Category: TargetCategory,
		Usage:    "`PEM_FIlE_PATH` is a PEM-Encoded PKCS#8 private key file, ",
	})

	_targetPEMPwdFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "target.senderpempwd",
		Category: TargetCategory,
		Usage:    "password of the private key file",
		Aliases:  []string{"pwd"},
	})

	_targetIsTKM = altsrc.NewBoolFlag(&cli.BoolFlag{
		Name:     "target.istkm",
		Category: TargetCategory,
		Usage:    "whether the target chain is a Thinkium chain (for testing)",
		Value:    false,
	})

	_targetGPTTL = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "target.gpttl",
		Category: TargetCategory,
		Usage:    "TTL of target chain suggest GasPrice cache (in seconds)",
		Value:    60 * 10,
	})

	_targetCheckBalance = altsrc.NewBoolFlag(&cli.BoolFlag{
		Name:     "target.checkbalance",
		Category: TargetCategory,
		Usage:    "whether to check the balance of sender before sending tx",
		Value:    true,
	})

	_maintainTargetLCFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "maintain.targetlc",
		Category: MaintainFlagCategory,
		Usage:    "the address of Thinkium Light-Client contract on target chain",
	})

	_syncTkmChainIDFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "sync.tkmchainid",
		Category: SyncFlagCategory,
		Usage:    "the TKM ETH-ChainID used for target mcs.transferIn function",
	})

	_syncTkmMCSFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "sync.tkmmcs",
		Category: SyncFlagCategory,
		Usage:    "the address of Map-Cross-Chain-Service contract on source Thinkium chain",
	})

	_syncTargetMCSFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "sync.targetmcs",
		Category: SyncFlagCategory,
		Usage:    "the address of Map-Cross-Chain-Service-Relay contract on target chain",
	})

	_syncTargetLCFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "sync.targetlc",
		Category: SyncFlagCategory,
		Usage:    "the address of Thinkium Light-Client contract on target chain",
	})

	_syncUpdatableLCFlag = altsrc.NewBoolFlag(&cli.BoolFlag{
		Name:     "sync.updatablelc",
		Category: SyncFlagCategory,
		Usage:    "whether the Thinkium Light-Client is updatable by admin",
	})

	_syncMaxHeightTTLFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "sync.maxheightttl",
		Category: SyncFlagCategory,
		Usage:    "TTL of cache for the max validatable sub-chain height (in seconds)",
		Value:    60,
	})

	// TODO: due to bug in urfave/cli/v2.25.7, which always get 0 for nested int64
	_updaterIntervalFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "update.interval",
		Category: UpdaterFlagCategory,
		Usage:    "interval seconds for every updates",
		Value:    60 * 60 * 6,
	})

	_updaterTargetLCFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "update.targetlc",
		Category: UpdaterFlagCategory,
		Usage:    "the address of Thinkium updatable Light-Client contract on target chain",
	})

	_updaterForceEpochFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "update.epoch",
		Aliases:  []string{"ep"},
		Category: UpdaterFlagCategory,
		Usage:    "force update epoch `EPOCH` and committee to TKM Light-Client on target chain",
	})

	_updaterPostponeFlag = &cli.Uint64Flag{
		Name:     "postpone",
		Category: UpdaterFlagCategory,
		Usage:    "postpone next updatation for `POSTPONE` seconds",
		Aliases:  []string{"p"},
	}

	_pemOutputFlag = &cli.StringFlag{
		Name:    "output",
		Usage:   "output PEM `FILE_PATH`",
		Aliases: []string{"o"},
	}

	_pemInputFlag = &cli.StringFlag{
		Name:    "input",
		Usage:   "input PEM `FILE_PATH`",
		Aliases: []string{"i"},
	}

	_xmaintainTargetLCFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "xmaintain.targetlc",
		Category: XMaintainFlagCategory,
		Usage:    "the address of X-Relay Light-Client contract on target chain",
	})

	_xmaintainSyncStartHeightKeyFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "xmaintain.xsyncstartheightkey",
		Category: XMaintainFlagCategory,
		Usage:    "the redis key of the corresponding xsync start height key",
	})

	_xSyncChainIDFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "xsync.chainid",
		Category: XSyncFlagCategory,
		Usage:    "the x-relay ETH-ChainID used for target mcs.transferIn function",
	})

	_xSyncMCSFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "xsync.mcs",
		Category: XSyncFlagCategory,
		Usage:    "the address of Map-Cross-Chain-Service-Relay contract on X-Relay chain",
	})

	_xSyncTargetMCSFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "xsync.targetmcs",
		Category: XSyncFlagCategory,
		Usage:    "the address of Map-Cross-Chain-Service contract on target chain",
	})

	_xSyncTargetLCFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:     "xsync.targetlc",
		Category: XSyncFlagCategory,
		Usage:    "the address of X-Relay Light-Client contract on target chain",
	})

	_xSyncMaxHeightTTLFlag = altsrc.NewUint64Flag(&cli.Uint64Flag{
		Name:     "xsync.maxheightttl",
		Category: XSyncFlagCategory,
		Usage:    "TTL of cache for the max validatable x-relay height (in seconds)",
		Value:    60,
	})

	_allFlags = []cli.Flag{
		_confFileFlag,
		_redisFlag,
		_ttlRunningLcokFlag,
		_ttlSendingLockFlag,
		_intervalFlag,
		_srcRpcFlag,
		_srcChainFlag,
		_srcBaseChainIDFlag,
		_srcBlocksInEpochFlag,
		_startHeightFlag,
		_srcIgnoreBlocks,
		_targetNameFlag,
		_targetApiFlag,
		_targetChainIDFlag,
		_targetPrivFlag,
		_targetPEMFlag,
		_targetPEMPwdFlag,
		_targetIsTKM,
		_targetGPTTL,
		_targetCheckBalance,
		_retryIntervalFlag,
		_logFileFlag,
		_checkLNCommFlag,
	}

	_maintainFlags = []cli.Flag{
		_maintainTargetLCFlag,
	}

	_syncFlags = []cli.Flag{
		_syncTkmChainIDFlag,
		_syncTkmMCSFlag,
		_syncTargetMCSFlag,
		_syncTargetLCFlag,
		_syncUpdatableLCFlag,
		_syncMaxHeightTTLFlag,
	}

	_updateFlags = []cli.Flag{
		_updaterTargetLCFlag,
		_updaterIntervalFlag,
		_updaterForceEpochFlag,
		_updaterPostponeFlag,
	}

	_pemFlags = []cli.Flag{
		_pemOutputFlag,
		_pemInputFlag,
	}

	_xmaintainFlags = []cli.Flag{
		_xmaintainTargetLCFlag,
	}

	_xSyncFlags = []cli.Flag{
		_xSyncChainIDFlag,
		_xSyncMCSFlag,
		_xSyncTargetMCSFlag,
		_xSyncTargetLCFlag,
		_xSyncMaxHeightTTLFlag,
	}
)

func stringToAddress(ctx *cli.Context, name string) (common.Address, error) {
	bs, err := hex.DecodeString(ctx.String(name))
	if err != nil || len(bs) != common.AddressLength {
		return common.Address{}, fmt.Errorf("invalid %s", name)
	}
	return common.BytesToAddress(bs), nil
}
