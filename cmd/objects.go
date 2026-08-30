package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var errNothingToUpdate = errors.New("no changes requested, see --help for flags")

var objectsCmd = &cobra.Command{
	Use:   "objects",
	Short: "Manage objects in the current space",
}

var objectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects",
	Args:  cobra.NoArgs,
	RunE: listInSpace(
		func(sc anytype.SpaceContext, _ []string) lister[anytype.Object] { return sc.Objects() },
		objectTable),
}

func objectTable(items []anytype.Object) *output.Table {
	t := output.NewTable("ID", "NAME", "TYPE", "LAYOUT").MaxWidth(1, 40).MaxWidth(2, 20)
	for _, o := range items {
		t.Row(o.ID, o.Name, typeKey(o.Type), o.Layout)
	}
	return t
}

func typeKey(t *anytype.Type) string {
	if t == nil {
		return ""
	}
	return t.Key
}

var objectsGetCmd = &cobra.Command{
	Use:   "get <object-id>",
	Short: "Show one object with its properties",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Object(args[0]).Get(s.ctx)
		if err != nil {
			return err
		}
		return s.out.Print(resp.Object, func(w io.Writer) { writeObject(w, resp.Object) })
	}),
}

func writeObject(w io.Writer, o *anytype.Object) {
	output.NewDetails().
		Add("ID", o.ID).
		Add("Name", o.Name).
		AddIf("Type", typeKey(o.Type)).
		Add("Layout", o.Layout).
		Add("Space ID", o.SpaceID).
		Add("Archived", strconv.FormatBool(o.Archived)).
		AddIf("Icon", output.Icon(o.Icon)).
		AddIf("Snippet", o.Snippet).
		Write(w)

	if len(o.Properties) == 0 {
		return
	}
	fmt.Fprintln(w, "\nProperties")
	t := output.NewTable("KEY", "NAME", "FORMAT", "VALUE").MaxWidth(3, 60)
	for _, p := range o.Properties {
		t.Row(p.Key, p.Name, string(p.Format), output.PropertyValue(p))
	}
	t.Write(w)
}

// bodyFlags returns --body, or the content of --body-file ("-" is stdin).
func bodyFlags(cmd *cobra.Command) (string, error) {
	body, _ := cmd.Flags().GetString("body")
	path, _ := cmd.Flags().GetString("body-file")
	if path == "" {
		return body, nil
	}
	if body != "" {
		return "", errors.New("--body and --body-file are exclusive")
	}
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(data), nil
}

var objectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an object",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		req := anytype.CreateObjectRequest{Icon: iconFlag(cmd)}
		req.Name, _ = f.GetString("name")
		req.TypeKey, _ = f.GetString("type")
		req.TemplateID, _ = f.GetString("template")
		props, _ := f.GetStringArray("property")
		body, err := bodyFlags(cmd)
		if err != nil {
			return err
		}
		req.Body = body

		return inSpace(func(s *session, sc anytype.SpaceContext, _ []string) error {
			if req.Properties, err = propertyValues(s, sc, props); err != nil {
				return err
			}
			resp, err := sc.Objects().Create(s.ctx, req)
			if err != nil {
				return err
			}
			return printCreated(s.out, resp.Object, "Object", resp.Object.ID, resp.Object.Name)
		})(cmd, args)
	},
}

var objectsUpdateCmd = &cobra.Command{
	Use:   "update <object-id>",
	Short: "Change an object's name, body, icon, type, or properties",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := cmd.Flags()
		req := anytype.UpdateObjectRequest{Icon: iconFlag(cmd)}
		req.Name, _ = f.GetString("name")
		req.TypeKey, _ = f.GetString("type")
		props, _ := f.GetStringArray("property")
		body, err := bodyFlags(cmd)
		if err != nil {
			return err
		}
		req.Markdown = body
		if req.Name == "" && req.TypeKey == "" && req.Markdown == "" && req.Icon == nil && len(props) == 0 {
			return errNothingToUpdate
		}

		return inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
			if req.Properties, err = propertyValues(s, sc, props); err != nil {
				return err
			}
			resp, err := sc.Object(args[0]).Update(s.ctx, req)
			if err != nil {
				return err
			}
			return s.out.Print(resp.Object, func(w io.Writer) { writeObject(w, resp.Object) })
		})(cmd, args)
	},
}

var objectsDeleteCmd = &cobra.Command{
	Use:   "delete <object-id>",
	Short: "Archive an object (moves it to the bin)",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Object(args[0]).Delete(s.ctx)
		if err != nil {
			return err
		}
		return printArchived(s.out, resp.Object, "Object", resp.Object.ID, resp.Object.Name)
	}),
}

var objectsExportCmd = &cobra.Command{
	Use:   "export <object-id>",
	Short: "Print an object as markdown",
	Args:  cobra.ExactArgs(1),
	RunE: inSpace(func(s *session, sc anytype.SpaceContext, args []string) error {
		resp, err := sc.Object(args[0]).Get(s.ctx, anytype.WithFormat("md"))
		if err != nil {
			return err
		}
		return s.out.Print(resp.Object, func(w io.Writer) { fmt.Fprintln(w, resp.Object.Markdown) })
	}),
}

func addBodyFlags(cmd *cobra.Command) {
	cmd.Flags().String("body", "", "markdown body")
	cmd.Flags().String("body-file", "", "read markdown body from a file, or stdin with -")
	cmd.Flags().StringArray("property", nil,
		"property as key=value, repeatable; lists are comma-separated, select/multi_select take tag IDs")
	cmd.Flags().String("icon", "", "emoji icon")
}

func init() {
	rootCmd.AddCommand(objectsCmd)
	objectsCmd.AddCommand(objectsListCmd, objectsGetCmd, objectsCreateCmd, objectsUpdateCmd, objectsDeleteCmd, objectsExportCmd)

	addListFlags(objectsListCmd)

	c := objectsCreateCmd.Flags()
	c.String("name", "", "object name")
	c.String("type", "page", "type key")
	c.String("template", "", "template ID")
	addBodyFlags(objectsCreateCmd)

	u := objectsUpdateCmd.Flags()
	u.String("name", "", "new name")
	u.String("type", "", "new type key")
	addBodyFlags(objectsUpdateCmd)
}
