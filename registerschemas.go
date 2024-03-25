// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"resenje.org/eas"
)

func registerSchemasCommand() error {
	cli := flag.NewFlagSet("schulzeoneas register-schemas", flag.ExitOnError)

	endpointFlag := cli.String("rpc-endpoint", "https://ethereum-sepolia-rpc.publicnode.com/", "")
	easContractAddressFlag := cli.String("eas-contract-address", "0xC2679fBD37d54388Ce493F1DB75320D236e1815e", "")
	privateKeyFlag := cli.String("private-key", "", "")
	configSchemaUIDFlag := cli.String("config-schema-uid", "", "")
	votingSchemaUIDFlag := cli.String("voting-schema-uid", "", "")
	votingSchemaBlockNumberFlag := cli.Uint64("voting-schema-block", 0, "")
	ballotSchemaUIDFlag := cli.String("ballot-schema-uid", "", "")
	ballotSchemaBlockNumberFlag := cli.Uint64("ballot-schema-block", 0, "")

	if err := cli.Parse(os.Args[2:]); err != nil {
		log.Println(err)
		cli.Usage()
	}

	configSchemaUID := eas.HexDecodeUID(*configSchemaUIDFlag)

	ctx := context.Background()

	provateKey, err := eas.HexParsePrivateKey(*privateKeyFlag)
	if err != nil {
		return err
	}

	log.Println("Wallet address:", crypto.PubkeyToAddress(provateKey.PublicKey))

	client, err := eas.NewClient(ctx, *endpointFlag, provateKey, common.HexToAddress(*easContractAddressFlag), nil)
	if err != nil {
		return err
	}

	config := configSchema{
		VotingSchemaUID:   eas.HexDecodeUID(*votingSchemaUIDFlag),
		VotingSchemaBlock: *votingSchemaBlockNumberFlag,
		BallotSchemaUID:   eas.HexDecodeUID(*ballotSchemaUIDFlag),
		BallotSchemaBlock: *ballotSchemaBlockNumberFlag,
	}

	if configSchemaUID.IsZero() {
		tx, wait, err := registerConfigSchema(ctx, client)
		if err != nil {
			return err
		}
		log.Println("Waiting Config schema registration:", tx)
		u, blockNumber, err := wait(ctx)
		if err != nil {
			return err
		}
		log.Println("Config Schema UID:", u, "at block", blockNumber)
		configSchemaUID = u
	}

	if config.VotingSchemaUID.IsZero() {
		tx, wait, err := registerVotingSchema(ctx, client)
		if err != nil {
			return err
		}
		log.Println("Waiting Voting schema registration:", tx)
		u, blockNumber, err := wait(ctx)
		if err != nil {
			return err
		}
		log.Println("Voting Schema UID:", u, "at block", blockNumber)
		config.VotingSchemaUID = u
		config.VotingSchemaBlock = blockNumber
	}

	if config.BallotSchemaUID.IsZero() {
		tx, wait, err := registerBallotSchema(ctx, client)
		if err != nil {
			return err
		}
		log.Println("Waiting Ballot schema registration:", tx)
		u, blockNumber, err := wait(ctx)
		if err != nil {
			return err
		}
		log.Println("Ballot Schema UID:", u, "at block", blockNumber)
		config.BallotSchemaUID = u
		config.BallotSchemaBlock = blockNumber
	}

	tx, wait, err := client.EAS.Attest(ctx, configSchemaUID, nil, config)
	if err != nil {
		return err
	}
	log.Println("Waiting Config attestation:", tx.Hash())
	r, err := wait(ctx)
	if err != nil {
		return err
	}
	log.Println("Config Schema attestation:", r.UID)

	return nil
}

var registerVotingSchema = registerSchema[votingSchema]
var registerBallotSchema = registerSchema[ballotSchema]
var registerConfigSchema = registerSchema[configSchema]

func registerSchema[T any](ctx context.Context, client *eas.Client) (common.Hash, func(context.Context) (eas.UID, uint64, error), error) {
	var t T
	schema, err := eas.NewSchema(t)
	if err != nil {
		return common.Hash{}, nil, err
	}
	tx, wait, err := client.SchemaRegistry.Register(ctx, schema, common.Address{}, true)
	if err != nil {
		return common.Hash{}, nil, err
	}
	return tx.Hash(), func(ctx context.Context) (eas.UID, uint64, error) {
		r, err := wait(ctx)
		if err != nil {
			return eas.UID{}, 0, err
		}

		return r.UID, r.Raw.BlockNumber, nil
	}, nil
}
