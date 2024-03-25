// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rivo/tview"
)

func (a *app) newSetAccountOptions() tview.Primitive {
	accounts := a.keystore.Accounts()

	switch len(accounts) {
	case 0:
		return a.newManageAccountsMenu()
	case 1:
		return a.newUnlockAccountForm(nil, accounts[0].Address)
	default:
		return a.newSelectAccountMenu()
	}
}

func (a *app) newManageAccountsMenu() tview.Primitive {
	list := tview.NewList()
	if len(a.keystore.Accounts()) > 0 {
		list.AddItem("Select account", "", 's', func() {
			a.render(a.newSelectAccountMenu())
		})
	}
	list.AddItem("Import existing account", "", 'i', func() {
		a.render(a.showImportAccount(list))
	})
	list.AddItem("Create a new account", "", 'c', func() {
		a.render(a.newCreateAccountForm(list))
	})
	if len(a.keystore.Accounts()) > 0 {
		list.AddItem("Delete account", "", 'd', func() {
			a.render(a.newDeleteAccountMenu(list))
		})
	}
	if a.client == nil {
		list.AddItem("Quit", "", 'q', func() {
			a.Stop()
		})
	} else {
		list.AddItem("Main menu", "", 'm', func() {
			a.render(a.newMainMenu())
		})
	}
	list.SetBorder(true).SetTitle(" Account ").SetTitleAlign(tview.AlignLeft)
	return list
}

func (a *app) newSelectAccountMenu() tview.Primitive {
	list := tview.NewList()
	for _, account := range a.keystore.Accounts() {
		list.AddItem(account.Address.String(), "", 0, func() {
			a.render(a.newUnlockAccountForm(list, account.Address))
		})
	}
	list.AddItem("Manage accounts", "", 'm', func() {
		a.render(a.newManageAccountsMenu())
	})
	if a.client == nil {
		list.AddItem("Quit", "", 'q', func() {
			a.Stop()
		})
	}
	list.SetBorder(true).SetTitle(" Select account ").SetTitleAlign(tview.AlignLeft)
	return list
}

func (a *app) newCreateAccountForm(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var password string
	form.AddPasswordField("Password", "", 16, '*', func(text string) {
		password = text
	})
	var confirmation string
	form.AddPasswordField("Confirm", "", 16, '*', func(text string) {
		confirmation = text
	})
	form.AddButton("Create", func() {
		if password != confirmation {
			a.render(a.newMessage(form, "Passwords do not match"))
			return
		}
		a.renderAsync(form, "Creating account...", func() (tview.Primitive, error) {
			account, err := a.keystore.NewAccount(password)
			if err != nil {
				return nil, err
			}
			if err := a.setAccount(account.Address, password); err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Create a new account ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) showImportAccount(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var hexKey string
	form.AddInputField("Hex key", "", 65, nil, func(text string) {
		hexKey = text
	})
	var password string
	form.AddPasswordField("Password", "", 16, '*', func(text string) {
		password = text
	})
	var confirmation string
	form.AddPasswordField("Confirm", "", 16, '*', func(text string) {
		confirmation = text
	})
	form.AddButton("Create", func() {
		if password != confirmation {
			a.render(a.newMessage(form, "Passwords do not match"))
			return
		}
		a.renderAsync(form, "Importing account...", func() (tview.Primitive, error) {
			pk, err := crypto.HexToECDSA(hexKey)
			if err != nil {
				return nil, err
			}
			account, err := a.keystore.ImportECDSA(pk, password)
			if err != nil {
				return nil, err
			}
			if err := a.setAccount(account.Address, password); err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Import account ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newUnlockAccountForm(previous tview.Primitive, address common.Address) tview.Primitive {
	form := tview.NewForm()
	var password string
	form.AddPasswordField("Password", "", 16, '*', func(text string) {
		password = text
	})
	form.AddButton("OK", func() {
		a.renderAsync(form, "Loading account...", func() (tview.Primitive, error) {
			if err := a.setAccount(address, password); err != nil {
				return nil, err
			}
			return nil, nil
		})
	})
	if previous != nil {
		form.AddButton("Cancel", func() {
			a.render(previous)
		})
	} else {
		form.AddButton("Manage accounts", func() {
			a.render(a.newManageAccountsMenu())
		})
		form.AddButton("Quit", func() {
			a.Stop()
		})
	}
	form.SetBorder(true).SetTitle(" Unlock account " + address.String() + " ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newDeleteAccountMenu(previous tview.Primitive) tview.Primitive {
	list := tview.NewList()
	for _, account := range a.keystore.Accounts() {
		list.AddItem(account.Address.String(), "", 0, func() {
			a.render(a.newDeleteAccountForm(list, account.Address))
		})
	}
	list.AddItem("Manage accounts", "", 'm', func() {
		a.render(previous)
	})
	list.SetBorder(true).SetTitle(" Select account ").SetTitleAlign(tview.AlignLeft)
	return list
}

func (a *app) newDeleteAccountForm(previous tview.Primitive, address common.Address) tview.Primitive {
	form := tview.NewForm()
	var password string
	form.AddPasswordField("Password", "", 16, '*', func(text string) {
		password = text
	})
	form.AddButton("Delete", func() {
		if err := a.deleteAccount(address, password); err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}
		if address == a.client.Address() {
			a.client = nil
			a.render(a.newSetAccountOptions())
			return
		}
		a.render(a.newMainMenu())
	})
	if previous == nil {
		form.AddButton("Cancel", func() {
			a.render(previous)
		})
	} else {
		form.AddButton("Quit", func() {
			a.Stop()
		})
	}
	form.SetBorder(true).SetTitle(" Delete account " + address.String() + " ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) setAccount(address common.Address, password string) error {
	var account *accounts.Account
	for _, a := range a.keystore.Accounts() {
		if address == a.Address {
			account = &a
			break
		}
	}
	if account == nil {
		return accounts.ErrUnknownAccount
	}

	keyJSON, err := a.keystore.Export(*account, password, "")
	if err != nil {
		return err
	}
	key, err := keystore.DecryptKey(keyJSON, "")
	if err != nil {
		return err
	}

	return a.setClient(context.Background(), key)
}

func (a *app) deleteAccount(address common.Address, password string) error {
	var account *accounts.Account
	for _, a := range a.keystore.Accounts() {
		if address == a.Address {
			account = &a
			break
		}
	}
	if account == nil {
		return accounts.ErrUnknownAccount
	}

	return a.keystore.Delete(*account, password)
}
