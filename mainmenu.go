// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/rivo/tview"
)

func (a *app) newMainMenu() tview.Primitive {
	list := tview.NewList()
	list.AddItem("Create a new voting", "", 'c', func() {
		a.render(a.newCreateVotingForm(list))
	})
	list.AddItem("Vote", "", 'v', func() {
		a.render(a.newOpenBallotForm(list))
	})
	list.AddItem("Read ballot", "", 'b', func() {
		a.render(a.newOpenSubmittedBallotForm(list))
	})
	list.AddItem("Voting results", "", 'r', func() {
		a.render(a.newOpenVotingResultsForm(list))
	})
	list.AddItem("Manage accounts", "", 'm', func() {
		a.render(a.newManageAccountsMenu())
	})
	list.AddItem("About Schulze on EAS", "", 'a', func() {
		a.render(a.newMessage(list,
			"Schulze voting method on Ethereum Attestation Service\nVersion: "+version,
		))
	})
	list.AddItem("Quit", "", 'q', func() {
		a.Stop()
	})
	list.SetBorder(true)
	return list
}
