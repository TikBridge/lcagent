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
	"fmt"
	"math/big"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/abi"
	"github.com/ThinkiumGroup/go-common/math"
)

// interface ILightNode {
// 		event UpdateBlockHeader(address indexed account, uint64 indexed blockHeight);
//		event UpdateCommittee(uint64 indexed epoch, bytes32 indexed commHash);
// 		function initialize(uint64 _currentEpoch, bytes[] memory _currentCommittee, bytes[] memory _nextCommittee) external;
// 		function verifyProofData(bytes memory proofBytes) external view returns (bool success, bytes32 txHash);
// 		function updateCommittee(bytes _epochCommittee) returns(bool);
// 		function verifyReceiptProof(TKM.ReceiptProof memory receiptProof) external view returns (bool success, bytes32 txHash);
// 		function updateNextCommittee(TKM.CommitteeProof memory commProof) returns(bool);
// 		function resetCommittee(bytes memory epochCommittee) external;
// 		function lastHeaderHeight() external view returns (uint64);
// 		function checkEpochCommittee(uint64 epoch) external view returns (address[] memory);
// }

// latest2Epoch is an array of epoch numbers with fixed size 2.
// latest2Epoch[1] is the latest epoch number, latest2Epoch[0] can be 0 just after initialization.
// get committee list by checkEpochCommittee(epoch)

var LightNodeABI abi.ABI

