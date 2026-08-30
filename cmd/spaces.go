package cmd

import (
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var spacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Manage spaces",
}

var spacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List spaces",
	Args:  cobra.NoArgs,
	RunE: listRun(
		func(s *session, _ []string) (lister[anytype.Space], error) { return s.client.Spaces(), nil },
		func(items []anytype.Space) *output.Table {
			t := output.NewTable("ID", "NAME", "DESCRIPTION").MaxWidth(1, 30).MaxWidth(2, 40)
			for _, sp := range items {
				t.Row(sp.ID, sp.Name, sp.Description)
			}
			return t
		}),
}

func writeSpace(w io.Writer, sp anytype.Space) {
	output.NewDetails().
		Add("ID", sp.ID).
		Add("Name", sp.Name).
		AddIf("Description", sp.Description).
		AddIf("Icon", output.Icon(sp.Icon)).
		AddIf("Network ID", sp.NetworkID).
		AddIf("Gateway URL", sp.GatewayURL).
		Write(w)
}

// spaceArg prefers the positional argument, else the --space fallbacks.
func spaceArg(s *session, args []string) (anytype.SpaceContext, error) {
	if len(args) == 1 {
		return s.resolveSpace(args[0])
	}
	return s.space()
}

var spacesGetCmd = &cobra.Command{
	Use:               "get [space]",
	Short:             "Show one space (defaults to the current space)",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeSpaceArg,
	RunE: run(func(s *session, args []string) error {
		sc, err := spaceArg(s, args)
		if err != nil {
			return err
		}
		resp, err := sc.Get(s.ctx)
		if err != nil {
			return err
		}
		return s.out.Print(resp.Space, func(w io.Writer) { writeSpace(w, resp.Space) })
	}),
}

var spacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a space",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req anytype.CreateSpaceRequest
		req.Name, _ = cmd.Flags().GetString("name")
		req.Description, _ = cmd.Flags().GetString("description")
		return run(func(s *session, _ []string) error {
			resp, err := s.client.Spaces().Create(s.ctx, req)
			if err != nil {
				return err
			}
			return printCreated(s.out, resp.Space, "Space", resp.Space.ID, resp.Space.Name)
		})(cmd, args)
	},
}

var spacesUpdateCmd = &cobra.Command{
	Use:               "update [space]",
	Short:             "Rename a space or change its description",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeSpaceArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req anytype.UpdateSpaceRequest
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			req.Name = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		if req.Name == nil && req.Description == nil {
			return errNothingToUpdate
		}
		return run(func(s *session, args []string) error {
			sc, err := spaceArg(s, args)
			if err != nil {
				return err
			}
			resp, err := sc.Update(s.ctx, req)
			if err != nil {
				return err
			}
			return s.out.Print(resp.Space, func(w io.Writer) { writeSpace(w, resp.Space) })
		})(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(spacesCmd)
	spacesCmd.AddCommand(spacesListCmd, spacesGetCmd, spacesCreateCmd, spacesUpdateCmd)

	addListFlags(spacesListCmd)

	spacesCreateCmd.Flags().String("name", "", "space name (required)")
	spacesCreateCmd.Flags().String("description", "", "space description")
	_ = spacesCreateCmd.MarkFlagRequired("name")

	spacesUpdateCmd.Flags().String("name", "", "new name")
	spacesUpdateCmd.Flags().String("description", "", "new description")
}
