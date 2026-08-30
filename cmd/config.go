package cmd

import (
	"fmt"
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect or change the config file",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file location",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		fmt.Println(cfg.Path())
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show the effective config (key redacted)",
	Args:  cobra.NoArgs,
	RunE: run(func(s *session, _ []string) error {
		view := struct {
			BaseURL      string `json:"base_url" yaml:"base_url"`
			DefaultSpace string `json:"default_space,omitempty" yaml:"default_space,omitempty"`
			HasAppKey    bool   `json:"has_app_key" yaml:"has_app_key"`
		}{s.cfg.BaseURL, s.cfg.DefaultSpace, s.cfg.Authenticated()}
		return s.out.Print(view, func(w io.Writer) {
			output.NewDetails().
				Add("Base URL", view.BaseURL).
				AddIf("Default space", view.DefaultSpace).
				Add("Has app key", fmt.Sprint(view.HasAppKey)).
				Write(w)
		})
	}),
}

var configSetCmd = &cobra.Command{
	Use:       "set <key> <value>",
	Short:     "Set default-space or base-url",
	Args:      cobra.ExactArgs(2),
	ValidArgs: []string{"default-space", "base-url"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		switch args[0] {
		case "default-space":
			cfg.DefaultSpace = args[1]
		case "base-url":
			cfg.BaseURL = args[1]
		default:
			return fmt.Errorf("unknown key %q (default-space, base-url)", args[0])
		}
		return cfg.Save()
	},
}

var configUnsetCmd = &cobra.Command{
	Use:       "unset <key>",
	Short:     "Clear default-space",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"default-space"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if args[0] != "default-space" {
			return fmt.Errorf("unknown key %q (default-space)", args[0])
		}
		cfg.DefaultSpace = ""
		return cfg.Save()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configPathCmd, configGetCmd, configSetCmd, configUnsetCmd)
}
