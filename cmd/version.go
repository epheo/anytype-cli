package cmd

import (
	"fmt"

	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

// version is set at build time with -ldflags "-X github.com/epheo/anytype-cli/cmd.version=..."
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI, SDK, and API versions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		sdk := anytype.GetVersionInfo()
		fmt.Printf("anytype-cli %s\nanytype-go  %s\napi         %s\n", version, sdk.Version, sdk.APIVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
