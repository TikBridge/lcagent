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
	"github.com/ThinkiumGroup/go-common/abi"
)

var UpdatableLightNodeAbi abi.ABI

const (
	tkmUpdateLcString = `[
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
    "inputs": [],
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
        "name": "_currentEpoch",
        "type": "uint64"
      },
      {
        "internalType": "bytes[]",
        "name": "_currentCommittee",
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
    "name": "lastEpoch",
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
    "name": "latestEpoch",
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
        "internalType": "uint64",
        "name": "epoch",
        "type": "uint64"
      },
      {
        "internalType": "bytes[]",
        "name": "committee",
        "type": "bytes[]"
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
  }
]
`

	uLastEpochName      = "lastEpoch"
	uCheckEpochCommName = "checkEpochCommittee"
	uUpdateCommName     = "updateCommittee"
	uUpdateCommEvent    = "UpdateCommittee"
)

func initUpdatableLNAbi() {
	UpdatableLightNodeAbi = mustInitAbi("updatable light node", tkmUpdateLcString)
}
