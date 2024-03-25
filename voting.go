// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"resenje.org/eas"
	"resenje.org/schulze"
)

func (a *app) setClient(ctx context.Context, key *keystore.Key) error {
	pk := key.PrivateKey
	c, err := eas.NewClient(ctx, a.ethereumEndpoint, pk, a.easContractAddress, nil)
	if err != nil {
		return err
	}
	a.client = c
	if a.config == nil {
		if err := a.getConfiguration(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) getConfiguration(ctx context.Context) error {
	attestation, err := a.client.EAS.GetAttestation(ctx, a.configUID)
	if err != nil {
		return err
	}
	var config configSchema
	if err := attestation.ScanValues(&config); err != nil {
		return err
	}
	a.config = &config
	return nil
}

func (a *app) newCreateVotingForm(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var title string
	form.AddInputField("Title", "", 40, nil, func(text string) {
		title = text
	})
	choices := make([]string, 2)
	form.AddInputField("Choice 1", "", 40, nil, func(text string) {
		choices[0] = text
	})
	form.AddInputField("Choice 2", "", 40, nil, func(text string) {
		choices[1] = text
	})
	form.AddButton("Add choice", func() {
		choices = append(choices, "")
		form.AddInputField("Choice "+strconv.Itoa(len(choices)), "", 40, nil, func(text string) {
			choices[len(choices)-1] = text
		})
		a.render(form)
	})
	form.AddButton("Remove choice", func() {
		if len(choices) > 2 {
			choices = choices[:len(choices)-1]
			form.RemoveFormItem(len(choices) + 1)
			a.render(form)
			a.SetFocus(form.GetFormItem(len(choices) - 1))
		}
	})
	form.AddButton("Create", func() {
		if strings.TrimSpace(title) == "" {
			a.render(a.newMessage(form, "Title is required"))
			return
		}
		for i, c := range choices {
			if strings.TrimSpace(c) == "" {
				a.render(a.newMessage(form, fmt.Sprintf("Choice %v cannot be empty", i+1)))
				return
			}
		}

		tx, wait, err := a.client.EAS.Attest(context.Background(), a.config.VotingSchemaUID, &eas.AttestOptions{
			RefUID:    a.configUID,
			Revocable: true,
		}, votingSchema{
			Title:   title,
			Choices: choices,
		})
		if err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}

		a.renderAsync(form, "Waiting transaction\n"+tx.Hash().String(), func() (tview.Primitive, error) {
			r, err := wait(context.Background())
			if err != nil {
				return nil, err
			}
			return a.newMessage(previous, "New voting UID\n"+r.UID.String()), nil
		})
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Create new voting ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newOpenBallotForm(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var votingUID eas.UID
	form.AddInputField("Voting", "", 67, nil, func(text string) {
		votingUID = eas.HexDecodeUID(text)
	})
	form.AddButton("Open ballot", func() {
		voting, err := a.client.EAS.GetAttestation(context.Background(), votingUID)
		if err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}
		a.render(a.newSubmitBallotForm(previous, votingUID, voting))
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Open ballot ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newSubmitBallotForm(previous tview.Primitive, votingUID eas.UID, attestation *eas.Attestation) tview.Primitive {
	form := tview.NewForm()
	var voting votingSchema
	if err := attestation.ScanValues(&voting); err != nil {
		return a.newMessage(form, "Error: "+err.Error())
	}
	ballot := make(map[uint16]uint16)
	for i, choice := range voting.Choices {
		form.AddInputField(choice, "", 2, func(textToCheck string, lastChar rune) bool {
			v, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}
			if v <= 0 {
				return false
			}
			if v > math.MaxUint16 {
				return false
			}
			return true
		}, func(text string) {
			v, _ := strconv.Atoi(text)
			ballot[uint16(i)] = uint16(v)
		})
	}
	form.AddButton("Vote", func() {
		var bs ballotSchema
		for index, rank := range ballot {
			bs = append(bs, ballotRanking{
				ChoiceIndex: index,
				Rank:        rank,
			})
		}
		tx, wait, err := a.client.EAS.Attest(context.Background(), a.config.BallotSchemaUID, &eas.AttestOptions{
			RefUID:    votingUID,
			Revocable: true,
		}, bs)
		if err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}
		a.renderAsync(form, "Waiting transaction\n"+tx.Hash().String(), func() (tview.Primitive, error) {
			r, err := wait(context.Background())
			if err != nil {
				return nil, err
			}
			return a.newMessage(previous, "Submitted ballot with UID\n"+r.UID.String()), nil
		})
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Ballot " + attestation.UID.String() + " ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newOpenSubmittedBallotForm(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var ballotUID eas.UID
	form.AddInputField("Ballot", "", 67, nil, func(text string) {
		ballotUID = eas.HexDecodeUID(text)
	})
	form.AddButton("Open ballot", func() {
		ballot, err := a.client.EAS.GetAttestation(context.Background(), ballotUID)
		if err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}
		voting, err := a.client.EAS.GetAttestation(context.Background(), ballot.RefUID)
		if err != nil {
			a.render(a.newMessage(form, "Error: "+err.Error()))
			return
		}
		a.render(a.newSubmittedBallotTable(previous, voting, ballot))
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Open ballot ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newSubmittedBallotTable(previous tview.Primitive, v *eas.Attestation, b *eas.Attestation) tview.Primitive {
	table := tview.NewTable()
	table.SetBorders(true)
	var voting votingSchema
	if err := v.ScanValues(&voting); err != nil {
		return a.newMessage(previous, "Error: "+err.Error())
	}
	var ballot ballotSchema
	if err := b.ScanValues(&ballot); err != nil {
		return a.newMessage(previous, "Error: "+err.Error())
	}
	bm := make(map[uint16]uint16)
	for _, rank := range ballot {
		bm[rank.ChoiceIndex] = rank.Rank
	}
	for i, choice := range voting.Choices {
		table.SetCell(i, 0, tview.NewTableCell(choice))
		rank, ok := bm[uint16(i)]
		if !ok {
			rank = 0
		}
		table.SetCell(i, 1, tview.NewTableCell(strconv.FormatUint(uint64(rank), 10)))
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		form := tview.NewForm()
		form.AddButton("OK", func() {
			a.render(previous)
		})
		form.AddButton("Change vote", func() {
			a.render(a.newSubmitBallotForm(previous, b.RefUID, v))
		})
		form.SetBorder(true).SetTitle(" Ballot " + b.UID.String() + " ").SetTitleAlign(tview.AlignLeft)
		a.render(form)
		return nil
	})

	table.SetBorder(true).SetTitle(" Ballot " + b.UID.String() + " ").SetTitleAlign(tview.AlignLeft)
	return table
}

func (a *app) newOpenVotingResultsForm(previous tview.Primitive) tview.Primitive {
	form := tview.NewForm()
	var votingUID eas.UID
	form.AddInputField("Voting", "", 67, nil, func(text string) {
		votingUID = eas.HexDecodeUID(text)
	})
	form.AddButton("Calculate results", func() {
		a.renderAsync(form, fmt.Sprintf("Calculating results for\n %s", votingUID), func() (tview.Primitive, error) {
			v, err := a.client.EAS.GetAttestation(context.Background(), votingUID)
			if err != nil {
				return nil, err
			}
			var voting votingSchema
			if err := v.ScanValues(&voting); err != nil {
				return nil, err
			}
			ballots := make(map[common.Address]ballotSchema)
			currentBlock, err := a.client.Backend().(ethereum.BlockNumberReader).BlockNumber(context.Background())
			if err != nil {
				return nil, err
			}
			for i := a.config.VotingSchemaBlock; i < currentBlock; i += 1000 {
				if err := func() error {
					it, err := a.client.EAS.FilterAttested(context.Background(), i, eas.Ptr(i+1000), nil, nil, []eas.UID{a.config.BallotSchemaUID})
					if err != nil {
						return err
					}
					defer it.Close()

					for it.Next() {
						r := it.Value()
						b, err := a.client.EAS.GetAttestation(context.Background(), r.UID)
						if err != nil {
							return err
						}
						if b.RefUID != votingUID {
							continue
						}
						var ballot ballotSchema
						if err := b.ScanValues(&ballot); err != nil {
							return err
						}
						ballots[r.Attester] = ballot
					}

					return nil
				}(); err != nil {
					return nil, err
				}
			}
			choices := make([]uint16, 0, len(voting.Choices))
			for i := range voting.Choices {
				choices = append(choices, uint16(i))
			}
			sch := schulze.NewVoting(choices)
			for _, ballot := range ballots {
				b := make(schulze.Ballot[uint16])
				for _, r := range ballot {
					b[r.ChoiceIndex] = int(r.Rank)
				}
				if _, err := sch.Vote(b); err != nil {
					return nil, err
				}
			}
			results, _, _ := sch.Compute()
			finalResults := make([]schulze.Result[string], 0, len(results))
			for _, r := range results {
				finalResults = append(finalResults, schulze.Result[string]{
					Choice:    voting.Choices[int(r.Choice)],
					Index:     r.Index,
					Wins:      r.Wins,
					Strength:  r.Strength,
					Advantage: r.Advantage,
				})
			}
			return a.newVotingResultsTable(previous, votingUID, finalResults), nil
		})
	})
	form.AddButton("Cancel", func() {
		a.render(previous)
	})
	form.SetBorder(true).SetTitle(" Open ballot ").SetTitleAlign(tview.AlignLeft)
	return form
}

func (a *app) newVotingResultsTable(previous tview.Primitive, votingUID eas.UID, results []schulze.Result[string]) tview.Primitive {
	table := tview.NewTable()
	table.SetBorders(true)
	table.SetCell(0, 0, tview.NewTableCell("Choice"))
	table.SetCell(0, 1, tview.NewTableCell("Wins"))
	for i, r := range results {
		table.SetCell(i+1, 0, tview.NewTableCell(r.Choice))
		table.SetCell(i+1, 1, tview.NewTableCell(strconv.FormatUint(uint64(r.Wins), 10)))
	}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		form := tview.NewForm()
		form.AddButton("OK", func() {
			a.render(previous)
		})
		form.SetBorder(true).SetTitle(" Voting " + votingUID.String() + " ").SetTitleAlign(tview.AlignLeft)
		a.render(form)
		return nil
	})

	table.SetBorder(true).SetTitle(" Voting " + votingUID.String() + " ").SetTitleAlign(tview.AlignLeft)
	return table
}
