package cmd

import (
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "Inspect members of the current space",
}

var membersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List members",
	Args:  cobra.NoArgs,
	RunE: listInSpace(
		func(sc anytype.SpaceContext, _ []string) lister[anytype.Member] { return sc.Members() },
		func(items []anytype.Member) *output.Table {
			t := output.NewTable("ID", "NAME", "ROLE", "STATUS")
			for _, m := range items {
				t.Row(m.ID, m.Name, string(m.Role), string(m.Status))
			}
			return t
		}),
}

var membersGetCmd = &cobra.Command{
	Use:   "get <member-id>",
	Short: "Show one member",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Member(args[0]).Get(s.ctx)
		if err != nil {
			return err
		}
		m := resp.Member
		return s.out.Print(m, func(w io.Writer) {
			output.NewDetails().
				Add("ID", m.ID).
				Add("Name", m.Name).
				AddIf("Global Name", m.GlobalName).
				Add("Identity", m.Identity).
				Add("Role", string(m.Role)).
				Add("Status", string(m.Status)).
				AddIf("Icon", output.Icon(m.Icon)).
				Write(w)
		})
	}),
}

func init() {
	rootCmd.AddCommand(membersCmd)
	membersCmd.AddCommand(membersListCmd, membersGetCmd)
	addListFlags(membersListCmd)
}
