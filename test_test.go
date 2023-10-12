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
	"math/big"
	"testing"
)

const (
	test_abi_string_1 = `[
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "bytes32",
						"name": "Hash",
						"type": "bytes32"
					},
					{
						"internalType": "bool",
						"name": "Pos",
						"type": "bool"
					},
					{
						"internalType": "uint8",
						"name": "Repeat",
						"type": "uint8"
					}
				],
				"internalType": "struct Mpts.MerkleProof[]",
				"name": "mps",
				"type": "tuple[]"
			},
			{
				"internalType": "uint32",
				"name": "index",
				"type": "uint32"
			}
		],
		"name": "HashAtIndex",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	}
]`

	test_abi_string_2 = `[
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "bytes",
						"name": "PostState",
						"type": "bytes"
					},
					{
						"internalType": "uint64",
						"name": "Status",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "CumulativeGasUsed",
						"type": "uint64"
					},
					{
						"components": [
							{
								"internalType": "address",
								"name": "Address",
								"type": "address"
							},
							{
								"internalType": "bytes32[]",
								"name": "Topics",
								"type": "bytes32[]"
							},
							{
								"internalType": "bytes",
								"name": "Data",
								"type": "bytes"
							},
							{
								"internalType": "uint64",
								"name": "BlockNumber",
								"type": "uint64"
							},
							{
								"internalType": "bytes32",
								"name": "TxHash",
								"type": "bytes32"
							},
							{
								"internalType": "uint32",
								"name": "TxIndex",
								"type": "uint32"
							},
							{
								"internalType": "uint32",
								"name": "Index",
								"type": "uint32"
							},
							{
								"internalType": "bytes32",
								"name": "BlockHash",
								"type": "bytes32"
							}
						],
						"internalType": "struct rlptest.Log[]",
						"name": "Logs",
						"type": "tuple[]"
					},
					{
						"internalType": "bytes",
						"name": "TxHash",
						"type": "bytes"
					},
					{
						"internalType": "address",
						"name": "ContractAddress",
						"type": "address"
					},
					{
						"internalType": "uint64",
						"name": "GasUsed",
						"type": "uint64"
					},
					{
						"internalType": "bytes",
						"name": "Out",
						"type": "bytes"
					},
					{
						"internalType": "string",
						"name": "Error",
						"type": "string"
					},
					{
						"components": [
							{
								"internalType": "address",
								"name": "Winner",
								"type": "address"
							},
							{
								"internalType": "uint256",
								"name": "Val",
								"type": "uint256"
							}
						],
						"internalType": "struct rlptest.Bonus[]",
						"name": "GasBonuses",
						"type": "tuple[]"
					},
					{
						"internalType": "uint16",
						"name": "Version",
						"type": "uint16"
					}
				],
				"internalType": "struct rlptest.Receipt",
				"name": "r",
				"type": "tuple"
			}
		],
		"name": "hashReceipt",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	}
]`
)

func Test_Uint256(t *testing.T) {
	i := new(big.Int)
	bs, _ := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	i.SetBytes(bs)
	t.Logf("%s", i)
}
