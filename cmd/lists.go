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

// listsCmd represents the lists command
var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage lists and views",
	Long:  `Interact with lists and views in Anytype spaces.`,
}

// listsViewsCmd represents the lists views command
var listsViewsCmd = &cobra.Command{
	Use:   "views [spaceID] [listID]",
	Short: "List views for a list",
	Long:  `List all available views for the specified list in an Anytype space.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		listID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).List(listID).Views().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list views: %v\n", err)
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
			table := output.NewTable([]string{"VIEW ID", "NAME", "LAYOUT"})
			for _, view := range resp.Data {
				table.AddRow([]string{view.ID, view.Name, view.Layout})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal views: %d\n", len(resp.Data))
			if resp.Pagination.HasMore {
				fmt.Printf("Has more views (Total: %d, Retrieved: %d)\n",
					resp.Pagination.Total,
					len(resp.Data))
			}
		}
	},
}

// listsObjectsCmd represents the lists objects command
var listsObjectsCmd = &cobra.Command{
	Use:   "objects [spaceID] [listID] [viewID]",
	Short: "List objects in a view",
	Long:  `List all objects in a specific view of a list in an Anytype space.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		listID := args[1]
		viewID := args[2]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).List(listID).View(viewID).Objects().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list objects in view: %v\n", err)
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
			table := output.NewTable([]string{"OBJECT ID", "NAME", "TYPE"})
			for _, obj := range resp.Data {
				table.AddRow([]string{obj.ID, obj.Name, obj.TypeKey})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal objects: %d\n", len(resp.Data))
			if resp.Pagination.HasMore {
				fmt.Printf("Has more objects (Total: %d, Retrieved: %d)\n",
					resp.Pagination.Total,
					len(resp.Data))
			}
		}
	},
}

// listsAddCmd represents the lists add command
var listsAddCmd = &cobra.Command{
	Use:   "add [spaceID] [listID] [objectIDs...]",
	Short: "Add objects to a list",
	Long:  `Add one or more objects to a list in an Anytype space.`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		listID := args[1]
		objectIDs := args[2:]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		err := anytypeClient.Space(spaceID).List(listID).Objects().Add(ctx, objectIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add objects to list: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added %d object(s) to list %s\n", len(objectIDs), listID)
		for i, id := range objectIDs {
			fmt.Printf("  %d. %s\n", i+1, id)
		}
	},
}

// listsRemoveCmd represents the lists remove command
var listsRemoveCmd = &cobra.Command{
	Use:   "remove [spaceID] [listID] [objectID]",
	Short: "Remove an object from a list",
	Long:  `Remove an object from a list in an Anytype space.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		listID := args[1]
		objectID := args[2]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		err := anytypeClient.Space(spaceID).List(listID).Object(objectID).Remove(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove object from list: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully removed object %s from list %s\n", objectID, listID)
	},
}

func init() {
	rootCmd.AddCommand(listsCmd)
	listsCmd.AddCommand(listsViewsCmd)
	listsCmd.AddCommand(listsObjectsCmd)
	listsCmd.AddCommand(listsAddCmd)
	listsCmd.AddCommand(listsRemoveCmd)
}
