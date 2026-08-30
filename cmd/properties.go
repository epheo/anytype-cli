package cmd

import (
	"io"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var propertyFormatList = []anytype.PropertyFormat{
	anytype.PropertyFormatText, anytype.PropertyFormatNumber, anytype.PropertyFormatSelect,
	anytype.PropertyFormatMultiSelect, anytype.PropertyFormatDate, anytype.PropertyFormatFiles,
	anytype.PropertyFormatCheckbox, anytype.PropertyFormatURL, anytype.PropertyFormatEmail,
	anytype.PropertyFormatPhone, anytype.PropertyFormatObjects,
}

var propertiesCmd = &cobra.Command{
	Use:   "properties",
	Short: "Manage properties of the current space",
}

var propertiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List properties",
	Args:  cobra.NoArgs,
	RunE: listInSpace(
		func(sc anytype.SpaceContext, _ []string) lister[anytype.Property] { return sc.Properties() },
		func(items []anytype.Property) *output.Table {
			t := output.NewTable("ID", "KEY", "NAME", "FORMAT")
			for _, p := range items {
				t.Row(p.ID, p.Key, p.Name, string(p.Format))
			}
			return t
		}),
}

func writeProperty(w io.Writer, p anytype.Property) {
	output.NewDetails().
		Add("ID", p.ID).
		Add("Key", p.Key).
		Add("Name", p.Name).
		Add("Format", string(p.Format)).
		Write(w)
}

var propertiesGetCmd = &cobra.Command{
	Use:   "get <property-id>",
	Short: "Show one property",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Property(args[0]).Get(s.ctx)
		if err != nil {
			return err
		}
		return s.out.Print(resp.Property, func(w io.Writer) { writeProperty(w, resp.Property) })
	}),
}

var propertiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a property",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		var req anytype.CreatePropertyRequest
		req.Name, _ = f.GetString("name")
		format, _ := f.GetString("format")
		req.Format = anytype.PropertyFormat(format)
		req.Key, _ = f.GetString("key")
		tags, _ := f.GetStringArray("tag")
		for _, t := range tags {
			name, color, _ := splitKV(t)
			if name == "" {
				name = t
			}
			req.Tags = append(req.Tags, anytype.CreateTagRequest{Name: name, Color: anytype.Color(color)})
		}
		return inSpace(func(s *session, sc anytype.SpaceContext, _ []string) error {
			resp, err := sc.Properties().Create(s.ctx, req)
			if err != nil {
				return err
			}
			return printCreated(s.out, resp.Property, "Property", resp.Property.ID, resp.Property.Key)
		})(cmd, args)
	},
}

var propertiesUpdateCmd = &cobra.Command{
	Use:   "update <property-id>",
	Short: "Rename a property or change its key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var req anytype.UpdatePropertyRequest
		req.Name, _ = cmd.Flags().GetString("name")
		req.Key, _ = cmd.Flags().GetString("key")
		if req.Name == "" && req.Key == "" {
			return errNothingToUpdate
		}
		return inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
			resp, err := sc.Property(args[0]).Update(s.ctx, req)
			if err != nil {
				return err
			}
			return s.out.Print(resp.Property, func(w io.Writer) { writeProperty(w, resp.Property) })
		})(cmd, args)
	},
}

var propertiesDeleteCmd = &cobra.Command{
	Use:   "delete <property-id>",
	Short: "Archive a property",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Property(args[0]).Delete(s.ctx)
		if err != nil {
			return err
		}
		return printArchived(s.out, resp.Property, "Property", resp.Property.ID, resp.Property.Key)
	}),
}

func init() {
	rootCmd.AddCommand(propertiesCmd)
	propertiesCmd.AddCommand(propertiesListCmd, propertiesGetCmd, propertiesCreateCmd, propertiesUpdateCmd, propertiesDeleteCmd)
	addListFlags(propertiesListCmd)

	c := propertiesCreateCmd.Flags()
	c.String("name", "", "property name (required)")
	c.String("format", "", "format: "+joinConsts(propertyFormatList...)+" (required)")
	c.String("key", "", "property key (generated when empty)")
	c.StringArray("tag", nil, "initial tag as name=color for select formats, repeatable")
	_ = propertiesCreateCmd.MarkFlagRequired("name")
	_ = propertiesCreateCmd.MarkFlagRequired("format")

	u := propertiesUpdateCmd.Flags()
	u.String("name", "", "new name")
	u.String("key", "", "new key")
}
