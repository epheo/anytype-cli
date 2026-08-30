package cmd

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var typeLayouts = []anytype.TypeLayout{
	anytype.TypeLayoutBasic, anytype.TypeLayoutProfile, anytype.TypeLayoutAction, anytype.TypeLayoutNote,
}

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "Manage object types in the current space",
}

var typesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List types",
	Args:  cobra.NoArgs,
	RunE: listInSpace(
		func(sc anytype.SpaceContext, _ []string) lister[anytype.Type] { return sc.Types() },
		func(items []anytype.Type) *output.Table {
			t := output.NewTable("ID", "KEY", "NAME", "LAYOUT", "ARCHIVED")
			for _, ty := range items {
				t.Row(ty.ID, ty.Key, ty.Name, ty.Layout, strconv.FormatBool(ty.Archived))
			}
			return t
		}),
}

// resolveType accepts a key, a display name, or an ID. Key and name come
// first because they are matched against the list; the API answers 500, not
// 404, for an unknown ID, so a direct GET cannot be used to probe.
func resolveType(s *session, sc anytype.SpaceContext, arg string) (*anytype.Type, error) {
	ty, err := sc.Types().Get(s.ctx, arg)
	if err == nil {
		return ty, nil
	}
	if !errors.Is(err, anytype.ErrNotFound) {
		return nil, err
	}
	key, err := sc.Types().GetKeyByName(s.ctx, arg)
	if err == nil {
		return sc.Types().Get(s.ctx, key)
	}
	if !errors.Is(err, anytype.ErrNotFound) {
		return nil, err
	}
	resp, err := sc.Type(arg).Get(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("type %q: no key or name matches, and lookup by ID failed: %w", arg, err)
	}
	return &resp.Type, nil
}

var typesGetCmd = &cobra.Command{
	Use:   "get <type>",
	Short: "Show one type by ID, key, or name",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		ty, err := resolveType(s, sc, args[0])
		if err != nil {
			return err
		}
		return s.out.Print(ty, func(w io.Writer) { writeType(w, ty) })
	}),
}

func writeType(w io.Writer, ty *anytype.Type) {
	output.NewDetails().
		Add("ID", ty.ID).
		Add("Key", ty.Key).
		Add("Name", ty.Name).
		AddIf("Plural Name", ty.PluralName).
		Add("Layout", ty.Layout).
		Add("Archived", strconv.FormatBool(ty.Archived)).
		AddIf("Icon", output.Icon(ty.Icon)).
		Write(w)
	if len(ty.Properties) == 0 {
		return
	}
	fmt.Fprintln(w, "\nProperties")
	t := output.NewTable("KEY", "NAME", "FORMAT")
	for _, p := range ty.Properties {
		t.Row(p.Key, p.Name, string(p.Format))
	}
	t.Write(w)
}

// parsePropertyDefs parses "key:name:format"; an empty name defaults to the key.
func parsePropertyDefs(specs []string) ([]anytype.PropertyDefinition, error) {
	defs := make([]anytype.PropertyDefinition, 0, len(specs))
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 3)
		if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
			return nil, fmt.Errorf("expected key:name:format, got %q", spec)
		}
		name := parts[1]
		if name == "" {
			name = parts[0]
		}
		defs = append(defs, anytype.PropertyDefinition{Key: parts[0], Name: name, Format: anytype.PropertyFormat(parts[2])})
	}
	return defs, nil
}

func typeRequestFromFlags(cmd *cobra.Command) (anytype.CreateTypeRequest, error) {
	f := cmd.Flags()
	req := anytype.CreateTypeRequest{Icon: iconFlag(cmd)}
	req.Key, _ = f.GetString("key")
	req.Name, _ = f.GetString("name")
	req.PluralName, _ = f.GetString("plural")
	layout, _ := f.GetString("layout")
	req.Layout = anytype.TypeLayout(layout)
	specs, _ := f.GetStringArray("property")
	defs, err := parsePropertyDefs(specs)
	if err != nil {
		return req, err
	}
	req.Properties = defs
	return req, nil
}

var typesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a type",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		req, err := typeRequestFromFlags(cmd)
		if err != nil {
			return err
		}
		if req.PluralName == "" {
			req.PluralName = req.Name + "s"
		}
		return inSpace(func(s *session, sc anytype.SpaceContext, _ []string) error {
			resp, err := sc.Types().Create(s.ctx, req)
			if err != nil {
				return err
			}
			return printCreated(s.out, resp.Type, "Type", resp.Type.ID, resp.Type.Key)
		})(cmd, args)
	},
}

var typesUpdateCmd = &cobra.Command{
	Use:   "update <type>",
	Short: "Change a type's name, layout, icon, or properties",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := typeRequestFromFlags(cmd)
		if err != nil {
			return err
		}
		req := anytype.UpdateTypeRequest{
			Key: c.Key, Name: c.Name, Icon: c.Icon, Layout: c.Layout,
			PluralName: c.PluralName, Properties: c.Properties,
		}
		if req.Key == "" && req.Name == "" && req.Icon == nil && req.Layout == "" && req.PluralName == "" && len(req.Properties) == 0 {
			return errNothingToUpdate
		}
		return inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
			ty, err := resolveType(s, sc, args[0])
			if err != nil {
				return err
			}
			resp, err := sc.Type(ty.ID).Update(s.ctx, req)
			if err != nil {
				return err
			}
			return s.out.Print(resp.Type, func(w io.Writer) { writeType(w, &resp.Type) })
		})(cmd, args)
	},
}

var typesDeleteCmd = &cobra.Command{
	Use:   "delete <type>",
	Short: "Archive a type",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		ty, err := resolveType(s, sc, args[0])
		if err != nil {
			return err
		}
		resp, err := sc.Type(ty.ID).Delete(s.ctx)
		if err != nil {
			return err
		}
		return printArchived(s.out, resp.Type, "Type", resp.Type.ID, resp.Type.Key)
	}),
}

func addTypeFlags(cmd *cobra.Command, create bool) {
	f := cmd.Flags()
	f.String("key", "", "type key")
	f.String("name", "", "type name")
	f.String("plural", "", "plural name (default: name + s)")
	f.String("layout", "", "layout: "+joinConsts(typeLayouts...))
	f.String("icon", "", "emoji icon")
	f.StringArray("property", nil, "property as key:name:format, repeatable")
	if create {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("layout")
	}
}

func init() {
	rootCmd.AddCommand(typesCmd)
	typesCmd.AddCommand(typesListCmd, typesGetCmd, typesCreateCmd, typesUpdateCmd, typesDeleteCmd)
	addListFlags(typesListCmd)
	addTypeFlags(typesCreateCmd, true)
	addTypeFlags(typesUpdateCmd, false)
}
