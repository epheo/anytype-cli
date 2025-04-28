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
	"github.com/epheo/anytype-cli/internal/spaces"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// typesCmd represents the types command
var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "Manage object types",
	Long:  `List and get information about object types in an Anytype space.`,
}

// typesListCmd represents the types list command
var typesListCmd = &cobra.Command{
	Use:   "list [spaceID|spaceName]",
	Short: "List all object types in a space",
	Long:  `List all available object types in the specified Anytype space.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceIdOrName := args[0]
		spaceID, err := spaces.ResolveSpace(cfg, spaceIdOrName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve space: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		types, err := anytypeClient.Space(spaceID).Types().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list object types: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(types, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(types)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Table format with dynamic column widths
			table := output.NewTable([]string{"KEY", "NAME", "LAYOUT", "DESCRIPTION"})
			for _, typ := range types {
				table.AddRow([]string{typ.Key, typ.Name, typ.RecommendedLayout, typ.Description})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal types: %d\n", len(types))
		}
	},
}

// typesGetCmd represents the types get command
var typesGetCmd = &cobra.Command{
	Use:   "get [spaceID|spaceName] [typeID]",
	Short: "Get details of a specific object type",
	Long:  `Retrieve detailed information about a specific object type in an Anytype space.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceIdOrName := args[0]
		spaceID, err := spaces.ResolveSpace(cfg, spaceIdOrName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve space: %v\n", err)
			os.Exit(1)
		}

		typeID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Type(typeID).Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get type details: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Type, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Type)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Detailed output
			typ := resp.Type
			fmt.Println("TYPE DETAILS")
			fmt.Println("------------")
			fmt.Printf("Key: %s\n", typ.Key)
			fmt.Printf("Name: %s\n", typ.Name)
			fmt.Printf("Description: %s\n", typ.Description)
			fmt.Printf("Layout: %s\n", typ.Layout)
			fmt.Printf("Recommended Layout: %s\n", typ.RecommendedLayout)
			fmt.Printf("Is Archived: %v\n", typ.IsArchived)
			fmt.Printf("Is Hidden: %v\n", typ.IsHidden)

			if len(typ.PropertyDefinitions) > 0 {
				fmt.Println("\nPROPERTY DEFINITIONS")
				fmt.Println("-------------------")
				fmt.Println("KEY                    NAME                   FORMAT           REQUIRED")
				fmt.Println("---------------------- ---------------------- ---------------- --------")
				for _, prop := range typ.PropertyDefinitions {
					fmt.Printf("%-20s  %-20s  %-12s  %v\n",
						output.Truncate(prop.Key, 20),
						output.Truncate(prop.Name, 20),
						output.Truncate(prop.Format, 12),
						prop.Required)
				}
			}
		}
	},
}

// templatesListCmd represents the templates list command
var templatesListCmd = &cobra.Command{
	Use:   "templates [spaceID|spaceName] [typeID]",
	Short: "List templates for an object type",
	Long:  `List all available templates for the specified object type in an Anytype space.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceIdOrName := args[0]
		spaceID, err := spaces.ResolveSpace(cfg, spaceIdOrName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve space: %v\n", err)
			os.Exit(1)
		}

		typeID := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		templates, err := anytypeClient.Space(spaceID).Type(typeID).Templates().List(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list templates: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(templates, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(templates)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Table format with dynamic column widths
			table := output.NewTable([]string{"TEMPLATE ID", "NAME", "ARCHIVED"})
			for _, template := range templates {
				table.AddRow([]string{template.ID, template.Name, fmt.Sprintf("%v", template.Archived)})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal templates: %d\n", len(templates))
		}
	},
}

// templatesGetCmd represents the templates get command
var templatesGetCmd = &cobra.Command{
	Use:   "template-get [spaceID|spaceName] [typeID] [templateID]",
	Short: "Get details of a specific template",
	Long:  `Retrieve detailed information about a specific template for an object type.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		spaceIdOrName := args[0]
		spaceID, err := spaces.ResolveSpace(cfg, spaceIdOrName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve space: %v\n", err)
			os.Exit(1)
		}

		typeID := args[1]
		templateID := args[2]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Space(spaceID).Type(typeID).Template(templateID).Get(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get template details: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(resp.Template, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		case "yaml":
			yamlOutput, err := yaml.Marshal(resp.Template)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(yamlOutput))
		default:
			// Detailed output
			template := resp.Template
			fmt.Println("TEMPLATE DETAILS")
			fmt.Println("----------------")
			fmt.Printf("ID: %s\n", template.ID)
			fmt.Printf("Name: %s\n", template.Name)
			fmt.Printf("Archived: %v\n", template.Archived)
			if template.Icon != nil {
				if template.Icon.Format == "emoji" {
					fmt.Printf("Icon: %s\n", template.Icon.Emoji)
				} else {
					fmt.Printf("Icon: %s (%s)\n", template.Icon.Name, template.Icon.Format)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(typesCmd)
	typesCmd.AddCommand(typesListCmd)
	typesCmd.AddCommand(typesGetCmd)
	typesCmd.AddCommand(templatesListCmd)
	typesCmd.AddCommand(templatesGetCmd)
}
