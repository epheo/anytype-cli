package cmd

import (
	"fmt"
	"os"

	"github.com/epheo/anytype-cli/internal/auth"
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Anytype",
	Long: `Authenticate with Anytype to obtain a session token and app key.
	
This command will initiate an authentication flow that requires you to enter
a verification code shown in your Anytype application.
	
Example:
  anytype-cli auth
  anytype-cli auth --base-url http://localhost:31009`,
	Run: func(cmd *cobra.Command, args []string) {
		if auth.IsAuthenticated(cfg) && !forceAuth {
			fmt.Println("You are already authenticated.")
			fmt.Println("To force re-authentication, use the --force flag.")
			return
		}

		newConfig, err := auth.RunAuthentication(cfg.BaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
			os.Exit(1)
		}

		// Update and save the config
		cfg.AppKey = newConfig.AppKey
		cfg.SessionToken = newConfig.SessionToken
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save credentials: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Authentication successful. Credentials saved.")
	},
}

var forceAuth bool

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().BoolVar(&forceAuth, "force", false, "Force re-authentication even if credentials exist")
}
