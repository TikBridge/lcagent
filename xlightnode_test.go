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
	"testing"
)

func TestXLNABI(t *testing.T) {
	_testShowAbi(t, XLightNodeAbi)
}

func TestXLNUpdate(t *testing.T) {
	bs, _ := hex.DecodeString("5f9facb0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000038000000000000000000000000000000000000000000000000000000000000006a00000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000079400000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000007c3e2b66640a5bfb6be75bc859e6bce2112279c67b78d0c942bc645152f1ff471000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000003730522229500082717bb6ba7aadfc88b3d9a51281aad53c76864c9a369e2b4a00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000a09e0280af575fa3a45ab718a90ea48d04da4912e4f0f5ebbe4d0f5f53dbd130000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006271cce8939a5076131b30ae46fc75097af365f7a50b1b528ba23a43020c59350000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000052d93d67601e8101069e3d477a70e4b4acbc12bc3c1d195af70ca961e4b70a4100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000d2df0e2711d300bd20302ea9be41ea26e5925669d63ff7e672028a25c5ab95fc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bb289e2798a98aee0f446263cd51b90eacdc0b1bb177171b2d352dfbbfef885000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000001e0000000000000000000000000000000000000000000000000000000000000024000000000000000000000000000000000000000000000000000000000000002a0000000000000000000000000000000000000000000000000000000000000004035229dbb226308059c5e0792bd6ce6aebfad874f7c2f15001f27886b2c9c76b53d9463a64eaf8095941ec7052e6e8332b347f784d374200b02095266961f16a00000000000000000000000000000000000000000000000000000000000000040722b8f54e785b09cde205a0bcedf9df844b9d6542d76ba66c3588d91a5fe83770109dd67a03545969a359289aff21e0dbbbc1c5dd33c92ce86b3fa1cf68f2b40000000000000000000000000000000000000000000000000000000000000004079d260d785858034d64641b9a5867bd08e3f6cae9a3f18e09dfd1a238808b97093d74a8f1a15d9cbab68b005383a2637cce73110f3dfb5faea0a7c371f0976b90000000000000000000000000000000000000000000000000000000000000040a93b150f11c422d8700554859281be8e34a91a859e0e021af186002c7e4a2661ea2467a63b417030d68e2fdddeb4342943dff13225da77124abf912fd092f71f0000000000000000000000000000000000000000000000000000000000000040db3e5b5ea24e1d760a59cf22cfafeed5a4e57af2108fc0df3bf457a82f754264b3fdf9d77fcab306a9809ebcd76de91e382d912a90e3f37edf4eb04f3f036d0b0000000000000000000000000000000000000000000000000000000000000040eba6f895f3e955582f10fc0f19efee43e1d3c2ee4240eb6fe106aaa387e6357e2a6a82f21624a5c0ffd9b47d00ab984206baec8614ee4d8bab1bdabe0435fea2000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000000000000000000000000000000000000000000041e9b439c1729e4ce6f96e1a1bd4be1cb16ee5cbc4202cd0e621749bbf6b56192c0390102f320d7e4e375e357d59c47bd5fc089aad448eb6bc35c89366f76cad9301000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041e5ad4156418671add27db7f6a0f7bcafa29c8d67c17acd76bc53dcfbb4a0d82260446004c9ad23c6f497ddb3d1218c4d4770ff65c69931e5b6fe7846d1e90ef4010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000417863162a0f6264309d0eb12a95749d1a81dc1af0759224ac3c5615294c53058649caa0e6caf992990fcc018fc97c69bb9ade7c8f9a82ca27ff2ee83a017f127a01000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041f08219133e599c46c53ce1081835b7ef78296651d462a8160af8736a3198115e3423008e51d1c6b0fc5bc5463c4dd39776bf0ba9f5bb88a9addcd7b5ba29bf52000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000412ebc3ab825d77c4381d65d09e6f522f9fb8d62346364fbb0a72a021cee81ec46419017e69a9107b9bd825e59e500f57e91cb1cd947fba2acbf7a37214d77f9f10100000000000000000000000000000000000000000000000000000000000000")
	obj := new(struct {
		Data XCommProofData `abi:"commData"`
	})
	if err := XLightNodeAbi.UnpackInput(obj, xUpdateCommName, bs[4:]); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", obj.Data)
}
