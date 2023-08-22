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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/abi"
	"github.com/ThinkiumGroup/go-tkmrpc/client"
	"github.com/ThinkiumGroup/go-tkmrpc/models"
	"golang.org/x/term"
)

const (
	defaultGas = 10000000
	timeFormat = "2006-01-02 15:04:05"
)

func mustInitAbi(name, abiString string) abi.ABI {
	a, err := abi.JSON(bytes.NewReader([]byte(abiString)))
	if err != nil {
		panic(fmt.Errorf("read %s abi failed: %w", name, err))
	}
	return a
}

func localip() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func unixSecondsString(us int64) string {
	return time.Unix(us, 0).Format(timeFormat)
}

func getSourceCommOfEpoch(ctx context.Context, src *client.Client, epoch common.EpochNum) (*models.Committee, error) {
	nids, err := src.Committee(ctx, epoch)
	if err != nil {
		return nil, fmt.Errorf("get committee of Epoch:%s failed: %w", epoch, err)
	}
	comm := models.NewCommittee().SetMembers(nids)
	if !comm.IsAvailable() {
		return comm, errors.New("committee not available")
	}
	return comm, nil
}

func committeeEquals(comm *models.Committee, addrs []common.Address) bool {
	if comm.Size() == 0 && len(addrs) == 0 {
		return true
	}
	if comm.Size() != len(addrs) {
		return false
	}

	addrMap := make(map[common.Address]struct{})
	for _, addr := range addrs {
		addrMap[addr] = struct{}{}
	}

	for _, nid := range comm.Members {
		ad, _ := common.AddressFromPubSlice(models.TKMCipher.PubFromNodeId(nid[:]))
		if _, exist := addrMap[ad]; !exist {
			return false
		}
		delete(addrMap, ad)
	}
	return len(addrMap) == 0
}

func readPwd(hint string, defaultVal string, mask string) ([]byte, error) {
	var ioBuf []rune
	if hint != "" {
		fmt.Print(hint)
	}
	if strings.Index(hint, "\n") >= 0 {
		hint = strings.TrimSpace(hint[strings.LastIndex(hint, "\n"):])
	}
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer fmt.Println()
	defer func() {
		_ = term.Restore(fd, state)
	}()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		b, _, err := inputReader.ReadRune()
		if err != nil {
			return nil, err
		}
		if b == 0x0d {
			strValue := strings.TrimSpace(string(ioBuf))
			if len(strValue) == 0 {
				strValue = defaultVal
			}
			return []byte(strValue), nil
		}
		if b == 0x08 || b == 0x7F {
			if len(ioBuf) > 0 {
				ioBuf = ioBuf[:len(ioBuf)-1]
			}
			fmt.Print("\r")
			for i := 0; i < len(ioBuf)+2+len(hint); i++ {
				fmt.Print(" ")
			}
		} else {
			ioBuf = append(ioBuf, b)
		}
		fmt.Print("\r")
		if hint != "" {
			fmt.Print(hint)
		}
		for i := 0; i < len(ioBuf); i++ {
			fmt.Print(mask)
		}
	}
}

func locateLog(logs models.Logs, contractAddr common.Address, topicId common.Hash) (int, *models.Log) {
	for i, l := range logs {
		if l.Address == contractAddr && len(l.Topics) > 0 && l.Topics[0] == topicId {
			return i, l
		}
	}
	return -1, nil
}
