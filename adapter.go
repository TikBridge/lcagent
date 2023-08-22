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

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/math"
	"github.com/ThinkiumGroup/go-common/trie"
	"github.com/ThinkiumGroup/go-tkmrpc/client"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	Eth2TKM       struct{}
	TKM2Eth       struct{}
	TKM2LightNode struct{}
)

var (
	E2T  = Eth2TKM{}
	T2E  = TKM2Eth{}
	T2LN = TKM2LightNode{}
)

func (te TKM2Eth) Address(addr common.Address) common2.Address {
	return common2.Address(addr)
}

func (te TKM2Eth) AddressP(addr *common.Address) *common2.Address {
	if addr == nil {
		return nil
	}
	a := te.Address(*addr)
	return &a
}

func (te TKM2Eth) Hash(h common.Hash) common2.Hash {
	return common2.Hash(h)
}

func (et Eth2TKM) Address(addr common2.Address) common.Address {
	return common.Address(addr)
}

func (et Eth2TKM) AddressP(addr *common2.Address) *common.Address {
	if addr == nil {
		return nil
	}
	r := et.Address(*addr)
	return &r
}

func (et Eth2TKM) Hash(h common2.Hash) common.Hash {
	return common.Hash(h)
}

func (et Eth2TKM) HashP(h *common2.Hash) *common.Hash {
	if h == nil {
		return nil
	}
	r := et.Hash(*h)
	return &r
}

func (et Eth2TKM) Hashs(hs []common2.Hash) []common.Hash {
	if hs == nil {
		return nil
	}
	ret := make([]common.Hash, len(hs))
	for i, h := range hs {
		ret[i] = et.Hash(h)
	}
	return ret
}

func (et Eth2TKM) AccessTuple(t types.AccessTuple) models.AccessTuple {
	return models.AccessTuple{
		Address:     et.Address(t.Address),
		StorageKeys: et.Hashs(t.StorageKeys),
	}
}

func (et Eth2TKM) AccessList(al types.AccessList) models.AccessList {
	if al == nil {
		return nil
	}
	ret := make(models.AccessList, len(al))
	for i, t := range al {
		ret[i] = et.AccessTuple(t)
	}
	return ret
}

func (et Eth2TKM) Log(l *types.Log) *models.Log {
	if l == nil {
		return nil
	}
	return &models.Log{
		Address:     et.Address(l.Address),
		Topics:      et.Hashs(l.Topics),
		Data:        l.Data,
		BlockNumber: l.BlockNumber,
		TxHash:      et.Hash(l.TxHash),
		TxIndex:     l.TxIndex,
		Index:       l.Index,
		BlockHash:   et.HashP(&l.BlockHash),
	}
}

func (et Eth2TKM) Logs(ls []*types.Log) []*models.Log {
	if ls == nil {
		return nil
	}
	ret := make([]*models.Log, len(ls))
	for i, l := range ls {
		ret[i] = et.Log(l)
	}
	return ret
}

func (et Eth2TKM) _txWithChainID(cid common.ChainID, tx *types.Transaction) (*models.Transaction, error) {
	var extrakeys = new(models.Extra)
	v, r, s := tx.RawSignatureValues()
	if !models.AvailableSignatureValues(v, r, s) {
		return nil, errors.New("available signature values are missing")
	}
	switch tx.Type() {
	case models.LegacyTxType:
		extrakeys = &models.Extra{
			Type:     models.LegacyTxType,
			GasPrice: tx.GasPrice(),
			V:        v,
			R:        r,
			S:        s,
		}
	case models.AccessListTxType:
		extrakeys = &models.Extra{
			Type:       models.AccessListTxType,
			GasPrice:   tx.GasPrice(),
			AccessList: et.AccessList(tx.AccessList()),
			V:          v,
			R:          r,
			S:          s,
		}
	case models.DynamicFeeTxType:
		extrakeys = &models.Extra{
			Type:       models.DynamicFeeTxType,
			GasTipCap:  tx.GasTipCap(),
			GasFeeCap:  tx.GasFeeCap(),
			AccessList: et.AccessList(tx.AccessList()),
			V:          v,
			R:          r,
			S:          s,
		}
	}
	extrakeys.Gas = tx.Gas()

	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return nil, err
	}

	gtkmtx := &models.Transaction{
		ChainID:   cid,
		From:      et.AddressP(&sender),
		To:        et.AddressP(tx.To()),
		Nonce:     tx.Nonce(),
		UseLocal:  false,
		Val:       tx.Value(),
		Input:     tx.Data(),
		Version:   models.TxVersion,
		MultiSigs: nil,
	}
	if err := gtkmtx.SetExtraKeys(extrakeys); err != nil {
		return nil, err
	}
	return gtkmtx, nil
}

