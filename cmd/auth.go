package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/epheo/anytype-cli/internal/auth"
	"github.com/epheo/anytype-cli/internal/client"
	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Log in to the local Anytype app",
	Long: `Request a verification code from the Anytype app, then store the
resulting API key in the config file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if cfg.Authenticated() && !force {
			fmt.Println("Already authenticated. Use --force to log in again.")
			return nil
		}

		// Typing the code takes longer than an API call; keep --timeout only when set explicitly.
		timeout := 5 * time.Minute
		if cmd.Flags().Changed("timeout") {
			timeout = flagTimeout
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		key, err := auth.Login(ctx, cfg.BaseURL, os.Stdin, os.Stdout)
		if err != nil {
			return err
		}
		cfg.AppKey = key
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save credentials: %w", err)
		}
		fmt.Printf("Credentials saved to %s\n", cfg.Path())
		return nil
	},
}

type authStatus struct {
	Config        string `json:"config" yaml:"config"`
	BaseURL       string `json:"base_url" yaml:"base_url"`
	Authenticated bool   `json:"authenticated" yaml:"authenticated"`
	KeyValid      bool   `json:"key_valid" yaml:"key_valid"`
	Error         string `json:"error,omitempty" yaml:"error,omitempty"`
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show where credentials come from and whether the key works",
	Args:  cobra.NoArgs,
	RunE: run(func(s *session, _ []string) error {
		st := authStatus{Config: s.cfg.Path(), BaseURL: s.cfg.BaseURL, Authenticated: s.cfg.Authenticated()}
		if st.Authenticated {
			_, err := client.New(s.cfg).Spaces().List(s.ctx, anytype.WithLimit(1))
			st.KeyValid = err == nil
			if err != nil {
				st.Error = describe(err)
			}
		}
		return s.out.Print(st, func(w io.Writer) {
			output.NewDetails().
				Add("Config", st.Config).
				Add("Base URL", st.BaseURL).
				Add("Authenticated", fmt.Sprint(st.Authenticated)).
				Add("Key valid", fmt.Sprint(st.KeyValid)).
				AddIf("Error", st.Error).
				Write(w)
		})
	}),
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the stored API key",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if !cfg.Authenticated() {
			return errors.New("no stored credentials")
		}
		cfg.AppKey = ""
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Credentials removed from %s\n", cfg.Path())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authStatusCmd, authLogoutCmd)
	authCmd.Flags().Bool("force", false, "log in again even if credentials exist")
}
