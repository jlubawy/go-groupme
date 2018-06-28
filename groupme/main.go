// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/jlubawy/go-cli"
)

const AccessTokenKey = "GROUPME_TOKEN"

var AccessToken string

func init() {
	AccessToken = os.Getenv(AccessTokenKey)
	if AccessToken == "" {
		cli.Fatalf("Must set access token environment variable '%s'.\n", AccessTokenKey)
	}
}

var program = cli.Program{
	Name: "groupme",
	Commands: []cli.Command{
		groupsCommand,
		messagesCommand,
	},
}

func main() { program.RunAndExit() }