const (
	tkmlcstring = `[
  {
    "inputs": [],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "address",
        "name": "previousAdmin",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "address",
        "name": "newAdmin",
        "type": "address"
      }
    ],
    "name": "AdminChanged",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "previous",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newAdmin",
        "type": "address"
      }
    ],
    "name": "AdminTransferred",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "beacon",
        "type": "address"
      }
    ],
    "name": "BeaconUpgraded",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "previousPending",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newPending",
        "type": "address"
      }
    ],
    "name": "ChangePendingAdmin",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "uint8",
        "name": "version",
        "type": "uint8"
      }
    ],
    "name": "Initialized",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "Paused",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "Unpaused",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "account",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "uint64",
        "name": "blockHeight",
        "type": "uint64"
      }
    ],
    "name": "UpdateBlockHeader",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "uint64",
        "name": "epoch",
        "type": "uint64"
      },
      {
        "indexed": true,
        "internalType": "bytes32",
        "name": "commHash",
        "type": "bytes32"
      }
    ],
    "name": "UpdateCommittee",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "implementation",
        "type": "address"
      }
    ],
    "name": "Upgraded",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "uint64",
        "name": "epoch",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "bytes[]",
        "name": "_currentCommittee",
        "type": "bytes[]"
      }
    ],
    "name": "initializeCommittee",
    "type": "event"
  },
  {
    "inputs": [],
    "name": "changeAdmin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint64",
        "name": "epoch",
        "type": "uint64"
      }
    ],
    "name": "checkEpochCommittee",
    "outputs": [
      {
        "internalType": "address[]",
        "name": "",
        "type": "address[]"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "getAdmin",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "getImplementation",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint64",
        "name": "_currentHeight",
        "type": "uint64"
      },
      {
        "internalType": "bytes[]",
        "name": "_currentCommittee",
        "type": "bytes[]"
      },
      {
        "internalType": "bytes[]",
        "name": "_nextCommittee",
        "type": "bytes[]"
      }
    ],
    "name": "initialize",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "lastHeaderHeight",
    "outputs": [
      {
        "internalType": "uint64",
        "name": "",
        "type": "uint64"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "lastHeight",
    "outputs": [
      {
        "internalType": "uint64",
        "name": "",
        "type": "uint64"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "name": "latest2Epoch",
    "outputs": [
      {
        "internalType": "uint64",
        "name": "",
        "type": "uint64"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint64",
        "name": "",
        "type": "uint64"
      },
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "name": "mainChainCommittee",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "paused",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "pendingAdmin",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "proxiableUUID",
    "outputs": [
      {
        "internalType": "bytes32",
        "name": "",
        "type": "bytes32"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes",
        "name": "epochCommittee",
        "type": "bytes"
      }
    ],
    "name": "resetCommittee",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "pure",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "pendingAdmin_",
        "type": "address"
      }
    ],
    "name": "setPendingAdmin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bool",
        "name": "_flag",
        "type": "bool"
      }
    ],
    "name": "togglePause",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes",
        "name": "_epochCommittee",
        "type": "bytes"
      }
    ],
    "name": "updateCommittee",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
	{
		"inputs": [
			{
				"components": [
					{
						"components": [
							{
								"internalType": "bytes",
								"name": "PreviousHash",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "HashHistory",
								"type": "bytes"
							},
							{
								"internalType": "uint32",
								"name": "ChainID",
								"type": "uint32"
							},
							{
								"internalType": "uint64",
								"name": "Height",
								"type": "uint64"
							},
							{
								"internalType": "bool",
								"name": "Empty",
								"type": "bool"
							},
							{
								"internalType": "uint64",
								"name": "ParentHeight",
								"type": "uint64"
							},
							{
								"internalType": "bytes",
								"name": "ParentHash",
								"type": "bytes"
							},
							{
								"internalType": "address",
								"name": "RewardAddress",
								"type": "address"
							},
							{
								"internalType": "bytes",
								"name": "AttendanceHash",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RewardedCursor",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "CommitteeHash",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "ElectedNextRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "NewCommitteeSeed",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RREra",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RRRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RRNextRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RRChangingRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "MergedDeltaRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "BalanceDeltaRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "StateRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "ChainInfoRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "WaterlinesRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "VCCRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "CashedRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "TransactionRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "ReceiptRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "HdsRoot",
								"type": "bytes"
							},
							{
								"internalType": "uint64",
								"name": "TimeStamp",
								"type": "uint64"
							},
							{
								"internalType": "bytes",
								"name": "ElectResultRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "PreElectRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "FactorRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RRReceiptRoot",
								"type": "bytes"
							},
							{
								"internalType": "uint16",
								"name": "Version",
								"type": "uint16"
							},
							{
								"internalType": "bytes",
								"name": "ConfirmedRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RewardedEra",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "BridgeRoot",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "RandomHash",
								"type": "bytes"
							},
							{
								"internalType": "bool",
								"name": "SeedGenerated",
								"type": "bool"
							},
							{
								"internalType": "bytes",
								"name": "TxParamsRoot",
								"type": "bytes"
							}
						],
						"internalType": "struct TKM.BlockHeader",
						"name": "header",
						"type": "tuple"
					},
					{
						"internalType": "bytes[]",
						"name": "committee",
						"type": "bytes[]"
					},
					{
						"internalType": "bytes[]",
						"name": "signatures",
						"type": "bytes[]"
					}
				],
				"internalType": "struct TKM.CommitteeProof",
				"name": "commProof",
				"type": "tuple"
			}
		],
		"name": "updateNextCommittee",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "newImplementation",
        "type": "address"
      }
    ],
    "name": "upgradeTo",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "newImplementation",
        "type": "address"
      },
      {
        "internalType": "bytes",
        "name": "data",
        "type": "bytes"
      }
    ],
    "name": "upgradeToAndCall",
    "outputs": [],
    "stateMutability": "payable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes",
        "name": "proofBytes",
        "type": "bytes"
      }
    ],
    "name": "verifyProofData",
    "outputs": [
      {
        "internalType": "bool",
        "name": "success",
        "type": "bool"
      },
      {
        "internalType": "string",
        "name": "message",
        "type": "string"
      },
      {
        "internalType": "bytes",
        "name": "logBytes",
        "type": "bytes"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "components": [
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
                "internalType": "struct TKM.Log[]",
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
                "internalType": "struct TKM.Bonus[]",
                "name": "GasBonuses",
                "type": "tuple[]"
              },
              {
                "internalType": "uint16",
                "name": "Version",
                "type": "uint16"
              }
            ],
            "internalType": "struct TKM.Receipt",
            "name": "r",
            "type": "tuple"
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
            "internalType": "struct TKM.Log",
            "name": "log",
            "type": "tuple"
          },
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
            "internalType": "struct MPT.MerkleProof[]",
            "name": "logProof",
            "type": "tuple[]"
          },
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
            "internalType": "struct MPT.MerkleProof[]",
            "name": "p",
            "type": "tuple[]"
          },
          {
            "components": [
              {
                "internalType": "bytes",
                "name": "PreviousHash",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "HashHistory",
                "type": "bytes"
              },
              {
                "internalType": "uint32",
                "name": "ChainID",
                "type": "uint32"
              },
              {
                "internalType": "uint64",
                "name": "Height",
                "type": "uint64"
              },
              {
                "internalType": "bool",
                "name": "Empty",
                "type": "bool"
              },
              {
                "internalType": "uint64",
                "name": "ParentHeight",
                "type": "uint64"
              },
              {
                "internalType": "bytes",
                "name": "ParentHash",
                "type": "bytes"
              },
              {
                "internalType": "address",
                "name": "RewardAddress",
                "type": "address"
              },
              {
                "internalType": "bytes",
                "name": "AttendanceHash",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RewardedCursor",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "CommitteeHash",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "ElectedNextRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "NewCommitteeSeed",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RREra",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RRRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RRNextRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RRChangingRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "MergedDeltaRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "BalanceDeltaRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "StateRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "ChainInfoRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "WaterlinesRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "VCCRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "CashedRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "TransactionRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "ReceiptRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "HdsRoot",
                "type": "bytes"
              },
              {
                "internalType": "uint64",
                "name": "TimeStamp",
                "type": "uint64"
              },
              {
                "internalType": "bytes",
                "name": "ElectResultRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "PreElectRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "FactorRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RRReceiptRoot",
                "type": "bytes"
              },
              {
                "internalType": "uint16",
                "name": "Version",
                "type": "uint16"
              },
              {
                "internalType": "bytes",
                "name": "ConfirmedRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RewardedEra",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "BridgeRoot",
                "type": "bytes"
              },
              {
                "internalType": "bytes",
                "name": "RandomHash",
                "type": "bytes"
              },
              {
                "internalType": "bool",
                "name": "SeedGenerated",
                "type": "bool"
              },
              {
                "internalType": "bytes",
                "name": "TxParamsRoot",
                "type": "bytes"
              }
            ],
            "internalType": "struct TKM.BlockHeader",
            "name": "h",
            "type": "tuple"
          },
          {
            "internalType": "bytes[]",
            "name": "s",
            "type": "bytes[]"
          }
        ],
        "internalType": "struct TKM.ReceiptProof",
        "name": "receiptProof",
        "type": "tuple"
      }
    ],
    "name": "verifyReceiptProof",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
`

	updateNextCommName = "updateNextCommittee"
	updateCommName     = "updateCommittee"
	updateCommEvent    = "UpdateCommittee"

	verifyReceiptStruct = "verifyReceiptProof"
	verifyReceiptData   = "verifyProofData"

	lastHeightName = "lastHeight"

	latest2EpochName   = "latest2Epoch"
	checkEpochCommName = "checkEpochCommittee"
)

