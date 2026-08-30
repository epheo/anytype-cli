package cmd

import (
	"io"
	"strconv"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Browse templates of a type in the current space",
}

var templatesListCmd = &cobra.Command{
	Use:   "list <type>",
	Short: "List templates of a type",
	Args:  cobra.ExactArgs(1),
	RunE: listRun(
		func(s *session, args []string) (lister[anytype.Object], error) {
			sc, err := s.space()
			if err != nil {
				return nil, err
			}
			ty, err := resolveType(s, sc, args[0])
			if err != nil {
				return nil, err
			}
			return sc.Type(ty.ID).Templates(), nil
		},
		func(items []anytype.Object) *output.Table {
			t := output.NewTable("ID", "NAME", "ARCHIVED")
			for _, o := range items {
				t.Row(o.ID, o.Name, strconv.FormatBool(o.Archived))
			}
			return t
		}),
}

var templatesGetCmd = &cobra.Command{
	Use:   "get <type> <template-id>",
	Short: "Show one template",
	Args:  cobra.ExactArgs(2),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		ty, err := resolveType(s, sc, args[0])
		if err != nil {
			return err
		}
		resp, err := sc.Type(ty.ID).Template(args[1]).Get(s.ctx)
		if err != nil {
			return err
		}
		return s.out.Print(resp.Template, func(w io.Writer) { writeObject(w, &resp.Template) })
	}),
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(templatesListCmd, templatesGetCmd)
	addListFlags(templatesListCmd)
}
