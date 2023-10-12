package main

import "github.com/ThinkiumGroup/go-common/abi"

var Eth2LightNodeAbi abi.ABI

const (
	eth2LightNodeString = `[
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
				"name": "beacon",
				"type": "address"
			}
		],
		"name": "BeaconUpgraded",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_admin",
				"type": "address"
			}
		],
		"name": "changeAdmin",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint64",
				"name": "_chainId",
				"type": "uint64"
			},
			{
				"internalType": "address",
				"name": "_controller",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "_mptVerify",
				"type": "address"
			},
			{
				"components": [
					{
						"internalType": "uint64",
						"name": "slot",
						"type": "uint64"
					},
					{
						"internalType": "uint64",
						"name": "proposerIndex",
						"type": "uint64"
					},
					{
						"internalType": "bytes32",
						"name": "parentRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "stateRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "bodyRoot",
						"type": "bytes32"
					}
				],
				"internalType": "struct Types.BeaconBlockHeader",
				"name": "_finalizedBeaconHeader",
				"type": "tuple"
			},
			{
				"internalType": "uint256",
				"name": "_finalizedExeHeaderNumber",
				"type": "uint256"
			},
			{
				"internalType": "bytes32",
				"name": "_finalizedExeHeaderHash",
				"type": "bytes32"
			},
			{
				"internalType": "bytes",
				"name": "_curSyncCommitteeAggPubKey",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "_nextSyncCommitteeAggPubKey",
				"type": "bytes"
			},
			{
				"internalType": "bytes32[]",
				"name": "_syncCommitteePubkeyHashes",
				"type": "bytes32[]"
			},
			{
				"internalType": "bool",
				"name": "_verifyUpdate",
				"type": "bool"
			}
		],
		"name": "initialize",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
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
		"inputs": [
			{
				"internalType": "bytes",
				"name": "_syncCommitteePubkeyPart",
				"type": "bytes"
			}
		],
		"name": "initSyncCommitteePubkey",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
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
		"inputs": [
			{
				"internalType": "bool",
				"name": "flag",
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
		"inputs": [
			{
				"internalType": "bytes",
				"name": "_blockHeader",
				"type": "bytes"
			}
		],
		"name": "updateBlockHeader",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
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
				"indexed": false,
				"internalType": "uint256",
				"name": "start",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "end",
				"type": "uint256"
			}
		],
		"name": "UpdateBlockHeader",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "_data",
				"type": "bytes"
			}
		],
		"name": "updateLightClient",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
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
				"indexed": false,
				"internalType": "uint256",
				"name": "slot",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "height",
				"type": "uint256"
			}
		],
		"name": "UpdateLightClient",
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
		"inputs": [],
		"name": "BLS_PUBKEY_LENGTH",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "chainId",
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
		"name": "clientState",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "",
				"type": "bytes"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "EXECUTION_PROOF_SIZE",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "exeHeaderEndHash",
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
		"inputs": [],
		"name": "exeHeaderEndNumber",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "exeHeaderStartNumber",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "FINALITY_PROOF_SIZE",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "finalizedBeaconHeader",
		"outputs": [
			{
				"internalType": "uint64",
				"name": "slot",
				"type": "uint64"
			},
			{
				"internalType": "uint64",
				"name": "proposerIndex",
				"type": "uint64"
			},
			{
				"internalType": "bytes32",
				"name": "parentRoot",
				"type": "bytes32"
			},
			{
				"internalType": "bytes32",
				"name": "stateRoot",
				"type": "bytes32"
			},
			{
				"internalType": "bytes32",
				"name": "bodyRoot",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "finalizedExeHeaderNumber",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
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
		"name": "finalizedExeHeaders",
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
		"inputs": [
			{
				"components": [
					{
						"components": [
							{
								"internalType": "bytes32",
								"name": "parentHash",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "sha3Uncles",
								"type": "bytes32"
							},
							{
								"internalType": "address",
								"name": "miner",
								"type": "address"
							},
							{
								"internalType": "bytes32",
								"name": "stateRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "transactionsRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "receiptsRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes",
								"name": "logsBloom",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "difficulty",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "number",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "gasLimit",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "gasUsed",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "timestamp",
								"type": "uint256"
							},
							{
								"internalType": "bytes",
								"name": "extraData",
								"type": "bytes"
							},
							{
								"internalType": "bytes32",
								"name": "mixHash",
								"type": "bytes32"
							},
							{
								"internalType": "bytes",
								"name": "nonce",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "baseFeePerGas",
								"type": "uint256"
							},
							{
								"internalType": "bytes32",
								"name": "withdrawalsRoot",
								"type": "bytes32"
							}
						],
						"internalType": "struct Types.BlockHeader",
						"name": "header",
						"type": "tuple"
					},
					{
						"components": [
							{
								"internalType": "uint256",
								"name": "receiptType",
								"type": "uint256"
							},
							{
								"internalType": "bytes",
								"name": "postStateOrStatus",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "cumulativeGasUsed",
								"type": "uint256"
							},
							{
								"internalType": "bytes",
								"name": "bloom",
								"type": "bytes"
							},
							{
								"components": [
									{
										"internalType": "address",
										"name": "addr",
										"type": "address"
									},
									{
										"internalType": "bytes[]",
										"name": "topics",
										"type": "bytes[]"
									},
									{
										"internalType": "bytes",
										"name": "data",
										"type": "bytes"
									}
								],
								"internalType": "struct Types.TxLog[]",
								"name": "logs",
								"type": "tuple[]"
							}
						],
						"internalType": "struct Types.TxReceipt",
						"name": "txReceipt",
						"type": "tuple"
					},
					{
						"internalType": "bytes",
						"name": "keyIndex",
						"type": "bytes"
					},
					{
						"internalType": "bytes[]",
						"name": "proof",
						"type": "bytes[]"
					}
				],
				"internalType": "struct Types.ReceiptProof",
				"name": "receiptProof",
				"type": "tuple"
			}
		],
		"name": "getBytes",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "",
				"type": "bytes"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"internalType": "bytes32",
						"name": "parentHash",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "sha3Uncles",
						"type": "bytes32"
					},
					{
						"internalType": "address",
						"name": "miner",
						"type": "address"
					},
					{
						"internalType": "bytes32",
						"name": "stateRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "transactionsRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes32",
						"name": "receiptsRoot",
						"type": "bytes32"
					},
					{
						"internalType": "bytes",
						"name": "logsBloom",
						"type": "bytes"
					},
					{
						"internalType": "uint256",
						"name": "difficulty",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "number",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "gasLimit",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "gasUsed",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "timestamp",
						"type": "uint256"
					},
					{
						"internalType": "bytes",
						"name": "extraData",
						"type": "bytes"
					},
					{
						"internalType": "bytes32",
						"name": "mixHash",
						"type": "bytes32"
					},
					{
						"internalType": "bytes",
						"name": "nonce",
						"type": "bytes"
					},
					{
						"internalType": "uint256",
						"name": "baseFeePerGas",
						"type": "uint256"
					},
					{
						"internalType": "bytes32",
						"name": "withdrawalsRoot",
						"type": "bytes32"
					}
				],
				"internalType": "struct Types.BlockHeader[]",
				"name": "_headers",
				"type": "tuple[]"
			}
		],
		"name": "getHeadersBytes",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "",
				"type": "bytes"
			}
		],
		"stateMutability": "pure",
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
				"components": [
					{
						"components": [
							{
								"internalType": "uint64",
								"name": "slot",
								"type": "uint64"
							},
							{
								"internalType": "uint64",
								"name": "proposerIndex",
								"type": "uint64"
							},
							{
								"internalType": "bytes32",
								"name": "parentRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "stateRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "bodyRoot",
								"type": "bytes32"
							}
						],
						"internalType": "struct Types.BeaconBlockHeader",
						"name": "attestedHeader",
						"type": "tuple"
					},
					{
						"components": [
							{
								"internalType": "bytes",
								"name": "pubkeys",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "aggregatePubkey",
								"type": "bytes"
							}
						],
						"internalType": "struct Types.SyncCommittee",
						"name": "nextSyncCommittee",
						"type": "tuple"
					},
					{
						"internalType": "bytes32[]",
						"name": "nextSyncCommitteeBranch",
						"type": "bytes32[]"
					},
					{
						"components": [
							{
								"internalType": "uint64",
								"name": "slot",
								"type": "uint64"
							},
							{
								"internalType": "uint64",
								"name": "proposerIndex",
								"type": "uint64"
							},
							{
								"internalType": "bytes32",
								"name": "parentRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "stateRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "bodyRoot",
								"type": "bytes32"
							}
						],
						"internalType": "struct Types.BeaconBlockHeader",
						"name": "finalizedHeader",
						"type": "tuple"
					},
					{
						"internalType": "bytes32[]",
						"name": "finalityBranch",
						"type": "bytes32[]"
					},
					{
						"components": [
							{
								"internalType": "bytes32",
								"name": "parentHash",
								"type": "bytes32"
							},
							{
								"internalType": "address",
								"name": "feeRecipient",
								"type": "address"
							},
							{
								"internalType": "bytes32",
								"name": "stateRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "receiptsRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes",
								"name": "logsBloom",
								"type": "bytes"
							},
							{
								"internalType": "bytes32",
								"name": "prevRandao",
								"type": "bytes32"
							},
							{
								"internalType": "uint256",
								"name": "blockNumber",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "gasLimit",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "gasUsed",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "timestamp",
								"type": "uint256"
							},
							{
								"internalType": "bytes",
								"name": "extraData",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "baseFeePerGas",
								"type": "uint256"
							},
							{
								"internalType": "bytes32",
								"name": "blockHash",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "transactionsRoot",
								"type": "bytes32"
							},
							{
								"internalType": "bytes32",
								"name": "withdrawalsRoot",
								"type": "bytes32"
							}
						],
						"internalType": "struct Types.Execution",
						"name": "finalizedExecution",
						"type": "tuple"
					},
					{
						"internalType": "bytes32[]",
						"name": "executionBranch",
						"type": "bytes32[]"
					},
					{
						"components": [
							{
								"internalType": "bytes",
								"name": "syncCommitteeBits",
								"type": "bytes"
							},
							{
								"internalType": "bytes",
								"name": "syncCommitteeSignature",
								"type": "bytes"
							}
						],
						"internalType": "struct Types.SyncAggregate",
						"name": "syncAggregate",
						"type": "tuple"
					},
					{
						"internalType": "uint64",
						"name": "signatureSlot",
						"type": "uint64"
					}
				],
				"internalType": "struct Types.LightClientUpdate",
				"name": "_update",
				"type": "tuple"
			}
		],
		"name": "getUpdateBytes",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "",
				"type": "bytes"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "headerHeight",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "initialized",
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
		"name": "initStage",
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
		"name": "MAX_BLOCK_SAVED",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "MAX_DELETE_COUNT",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "mptVerify",
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
		"name": "NEXT_SYNC_COMMITTEE_PROOF_SIZE",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
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
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"name": "syncCommittees",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "pubkeys",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "aggregatePubkey",
				"type": "bytes"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "verifiableHeaderRange",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "_receiptProof",
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
				"name": "logs",
				"type": "bytes"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

	lightManagerString = `[
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
		"inputs": [
			{
				"internalType": "address",
				"name": "_admin",
				"type": "address"
			}
		],
		"name": "changeAdmin",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			}
		],
		"name": "clientState",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "",
				"type": "bytes"
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
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			}
		],
		"name": "headerHeight",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "initialize",
		"outputs": [],
		"stateMutability": "nonpayable",
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
		"name": "lightClientContract",
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
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			},
			{
				"internalType": "address",
				"name": "_contract",
				"type": "address"
			}
		],
		"name": "register",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "_blockHeader",
				"type": "bytes"
			}
		],
		"name": "updateBlockHeader",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "_data",
				"type": "bytes"
			}
		],
		"name": "updateLightClient",
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
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			}
		],
		"name": "verifiableHeaderRange",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_chainId",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "_receiptProof",
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
				"name": "logs",
				"type": "bytes"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`
)
