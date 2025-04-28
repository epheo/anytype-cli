package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script for your shell",
	Long: `Generate shell completion script for anytype-cli.

To load completions:

Bash:
  $ source <(anytype-cli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ anytype-cli completion bash > /etc/bash_completion.d/anytype-cli
  # macOS:
  $ anytype-cli completion bash > /usr/local/etc/bash_completion.d/anytype-cli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ anytype-cli completion zsh > "${fpath[1]}/_anytype-cli"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ anytype-cli completion fish | source

  # To load completions for each session, execute once:
  $ anytype-cli completion fish > ~/.config/fish/completions/anytype-cli.fish

PowerShell:
  PS> anytype-cli completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> anytype-cli completion powershell > anytype-cli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
