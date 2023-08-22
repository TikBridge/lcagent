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
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"hash"
	"math/big"
	sc "sync"
	"time"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/ThinkiumGroup/go-common/trie"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"github.com/bsm/redislock"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/sha3"
)

const (
	ExitUnknown          = 0xff
	ExitByInput          = 0x40
	ExitByContext        = 0x41
	ExitRunningLockErr   = 0x42
	ExitLCErr            = 0x43
	ExitByConfig         = 0x44
	ExitSourceErr        = 0x51
	ExitTargetErr        = 0x52
	ExitRedisErr         = 0x53
	ExitBasicHandlerErr  = 0x54
	ExitLooperHandlerErr = 0x55
)

type CommitteeProof struct {
	Header *models.BlockHeader
	Comm   *models.Committee
	PaSs   models.PubAndSigs
	// for xmaintainer: the epoch where XRelay is currently synchronizing transactions with
	// the Target chain, X-LightNode will delete all committees before this epoch
	SyncingEpoch common.EpochNum
}

func (p *CommitteeProof) String() string {
	if p == nil {
		return "CommProof<nil>"
	}
	return fmt.Sprintf("CommProof{%s %s PaSs:%d SyncingEpoch:%d}", p.Header.Summary(), p.Comm, len(p.PaSs), p.SyncingEpoch)
}

func (p *CommitteeProof) InfoString(level common.IndentLevel) string {
	if p == nil {
		return "CommProof<nil>"
	}
	next := level + 1
	indent := next.IndentString()
	return fmt.Sprintf("CommProof{"+
		"\n%sHeader: %s"+
		"\n%sComm: %s"+
		"\n%sPaSs: %s"+
		"\n%sSyncingEpoch: %d"+
		"\n%s}",
		indent, p.Header.InfoString(next),
		indent, p.Comm.InfoString(next),
		indent, p.PaSs.InfoString(next),
		indent, p.SyncingEpoch, level.IndentString())
}

func (p *CommitteeProof) Verify(mainchainNeeded bool) error {
	if p.Header == nil || !p.Comm.IsAvailable() {
		return errors.New("missing header or committee")
	}
	if mainchainNeeded && !p.Header.ChainID.IsMain() {
		return errors.New("not a main chain header")
	}
	if commHash := p.Comm.Hash(); !commHash.Equal(p.Header.ElectedNextRoot) {
		return errors.New("committee not match with Header.ElectedNextRoot")
	}
	boh := p.Header.Hash()
	if n, err := p.PaSs.Verify(boh[:]); err != nil || n <= 0 {
		return fmt.Errorf("signature list verify failed: %w, passed:%d", err, n)
	}
	return nil
}

func (p *CommitteeProof) ForABI() *TKMCommProof {
	if p == nil {
		return nil
	}
	return &TKMCommProof{
		Header:    T2LN.Header(*(p.Header)),
		Committee: common.NodeIDs(p.Comm.Members).ToBytesSlice(),
		Sigs:      T2LN.PaSs(p.PaSs),
	}
}

func (p *CommitteeProof) ForXABI() *XCommProof {
	if p == nil {
		return nil
	}
	return &XCommProof{
		Header:       T2LN.Header(*(p.Header)),
		Committee:    common.NodeIDs(p.Comm.Members).ToBytesSlice(),
		Sigs:         T2LN.PaSs(p.PaSs),
		SyncingEpoch: uint64(p.SyncingEpoch),
	}
}

func (p *CommitteeProof) ForXDataABI() (*XCommProofData, error) {
	if p == nil {
		return nil, nil
	}
	proof := new(trie.ProofChain)
	if _, err := p.Header.MakeProof(trie.ProofHeaderBase+models.BHElectedNextRoot, proof); err != nil {
		return nil, err
	}
	return &XCommProofData{
		Proofs:       T2LN.MerkleProofs(*proof),
		Committee:    common.NodeIDs(p.Comm.Members).ToBytesSlice(),
		Sigs:         T2LN.PaSs(p.PaSs),
		ChainID:      uint32(p.Header.ChainID),
		Height:       uint64(p.Header.Height),
		SyncingEpoch: uint64(p.SyncingEpoch),
	}, nil
}

type updateEvent struct {
	Epoch    common.EpochNum `abi:"epoch"`
	CommHash common.Hash     `abi:"commHash"`
}

func (e *updateEvent) String() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s(epoch:%d, commHash:%x)", updateCommEvent, e.Epoch, e.CommHash[:])
}

