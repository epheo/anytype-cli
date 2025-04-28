package cmd

import (
	"fmt"

	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

// CLI version information
const (
	AppVersion = "0.1.0"
	AppName    = "anytype-cli"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display version information for the CLI and the Anytype SDK it's using.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get SDK version info
		sdkVersion := anytype.GetVersionInfo()

		fmt.Printf("%s version: %s\n", AppName, AppVersion)
		fmt.Printf("Anytype SDK version: %s\n", sdkVersion.Version)
		fmt.Printf("Anytype API version: %s\n", sdkVersion.APIVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
