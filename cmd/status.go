// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/McKael/madon/v3"
)

var statusPostFlags *flag.FlagSet

var statusOpts struct {
	statusID int64
	unset    bool // TODO remove eventually?

	// The following fields are used for the post/toot command
	visibility     string
	sensitive      bool
	spoiler        string
	inReplyToID    int64
	mediaIDs       string
	mediaFilePath  string
	textFilePath   string
	stdin          bool
	addMentions    bool
	sameVisibility bool

	// Used for several subcommands to limit the number of results
	limit, keep uint
	//sinceID, maxID int64
	all bool
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Subcommands
	statusCmd.AddCommand(statusSubcommands...)

	// Global flags
	statusCmd.PersistentFlags().Int64VarP(&statusOpts.statusID, "status-id", "s", 0, "Status ID number")
	statusCmd.PersistentFlags().UintVarP(&statusOpts.limit, "limit", "l", 0, "Limit number of API results")
	statusCmd.PersistentFlags().UintVarP(&statusOpts.keep, "keep", "k", 0, "Limit number of results")
	//statusCmd.PersistentFlags().Int64Var(&statusOpts.sinceID, "since-id", 0, "Request IDs greater than a value")
	//statusCmd.PersistentFlags().Int64Var(&statusOpts.maxID, "max-id", 0, "Request IDs less (or equal) than a value")
	statusCmd.PersistentFlags().BoolVar(&statusOpts.all, "all", false, "Fetch all results (for reblogged-by/favourited-by)")

	// Subcommand flags
	statusReblogSubcommand.Flags().BoolVar(&statusOpts.unset, "unset", false, "Unreblog the status (deprecated)")
	statusFavouriteSubcommand.Flags().BoolVar(&statusOpts.unset, "unset", false, "Remove the status from the favourites (deprecated)")
	statusPinSubcommand.Flags().BoolVar(&statusOpts.unset, "unset", false, "Unpin the status (deprecated)")
	statusPostSubcommand.Flags().BoolVar(&statusOpts.sensitive, "sensitive", false, "Mark post as sensitive (NSFW)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.visibility, "visibility", "", "Visibility (direct|private|unlisted|public)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.spoiler, "spoiler", "", "Spoiler warning (CW)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.mediaIDs, "media-ids", "", "Comma-separated list of media IDs")
	statusPostSubcommand.Flags().StringVarP(&statusOpts.mediaFilePath, "file", "f", "", "Media file name")
	statusPostSubcommand.Flags().StringVar(&statusOpts.textFilePath, "text-file", "", "Text file name (message content)")
	statusPostSubcommand.Flags().Int64VarP(&statusOpts.inReplyToID, "in-reply-to", "r", 0, "Status ID to reply to")
	statusPostSubcommand.Flags().BoolVar(&statusOpts.stdin, "stdin", false, "Read message content from standard input")
	statusPostSubcommand.Flags().BoolVar(&statusOpts.addMentions, "add-mentions", false, "Add mentions when replying")
	statusPostSubcommand.Flags().BoolVar(&statusOpts.sameVisibility, "same-visibility", false, "Use same visibility as original message (for replies)")

	// Deprecated flags
	statusReblogSubcommand.Flags().MarkDeprecated("unset", "please use unboost instead")
	statusFavouriteSubcommand.Flags().MarkDeprecated("unset", "please use unfavourite instead")
	statusPinSubcommand.Flags().MarkDeprecated("unset", "please use unpin instead")

	// Flag completion
	annotation := make(map[string][]string)
	annotation[cobra.BashCompCustom] = []string{"__madonctl_visibility"}

	statusPostSubcommand.Flags().Lookup("visibility").Annotations = annotation

	// This one will be used to check if the options were explicitly set or not
	statusPostFlags = statusPostSubcommand.Flags()
}

// statusCmd represents the status command
// This command does nothing without a subcommand
var statusCmd = &cobra.Command{
	Use:     "status --status-id ID subcommand",
	Aliases: []string{"st"},
	Short:   "Get status details",
	//Long:    `TBW...`, // TODO
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This is common to status and all status subcommands but "post"
		if statusOpts.statusID < 1 && cmd.Name() != "post" {
			return errors.New("missing status ID")
		}
		return madonInit(true)
	},
}

