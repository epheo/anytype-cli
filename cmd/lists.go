package cmd

import (
	"fmt"
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage collections (lists) and their views in the current space",
}

var listsViewsCmd = &cobra.Command{
	Use:   "views <list-id>",
	Short: "List views of a list",
	Args:  cobra.ExactArgs(1),
	RunE: listInSpace(
		func(sc anytype.SpaceContext, args []string) lister[anytype.ListView] { return sc.List(args[0]).Views() },
		func(items []anytype.ListView) *output.Table {
			t := output.NewTable("ID", "NAME", "LAYOUT", "FILTERS", "SORTS")
			for _, v := range items {
				t.Row(v.ID, v.Name, v.Layout, fmt.Sprint(len(v.Filters)), fmt.Sprint(len(v.Sorts)))
			}
			return t
		}),
}

var listsObjectsCmd = &cobra.Command{
	Use:   "objects <list-id> <view-id>",
	Short: "List objects shown by a view",
	Args:  cobra.ExactArgs(2),
	RunE: listInSpace(
		func(sc anytype.SpaceContext, args []string) lister[anytype.Object] {
			return sc.List(args[0]).View(args[1]).Objects()
		},
		objectTable),
}

var listsAddCmd = &cobra.Command{
	Use:   "add <list-id> <object-id>...",
	Short: "Add objects to a list",
	Args:  cobra.MinimumNArgs(2),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		ids := args[1:]
		if err := sc.List(args[0]).Objects().Add(s.ctx, ids); err != nil {
			return err
		}
		return s.out.Print(ids, func(w io.Writer) {
			fmt.Fprintf(w, "Added %d object(s) to list %s\n", len(ids), args[0])
		})
	}),
}

var listsRemoveCmd = &cobra.Command{
	Use:   "remove <list-id> <object-id>",
	Short: "Remove an object from a list",
	Args:  cobra.ExactArgs(2),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		if err := sc.List(args[0]).Object(args[1]).Remove(s.ctx); err != nil {
			return err
		}
		return s.out.Print(args[1], func(w io.Writer) {
			fmt.Fprintf(w, "Removed object %s from list %s\n", args[1], args[0])
		})
	}),
}

func init() {
	rootCmd.AddCommand(listsCmd)
	listsCmd.AddCommand(listsViewsCmd, listsObjectsCmd, listsAddCmd, listsRemoveCmd)
	addListFlags(listsViewsCmd)
	addListFlags(listsObjectsCmd)
}
