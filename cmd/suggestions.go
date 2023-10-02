// Copyright © 2018-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon/v3"
)

var suggestionsOpts struct {
	accountID  madon.ActivityID
	accountIDs string

	//limit uint
	keep uint
	//all bool
}

// suggestionsCmd represents the suggestions command
var suggestionsCmd = &cobra.Command{
	Use:     "suggestions",
	Aliases: []string{"suggestion"},
	Short:   "Display and remove the follow suggestions",
	RunE:    suggestionsGetRunE, // Defaults to list
}

func init() {
	RootCmd.AddCommand(suggestionsCmd)

	// Subcommands
	suggestionsCmd.AddCommand(suggestionsSubcommands...)

	//suggestionsGetSubcommand.Flags().UintVarP(&suggestionsOpts.limit, "limit", "l", 0, "Limit number of API results")
	suggestionsGetSubcommand.Flags().UintVarP(&suggestionsOpts.keep, "keep", "k", 0, "Limit number of results")
	//suggestionsGetSubcommand.Flags().BoolVar(&suggestionsOpts.all, "all", false, "Fetch all results")

	suggestionsDeleteSubcommand.Flags().StringVarP(&suggestionsOpts.accountID, "account-id", "a", "", "Account ID number")
	suggestionsDeleteSubcommand.Flags().StringVar(&suggestionsOpts.accountIDs, "account-ids", "", "Comma-separated list of account IDs")
}

var suggestionsSubcommands = []*cobra.Command{
	suggestionsGetSubcommand,
	suggestionsDeleteSubcommand,
}

var suggestionsGetSubcommand = &cobra.Command{
	Use:     "list",
	Short:   "Display the suggestions (default subcommand)",
	Long:    `Display the list of account suggestions.`,
	Aliases: []string{"ls", "get", "display", "show"},
	RunE:    suggestionsGetRunE,
}

var suggestionsDeleteSubcommand = &cobra.Command{
	Use:     "delete",
	Short:   "Remove an account from the suggestion list",
	Aliases: []string{"remove", "del", "rm"},
	RunE:    suggestionsDeleteRunE,
}

func suggestionsGetRunE(cmd *cobra.Command, args []string) error {
	opt := suggestionsOpts

	/*
		// Note: The API currently does not support pagination
		// Set up LimitParams
		var limOpts *madon.LimitParams
		if opt.all || opt.limit > 0 {
			limOpts = new(madon.LimitParams)
			limOpts.All = opt.all
		}
		if opt.limit > 0 {
			limOpts.Limit = int(opt.limit)
		}
	*/

	// We need to be logged in
	if err := madonInit(true); err != nil {
		return err
	}

	var obj interface{}
	var err error

	var accountList []madon.Account
	accountList, err = gClient.GetSuggestions(nil)

	if opt.keep > 0 && len(accountList) > int(opt.keep) {
		accountList = accountList[:opt.keep]
	}

	obj = accountList

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.printObj(obj)
}

func suggestionsDeleteRunE(cmd *cobra.Command, args []string) error {
	opt := suggestionsOpts
	var ids []madon.ActivityID
	var err error

	if opt.accountID == "" && len(opt.accountIDs) == 0 {
		return errors.New("missing account IDs")
	}
	if opt.accountID != "" && len(opt.accountIDs) > 0 {
		return errors.New("incompatible options")
	}

	ids, err = splitIDs(opt.accountIDs)
	if err != nil {
		return errors.New("cannot parse account IDs")
	}
	if opt.accountID != "" { // Allow --account-id
		ids = []madon.ActivityID{opt.accountID}
	}
	if len(ids) < 1 {
		return errors.New("missing account IDs")
	}

	// We need to be logged in
	if err := madonInit(true); err != nil {
		return err
	}

	for _, id := range ids {
		if e := gClient.DeleteSuggestion(id); err != nil {
			errPrint("Cannot remove account %d: %s", id, e)
			err = e
		}
	}

	if err != nil {
		os.Exit(1)
	}
	return nil
}