func initTKMLNAbi() {
	LightNodeABI = *abi.MustInitAbi("tkm light node", tkmlcstring)
}

type (
	TKMHeader struct {
		PreviousHash     []byte         `abi:"PreviousHash"`
		HashHistory      []byte         `abi:"HashHistory"`
		ChainID          uint32         `abi:"ChainID"`
		Height           uint64         `abi:"Height"`
		Empty            bool           `abi:"Empty"`
		ParentHeight     uint64         `abi:"ParentHeight"`
		ParentHash       []byte         `abi:"ParentHash"`
		RewardAddress    common.Address `abi:"RewardAddress"`
		AttendanceHash   []byte         `abi:"AttendanceHash"`
		RewardedCursor   []byte         `abi:"RewardedCursor"` // nil: *RewardedCursor==nil, 8bytes big-endian: *RewardedCursor not nil
		CommitteeHash    []byte         `abi:"CommitteeHash"`
		ElectedNextRoot  []byte         `abi:"ElectedNextRoot"`
		NewCommitteeSeed []byte         `abi:"NewCommitteeSeed"`
		RREra            []byte         `abi:"RREra"` // nil: *RREra==nil, 8bytes big-endian: *RREra not nil
		RRRoot           []byte         `abi:"RRRoot"`
		RRNextRoot       []byte         `abi:"RRNextRoot"`
		RRChangingRoot   []byte         `abi:"RRChangingRoot"`
		MergedDeltaRoot  []byte         `abi:"MergedDeltaRoot"`
		BalanceDeltaRoot []byte         `abi:"BalanceDeltaRoot"`
		StateRoot        []byte         `abi:"StateRoot"`
		ChainInfoRoot    []byte         `abi:"ChainInfoRoot"`
		WaterlinesRoot   []byte         `abi:"WaterlinesRoot"`
		VCCRoot          []byte         `abi:"VCCRoot"`
		CashedRoot       []byte         `abi:"CashedRoot"`
		TransactionRoot  []byte         `abi:"TransactionRoot"`
		ReceiptRoot      []byte         `abi:"ReceiptRoot"`
		HdsRoot          []byte         `abi:"HdsRoot"`
		TimeStamp        uint64         `abi:"TimeStamp"`
		ElectResultRoot  []byte         `abi:"ElectResultRoot"`
		PreElectRoot     []byte         `abi:"PreElectRoot"`
		FactorRoot       []byte         `abi:"FactorRoot"`
		RRReceiptRoot    []byte         `abi:"RRReceiptRoot"`
		Version          uint16         `abi:"Version"`
		ConfirmedRoot    []byte         `abi:"ConfirmedRoot"`
		RewardedEra      []byte         `abi:"RewardedEra"` // nil: *RewardedEra==nil, 8bytes big-endian: *RewardedEra not nil
		BridgeRoot       []byte         `abi:"BridgeRoot"`
		RandomHash       []byte         `abi:"RandomHash"`
		SeedGenerated    bool           `abi:"SeedGenerated"`
		TxParamsRoot     []byte         `abi:"TxParamsRoot"`
	}

	TKMCommProof struct {
		Header    TKMHeader `abi:"header"`
		Committee [][]byte  `abi:"committee"`
		Sigs      [][]byte  `abi:"signatures"`
	}

	TKMLog struct {
		Address     common.Address `abi:"Address"`
		Topics      []common.Hash  `abi:"Topics"`
		Data        []byte         `abi:"Data"`
		BlockNumber uint64         `abi:"BlockNumber"`
		TxHash      common.Hash    `abi:"TxHash"`
		TxIndex     uint32         `abi:"TxIndex"`
		Index       uint32         `abi:"Index"`
		BlockHash   common.Hash    `abi:"BlockHash"`
	}

	TKMBonus struct {
		Winner common.Address `abi:"Winner"`
		Val    *big.Int       `abi:"Val"`
	}

	TKMReceipt struct {
		PostState         []byte         `abi:"PostState"`
		Status            uint64         `abi:"Status"`
		CumulativeGasUsed uint64         `abi:"CumulativeGasUsed"`
		Logs              []TKMLog       `abi:"Logs"`
		TxHash            []byte         `abi:"TxHash"`
		ContractAddress   common.Address `abi:"ContractAddress"`
		GasUsed           uint64         `abi:"GasUsed"`
		Out               []byte         `abi:"Out"`
		Error             string         `abi:"Error"`
		GasBonuses        []TKMBonus     `abi:"GasBonuses"`
		Version           uint16         `abi:"Version"`
	}

	MerkleProof struct {
		Hash     common.Hash `abi:"Hash"`
		Position bool        `abi:"Pos"`    // Proof.Hash on the left is true
		Repeat   uint8       `abi:"Repeat"` // number of consecutive repetitions -1
	}

	MerkleProofs []MerkleProof

	TKMReceiptProof struct {
		Receipt    TKMReceipt    `abi:"r"`
		Log        TKMLog        `abi:"log"`
		LogProof   []MerkleProof `abi:"logProof"`
		Proofs     []MerkleProof `abi:"p"`
		Header     TKMHeader     `abi:"h"`
		Signatures [][]byte      `abi:"s"`
	}

	TKMReceiptData struct {
		Receipt    TKMReceipt    `abi:"rcpt"`     // without Logs
		Log        TKMLog        `abi:"log"`      // transferOut Log
		LogProof   []MerkleProof `abi:"logProof"` // use log proof to calculate logRoot
		Proofs     []MerkleProof `abi:"proof"`    // receipt proof to block header hash
		ChainID    uint32        `abi:"chainid"`  // block chain id
		Height     uint64        `abi:"height"`   // block height
		Signatures [][]byte      `abi:"sigs"`     // signature list of the block
	}
)

