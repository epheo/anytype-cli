// Package cmd wires cobra commands to the Anytype SDK.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var (
	flagConfig  string
	flagBaseURL string
	flagOutput  string
	flagSpace   string
	flagTimeout time.Duration
)

var errNotAuthenticated = errors.New("not authenticated, run 'anytype-cli auth' first")

var rootCmd = &cobra.Command{
	Use:   "anytype-cli",
	Short: "Command line client for the Anytype local API",
	Long: `anytype-cli manages spaces, objects, types, properties, tags, lists,
templates, and members of a locally running Anytype app.

Commands that work inside a space take it from --space, then ` + config.EnvSpace + `,
then the default_space set with 'anytype-cli config set default-space'.
A space may be given by ID or by name.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := output.ParseFormat(flagOutput); err != nil {
			return err
		}
		if !needsAuth(cmd) {
			return nil
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if !cfg.Authenticated() {
			return errNotAuthenticated
		}
		return nil
	},
}

// needsAuth exempts commands that must work before login, and cobra's
// hidden "__complete" helpers so shell completion never prints errors.
func needsAuth(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "auth", "version", "help", "completion", "config":
			return false
		}
		if strings.HasPrefix(c.Name(), "__") {
			return false
		}
	}
	return true
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", describe(err))
		os.Exit(1)
	}
}

// describe keeps the API's own message but drops the HTTP framing.
func describe(err error) string {
	var apiErr *anytype.APIError
	if errors.As(err, &apiErr) {
		prefix := strings.TrimSuffix(err.Error(), apiErr.Error())
		return fmt.Sprintf("%s%s (HTTP %d)", prefix, apiErr.Message, apiErr.Status)
	}
	return err.Error()
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagConfig, "config", "", "config file (default ~/.anytype-cli/config.yaml)")
	pf.StringVar(&flagBaseURL, "base-url", "", "Anytype API base URL (default "+config.DefaultBaseURL+", env "+config.EnvBaseURL+")")
	pf.StringVarP(&flagOutput, "output", "o", string(output.FormatTable), "output format: table, json, yaml")
	pf.StringVarP(&flagSpace, "space", "s", "", "space ID or name (env "+config.EnvSpace+", or config default_space)")
	pf.DurationVar(&flagTimeout, "timeout", 30*time.Second, "timeout for each API call")
	_ = rootCmd.RegisterFlagCompletionFunc("space", completeSpaceFlag)
}
