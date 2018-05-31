// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jlubawy/go-groupme"
)

const AccessTokenKey = "GROUPME_TOKEN"

type Service struct {
	Name    string
	Command func(client groupme.Client, args []string)
}

var services = []Service{
	{
		Name: "groups",
		Command: func(client groupme.Client, args []string) {
			var options groupme.GroupsIndexOptions

			fs := flag.NewFlagSet("groups", flag.ExitOnError)
			fs.Usage = func() {
				infof("format")
			}
			fs.IntVar(&options.Offset, "offset", 0, "the page offset to start at")
			fs.IntVar(&options.Limit, "limit", 0, "limit the number of groups returned")
			fs.Parse(args)

			service := groupme.NewGroupsService(client)
			groups, err := service.Index(&options)
			if err != nil {
				fatalf("Error indexing groups: %+v\n", err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(&groups); err != nil {
				fatalf("Error encoding groups: %v\n", err)
			}
		},
	},
	{
		Name: "messages",
		Command: func(client groupme.Client, args []string) {
			var (
				flagGroupID string
				options     groupme.MessagesIndexOptions
			)

			fs := flag.NewFlagSet("messages", flag.ExitOnError)
			fs.StringVar(&flagGroupID, "groupID", "", "group ID to list messages for")
			fs.StringVar(&options.BeforeID, "before", "", "returns messages created before the given message ID")
			fs.StringVar(&options.SinceID, "since", "", "returns most recent messages created after the given message ID")
			fs.StringVar(&options.AfterID, "after", "", "returns messages created immediately after the given message ID")
			fs.IntVar(&options.Limit, "limit", 0, "limit the number of messages returned, the maximum is 100")
			fs.Parse(args)

			if flagGroupID == "" {
				fatalf("Must provide a group ID.\n")
			}

			service := groupme.NewMessagesService(client)
			messages, err := service.Index(flagGroupID, &options)
			if err != nil {
				fatalf("Error indexing messages: %+v\n", err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(&messages); err != nil {
				fatalf("Error encoding messages: %v\n", err)
			}
		},
	},
}

func main() {
	flag.Parse()

	accessToken := os.Getenv(AccessTokenKey)
	if accessToken == "" {
		fatalf("Must set access token environment variable '%s'.\n\n", AccessTokenKey)
	}

	if flag.NArg() == 0 {
		fatalf("Must provide a service name.\n\n")
	}

	for _, service := range services {
		if service.Name == flag.Arg(0) {
			client := groupme.NewClient(context.Background(), accessToken)
			service.Command(client, flag.Args()[1:])
			os.Exit(0)
		}
	}

	fatalf("Unsupported service '%s'.\n\n", flag.Arg(0))
}

func infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func fatalf(format string, args ...interface{}) {
	infof(format, args...)
	os.Exit(1)
}
