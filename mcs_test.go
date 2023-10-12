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

func TestMCSABI(t *testing.T) {
	_testShowAbi(t, MCSAbi)
}

func TestMCSRelayABI(t *testing.T) {
	_testShowAbi(t, MCSRelayAbi)
}

func TestTransferInInput(t *testing.T) {
	bs, _ := hex.DecodeString("d24c6944000000000000000000000000000000000000000000000000000000000000c35100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000001fc0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000ac00000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000c2000000000000000000000000000000000000000000000000000000000000012a00000000000000000000000000000000000000000000000000000000000001d0000000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000001015600000000000000000000000000000000000000000000000000000000000001c00000000000000000000000000000000000000000000000000000000000000920000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000101560000000000000000000000000000000000000000000000000000000000000960000000000000000000000000000000000000000000000000000000000000098000000000000000000000000000000000000000000000000000000000000009a0000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000277b22666565223a223236333531323030303030303030303030222c22726f6f74223a6e756c6c7d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000003e0000000000000000000000000039838258a99701932bc0a3308070025fff4b90f000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000001a5ed54daa132fff621be5815bb5fb627be80013eeaa4c9f173c299f3384c9bc4487ed100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000038c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92500000000000000000000000036436492ec5f234b1d0a5d1b61d0fe621c3fe90e0000000000000000000000003d8d602d3628efefe95a4d47f89518453d0833ce00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000039838258a99701932bc0a3308070025fff4b90f000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000001a5ed54daa132fff621be5815bb5fb627be80013eeaa4c9f173c299f3384c9bc4487ed10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000036436492ec5f234b1d0a5d1b61d0fe621c3fe90e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000008ac7230489e800000000000000000000000000003d8d602d3628efefe95a4d47f89518453d0833ce000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000001a5ed54daa132fff621be5815bb5fb627be80013eeaa4c9f173c299f3384c9bc4487ed1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000344ff77018688dad4b245e8ab97358ed57ed92269952ece7ffd321366ce078622000000000000000000000000000000000000000000000000000000000000c351000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000001c03e45fd66107400742f43526fa14bbb5537643cce95f88ec32287db869a72366100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000008ac7230489e8000000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000000000014039838258a99701932bc0a3308070025fff4b90f000000000000000000000000000000000000000000000000000000000000000000000000000000000000001436436492ec5f234b1d0a5d1b61d0fe621c3fe90e000000000000000000000000000000000000000000000000000000000000000000000000000000000000001436436492ec5f234b1d0a5d1b61d0fe621c3fe90e000000000000000000000000000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020daa132fff621be5815bb5fb627be80013eeaa4c9f173c299f3384c9bc4487ed10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000007886d5fdf034cdb4a6be13c7c567a7a97ca1d502000000000000000000000000000000000000000000000000001c15e20e9f2000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000549ec2f4f56e2d56a708b8c559704826c0e3cb5c0c21d23a0475e8c6c6cf6c8900000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000c4177f38691dd52b2dacd27962d0357a67fc29ac1604946901827b26908257270000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000089c8c0cf74358ea18c391fb8e6fb740894da4dc50d3b83a56b534aa898536da300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000ed1e24f13a667e1aaa668fd9e3f4d6b79fe1def572e9ffae7c4888c90597f2f3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005f03e62c1f4df1a002ac901ad9c07f6fce5f1c1c481cd9074fda56d14f4dbfa100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000652d4fa39dba8c042f82912baba904616ec6c417638b2f4f38b6353cf12682c000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000eae270252033b30fdeb4b5ff219389f898185e60707cb61fd9626e1772761fd2000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a04b7a7857236717dc2eb910e49a6f95543a4d38fd5e8cc0a51c9309ff056bf000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002b3420931c3f4000c71bf1ced57e94e0c6dbc42cb1cf7b95f031754fe57102e200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000f2e03ee0596cc9b982ed667d4a3e7f51d1cd66f53472f6b6cba3059f601904ca00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000681afa780d17da29203322b473d3f210a7d621259a4e6ce9e403f5a266ff719a00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000ba067d6ae5f2617278b1fa54d6ee728758d6bfdf5a527030dd26deede9d33e1e0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000038ad85309f797e145e6b37ed88e6bd4452204c751677f297fbf9c407e32029c600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b66f71cccf1f478c99c7565ee15804ac6cb57d3718a670b31e2c1fc129920bd6000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000007481638863f6bffd2117a897be8eb35fdcf825552683d32fcc6a8bca51e031ff0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007ade712a257c68b9deea0a8a82893f97b9d86e7ad31fbab3b90a73806b516fd0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004e0000000000000000000000000000000000000000000000000000000000000052000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a8250c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005600000000000000000000000000b70e6f67512bcd07b7d1cbbd04dbbfadfbeaf37000000000000000000000000000000000000000000000000000000000000058000000000000000000000000000000000000000000000000000000000000005c000000000000000000000000000000000000000000000000000000000000005e000000000000000000000000000000000000000000000000000000000000006200000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000068000000000000000000000000000000000000000000000000000000000000006c0000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000007400000000000000000000000000000000000000000000000000000000000000760000000000000000000000000000000000000000000000000000000000000078000000000000000000000000000000000000000000000000000000000000007a000000000000000000000000000000000000000000000000000000000000007e0000000000000000000000000000000000000000000000000000000000000082000000000000000000000000000000000000000000000000000000000000008400000000000000000000000000000000000000000000000000000000000000860000000000000000000000000000000000000000000000000000000000000088000000000000000000000000000000000000000000000000000000000000008a000000000000000000000000000000000000000000000000000000000000008c00000000000000000000000000000000000000000000000000000000064fab0430000000000000000000000000000000000000000000000000000000000000900000000000000000000000000000000000000000000000000000000000000092000000000000000000000000000000000000000000000000000000000000009400000000000000000000000000000000000000000000000000000000000000980000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000009a000000000000000000000000000000000000000000000000000000000000009e00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000a2000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a400000000000000000000000000000000000000000000000000000000000000020141cae1c0891d42b7af0f89c3ab34ecb12a6f3e9d3e4ce7b2c2ce85bf2662cbc0000000000000000000000000000000000000000000000000000000000000020a676599042a22e3e4e8c1f4ea4e8daa99f10255b7969127b020509d31a1244450000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002034e416b71d52da72a025c844adc242efda0fc8ebbd0b6d27938d9cce41690fb7000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000206ce67f98de967892b5ed540bb129480e1ad9f6816ed38881e4110d2d7310647f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000014f72b5456693eb329f80d2ec07e211dea8240511900000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000001b250000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020cf1c76b72c1d1bd7c488256c0d7169ef24e94220fce2df8db94a5353388082ea0000000000000000000000000000000000000000000000000000000000000020cf1c76b72c1d1bd7c488256c0d7169ef24e94220fce2df8db94a5353388082ea0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020807eb71e12d012f1ad3b00c7d78d683128250551c28c4a60502c5cc823c20ceb00000000000000000000000000000000000000000000000000000000000000205d132ab1556ce3c597e5313371e924e4d79b2cdc2f53a3cc657c1a0bb7c4d39e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020ccf256e5cdc25bb9fdc38010690d3d875c623bf604189de212467915a08a3020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020429ff17a22fc482b943e8bddcbed8f5afc1a6ca6395dbfa1ad7bd8fb66a2898f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000203c65573510ff371f2b654e2f35c0cf38331eb73109f3e8a5925e311586b23acb00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000416b77bdf5448010955dfee83a6f0a90d828054ce491610f42166f2da97d02e8724d998814a87d35fd118889854ab9def427c3c1bd67d6eafcf8d5110aa99e76700100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004176c58e0ee9d5b25133a24e5deba8003fa3a4e1051ed931eff1f036af842f17f006b8cf6faaf9135fa8ca5e215305352c834a7f4acb428e14b3b7e41f085e69cb010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000415a768c6653a475a16ad698a242dba5d3502a3450c00df85a1d08e47df8a1f46d7d4ada58bd13d177ae17dd166301ade4db26c0b4217fa1e45f152c8b66f996d600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041c03836f5adecc27278a63b6a4f9ee385be8b8a26cc76cae84ba8db84028752655928a81609ad49b7303564008054f80c117e3294d5efd9113dc365c4ca74cfd20100000000000000000000000000000000000000000000000000000000000000")
	one := new(struct {
		ChainID   *big.Int `abi:"_chainId"`
		ProofData []byte   `abi:"_receiptProof"`
	})
	if err := MCSAbi.UnpackInput(one, transferInName, bs[4:]); err != nil {
		t.Fatal(err)
	}
	t.Log("chainid:", one.ChainID)
	b := new(struct {
		Proof TKMReceiptProof `abi:"receiptProof"`
	})
	if err := LightNodeABI.UnpackInput(b, verifyReceiptStruct, one.ProofData); err != nil {
		t.Fatal("receipt parse failed", err)
	}
	t.Logf("%+v", b.Proof)
}
