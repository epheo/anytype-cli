package cmd

import (
	"context"
	"fmt"
	"iter"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

// searchLister adapts a SearchRequest to the lister interface so search
// shares the pagination flags of every list command.
type searchLister struct {
	req    anytype.SearchRequest
	search func(context.Context, anytype.SearchRequest, ...anytype.ListOption) (*anytype.Page[anytype.Object], error)
	all    func(context.Context, anytype.SearchRequest, ...anytype.ListOption) iter.Seq2[anytype.Object, error]
}

func (l searchLister) List(ctx context.Context, opts ...anytype.ListOption) (*anytype.Page[anytype.Object], error) {
	return l.search(ctx, l.req, opts...)
}

func (l searchLister) All(ctx context.Context, opts ...anytype.ListOption) iter.Seq2[anytype.Object, error] {
	return l.all(ctx, l.req, opts...)
}

var sortProperties = []anytype.SortProperty{
	anytype.SortPropertyCreatedDate, anytype.SortPropertyLastModifiedDate,
	anytype.SortPropertyLastOpenedDate, anytype.SortPropertyName,
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search objects in the current space, or everywhere with --all-spaces",
	Long: `Search objects. Without a current space, or with --all-spaces, the search
spans every space. Filters need a space because property keys belong to one.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		req := anytype.SearchRequest{}
		if len(args) == 1 {
			req.Query = args[0]
		}
		req.Types, _ = f.GetStringSlice("types")
		sort, _ := f.GetString("sort")
		dir, _ := f.GetString("direction")
		if dir != string(anytype.SortDirectionAsc) && dir != string(anytype.SortDirectionDesc) {
			return fmt.Errorf("--direction must be asc or desc")
		}
		if sort != "" {
			req.Sort = &anytype.SortOptions{Property: anytype.SortProperty(sort), Direction: anytype.SortDirection(dir)}
		}
		everywhere, _ := f.GetBool("all-spaces")
		filters, _ := f.GetStringArray("filter")
		match, _ := f.GetString("match")
		var op anytype.FilterOperator
		switch match {
		case "all":
			op = anytype.FilterOperatorAnd
		case "any":
			op = anytype.FilterOperatorOr
		default:
			return fmt.Errorf("--match must be all or any")
		}

		return listRun(
			func(s *session, _ []string) (lister[anytype.Object], error) {
				if everywhere || s.spaceQuery() == "" {
					if len(filters) > 0 {
						return nil, fmt.Errorf("--filter needs a space, drop --all-spaces or set --space")
					}
					return searchLister{req: req, search: s.client.Search().Search, all: s.client.Search().All}, nil
				}
				sc, err := s.space()
				if err != nil {
					return nil, err
				}
				if len(filters) > 0 {
					formats, err := propertyFormats(s, sc)
					if err != nil {
						return nil, err
					}
					items, err := parseFilters(filters, formats)
					if err != nil {
						return nil, err
					}
					req.Filters = &anytype.FilterExpression{Operator: op, Conditions: items}
				}
				return searchLister{req: req, search: sc.Search, all: sc.SearchAll}, nil
			},
			func(items []anytype.Object) *output.Table {
				t := output.NewTable("ID", "NAME", "TYPE", "SPACE ID").MaxWidth(1, 40).MaxWidth(2, 20)
				for _, o := range items {
					t.Row(o.ID, o.Name, typeKey(o.Type), o.SpaceID)
				}
				return t
			})(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	f := searchCmd.Flags()
	f.StringSlice("types", nil, "type keys to include, comma-separated (e.g. page,task)")
	f.String("sort", "", "sort property: "+joinConsts(sortProperties...))
	f.String("direction", string(anytype.SortDirectionDesc), "sort direction: asc, desc")
	f.Bool("all-spaces", false, "search every space even when a current space is set")
	f.StringArray("filter", nil, "property filter as key:condition[:value], repeatable; conditions: "+joinConsts(filterConditions...))
	f.String("match", "all", "combine filters with all (and) or any (or)")
	addListFlags(searchCmd)
}