func (et Eth2TKM) ChainID(ethChainId *big.Int) (common.ChainID, error) {
	if ethChainId == nil {
		return common.NilChainID, errors.New("nil chain id")
	}
	if !ethChainId.IsUint64() {
		return common.NilChainID, errors.New("chain id not available")
	}
	ethcid := ethChainId.Uint64()
	if ethcid > uint64(math.MaxUint32) {
		return common.NilChainID, errors.New("chain id out of range")
	}
	cid := uint32(ethcid)
	return common.ChainID(cid), nil
}

func (et Eth2TKM) TxSameChainID(tx *types.Transaction) (*models.Transaction, error) {
	cid, err := et.ChainID(tx.ChainId())
	if err != nil {
		return nil, err
	}
	return et._txWithChainID(cid, tx)
}

func (et Eth2TKM) Tx(tx *types.Transaction) (*models.Transaction, error) {
	cid, err := models.FromETHChainID(tx.ChainId())
	if err != nil {
		return nil, err
	}
	return et._txWithChainID(cid, tx)
}

func (et Eth2TKM) Receipt(ethrpt *types.Receipt) *models.Receipt {
	if ethrpt == nil {
		return nil
	}
	root := ethrpt.PostState
	poststate, _ := models.NewPostState(nil, root).Bytes()
	return &models.Receipt{
		PostState:       poststate,
		Status:          ethrpt.Status,
		Logs:            et.Logs(ethrpt.Logs),
		GasBonuses:      nil,
		TxHash:          et.Hash(ethrpt.TxHash),
		ContractAddress: et.AddressP(&ethrpt.ContractAddress),
		Out:             nil,
		GasUsed:         ethrpt.GasUsed,
		Error:           "",
	}
}

func (et Eth2TKM) TxReceipt(ethtx *types.Transaction, ethrpt *types.Receipt, tkmChainTx ...bool) (*client.TransactionReceipt, error) {
	if ethrpt == nil {
		return nil, nil
	}
	root := ethrpt.PostState
	poststate, _ := models.NewPostState(nil, root).Bytes()
	height := common.NilHeight
	if ethrpt.BlockNumber != nil && ethrpt.BlockNumber.IsUint64() {
		height = common.Height(ethrpt.BlockNumber.Uint64())
	}
	var tx *models.Transaction
	var err error
	if len(tkmChainTx) > 0 && tkmChainTx[0] {
		tx, err = et.Tx(ethtx)
	} else {
		tx, err = et.TxSameChainID(ethtx)
	}
	if err != nil {
		return nil, err
	}
	var sig *models.PubAndSig
	if tx != nil {
		sig, err = tx.GetSignature()
		if err != nil {
			return nil, err
		}
	}
	return &client.TransactionReceipt{
		Transaction:     tx,
		Sig:             sig,
		PostState:       poststate,
		Status:          ethrpt.Status,
		Logs:            et.Logs(ethrpt.Logs),
		GasBonuses:      nil,
		TxHash:          et.Hash(ethrpt.TxHash),
		ContractAddress: et.Address(ethrpt.ContractAddress),
		Out:             nil,
		Height:          height,
		GasUsed:         ethrpt.GasUsed,
		GasFee:          "",
		PostRoot:        root,
		Error:           "",
		Param:           nil,
	}, nil
}

