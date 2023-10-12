package main

import (
	"testing"

	"github.com/ThinkiumGroup/go-common/abi"
)

func TestEth2LightNodeAbi(t *testing.T) {
	{
		ab := *abi.MustInitAbi("eth2 light node", eth2LightNodeString)
		_testShowAbi(t, ab)
	}
	{
		ab := *abi.MustInitAbi("light manager", lightManagerString)
		_testShowAbi(t, ab)
	}
}
