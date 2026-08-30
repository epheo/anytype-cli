package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"
	"sync"

	"github.com/epheo/anytype-cli/internal/client"
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-cli/internal/output"
	"github.com/epheo/anytype-cli/internal/spaces"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

var (
	cfgOnce sync.Once
	cfgVal  *config.Config
	cfgErr  error
)

var errNoSpace = errors.New("no space given: use --space, " + config.EnvSpace + ", or 'anytype-cli config set default-space'")

func loadConfig() (*config.Config, error) {
	cfgOnce.Do(func() {
		cfgVal, cfgErr = config.Load(flagConfig)
		if cfgErr == nil && flagBaseURL != "" {
			cfgVal.BaseURL = flagBaseURL
		}
	})
	return cfgVal, cfgErr
}

func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), flagTimeout)
}

// session bundles what nearly every command needs after argument parsing.
type session struct {
	ctx    context.Context
	cfg    *config.Config
	client anytype.Client
	out    output.Printer
}

func newSession() (*session, context.CancelFunc, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}
	format, _ := output.ParseFormat(flagOutput)
	ctx, cancel := newContext()
	return &session{
		ctx:    ctx,
		cfg:    cfg,
		client: client.New(cfg),
		out:    output.Printer{Format: format, Out: os.Stdout},
	}, cancel, nil
}

// resolveSpace turns an ID or name into a SpaceContext.
func (s *session) resolveSpace(query string) (anytype.SpaceContext, error) {
	id, err := spaces.Resolve(s.ctx, s.client, query)
	if err != nil {
		return nil, err
	}
	return s.client.Space(id), nil
}

// spaceQuery is the raw --space value after fallbacks; empty when none set.
func (s *session) spaceQuery() string {
	if flagSpace != "" {
		return flagSpace
	}
	return s.cfg.DefaultSpace
}

// space is the current space from --space, env, or config.
func (s *session) space() (anytype.SpaceContext, error) {
	q := s.spaceQuery()
	if q == "" {
		return nil, errNoSpace
	}
	return s.resolveSpace(q)
}

// run wraps a command body with session setup so RunE bodies stay short.
func run(fn func(s *session, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		s, cancel, err := newSession()
		if err != nil {
			return err
		}
		defer cancel()
		return fn(s, args)
	}
}

// inSpace is run for commands that operate inside the current space.
func inSpace(fn func(s *session, sc anytype.SpaceContext, args []string) error) func(*cobra.Command, []string) error {
	return run(func(s *session, args []string) error {
		sc, err := s.space()
		if err != nil {
			return err
		}
		return fn(s, sc, args)
	})
}

// completeSpaceArg completes the first positional argument of spaces get/update.
func completeSpaceArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSpaceFlag(cmd, args, toComplete)
}

func completeSpaceFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	cfg, err := loadConfig()
	if err != nil || !cfg.Authenticated() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ctx, cancel := newContext()
	defer cancel()
	list, err := spaces.Completions(ctx, client.New(cfg))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return list, cobra.ShellCompDirectiveNoFileComp
}

// Pagination flags shared by every list command.

func addListFlags(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 0, "maximum items per page (server default when 0)")
	cmd.Flags().Int("offset", 0, "number of items to skip")
	cmd.Flags().Bool("all", false, "fetch every page")
}

type lister[T any] interface {
	List(ctx context.Context, opts ...anytype.ListOption) (*anytype.Page[T], error)
	All(ctx context.Context, opts ...anytype.ListOption) iter.Seq2[T, error]
}

// fetch honors --limit/--offset/--all. Pagination is nil when --all was used.
func fetch[T any](cmd *cobra.Command, ctx context.Context, l lister[T]) ([]T, *anytype.Pagination, error) {
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	all, _ := cmd.Flags().GetBool("all")

	var opts []anytype.ListOption
	if limit > 0 {
		opts = append(opts, anytype.WithLimit(limit))
	}
	if offset > 0 {
		opts = append(opts, anytype.WithOffset(offset))
	}

	if all {
		var items []T
		for item, err := range l.All(ctx, opts...) {
			if err != nil {
				return nil, nil, err
			}
			items = append(items, item)
		}
		return items, nil, nil
	}

	page, err := l.List(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}
	return page.Data, &page.Pagination, nil
}

// listRun binds fetch to a command; pick chooses the SDK lister, show builds the table.
func listRun[T any](pick func(s *session, args []string) (lister[T], error), show func(items []T) *output.Table) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		s, cancel, err := newSession()
		if err != nil {
			return err
		}
		defer cancel()

		l, err := pick(s, args)
		if err != nil {
			return err
		}
		items, pg, err := fetch(cmd, s.ctx, l)
		if err != nil {
			return err
		}
		if items == nil {
			items = []T{}
		}
		return s.out.Print(items, func(w io.Writer) {
			show(items).Write(w)
			writeFooter(w, len(items), pg)
		})
	}
}

// listInSpace is listRun for listers that hang off the current space.
func listInSpace[T any](pick func(sc anytype.SpaceContext, args []string) lister[T], show func(items []T) *output.Table) func(*cobra.Command, []string) error {
	return listRun(func(s *session, args []string) (lister[T], error) {
		sc, err := s.space()
		if err != nil {
			return nil, err
		}
		return pick(sc, args), nil
	}, show)
}

func writeFooter(w io.Writer, n int, pg *anytype.Pagination) {
	if pg == nil || !pg.HasMore {
		fmt.Fprintf(w, "\nTotal: %d\n", n)
		return
	}
	fmt.Fprintf(w, "\nShowing %d of %d (offset %d). Use --offset %d or --all.\n",
		n, pg.Total, pg.Offset, pg.Offset+n)
}

// Shared flag parsing helpers.

func iconFlag(cmd *cobra.Command) *anytype.Icon {
	emoji, _ := cmd.Flags().GetString("icon")
	if emoji == "" {
		return nil
	}
	return anytype.EmojiIcon(emoji)
}

// splitKV parses "key=value"; the value may itself contain "=".
func splitKV(s string) (string, string, error) {
	k, v, ok := strings.Cut(s, "=")
	if !ok || k == "" {
		return "", "", fmt.Errorf("expected key=value, got %q", s)
	}
	return k, v, nil
}

func joinConsts[T ~string](values ...T) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = string(v)
	}
	return strings.Join(parts, ", ")
}

func printCreated(out output.Printer, data any, kind, id, name string) error {
	return out.Print(data, func(w io.Writer) {
		fmt.Fprintf(w, "%s created: %s (%s)\n", kind, name, id)
	})
}

func printArchived(out output.Printer, data any, kind, id, name string) error {
	return out.Print(data, func(w io.Writer) {
		fmt.Fprintf(w, "%s archived: %s (%s)\n", kind, name, id)
	})
}
