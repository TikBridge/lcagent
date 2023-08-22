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
	"math/big"
	"testing"
	"time"
)

func TestExpirable(t *testing.T) {
	e := NewExpirable[*big.Int]((*big.Int)(nil), 1000, 0)
	v, exist := e.Get()
	if v == nil && exist == false {
		t.Log("check nil *big.Int")
	} else {
		t.Fatalf("v=%s exist=%t", v, exist)
	}

	e.Update(nil)
	v, exist = e.Get()
	if v == nil && exist {
		t.Log("check nil *big.Int set")
	} else {
		t.Fatalf("v=%s exist=%t", v, exist)
	}
	time.Sleep(500 * time.Millisecond)
	v, exist = e.Get()
	if v == nil && exist {
		t.Log("check nil *big.Int not-expired")
	} else {
		t.Fatalf("not-expired failed: v=%s exist=%t", v, exist)
	}
	time.Sleep(time.Second)
	v, exist = e.Get()
	if v == nil && exist == false {
		t.Log("check nil *big.Int expired")
	} else {
		t.Fatalf("timeout failed: v=%s exist=%t", v, exist)
	}

	i := big.NewInt(10)
	e.Update(i)
	v, exist = e.Get()
	if v == i && exist {
		t.Log("check not-nil *big.Int set")
	} else {
		t.Fatalf("set not-nil failed: v=%s exist=%t", v, exist)
	}
}
