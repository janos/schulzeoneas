// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"resenje.org/eas"
)

type app struct {
	*tview.Application

	ethereumEndpoint   string
	easContractAddress common.Address
	configUID          eas.UID

	keystore *keystore.KeyStore
	client   *eas.Client
	config   *configSchema
}

func newApp(
	configDir string,
	ethereumEndpoint string,
	easContractAddress common.Address,
	configUID eas.UID,
) error {
	keystoreDir := filepath.Join(configDir, "SchulzeOnEAS", "keystore")

	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return err
	}

	a := &app{
		Application: tview.NewApplication(),

		ethereumEndpoint:   ethereumEndpoint,
		easContractAddress: easContractAddress,
		configUID:          configUID,

		keystore: keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP),
	}
	a.render(a.newSetAccountOptions())
	return a.Run()
}

func (a *app) render(primitive tview.Primitive) {
	frame := tview.NewFrame(primitive)
	frame.AddText("Schulze on EAS", true, tview.AlignCenter, tcell.ColorWhite)
	if a.client != nil {
		frame.AddText(shortAddress(a.client.Address()), true, tview.AlignRight, tcell.ColorWhite)
	}
	a.SetRoot(frame, true)
}

func (a *app) renderAsync(previous tview.Primitive, message string, wait func() (tview.Primitive, error)) {
	modal := tview.NewModal()
	modal.SetText(message)
	go func() {
		defer a.Draw()

		p, err := wait()
		if err != nil {
			a.render(a.newMessage(previous, "Error: "+err.Error()))
			return
		}
		if p != nil {
			a.render(p)
		} else {
			a.render(a.newMainMenu())
		}
	}()
	a.render(modal)
}

func (a *app) newMessage(previous tview.Primitive, message string) tview.Primitive {
	modal := tview.NewModal()
	modal.SetText(message)
	modal.AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if previous != nil {
			a.render(previous)
		}
		a.render(a.newMainMenu())
	})
	return modal
}

func shortAddress(a common.Address) string {
	s := a.Hex()
	if l := len(s); l > 8 {
		return s[:6] + ".." + s[l-3:]
	}
	return s
}
