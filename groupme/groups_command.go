// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/jlubawy/go-cli"
	"github.com/jlubawy/go-groupme"
)

var groupsOptions struct {
	groupme.GroupsIndexOptions
	Compact bool
}

var groupsCommand = cli.Command{
	Name:             "groups",
	ShortDescription: "query groups that the authenticated user belongs to",
	Description:      `Query groups that the authenticated user belongs to.`,
	ShortUsage:       "[-offset=0] [-limit=0] [-compact=false]",
	SetupFlags: func(fs *flag.FlagSet) {
		fs.IntVar(&groupsOptions.Offset, "offset", 0, "the page offset to start at")
		fs.IntVar(&groupsOptions.Limit, "limit", 0, "limit the number of groups returned")
		fs.BoolVar(&groupsOptions.Compact, "compact", false, "output compact JSON")
	},
	Run: func(args []string) {
		client := groupme.NewClient(context.Background(), AccessToken)

		service := groupme.NewGroupsService(client)
		groups, err := service.Index(&groupsOptions.GroupsIndexOptions)
		if err != nil {
			cli.Fatalf("Error indexing groups: %+v\n", err)
		}

		enc := json.NewEncoder(os.Stdout)
		if !groupsOptions.Compact {
			enc.SetIndent("", "  ")
		}
		if err := enc.Encode(&groups); err != nil {
			cli.Fatalf("Error encoding groups: %v\n", err)
		}
	},
}