func (tl TKM2LightNode) Header(h models.BlockHeader) TKMHeader {
	return TKMHeader{
		PreviousHash:     h.PreviousHash.Bytes(),
		HashHistory:      h.HashHistory.Bytes(),
		ChainID:          uint32(h.ChainID),
		Height:           uint64(h.Height),
		Empty:            h.Empty,
		ParentHeight:     uint64(h.ParentHeight),
		ParentHash:       h.ParentHash.Slice(),
		RewardAddress:    h.RewardAddress,
		AttendanceHash:   h.AttendanceHash.Slice(),
		RewardedCursor:   h.RewardedCursor.Slice(),
		CommitteeHash:    h.CommitteeHash.Slice(),
		ElectedNextRoot:  h.ElectedNextRoot.Slice(),
		NewCommitteeSeed: h.Seed.Slice(),
		RREra:            h.RREra.Slice(),
		RRRoot:           h.RRRoot.Slice(),
		RRNextRoot:       h.RRNextRoot.Slice(),
		RRChangingRoot:   h.RRChangingRoot.Slice(),
		MergedDeltaRoot:  h.MergedDeltaRoot.Slice(),
		BalanceDeltaRoot: h.BalanceDeltaRoot.Slice(),
		StateRoot:        h.StateRoot.Bytes(),
		ChainInfoRoot:    h.ChainInfoRoot.Slice(),
		WaterlinesRoot:   h.WaterlinesRoot.Slice(),
		VCCRoot:          h.VCCRoot.Slice(),
		CashedRoot:       h.CashedRoot.Slice(),
		TransactionRoot:  h.TransactionRoot.Slice(),
		ReceiptRoot:      h.ReceiptRoot.Slice(),
		HdsRoot:          h.HdsRoot.Slice(),
		TimeStamp:        h.TimeStamp,
		ElectResultRoot:  h.ElectResultRoot.Slice(),
		PreElectRoot:     h.PreElectRoot.Slice(),
		FactorRoot:       h.FactorRoot.Slice(),
		RRReceiptRoot:    h.RRReceiptRoot.Slice(),
		Version:          h.Version,
		ConfirmedRoot:    h.ConfirmedRoot.Slice(),
		RewardedEra:      h.RewardedEra.Slice(),
		BridgeRoot:       h.BridgeRoot.Slice(),
		RandomHash:       h.RandomHash.Slice(),
		SeedGenerated:    h.SeedGenerated,
		TxParamsRoot:     h.TxParamsRoot.Slice(),
	}
}

func (tl TKM2LightNode) HeaderP(h *models.BlockHeader) *TKMHeader {
	if h == nil {
		return nil
	}
	r := tl.Header(*h)
	return &r
}

func (tl TKM2LightNode) PaSs(pass models.PubAndSigs) [][]byte {
	if pass == nil {
		return nil
	}
	ret := make([][]byte, 0, len(pass))
	for _, pas := range pass {
		if pas != nil && len(pas.Signature) > 0 {
			ret = append(ret, common.CopyBytes(pas.Signature))
		}
	}
	return ret
}

func (tl TKM2LightNode) MustHash(h *common.Hash) common.Hash {
	if h == nil {
		return common.Hash{}
	}
	hh := *h
	return hh
}

func (tl TKM2LightNode) Hashs(hs []common.Hash) []common.Hash {
	if hs == nil {
		return []common.Hash{}
	}
	return hs
}

func (tl TKM2LightNode) MustAddress(addr *common.Address) common.Address {
	if addr == nil {
		return common.Address{}
	}
	b := *addr
	return b
}

func (tl TKM2LightNode) Bytes(bs []byte) []byte {
	if bs == nil {
		return []byte{}
	}
	return bs
}

func (tl TKM2LightNode) ReceiptLog(rlog *models.Log) TKMLog {
	if rlog == nil {
		return TKMLog{}
	}
	return TKMLog{
		Address:     rlog.Address,
		Topics:      tl.Hashs(rlog.Topics),
		Data:        tl.Bytes(rlog.Data),
		BlockNumber: rlog.BlockNumber,
		TxHash:      rlog.TxHash,
		TxIndex:     uint32(rlog.TxIndex),
		Index:       uint32(rlog.Index),
		BlockHash:   tl.MustHash(rlog.BlockHash),
	}
}

func (tl TKM2LightNode) ReceiptLogs(rlogs []*models.Log) []TKMLog {
	if rlogs == nil {
		return nil
	}
	ret := make([]TKMLog, len(rlogs))
	for i, rlog := range rlogs {
		ret[i] = tl.ReceiptLog(rlog)
	}
	return ret
}

func (tl TKM2LightNode) Bonus(b *models.Bonus) TKMBonus {
	if b == nil {
		return TKMBonus{}
	}
	return TKMBonus{
		Winner: b.Winner,
		Val:    b.Val,
	}
}

func (tl TKM2LightNode) Bonuses(bs []*models.Bonus) []TKMBonus {
	if bs == nil {
		return nil
	}
	rs := make([]TKMBonus, len(bs))
	for i, b := range bs {
		rs[i] = tl.Bonus(b)
	}
	return rs
}

func (tl TKM2LightNode) Receipt(rpt models.Receipt) TKMReceipt {
	return TKMReceipt{
		PostState:         rpt.PostState,
		Status:            rpt.Status,
		CumulativeGasUsed: rpt.CumulativeGasUsed,
		Logs:              tl.ReceiptLogs(rpt.Logs),
		TxHash:            rpt.TxHash[:],
		ContractAddress:   tl.MustAddress(rpt.ContractAddress),
		GasUsed:           rpt.GasUsed,
		Out:               rpt.Out,
		Error:             rpt.Error,
		GasBonuses:        tl.Bonuses(rpt.GasBonuses),
		Version:           rpt.Version,
	}
}

