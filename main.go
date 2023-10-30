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
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/ThinkiumGroup/go-common"
	"github.com/ThinkiumGroup/go-common/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

const (
	version common.Version = 1002003
)

func main() {
	app := &cli.App{
		Usage:     "maintain light-node or mcs on ethereum-like chains",
		Version:   version.String(),
		Copyright: "Copyright 2023 TikBridge",
		Commands: []*cli.Command{
			{
				Name:     "maintain",
				Aliases:  []string{"mt"},
				Usage:    "maintain the TKM Light Node on Ethereum-like chains",
				Category: "TKM-SOURCE",
				Action:   maintain,
				Flags:    _maintainFlags,
				Before:   altsrc.InitInputSourceWithContext(_maintainFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
			},
			{
				Name:     "sync",
				Aliases:  []string{"sc"},
				Usage:    "synchronize TKM txs which including mapTransferOut event to Ethereum-like chains (X-RELAY or BSC)",
				Category: "TKM-SOURCE",
				Action:   sync,
				Flags:    _syncFlags,
				Before:   altsrc.InitInputSourceWithContext(_syncFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
			},
			{
				Name:     "update",
				Aliases:  []string{"u"},
				Usage:    "use admin to update committee info to the TKM Light Node on Ethereum-like chains",
				Category: "TKM-SOURCE",
				Action:   update,
				Flags:    _updateFlags,
				Before:   altsrc.InitInputSourceWithContext(_updateFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
			},
			{
				Name:     "pem",
				Aliases:  []string{"p"},
				Usage:    "generate a PEM-Encoded PKCS#8 private key file",
				Category: "MISC",
				Action:   pemfile,
				Flags:    _pemFlags,
			},
			{
				Name:     "xmaintain",
				Aliases:  []string{"xm"},
				Usage:    "maintain the X-Relay Light Node on Ethereum-like 3rd-party chains",
				Category: "X-RELAY-SOURCE",
				Action:   xmaintain,
				Flags:    _xmaintainFlags,
				Before:   altsrc.InitInputSourceWithContext(_xmaintainFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
			},
			{
				Name:     "xsync",
				Aliases:  []string{"xs"},
				Usage:    "synchronize X-Relay txs which including mapTransferOut event to Ethereum-like 3rd-party chains",
				Category: "X-RELAY-SOURCE",
				Action:   xsync,
				Flags:    _xSyncFlags,
				Before:   altsrc.InitInputSourceWithContext(_xSyncFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
			},
		},
		Flags:  _allFlags,
		Before: altsrc.InitInputSourceWithContext(_allFlags, altsrc.NewYamlSourceFromFlagFunc(_confFileFlag.Name)),
	}

	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))
	for _, cmd := range app.Commands {
		sort.Sort(cli.FlagsByName(cmd.Flags))
	}

	baseCtx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		ss := <-sigs
		log.Errorf("GOT A SYSTEM SIGNAL[%s]", ss.String())
	}()
	if err := app.RunContext(baseCtx, os.Args); err != nil {
		log.Error(err)
	}
}

func checkerror(err error) error {
	if err != nil {
		var exitErr cli.ExitCoder
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 0 {
				// want to exit the process
				log.Info(err)
				return nil
			}
			log.Errorf("%v", err)
			return exitErr
		} else {
			return cli.Exit(err, ExitUnknown)
		}
	} else {
		return nil
	}
}

func sync(ctx *cli.Context) error {
	a := &syncer{}
	a.bHandler = a
	a.lHander = a
	return checkerror(a.run(ctx))
}

func maintain(ctx *cli.Context) error {
	a := &maintainer{}
	a.bHandler = a
	a.lHander = a
	return checkerror(a.run(ctx))
}

func update(ctx *cli.Context) error {
	a := &updater{}
	a.bHandler = a
	return checkerror(a.run(ctx))
}

func xmaintain(ctx *cli.Context) error {
	a := &xmaintainer{}
	a.bHandler = a
	a.lHander = a
	return checkerror(a.run(ctx))
}

func xsync(ctx *cli.Context) error {
	a := &xsyncer{}
	a.bHandler = a
	a.lHander = a
	return checkerror(a.run(ctx))
}

func pemfile(ctx *cli.Context) error {
	if path := ctx.String(_pemInputFlag.Name); path != "" {
		log.Infof("Input PATH: %s", path)
		filebytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read input file failed: %w", err)
		}
		needPwd, pkcs8Bytes, err := ValidPEM(filebytes)
		if err != nil {
			return fmt.Errorf("valid input file failed: %w", err)
		}
		var pwd []byte
		if needPwd {
			pwd, err = readPwd("PEM need a password: ", "", "*")
			if err != nil {
				return fmt.Errorf("read password failed: %w", err)
			}
		}
		sk, err := ParsePKCS8PrivateKey(pkcs8Bytes, pwd)
		if err != nil {
			return err
		}
		priv := ETHSigner.PrivToBytes(sk)
		log.Infof("PRIV: %x", priv)
	} else {
		path = ctx.String(_pemOutputFlag.Name)
		if path == "" {
			return errors.New("output path is missing")
		}
		log.Infof("Ouput PATH: %s", path)
		privHex, err := readPwd("HEX of private key: ", "", "*")
		if err != nil {
			return fmt.Errorf("read private key failed: %w", err)
		}
		priv, err := hex.DecodeString(string(privHex))
		if err != nil {
			return fmt.Errorf("decode hex failed: %w", err)
		}
		if len(priv) != ETHSigner.LengthOfPrivateKey() {
			return errors.New("invalid input private key")
		}
		log.Infof("PRIV: %x", priv)
		privkey, err := ETHSigner.BytesToPriv(priv)
		if err != nil || privkey == nil {
			return fmt.Errorf("to ecdsa key failed: %w", err)
		}
		pwd, err := readPwd("password of file: ", "", "")
		if err != nil {
			return fmt.Errorf("read password failed: %w", err)
		}
		pwd1, err := readPwd("password of file again: ", "", "")
		if err != nil {
			return fmt.Errorf("read password the 2nd time failed: %w", err)
		}
		if !bytes.Equal(pwd, pwd1) {
			return errors.New("password not match")
		}
		pembytes, err := MarshalPrivateKeyPEM(privkey, pwd)
		if err != nil {
			return fmt.Errorf("convert to PEM failed: %w", err)
		}
		err = os.WriteFile(path, pembytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("write PEM failed: %w", err)
		}
	}
	return nil
}
