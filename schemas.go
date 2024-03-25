// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"resenje.org/eas"
)

type votingSchema struct {
	Title   string   `abi:"title"`
	Choices []string `abi:"choices"`
}

type ballotSchema []ballotRanking

type ballotRanking struct {
	ChoiceIndex uint16 `abi:"choiceIndex"`
	Rank        uint16 `abi:"rank"`
}

type configSchema struct {
	VotingSchemaUID   eas.UID `abi:"votingSchemaUID"`
	VotingSchemaBlock uint64  `abi:"votingSchemaBlock"`
	BallotSchemaUID   eas.UID `abi:"ballotSchemaUID"`
	BallotSchemaBlock uint64  `abi:"ballotSchemaBlock"`
}
