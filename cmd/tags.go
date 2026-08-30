package cmd

import (
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var tagColors = []anytype.Color{
	anytype.ColorGrey, anytype.ColorYellow, anytype.ColorOrange, anytype.ColorRed, anytype.ColorPink,
	anytype.ColorPurple, anytype.ColorBlue, anytype.ColorIce, anytype.ColorTeal, anytype.ColorLime,
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags of a select or multi_select property",
}

var tagsListCmd = &cobra.Command{
	Use:   "list <property-id>",
	Short: "List tags of a property",
	Args:  cobra.ExactArgs(1),
	RunE: listInSpace(
		func(sc anytype.SpaceContext, args []string) lister[anytype.Tag] { return sc.Property(args[0]).Tags() },
		func(items []anytype.Tag) *output.Table {
			t := output.NewTable("ID", "KEY", "NAME", "COLOR")
			for _, tag := range items {
				t.Row(tag.ID, tag.Key, tag.Name, string(tag.Color))
			}
			return t
		}),
}

func writeTag(w io.Writer, t anytype.Tag) {
	output.NewDetails().
		Add("ID", t.ID).
		Add("Key", t.Key).
		Add("Name", t.Name).
		AddIf("Color", string(t.Color)).
		Write(w)
}

var tagsGetCmd = &cobra.Command{
	Use:   "get <property-id> <tag-id>",
	Short: "Show one tag",
	Args:  cobra.ExactArgs(2),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Property(args[0]).Tag(args[1]).Get(s.ctx)
		if err != nil {
			return err
		}
		return s.out.Print(resp.Tag, func(w io.Writer) { writeTag(w, resp.Tag) })
	}),
}

func tagRequest(cmd *cobra.Command) (name string, color anytype.Color, key string) {
	name, _ = cmd.Flags().GetString("name")
	c, _ := cmd.Flags().GetString("color")
	key, _ = cmd.Flags().GetString("key")
	return name, anytype.Color(c), key
}

var tagsCreateCmd = &cobra.Command{
	Use:   "create <property-id>",
	Short: "Create a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, color, key := tagRequest(cmd)
		req := anytype.CreateTagRequest{Name: name, Color: color, Key: key}
		return inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
			resp, err := sc.Property(args[0]).Tags().Create(s.ctx, req)
			if err != nil {
				return err
			}
			return printCreated(s.out, resp.Tag, "Tag", resp.Tag.ID, resp.Tag.Name)
		})(cmd, args)
	},
}

var tagsUpdateCmd = &cobra.Command{
	Use:   "update <property-id> <tag-id>",
	Short: "Change a tag's name, color, or key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, color, key := tagRequest(cmd)
		if name == "" && color == "" && key == "" {
			return errNothingToUpdate
		}
		req := anytype.UpdateTagRequest{Name: name, Color: color, Key: key}
		return inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
			resp, err := sc.Property(args[0]).Tag(args[1]).Update(s.ctx, req)
			if err != nil {
				return err
			}
			return s.out.Print(resp.Tag, func(w io.Writer) { writeTag(w, resp.Tag) })
		})(cmd, args)
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <property-id> <tag-id>",
	Short: "Archive a tag",
	Args:  cobra.ExactArgs(2),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Property(args[0]).Tag(args[1]).Delete(s.ctx)
		if err != nil {
			return err
		}
		return printArchived(s.out, resp.Tag, "Tag", resp.Tag.ID, resp.Tag.Name)
	}),
}

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.AddCommand(tagsListCmd, tagsGetCmd, tagsCreateCmd, tagsUpdateCmd, tagsDeleteCmd)
	addListFlags(tagsListCmd)

	c := tagsCreateCmd.Flags()
	c.String("name", "", "tag name (required)")
	c.String("color", string(anytype.ColorGrey), "tag color: "+joinConsts(tagColors...))
	c.String("key", "", "tag key (generated when empty)")
	_ = tagsCreateCmd.MarkFlagRequired("name")

	u := tagsUpdateCmd.Flags()
	u.String("name", "", "new name")
	u.String("color", "", "new color: "+joinConsts(tagColors...))
	u.String("key", "", "new key")
}
