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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stephenfire/pkcs8"
)

var (
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	oidSecp256k1      = asn1.ObjectIdentifier{1, 3, 132, 0, 10}
)

type (
	pkcs8Format struct {
		Version    int
		Algo       pkix.AlgorithmIdentifier
		PrivateKey []byte
		// optional attributes omitted.
	}

	ecPrivateKey struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}
)

func sk2pkcs8(key *ecdsa.PrivateKey) ([]byte, error) {
	if !key.Curve.IsOnCurve(key.X, key.Y) {
		return nil, errors.New("invalid elliptic key public key")
	}
	privateKey := make([]byte, (key.Curve.Params().N.BitLen()+7)/8)
	privKeyDer, err := asn1.Marshal(ecPrivateKey{
		Version:       1,
		PrivateKey:    key.D.FillBytes(privateKey),
		NamedCurveOID: oidSecp256k1,
		PublicKey:     asn1.BitString{Bytes: elliptic.Marshal(key.Curve, key.X, key.Y)},
	})
	if err != nil {
		return nil, err
	}
	oidBytes, err := asn1.Marshal(oidSecp256k1)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal curve OID: %w", err)
	}
	privKey := pkcs8Format{
		Algo: pkix.AlgorithmIdentifier{
			Algorithm: oidPublicKeyECDSA,
			Parameters: asn1.RawValue{
				FullBytes: oidBytes,
			},
		},
		PrivateKey: privKeyDer,
	}
	return asn1.Marshal(privKey)
}

func pkcs82sk(der []byte) (*ecdsa.PrivateKey, error) {
	var pkcs8Asn1 pkcs8Format
	if _, err := asn1.Unmarshal(der, &pkcs8Asn1); err != nil {
		return nil, err
	}
	if !pkcs8Asn1.Algo.Algorithm.Equal(oidPublicKeyECDSA) {
		return nil, errors.New("not an ECDSA key file")
	}
	bytes := pkcs8Asn1.Algo.Parameters.FullBytes
	namedCurveOID := new(asn1.ObjectIdentifier)
	if _, err := asn1.Unmarshal(bytes, namedCurveOID); err != nil {
		return nil, fmt.Errorf("failed to unmarshal curve OID: %w", err)
	}
	if !namedCurveOID.Equal(oidSecp256k1) {
		return nil, errors.New("not an secp256k1 key")
	}
	var privKey ecPrivateKey
	if _, err := asn1.Unmarshal(pkcs8Asn1.PrivateKey, &privKey); err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}
	if privKey.Version != 1 {
		return nil, fmt.Errorf("unknown EC private key version %d", privKey.Version)
	}

	curve := secp256k1.S256()

	k := new(big.Int).SetBytes(privKey.PrivateKey)
	curveOrder := curve.Params().N
	if k.Cmp(curveOrder) >= 0 {
		return nil, errors.New("x509: invalid elliptic curve private key value")
	}
	priv := new(ecdsa.PrivateKey)
	priv.Curve = curve
	priv.D = k

	privateKey := make([]byte, (curveOrder.BitLen()+7)/8)

	// Some private keys have leading zero padding. This is invalid
	// according to [SEC1], but this code will ignore it.
	for len(privKey.PrivateKey) > len(privateKey) {
		if privKey.PrivateKey[0] != 0 {
			return nil, errors.New("x509: invalid private key length")
		}
		privKey.PrivateKey = privKey.PrivateKey[1:]
	}

	// Some private keys remove all leading zeros, this is also invalid
	// according to [SEC1] but since OpenSSL used to do this, we ignore
	// this too.
	copy(privateKey[len(privateKey)-len(privKey.PrivateKey):], privKey.PrivateKey)
	priv.X, priv.Y = curve.ScalarBaseMult(privateKey)

	return priv, nil
}

func MarshalPrivateKeyPKCS8(sk *ecdsa.PrivateKey, password []byte) ([]byte, error) {
	pkcs8Bytes, err := sk2pkcs8(sk)
	if err != nil {
		return nil, err
	}
	if len(password) == 0 {
		return pkcs8Bytes, nil
	}
	return pkcs8.EncryptPKCS8(pkcs8Bytes, password, nil)
}

func ParsePKCS8PrivateKey(der []byte, password []byte) (*ecdsa.PrivateKey, error) {
	if len(password) == 0 {
		return pkcs82sk(der)
	}
	decrypted, _, err := pkcs8.DecryptPKCS8(der, password)
	if err != nil {
		return nil, err
	}
	return pkcs82sk(decrypted)
}

func MarshalPrivateKeyPEM(sk *ecdsa.PrivateKey, password []byte) ([]byte, error) {
	pkcs8Bytes, err := MarshalPrivateKeyPKCS8(sk, password)
	if err != nil {
		return nil, err
	}
	pblock := &pem.Block{Bytes: pkcs8Bytes}
	if len(password) > 0 {
		pblock.Type = "ENCRYPTED EC PRIVATE KEY"
	} else {
		pblock.Type = "EC PRIVATE KEY"
	}
	return pem.EncodeToMemory(pblock), nil
}

var (
	ErrNoPemBlock     = errors.New("no pem block found")
	ErrNotAPrivateKey = errors.New("not a private key")
	// ErrMissingPwd     = errors.New("password is missing")
)

func ValidPEM(pemBytes []byte) (needPwd bool, pkcs8Bytes []byte, err error) {
	pblock, _ := pem.Decode(pemBytes)
	if pblock == nil {
		return false, nil, ErrNoPemBlock
	}
	if !strings.Contains(pblock.Type, "PRIVATE KEY") {
		return false, nil, ErrNotAPrivateKey
	}
	return strings.Contains(pblock.Type, "ENCRYPTED"), pblock.Bytes, nil
}
