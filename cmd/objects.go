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

// objectsCmd represents the objects command
var objectsCmd = &cobra.Command{
	Use:   "objects",
	Short: "Manage Anytype objects",
	Long:  `Create, read, update, and delete Anytype objects.`,
}

// objectsListCmd represents the objects list command
var objectsListCmd = &cobra.Command{
	Use:   "list [spaceID]",
	Short: "List objects in a space",
	Long:  `List all objects available in the specified space.`,
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
		objects, err := anytypeClient.Space(spaceID).Objects().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list objects: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(objects, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(objects)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Table format with dynamic column widths
			table := output.NewTable([]string{"OBJECT ID", "NAME", "TYPE", "LAYOUT"})
			// Don't truncate OBJECT ID as it's used for command line arguments
			table.SetColumnWidth(1, 30)
			table.SetColumnTruncate(1, true) // NAME column
			table.SetColumnWidth(2, 20)
			table.SetColumnTruncate(2, true) // TYPE column
			table.SetColumnWidth(3, 20)
			table.SetColumnTruncate(3, true) // LAYOUT column

			for _, obj := range objects {
				table.AddRow([]string{obj.ID, obj.Name, obj.TypeKey, obj.Layout})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal objects: %d\n", len(objects))
		}
	},
}

// objectsGetCmd represents the objects get command
var objectsGetCmd = &cobra.Command{
	Use:   "get [spaceID] [objectID]",
	Short: "Get details of a specific object",
	Long:  `Retrieve detailed information about a specific Anytype object.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		objectID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Object(objectID).Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get object: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Object, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Object)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Detailed output
			obj := resp.Object
			fmt.Println("OBJECT DETAILS")
			fmt.Println("--------------")
			fmt.Printf("ID: %s\n", obj.ID)
			fmt.Printf("Name: %s\n", obj.Name)
			fmt.Printf("Type: %s\n", obj.TypeKey)
			if obj.Type != nil {
				fmt.Printf("Type Name: %s\n", obj.Type.Name)
			}
			fmt.Printf("Layout: %s\n", obj.Layout)
			fmt.Printf("Space ID: %s\n", obj.SpaceID)
			fmt.Printf("Archived: %v\n", obj.Archived)
			if obj.Icon != nil {
				fmt.Printf("Icon: %s (%s)\n", obj.Icon.Emoji, obj.Icon.Format)
			}

			if len(obj.Properties) > 0 {
				fmt.Println("\nPROPERTIES")
				fmt.Println("----------")
				for _, prop := range obj.Properties {
					fmt.Printf("%s: ", prop.Name)

					switch {
					case prop.Text != "":
						fmt.Printf("%s\n", prop.Text)
					case prop.Number != 0:
						fmt.Printf("%f\n", prop.Number)
					case prop.Select != nil:
						fmt.Printf("%s\n", prop.Select.Name)
					case len(prop.MultiSelect) > 0:
						fmt.Print("[")
						for i, sel := range prop.MultiSelect {
							if i > 0 {
								fmt.Print(", ")
							}
							fmt.Print(sel.Name)
						}
						fmt.Println("]")
					default:
						fmt.Println("[complex type]")
					}
				}
			}
		}
	},
}

// objectsCreateCmd represents the objects create command
var objectsCreateCmd = &cobra.Command{
	Use:   "create [spaceID]",
	Short: "Create a new object",
	Long:  `Create a new object in the specified Anytype space.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		// Validate inputs
		if objectName == "" {
			fmt.Println("Object name is required")
			os.Exit(1)
		}
		if objectTypeKey == "" {
			fmt.Println("Type key is required. Use 'ot-page' for a basic page.")
			os.Exit(1)
		}

		spaceID := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)

		var icon *anytype.Icon
		if objectIcon != "" {
			icon = &anytype.Icon{
				Format: anytype.IconFormatEmoji,
				Emoji:  objectIcon,
			}
		}

		createReq := anytype.CreateObjectRequest{
			TypeKey:     objectTypeKey,
			Name:        objectName,
			Description: objectDesc,
			Body:        objectBody,
			Icon:        icon,
		}

		if objectTemplateID != "" {
			createReq.TemplateID = objectTemplateID
		}

		resp, err := anytypeClient.Space(spaceID).Objects().Create(ctx, createReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create object: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Object, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Object)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			fmt.Println("Object created successfully:")
			fmt.Printf("ID: %s\n", resp.Object.ID)
			fmt.Printf("Name: %s\n", resp.Object.Name)
			fmt.Printf("Type: %s\n", resp.Object.TypeKey)
		}
	},
}

// objectsDeleteCmd represents the objects delete command
var objectsDeleteCmd = &cobra.Command{
	Use:   "delete [spaceID] [objectID]",
	Short: "Delete an object",
	Long:  `Delete an Anytype object from the specified space.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		objectID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Object(objectID).Delete(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete object: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Object, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Object)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			fmt.Printf("Object '%s' (ID: %s) deleted successfully.\n", resp.Object.Name, resp.Object.ID)
			fmt.Printf("Archive status: %v\n", resp.Object.Archived)
		}
	},
}

// objectsExportCmd represents the objects export command
var objectsExportCmd = &cobra.Command{
	Use:   "export [spaceID] [objectID]",
	Short: "Export an object",
	Long:  `Export an Anytype object in markdown format.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceID := args[0]
		objectID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Object(objectID).Export(ctx, "markdown")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export object: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		default:
			fmt.Println(resp.Markdown)
		}
	},
}

var (
	objectName       string
	objectTypeKey    string
	objectDesc       string
	objectIcon       string
	objectBody       string
	objectTemplateID string
)

func init() {
	rootCmd.AddCommand(objectsCmd)
	objectsCmd.AddCommand(objectsListCmd)
	objectsCmd.AddCommand(objectsGetCmd)
	objectsCmd.AddCommand(objectsCreateCmd)
	objectsCmd.AddCommand(objectsDeleteCmd)
	objectsCmd.AddCommand(objectsExportCmd)

	// Flags for create command
	objectsCreateCmd.Flags().StringVar(&objectName, "name", "", "Name for the new object (required)")
	objectsCreateCmd.Flags().StringVar(&objectTypeKey, "type", "ot-page", "Type key for the object (default: ot-page)")
	objectsCreateCmd.Flags().StringVar(&objectDesc, "description", "", "Description for the new object")
	objectsCreateCmd.Flags().StringVar(&objectIcon, "icon", "", "Emoji icon for the object (e.g. '📄')")
	objectsCreateCmd.Flags().StringVar(&objectBody, "body", "", "Markdown body content for the object")
	objectsCreateCmd.Flags().StringVar(&objectTemplateID, "template", "", "Template ID to use for creating the object")
	objectsCreateCmd.MarkFlagRequired("name")
}
