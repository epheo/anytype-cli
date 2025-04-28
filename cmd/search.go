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
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for objects",
	Long:  `Search for objects in Anytype spaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated(cfg) {
			fmt.Println("You are not authenticated. Please run 'anytype-cli auth' first.")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)

		var searchReq anytype.SearchRequest
		searchReq.Query = searchQuery

		if len(searchTypes) > 0 {
			searchReq.Types = searchTypes
		}

		if searchSortProperty != "" {
			searchReq.Sort = &anytype.SortOptions{
				Property:  anytype.SortProperty(searchSortProperty),
				Direction: anytype.SortDirection(searchSortDirection),
			}
		}

		var resp *anytype.SearchResponse
		var err error

		if searchSpaceID != "" {
			// Search within a specific space
			// Resolve space ID if it's a name
			spaceID, spaceErr := spaces.ResolveSpace(cfg, searchSpaceID)
			if spaceErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to resolve space: %v\n", spaceErr)
				os.Exit(1)
			}
			resp, err = anytypeClient.Space(spaceID).Search(ctx, searchReq)
		} else {
			// Global search across all spaces
			resp, err = anytypeClient.Search().Search(ctx, searchReq)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to search: %v\n", err)
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
			table := output.NewTable([]string{"OBJECT ID", "NAME", "TYPE", "SPACE ID"})
			// Don't truncate ID columns (0 and 3) as they're used for command line arguments
			table.SetColumnWidth(1, 30)
			table.SetColumnTruncate(1, true) // NAME column
			table.SetColumnWidth(2, 20)
			table.SetColumnTruncate(2, true) // TYPE column

			for _, obj := range resp.Data {
				table.AddRow([]string{obj.ID, obj.Name, obj.TypeKey, obj.SpaceID})
			}
			fmt.Print(table.String())
			fmt.Printf("\nTotal results: %d\n", len(resp.Data))

			// Print search details
			fmt.Printf("\nSearch details:\n")
			fmt.Printf("  Query: '%s'\n", searchQuery)
			if len(searchTypes) > 0 {
				fmt.Printf("  Types: %v\n", searchTypes)
			}
			if searchSortProperty != "" {
				fmt.Printf("  Sorted by: %s (%s)\n", searchSortProperty, searchSortDirection)
			}
			if searchSpaceID != "" {
				fmt.Printf("  Limited to space: %s\n", searchSpaceID)
			} else {
				fmt.Printf("  Searched across all spaces\n")
			}
		}
	},
}

var (
	searchQuery         string
	searchTypes         []string
	searchSortProperty  string
	searchSortDirection string
	searchSpaceID       string
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchQuery, "query", "", "Search query string")
	searchCmd.Flags().StringSliceVar(&searchTypes, "types", []string{}, "Filter by object types (comma-separated, e.g. 'ot-page,ot-note')")
	searchCmd.Flags().StringVar(&searchSortProperty, "sort", "", "Property to sort results by (created_date, last_modified_date, last_opened_date, name)")
	searchCmd.Flags().StringVar(&searchSortDirection, "direction", "desc", "Sort direction (asc or desc)")
	searchCmd.Flags().StringVar(&searchSpaceID, "space", "", "Limit search to this space (can be either ID or name, default: search all spaces)")

	// Set up completion functions after config is loaded
	cobra.OnInitialize(func() {
		if cfg != nil {
			spaceCompletion := spaces.GetSpaceCompletionFunc(cfg)
			searchCmd.RegisterFlagCompletionFunc("space", spaceCompletion)
		}
	})
}