var statusSubcommands = []*cobra.Command{
	&cobra.Command{
		Use:     "show",
		Aliases: []string{"display"},
		Short:   "Get the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "context",
		Short: "Get the status context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "card",
		Short: "Get the status card",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "reblogged-by",
		Short: "Display accounts which reblogged the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "favourited-by",
		Aliases: []string{"favorited-by"},
		Short:   "Display accounts which favourited the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "mute-conversation",
		Aliases: []string{"mute"},
		Short:   "Mute the conversation containing the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "unmute-conversation",
		Aliases: []string{"unmute"},
		Short:   "Unmute the conversation containing the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	statusReblogSubcommand,
	statusUnreblogSubcommand,
	statusFavouriteSubcommand,
	statusUnfavouriteSubcommand,
	statusPinSubcommand,
	statusUnpinSubcommand,
	statusPostSubcommand,
}

var statusReblogSubcommand = &cobra.Command{
	Use:     "boost",
	Aliases: []string{"reblog"},
	Short:   "Boost (reblog) a status message",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusUnreblogSubcommand = &cobra.Command{
	Use:     "unboost",
	Aliases: []string{"unreblog"},
	Short:   "Cancel boost (reblog) of a status message",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusFavouriteSubcommand = &cobra.Command{
	Use:     "favourite",
	Aliases: []string{"favorite", "fave"},
	Short:   "Mark the status as favourite",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusUnfavouriteSubcommand = &cobra.Command{
	Use:     "unfavourite",
	Aliases: []string{"unfavorite", "unfave"},
	Short:   "Unmark the status as favourite",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusPinSubcommand = &cobra.Command{
	Use:   "pin",
	Short: "Pin a status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusUnpinSubcommand = &cobra.Command{
	Use:   "unpin",
	Short: "Unpin a status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusPostSubcommand = &cobra.Command{
	Use:     "post",
	Aliases: []string{"toot", "pouet"},
	Short:   "Post a message (same as 'madonctl toot')",
	Example: `  madonctl status post "Hello, World"
  madonctl status post --spoiler Warning "Spoiled"
  madonctl status toot --visibility private "To my followers only"
  madonctl status toot --sensitive --file image.jpg Image
  madonctl status post --media-ids ID1,ID2,ID3 Image
  madonctl status toot --text-file message.txt
  madonctl status post --in-reply-to STATUSID "@user response"
  madonctl status post --in-reply-to STATUSID --add-mentions "response"
  echo "Hello from #madonctl" | madonctl status toot --stdin

The default visibility can be set in the configuration file with the option
'default_visibility' (or with an environmnent variable).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

func statusSubcommandRunE(subcmd string, args []string) error {
	opt := statusOpts

	var obj interface{}
	var err error

	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 /* || opt.sinceID > 0 || opt.maxID > 0 */ {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
	}

	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}
	/*
		if opt.maxID > 0 {
			limOpts.MaxID = int64(opt.maxID)
		}
		if opt.sinceID > 0 {
			limOpts.SinceID = int64(opt.sinceID)
		}
	*/

	switch subcmd {
	case "show":
		var status *madon.Status
		status, err = gClient.GetStatus(opt.statusID)
		obj = status
	case "context":
		var context *madon.Context
		context, err = gClient.GetStatusContext(opt.statusID)
		obj = context
	case "card":
		var context *madon.Card
		context, err = gClient.GetStatusCard(opt.statusID)
		obj = context
	case "reblogged-by":
		var accountList []madon.Account
		accountList, err = gClient.GetStatusRebloggedBy(opt.statusID, limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "favourited-by":
		var accountList []madon.Account
		accountList, err = gClient.GetStatusFavouritedBy(opt.statusID, limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "delete":
		err = gClient.DeleteStatus(opt.statusID)
	case "boost", "unboost":
		if opt.unset || subcmd == "unboost" {
			err = gClient.UnreblogStatus(opt.statusID)
		} else {
			err = gClient.ReblogStatus(opt.statusID)
		}
	case "favourite", "unfavourite":
		if opt.unset || subcmd == "unfavourite" {
			err = gClient.UnfavouriteStatus(opt.statusID)
		} else {
			err = gClient.FavouriteStatus(opt.statusID)
		}
	case "pin", "unpin":
		if opt.unset || subcmd == "unpin" {
			err = gClient.UnpinStatus(opt.statusID)
		} else {
			err = gClient.PinStatus(opt.statusID)
		}
	case "mute-conversation":
		var s *madon.Status
		s, err = gClient.MuteConversation(opt.statusID)
		obj = s
	case "unmute-conversation":
		var s *madon.Status
		s, err = gClient.UnmuteConversation(opt.statusID)
		obj = s
	case "post": // toot
		var s *madon.Status
		text := strings.Join(args, " ")
		if opt.textFilePath != "" {
			var b []byte
			if b, err = ioutil.ReadFile(opt.textFilePath); err != nil {
				break
			}
			text = string(b)
		} else if opt.stdin {
			var b []byte
			if b, err = ioutil.ReadAll(os.Stdin); err != nil {
				break
			}
			text = string(b)
		}
		s, err = toot(text)
		obj = s
	default:
		return errors.New("statusSubcommand: internal error")
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return p.printObj(obj)
}