func (h TKMHeader) String() string {
	return fmt.Sprintf("TKMHeader{ChainID:%d Height:%d}", h.ChainID, h.Height)
}

func (c TKMCommProof) String() string {
	return fmt.Sprintf("TKMComm{%s Comm:%s len(Sigs):%d}", c.Header.String(), common.PrintBytesSlice(c.Committee, 5), len(c.Sigs))
}

func (l TKMLog) String() string {
	return fmt.Sprintf("TKMLog{Addr:%x Topics:%d Data:%d Number:%d Tx:(Hash:%x Index:%d) Index:%d BlockHash:%x}",
		l.Address[:], len(l.Topics), len(l.Data), l.BlockNumber, l.TxHash[:], l.TxIndex, l.Index, l.BlockHash[:])
}

func (b TKMBonus) String() string {
	return fmt.Sprintf("TKMBonus{Winner:%x Val:%s}", b.Winner[:], math.BigForPrint(b.Val))
}

func (r TKMReceipt) String() string {
	return fmt.Sprintf("TKMReceipt{Status:%d Logs:%s TxHash:%x Contract:%x Gas:%d Version:%d}",
		r.Status, r.Logs, r.TxHash, r.ContractAddress[:], r.GasUsed, r.Version)
}

func (mp *MerkleProof) Same(oneHash common.Hash, pos bool) bool {
	if mp == nil {
		return false
	}
	return mp.Hash == oneHash && mp.Position == pos
}