const (
	secpPubSize  = 65 // ((Params().BitSize + 7) >> 3)*2 + 1
	secpPrivSize = 32 // (Params().BitSize + 7) >> 3
	secpSigSize  = 65
)

var ETHSigner EthSecp256k1Signer

func init() {
	ETHSigner = EthSecp256k1Signer{}
	models.TKMCipher = ETHSigner
	log.Infof("models.TKMCipher set to %s", models.TKMCipher)
}

type EthSecp256k1Signer struct{}

func (s EthSecp256k1Signer) Name() string {
	return "secp256k1_sha3"
}

func (s EthSecp256k1Signer) GenerateKey() (*ecdsa.PrivateKey, error) {
	pv, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (s EthSecp256k1Signer) Sign(priv []byte, hash []byte) (sig []byte, err error) {
	key, err := crypto.ToECDSA(priv)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(hash, key)
}

func (s EthSecp256k1Signer) Verify(pub []byte, hash []byte, sig []byte) bool {
	p := pub
	if len(pub) == 0 {
		var err error
		p, err = secp256k1.RecoverPubkey(hash, sig)
		if err != nil {
			return false
		}
	}
	if len(p) != s.LengthOfPublicKey() || len(sig) != s.LengthOfSignature() {
		return false
	}
	return crypto.VerifySignature(p, hash, sig[:64])
}

func (s EthSecp256k1Signer) RecoverPub(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

func (s EthSecp256k1Signer) PrivToBytes(priv *ecdsa.PrivateKey) []byte {
	return crypto.FromECDSA(priv)
}

func (s EthSecp256k1Signer) PubToBytes(pub *ecdsa.PublicKey) []byte {
	return crypto.FromECDSAPub(pub)
}

func (s EthSecp256k1Signer) BytesToPriv(d []byte) (*ecdsa.PrivateKey, error) {
	return crypto.ToECDSA(d)
}

func (s EthSecp256k1Signer) BytesToPub(pub []byte) (*ecdsa.PublicKey, error) {
	return crypto.UnmarshalPubkey(pub)
}

func (s EthSecp256k1Signer) PubFromNodeId(nid []byte) []byte {
	// 最安全的方法是先反解为ECCPublicKey，再将其转换为[]byte。但是因为转换都使用
	// elliptic.Marshal/Unmarshal，所以这里为了节省中间转换，直接使用elliptic中的参数进行转换
	pk := make([]byte, secpPubSize)
	pk[0] = 4
	copy(pk[1:], nid[:])
	return pk
}

func (s EthSecp256k1Signer) PubToNodeIdBytes(pub []byte) ([]byte, error) {
	id := make([]byte, 64)
	if len(pub)-1 != len(id) {
		return id, fmt.Errorf("need %d bytes, got %d bytes", len(id)+1, len(pub))
	}
	copy(id[:], pub[1:])
	return id, nil
}

func (s EthSecp256k1Signer) PubFromPriv(priv []byte) ([]byte, error) {
	eccpriv, err := s.BytesToPriv(priv)
	if err != nil {
		return nil, err
	}
	pub, ok := eccpriv.Public().(*ecdsa.PublicKey)
	if !ok || pub == nil {
		return nil, errors.New("failed get public key")
	}
	return s.PubToBytes(pub), nil
}

func (s EthSecp256k1Signer) Hasher() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

func (s EthSecp256k1Signer) LengthOfPublicKey() int {
	return secpPubSize
}

func (s EthSecp256k1Signer) LengthOfPrivateKey() int {
	return secpPrivSize
}

func (s EthSecp256k1Signer) LengthOfSignature() int {
	return secpSigSize
}

func (s EthSecp256k1Signer) LengthOfHash() int {
	return 32
}

func (s EthSecp256k1Signer) ValidateSignatureValues(V byte, R, S *big.Int, homestead bool) bool {
	return crypto.ValidateSignatureValues(V, R, S, homestead)
}

func (s EthSecp256k1Signer) String() string {
	return fmt.Sprintf("Secp256k1(sk:%d pk:%d sig:%d hash:%d)",
		s.LengthOfPrivateKey(), s.LengthOfPublicKey(), s.LengthOfSignature(), s.LengthOfHash())
}

type redisKeys struct {
	startHeightKey  string // key of saving start height value
	runnerLockKey   string // the key of the lock for running one loop
	runnerLockValue string // locked value IP+"@"+PID
	senderLockKey   string // prefix+sender.Address
}

func (k redisKeys) String() string {
	return fmt.Sprintf("REDIS{startHeight: %s Lock{runner:(%s = %s) sender: %s}}",
		k.startHeightKey, k.runnerLockKey, k.runnerLockValue, k.senderLockKey)
}

type DistributedLock interface {
	Fetch(cctx context.Context) (lockingValue string, err error)
	Release() error
	Refresh(cctx context.Context) error
}

type redisLock struct {
	redis  *redis.Client
	locker *redislock.Client
	key    string
	value  string
	ttl    time.Duration
	rlock  *redislock.Lock
	lock   sc.Mutex
}

func newRedisLock(client *redis.Client, locker *redislock.Client, key, value string, ttl time.Duration) *redisLock {
	return &redisLock{
		redis:  client,
		locker: locker,
		key:    key,
		value:  value,
		ttl:    ttl,
		rlock:  nil,
	}
}

func (l *redisLock) String() string {
	if l == nil {
		return "RedisLock<nil>"
	}
	return fmt.Sprintf("RedisLock{%s}", l.key)
}

func (l *redisLock) _fetch(cctx context.Context) (lockingValue string, err error) {
	ctx, cancel := context.WithTimeout(cctx, redisTimeout)
	defer cancel()
	var lock *redislock.Lock
	lock, err = l.locker.Obtain(ctx, l.key, l.ttl, &redislock.Options{Token: l.value})
	if err != nil {
		if err == redislock.ErrNotObtained {
			lockingValue = l.redis.Get(ctx, l.key).Val()
		}
	} else {
		l.rlock = lock
	}
	return
}

func (l *redisLock) Fetch(cctx context.Context) (lockingValue string, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.rlock != nil {
		return l.rlock.Token(), errors.New("lock already fetched")
	}
	defer func() {
		if err == nil {
			log.Debugf("%s lock success", l)
		}
	}()
	return l._fetch(cctx)
}

func (l *redisLock) Release() (err error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	defer func() {
		if err != nil {
			log.Warnf("%s release failed: %v", l, err)
		} else {
			log.Debugf("%s released", l)
		}
	}()
	if l.rlock == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	err = l.rlock.Release(ctx)
	l.rlock = nil
	return err
}

func (l *redisLock) _refresh(cctx context.Context) error {
	ctx, cancel := context.WithTimeout(cctx, redisTimeout)
	defer cancel()
	return l.rlock.Refresh(ctx, l.ttl, nil)
}

func (l *redisLock) Refresh(cctx context.Context) (err error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	defer func() {
		if err != nil {
			log.Warnf("%s refresh failed: %v", l, err)
		} else {
			log.Debugf("%s refreshed", l)
		}
	}()
	if l.rlock == nil {
		return errors.New("lock not fetched")
	}
	return l._refresh(cctx)
}

func (l *redisLock) FetchOrRefresh(cctx context.Context) (lockingValue string, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.rlock == nil {
		return l._fetch(cctx)
	} else {
		err = l._refresh(cctx)
		if err != nil {
			return
		}
		return l.value, nil
	}
}

type redisLocks []*redisLock

func (r redisLocks) Fetch(_ context.Context) (string, error) {
	return "", common.ErrUnsupported
}

func (r redisLocks) Release() error {
	return common.ErrUnsupported
}

func (r redisLocks) Refresh(cctx context.Context) error {
	var errs []error
	for _, l := range r {
		if l == nil {
			continue
		}
		if err := l.Refresh(cctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors: %v", errs)
	}
	return nil
}

type Expirable[T any] struct {
	value   T
	ttlms   int64
	expires int64 // a time stamp in time.UnixMilli()
	lock    sc.RWMutex
}

func NewExpirable[T any](val T, ttlms int64, expires int64) *Expirable[T] {
	return &Expirable[T]{
		value:   val,
		ttlms:   ttlms,
		expires: expires,
	}
}

func (e *Expirable[T]) Get() (v T, exist bool) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	n := time.Now()
	if n.UnixMilli() >= e.expires {
		return e.value, false
	} else {
		return e.value, true
	}
}

func (e *Expirable[T]) Update(val T) {
	e.lock.Lock()
	defer e.lock.Unlock()
	n := time.Now()
	e.value = val
	e.expires = n.UnixMilli() + e.ttlms
}

type LockError interface {
	error
	Unlock() bool
}

type notUnlockError struct {
	err error
}

func (ue *notUnlockError) Error() string {
	return ue.err.Error()
}

func (ue *notUnlockError) Unlock() bool {
	return false
}

func NotUnlockError(err error) LockError {
	return &notUnlockError{err: err}
}
