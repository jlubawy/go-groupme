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

var messagesOptions = struct {
	groupme.MessagesIndexOptions
	Compact bool
}{}

var messagesCommand = cli.Command{
	Name:             "messages",
	ShortDescription: "query messages from a particular group",
	Description:      `Query messages from a particular group.`,
	ShortUsage:       "[group ID]",
	SetupFlags: func(fs *flag.FlagSet) {
		fs.StringVar(&messagesOptions.BeforeID, "before", "", "returns messages created before the given message ID")
		fs.StringVar(&messagesOptions.SinceID, "since", "", "returns most recent messages created after the given message ID")
		fs.StringVar(&messagesOptions.AfterID, "after", "", "returns messages created immediately after the given message ID")
		fs.IntVar(&messagesOptions.Limit, "limit", 0, "limit the number of messages returned, the maximum is 100")
		fs.BoolVar(&messagesOptions.Compact, "compact", false, "output compact JSON")
	},
	Run: func(args []string) {
		if len(args) == 0 {
			cli.Fatal("Must provide a group ID.\n")
		} else if len(args) > 1 {
			cli.Fatal("Multiple group IDs provided.\n")
		}

		client := groupme.NewClient(context.Background(), AccessToken)

		service := groupme.NewMessagesService(client)
		messages, err := service.Index(args[0], &messagesOptions.MessagesIndexOptions)
		if err != nil {
			cli.Fatalf("Error indexing messages: %v\n", err)
		}

		enc := json.NewEncoder(os.Stdout)
		if !messagesOptions.Compact {
			enc.SetIndent("", "  ")
		}
		if err := enc.Encode(&messages); err != nil {
			cli.Fatalf("Error encoding messages: %v\n", err)
		}
	},
}
