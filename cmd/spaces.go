package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/epheo/anytype-cli/internal/auth"
	"github.com/epheo/anytype-cli/internal/client"
	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// spacesCmd represents the spaces command
var spacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Manage Anytype spaces",
	Long:  `List, create, and manage Anytype spaces.`,
}

// spacesListCmd represents the spaces list command
var spacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all spaces",
	Long:  `List all spaces accessible to the authenticated user.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Spaces().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list spaces: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Data, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Table format
			fmt.Println("SPACE ID                             NAME                  DESCRIPTION")
			fmt.Println("-----------------------------------  --------------------  --------------------")
			for _, space := range resp.Data {
				fmt.Printf("%-35s  %-20s  %s\n", space.ID, output.Truncate(space.Name, 20), output.Truncate(space.Description, 30))
			}
			fmt.Printf("\nTotal spaces: %d\n", len(resp.Data))
		}
	},
}

// spacesCreateCmd represents the spaces create command
var spacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new space",
	Long:  `Create a new Anytype space with the specified name and description.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		// Validate inputs
		if spaceName == "" {
			fmt.Println("Space name is required")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)

		var icon *anytype.Icon
		if spaceIcon != "" {
			icon = &anytype.Icon{
				Format: anytype.IconFormatEmoji,
				Emoji:  spaceIcon,
			}
		}

		createReq := anytype.CreateSpaceRequest{
			Name:        spaceName,
			Description: spaceDesc,
			Icon:        icon,
		}

		resp, err := anytypeClient.Spaces().Create(ctx, createReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create space: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Space, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Space)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			fmt.Println("Space created successfully:")
			fmt.Printf("ID: %s\n", resp.Space.ID)
			fmt.Printf("Name: %s\n", resp.Space.Name)
			fmt.Printf("Description: %s\n", resp.Space.Description)
		}
	},
}

// spacesGetCmd represents the spaces get command
var spacesGetCmd = &cobra.Command{
	Use:   "get [spaceID]",
	Short: "Get details of a specific space",
	Long:  `Retrieve detailed information about a specific Anytype space.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get space: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Space, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Space)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Detailed output
			space := resp.Space
			fmt.Println("SPACE DETAILS")
			fmt.Println("------------")
			fmt.Printf("ID: %s\n", space.ID)
			fmt.Printf("Name: %s\n", space.Name)
			fmt.Printf("Description: %s\n", space.Description)
			fmt.Printf("Home Object ID: %s\n", space.HomeID)
			fmt.Printf("Archive ID: %s\n", space.ArchiveID)
			fmt.Printf("Profile ID: %s\n", space.ProfileID)
			fmt.Printf("Created At: %s\n", formatTime(space.CreatedAt))
			fmt.Printf("Last Opened At: %s\n", formatTime(space.LastOpenedAt))
			if space.Icon != nil {
				fmt.Printf("Icon: %s (%s)\n", space.Icon.Emoji, space.Icon.Format)
			}
		}
	},
}

var (
	spaceName string
	spaceDesc string
	spaceIcon string
)

func init() {
	rootCmd.AddCommand(spacesCmd)
	spacesCmd.AddCommand(spacesListCmd)
	spacesCmd.AddCommand(spacesCreateCmd)
	spacesCmd.AddCommand(spacesGetCmd)

	// Flags for create command
	spacesCreateCmd.Flags().StringVar(&spaceName, "name", "", "Name for the new space (required)")
	spacesCreateCmd.Flags().StringVar(&spaceDesc, "description", "", "Description for the new space")
	spacesCreateCmd.Flags().StringVar(&spaceIcon, "icon", "", "Emoji icon for the space (e.g. '🚀')")
	spacesCreateCmd.MarkFlagRequired("name")
}

// Helper functions

// formatTime is a wrapper around output.FormatTime for backward compatibility
func formatTime(unixTime int64) string {
	return output.FormatTime(unixTime)
}
