package cmd

import (
	"fmt"
	"os"

	"github.com/epheo/anytype-cli/internal/auth"
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfg          *config.Config
	cfgFile      string
	baseURL      string
	verbose      bool
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "anytype-cli",
	Short: "A comprehensive CLI for interacting with Anytype",
	Long: `anytype-cli is a command line tool for interacting with Anytype
	
This CLI allows you to manage spaces, objects, and perform searches in Anytype,
all from your terminal using the Anytype-Go SDK.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip auth check for these commands
		if cmd.Name() == "auth" || cmd.Name() == "version" || cmd.Name() == "help" {
			return
		}

		// Parent command check - if this is a parent command, skip the auth check
		// as the actual subcommand will do the check
		if cmd.HasSubCommands() && len(args) == 0 {
			return
		}

		// Check if authenticated (except for auth command)
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("Error: You are not authenticated. Run 'anytype-cli auth' first.")
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.anytype-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "Anytype API base URL (default is http://localhost:31009)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table",
		fmt.Sprintf("output format (%s, %s, %s)", output.FormatTable, output.FormatJSON, output.FormatYAML))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	var err error

	// Load config from file or create a default one
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config from command-line flags
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
}