func (mp MerkleProof) String() string {
	p := 0
	if mp.Position {
		p = 1
	}
	if mp.Repeat > 0 {
		return fmt.Sprintf("(%d)%x +%d", p, mp.Hash[:], mp.Repeat)
	}
	return fmt.Sprintf("(%d)%x", p, mp.Hash[:])
}

func (mp *MerkleProof) Proof(val []byte) []byte {
	if mp == nil {
		return val
	}
	root := val
	for i := uint8(0); i <= mp.Repeat; i++ {
		root, _ = common.HashPairOrder(mp.Position, mp.Hash[:], root)
	}
	return root
}

func (mps MerkleProofs) Proof(start []byte) []byte {
	if len(mps) == 0 {
		return start
	}
	root := start
	for _, mp := range mps {
		root = mp.Proof(root)
	}
	return root
}

func (mps MerkleProofs) Size() int {
	if len(mps) == 0 {
		return 0
	}
	c := len(mps)
	for _, mp := range mps {
		if mp.Repeat > 0 {
			c += int(mp.Repeat)
		}
	}
	return c
}

func (rp *TKMReceiptProof) String() string {
	if rp == nil {
		return "TKMRptProof<nil>"
	}
	return fmt.Sprintf("TKMRptProof{%s(%s LogProof:%d) =(Proofs:%d)=> Header{C:%d H:%d}}",
		rp.Receipt, rp.Log, MerkleProofs(rp.LogProof).Size(), MerkleProofs(rp.Proofs).Size(), rp.Header.ChainID, rp.Header.Height)
}

func (rb *TKMReceiptData) String() string {
	if rb == nil {
		return "TKMRptData<nil>"
	}
	return fmt.Sprintf("TKMRptData{%s(%s LogProof:%d) =(Proofs:%d)=> Header{C:%d H:%d}}",
		rb.Receipt, rb.Log, MerkleProofs(rb.LogProof).Size(), MerkleProofs(rb.Proofs).Size(), rb.ChainID, rb.Height)
}