func (tl TKM2LightNode) MerkleProofs(pchain trie.ProofChain) []MerkleProof {
	var mps []*MerkleProof
	var last *MerkleProof
	_ = pchain.Iterate(func(val []byte, order bool) error {
		oneHash := common.BytesToHash(val)
		if last.Same(oneHash, order) {
			last.Repeat += 1
		} else {
			last = &MerkleProof{
				Hash:     oneHash,
				Position: order,
				Repeat:   0,
			}
			mps = append(mps, last)
		}
		return nil
	})
	ret := make([]MerkleProof, len(mps))
	for i, mp := range mps {
		ret[i] = *mp
	}
	return ret
}

func (tl TKM2LightNode) RcptAndTargetLog(receipt *models.Receipt, contractAddr common.Address,
	topicId common.Hash) (*models.Receipt, *models.Log, trie.ProofChain, error) {
	if receipt == nil || receipt.Version < models.ReceiptV2 || len(receipt.Logs) == 0 {
		return nil, nil, nil, errors.New("invalid receipt version")
	}
	rcpt := receipt.Clone()
	logs := models.Logs(rcpt.Logs)
	rcpt.Logs = nil
	i, rlog := locateLog(logs, contractAddr, topicId)
	if i < 0 {
		return nil, nil, nil, errors.New("no target Log found")
	}
	logProof := make(trie.ProofChain, 0)
	_, err := logs.MerkleRoot(i, &logProof)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("calculate log root of the Receipt failed: %w", err)
	}
	return rcpt, rlog, logProof, nil
}

func (tl TKM2LightNode) ReceiptProof(txProof *models.TxFinalProof, contractAddr common.Address,
	topicId common.Hash) (*TKMReceiptProof, error) {
	if txProof == nil {
		return nil, nil
	}
	if txProof.Header == nil || txProof.Receipt == nil || len(txProof.ReceiptProof) == 0 {
		return nil, errors.New("invalid tx final proof")
	}

	// txProof = _maliciousFinalProof(txProof)

	if txProof.Receipt.Version >= models.ReceiptV2 {
		rcpt, rlog, logProof, err := tl.RcptAndTargetLog(txProof.Receipt, contractAddr, topicId)
		if err != nil {
			return nil, err
		}
		return &TKMReceiptProof{
			Receipt:    tl.Receipt(*rcpt),
			Log:        tl.ReceiptLog(rlog),
			LogProof:   tl.MerkleProofs(logProof),
			Proofs:     tl.MerkleProofs(txProof.ReceiptProof),
			Header:     tl.Header(*txProof.Header),
			Signatures: tl.PaSs(txProof.Sigs),
		}, nil
	} else {
		return &TKMReceiptProof{
			Receipt:    tl.Receipt(*txProof.Receipt),
			Log:        TKMLog{},
			LogProof:   make([]MerkleProof, 0),
			Proofs:     tl.MerkleProofs(txProof.ReceiptProof),
			Header:     tl.Header(*txProof.Header),
			Signatures: tl.PaSs(txProof.Sigs),
		}, nil
	}
}

func (tl TKM2LightNode) ReceiptData(txProof *models.TxFinalProof, contractAddr common.Address,
	topicId common.Hash) (*TKMReceiptData, error) {
	if txProof == nil {
		return nil, nil
	}
	if txProof.Header == nil || txProof.Receipt == nil || len(txProof.ReceiptProof) == 0 {
		return nil, errors.New("invalid tx final proof")
	}

	if txProof.Receipt.Version >= models.ReceiptV2 {
		rcpt, rlog, logProof, err := tl.RcptAndTargetLog(txProof.Receipt, contractAddr, topicId)
		if err != nil {
			return nil, err
		}
		return &TKMReceiptData{
			Receipt:    tl.Receipt(*rcpt),
			Log:        tl.ReceiptLog(rlog),
			LogProof:   tl.MerkleProofs(logProof),
			Proofs:     tl.MerkleProofs(txProof.ReceiptProof),
			ChainID:    uint32(txProof.Header.ChainID),
			Height:     uint64(txProof.Header.Height),
			Signatures: tl.PaSs(txProof.Sigs),
		}, nil
	} else {
		return &TKMReceiptData{
			Receipt:    tl.Receipt(*txProof.Receipt),
			Log:        TKMLog{},
			LogProof:   make([]MerkleProof, 0),
			Proofs:     tl.MerkleProofs(txProof.ReceiptProof),
			ChainID:    uint32(txProof.Header.ChainID),
			Height:     uint64(txProof.Header.Height),
			Signatures: tl.PaSs(txProof.Sigs),
		}, nil
	}
}
