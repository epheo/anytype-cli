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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// membersCmd represents the members command
var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage space members",
	Long:  `List and get information about members in an Anytype space.`,
}

// membersListCmd represents the members list command
var membersListCmd = &cobra.Command{
	Use:   "list [spaceID]",
	Short: "List all members in a space",
	Long:  `List all members in the specified Anytype space.`,
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
		resp, err := anytypeClient.Space(spaceID).Members().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list members: %v\n", err)
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
			// Table format with dynamic column widths
			table := output.NewTable([]string{"MEMBER ID", "NAME", "ROLE", "STATUS"})
			for _, member := range resp.Data {
				table.AddRow([]string{member.ID, member.Name, member.Role, member.Status})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal members: %d\n", len(resp.Data))
		}
	},
}

// membersGetCmd represents the members get command
var membersGetCmd = &cobra.Command{
	Use:   "get [spaceID] [memberID]",
	Short: "Get details of a specific member",
	Long:  `Retrieve detailed information about a specific member in an Anytype space.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		memberID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Member(memberID).Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get member details: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Member, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Member)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Detailed output
			member := resp.Member
			fmt.Println("MEMBER DETAILS")
			fmt.Println("--------------")
			fmt.Printf("ID: %s\n", member.ID)
			fmt.Printf("Name: %s\n", member.Name)
			fmt.Printf("Global Name: %s\n", member.GlobalName)
			fmt.Printf("Identity: %s\n", member.Identity)
			fmt.Printf("Role: %s\n", member.Role)
			fmt.Printf("Status: %s\n", member.Status)
			if member.Icon != nil {
				if member.Icon.Format == "emoji" {
					fmt.Printf("Icon: %s\n", member.Icon.Emoji)
				} else {
					fmt.Printf("Icon: %s (%s)\n", member.Icon.Name, member.Icon.Format)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(membersCmd)
	membersCmd.AddCommand(membersListCmd)
	membersCmd.AddCommand(membersGetCmd)
}
