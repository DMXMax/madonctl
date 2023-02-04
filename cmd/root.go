// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/McKael/madon/v3"
)

// AppName is the CLI application name
const AppName = "madonctl"

// AppWebsite is the application website URL
const AppWebsite = "https://github.com/McKael/madonctl"

// defaultConfigFile is the path to the default configuration file
const defaultConfigFile = "$HOME/.config/" + AppName + "/" + AppName + ".yaml"

// Madon API client
var gClient *madon.Client

// Options
var cfgFile string
var safeMode bool
var instanceURL, appID, appSecret string
var login, password, token string
var verbose bool
var outputFormat string
var outputTemplate, outputTemplateFile, outputTheme string
var colorMode string

// Shell completion functions
const shellComplFunc = `
__madonctl_visibility() {
	COMPREPLY=( direct private unlisted public )
}
__madonctl_output() {
	COMPREPLY=( plain json yaml template theme )
}
__madonctl_color() {
	COMPREPLY=( auto on off )
}
__madonctl_theme() {
	local madonctl_output out
	# This doesn't handle spaces or special chars...
	if out=$(madonctl config themes 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
	fi
}
`

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               AppName,
	Short:             "A CLI utility for Mastodon API",
	PersistentPreRunE: checkOutputFormat,
	Long: `madonctl is a CLI tool for the Mastodon REST API.

You can use a configuration file to store common options.
For example, create ` + defaultConfigFile + ` with the following
contents:

	---
	instance: "INSTANCE"
	login: "USERNAME"
	password: "USERPASSWORD"
	...

The simplest way to generate a configuration file is to use the 'config dump'
command.

(Configuration files in JSON are also accepted.)

If you want shell auto-completion (for bash or zsh), you can generate the
completion scripts with "madonctl completion $SHELL".
For example if you use bash:

	madonctl completion bash > _bash_madonctl
	source _bash_madonctl

Now you should have tab completion for subcommands and flags.

Note: Most examples assume the user's credentials are set in the configuration
file.
`,
	Example: `  madonctl instance
  madonctl toot "Hello, World"
  madonctl toot --visibility direct "@McKael Hello, You"
  madonctl toot --visibility private --spoiler CW "The answer was 42"
  madonctl post --file image.jpg Selfie
  madonctl --instance INSTANCE --login USERNAME --password PASS timeline
  madonctl account notifications --list --clear
  madonctl account blocked
  madonctl account search Gargron
  madonctl search --resolve https://mastodon.social/@Gargron
  madonctl account follow 37
  madonctl account follow Gargron@mastodon.social
  madonctl account follow https://mastodon.social/@Gargron
  madonctl account --account-id 399 statuses
  madonctl status --status-id 416671 show
  madonctl status --status-id 416671 favourite
  madonctl status --status-id 416671 boost
  madonctl account show
  madonctl account show Gargron@mastodon.social
  madonctl account show -o yaml
  madonctl account --account-id 1 followers --template '{{.acct}}{{"\n"}}'
  madonctl config whoami
  madonctl timeline :mastodon`,
	BashCompletionFunction: shellComplFunc,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is "+defaultConfigFile+")")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	RootCmd.PersistentFlags().StringVarP(&instanceURL, "instance", "i", "", "Mastodon instance")
	RootCmd.PersistentFlags().StringVarP(&login, "login", "L", "", "Instance user login")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "P", "", "Instance user password")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "User token")
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "",
		"Output format (plain|json|yaml|template|theme)")
	RootCmd.PersistentFlags().StringVar(&outputTemplate, "template", "",
		"Go template (for output=template)")
	RootCmd.PersistentFlags().StringVar(&outputTemplateFile, "template-file", "",
		"Go template file (for output=template)")
	RootCmd.PersistentFlags().StringVar(&outputTheme, "theme", "",
		"Theme name (for output=theme)")
	RootCmd.PersistentFlags().StringVar(&colorMode, "color", "",
		"Color mode (auto|on|off; for output=template)")

	// Configuration file bindings
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("instance", RootCmd.PersistentFlags().Lookup("instance"))
	viper.BindPFlag("login", RootCmd.PersistentFlags().Lookup("login"))
	viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("color", RootCmd.PersistentFlags().Lookup("color"))

	// Flag completion
	annotationOutput := make(map[string][]string)
	annotationOutput[cobra.BashCompCustom] = []string{"__madonctl_output"}
	annotationColor := make(map[string][]string)
	annotationColor[cobra.BashCompCustom] = []string{"__madonctl_color"}
	annotationTheme := make(map[string][]string)
	annotationTheme[cobra.BashCompCustom] = []string{"__madonctl_theme"}

	RootCmd.PersistentFlags().Lookup("output").Annotations = annotationOutput
	RootCmd.PersistentFlags().Lookup("color").Annotations = annotationColor
	RootCmd.PersistentFlags().Lookup("theme").Annotations = annotationTheme
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "/dev/null" {
		return
	}

	viper.SetConfigName(AppName) // name of config file (without extension)
	viper.AddConfigPath("$HOME/.config/" + AppName)
	viper.AddConfigPath("$HOME/." + AppName)

	// Read in environment variables that match, with a prefix
	viper.SetEnvPrefix(AppName)
	viper.AutomaticEnv()

	// Enable ability to specify config file via flag
	viper.SetConfigFile(cfgFile)

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		if cfgFile != "" {
			errPrint("Error: cannot read configuration file '%s': %v", cfgFile, err)
			os.Exit(-1)
		}
	} else if viper.GetBool("verbose") {
		errPrint("Using config file: %s", viper.ConfigFileUsed())
	}
}
