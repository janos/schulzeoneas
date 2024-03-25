// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
)

func main() {
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var err error
	switch command {
	case "register-schemas":
		err = registerSchemasCommand()
	default:
		err = runApp()
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
