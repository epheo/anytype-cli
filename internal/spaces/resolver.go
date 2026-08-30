// Package spaces maps user-typed space names or IDs to space IDs.
package spaces

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/epheo/anytype-go"
)

var ErrAmbiguous = errors.New("ambiguous space name")

// Match picks a space by ID, then exact name, then unique substring.
// An unmatched query is returned unchanged so the API can reject unknown IDs.
func Match(list []anytype.Space, query string) (string, error) {
	for _, s := range list {
		if s.ID == query {
			return s.ID, nil
		}
	}
	for _, s := range list {
		if strings.EqualFold(s.Name, query) {
			return s.ID, nil
		}
	}

	q := strings.ToLower(query)
	var hits []anytype.Space
	for _, s := range list {
		if strings.Contains(strings.ToLower(s.Name), q) {
			hits = append(hits, s)
		}
	}
	switch len(hits) {
	case 0:
		return query, nil
	case 1:
		return hits[0].ID, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%q matches %d spaces, use the ID or a longer name:", query, len(hits))
	for i, s := range hits {
		if i == 5 {
			fmt.Fprintf(&b, "\n  ... and %d more", len(hits)-i)
			break
		}
		fmt.Fprintf(&b, "\n  %s  %s", s.ID, s.Name)
	}
	return "", fmt.Errorf("%w: %s", ErrAmbiguous, b.String())
}

// Resolve tries the SDK's exact-name lookup first and falls back to fuzzy
// matching over the full list only when that misses.
func Resolve(ctx context.Context, client anytype.Client, query string) (string, error) {
	sp, err := client.Spaces().GetByName(ctx, query)
	if err == nil {
		return sp.ID, nil
	}
	if !errors.Is(err, anytype.ErrNotFound) {
		return "", fmt.Errorf("look up space: %w", err)
	}

	var list []anytype.Space
	for s, err := range client.Spaces().All(ctx) {
		if err != nil {
			return "", fmt.Errorf("list spaces: %w", err)
		}
		list = append(list, s)
	}
	return Match(list, query)
}

// Completions returns "value<TAB>description" pairs for shell completion.
func Completions(ctx context.Context, client anytype.Client) ([]string, error) {
	var out []string
	for s, err := range client.Spaces().All(ctx) {
		if err != nil {
			return nil, err
		}
		out = append(out, s.ID+"\t"+s.Name)
		if !strings.ContainsAny(s.Name, " \t") {
			out = append(out, s.Name+"\t"+s.ID)
		}
	}
	return out, nil
}
