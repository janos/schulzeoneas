// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"resenje.org/eas"
)

const (
	defaultEndpoint           = "https://ethereum-sepolia-rpc.publicnode.com/"
	defaultEASContractAddress = "0xC2679fBD37d54388Ce493F1DB75320D236e1815e"
	defaultConfigUID          = "0x95061c892e6fad7afc7dc9d625d39e652ba221efeb1afa5e1dd4266c11a145c8"
)

func runApp() error {
	cli := flag.NewFlagSet("schulzeoneas", flag.ExitOnError)

	configDirFlag := cli.String("config-dir", "", "Local configuration directory")
	endpointFlag := cli.String("rpc-endpoint", defaultEndpoint, "Ethereum RPC URL")
	easContractAddressFlag := cli.String("eas-contract-address", defaultEASContractAddress, "Ethereum Attestation Service EAS contract address")
	configUIDFlag := cli.String("uid", defaultConfigUID, "UID of the SchulzeOnEAS config attestation")

	if err := cli.Parse(os.Args[1:]); err != nil {
		log.Println(err)
		cli.Usage()
	}

	configDir := *configDirFlag
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		configDir = dir
	}

	return newApp(configDir, *endpointFlag, common.HexToAddress(*easContractAddressFlag), eas.HexDecodeUID(*configUIDFlag))
}
