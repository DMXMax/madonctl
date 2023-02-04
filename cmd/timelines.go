// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/McKael/madon/v3"
)

var timelineOpts struct {
	local, onlyMedia bool
	limit, keep      uint
	sinceID, maxID   madon.ActivityID
}

// timelineCmd represents the timelines command
var timelineCmd = &cobra.Command{
	Use:     "timeline [home|public|direct|:HASHTAG|!list_id] [--local]",
	Aliases: []string{"tl"},
	Short:   "Fetch a timeline",
	Long: `
The timeline command fetches a timeline (home, local or federated).
The timeline "direct" contains only direct messages (that is, messages with
visibility set to "direct").
It can also get a hashtag-based timeline if the keyword or prefixed with
':' or '#', or a list-based timeline (use !ID with the list ID).`,
	Example: `  madonctl timeline
  madonctl timeline public --local
  madonctl timeline '!42'
  madonctl timeline :mastodon
  madonctl timeline direct`,
	RunE:      timelineRunE,
	ValidArgs: []string{"home", "public", "direct"},
}

func init() {
	RootCmd.AddCommand(timelineCmd)

	timelineCmd.Flags().BoolVar(&timelineOpts.local, "local", false, "Posts from the local instance")
	timelineCmd.Flags().BoolVar(&timelineOpts.onlyMedia, "only-media", false, "Only statuses with media attachments")
	timelineCmd.Flags().UintVarP(&timelineOpts.limit, "limit", "l", 0, "Limit number of API results")
	timelineCmd.Flags().UintVarP(&timelineOpts.keep, "keep", "k", 0, "Limit number of results")
	timelineCmd.PersistentFlags().StringVar(&timelineOpts.sinceID, "since-id", "", "Request IDs greater than a value")
	timelineCmd.PersistentFlags().StringVar(&timelineOpts.maxID, "max-id", "", "Request IDs less (or equal) than a value")
}

func timelineRunE(cmd *cobra.Command, args []string) error {
	opt := timelineOpts
	var limOpts *madon.LimitParams

	if opt.limit > 0 || opt.sinceID != "" || opt.maxID != "" {
		limOpts = new(madon.LimitParams)
	}

	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}
	if opt.maxID != "" {
		limOpts.MaxID = opt.maxID
	}
	if opt.sinceID != "" {
		limOpts.SinceID = opt.sinceID
	}

	tl := "home"
	if len(args) > 0 {
		tl = args[0]
	}

	// Home timeline and list-based timeline require to be logged in
	if err := madonInit(tl == "home" || tl == "direct" || strings.HasPrefix(tl, "!")); err != nil {
		return err
	}

	sl, err := gClient.GetTimelines(tl, opt.local, opt.onlyMedia, limOpts)
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	if opt.keep > 0 && len(sl) > int(opt.keep) {
		sl = sl[:opt.keep]
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return p.printObj(sl)
}
