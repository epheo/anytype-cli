// Package spaces provides helper functions for resolving space references
package spaces

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/epheo/anytype-cli/internal/auth"
	"github.com/epheo/anytype-cli/internal/client"
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-go"
	"github.com/spf13/cobra"
)

// ErrSpaceNotFound indicates the space could not be found
var ErrSpaceNotFound = errors.New("space not found")

// GetSpaceCompletionFunc returns a function that can be used for shell completion of space IDs and names
func GetSpaceCompletionFunc(cfg *config.Config) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Skip if we're not authenticated
		if !auth.IsAuthenticated(cfg) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Get all spaces
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		anytypeClient := client.GetClient(cfg)
		resp, err := anytypeClient.Spaces().List(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Return both IDs and names for completion
		var completions []string
		for _, space := range resp.Data {
			// Add space ID with description
			completions = append(completions, space.ID+"\t"+space.Name)
			// Add space name if it doesn't contain special characters
			if !strings.ContainsAny(space.Name, " \t\n\r") {
				completions = append(completions, space.Name+"\t"+space.ID)
			} else {
				// Add quoted name for spaces with special characters
				quotedName := fmt.Sprintf("%q", space.Name)
				completions = append(completions, quotedName+"\t"+space.ID)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ResolveSpace takes either a space ID or space name and returns the corresponding space ID.
// It first tries to find an exact match by name, then a partial match, and if not found,
// it falls back to treating the input as a direct ID.
func ResolveSpace(cfg *config.Config, spaceIdOrName string) (string, error) {
	// Get all spaces
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	anytypeClient := client.GetClient(cfg)
	resp, err := anytypeClient.Spaces().List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list spaces: %w", err)
	}

	// First: Check if the input directly matches a space ID
	// This is an optimization for users who know the exact space ID
	for _, space := range resp.Data {
		if space.ID == spaceIdOrName {
			return space.ID, nil
		}
	}

	// Second: Try to find an exact case-insensitive name match
	for _, space := range resp.Data {
		if strings.EqualFold(space.Name, spaceIdOrName) {
			return space.ID, nil
		}
	}

	// Third: Collect partial name matches
	matchedSpaces := []anytype.Space{}
	for _, space := range resp.Data {
		if strings.Contains(strings.ToLower(space.Name), strings.ToLower(spaceIdOrName)) {
			matchedSpaces = append(matchedSpaces, space)
		}
	}

	// If exactly one partial match, use it
	if len(matchedSpaces) == 1 {
		return matchedSpaces[0].ID, nil
	}

	// If multiple matches, provide a helpful error message
	if len(matchedSpaces) > 1 {
		msg := fmt.Sprintf("%s: multiple spaces matched '%s', please use space ID or a more specific name. Matched spaces:",
			ErrSpaceNotFound.Error(), spaceIdOrName)
		for i, space := range matchedSpaces {
			if i < 5 { // Limit to first 5 matches to avoid overwhelming output
				msg += fmt.Sprintf("\n  - '%s' (ID: %s)", space.Name, space.ID)
			}
		}
		if len(matchedSpaces) > 5 {
			msg += fmt.Sprintf("\n  ... and %d more", len(matchedSpaces)-5)
		}
		return "", errors.New(msg)
	}

	// Fourth: As a fallback, treat the input as a direct space ID
	// This handles cases where the user provided an ID that doesn't match any spaces
	// (which might be an error, but we'll let the API handle that)
	return spaceIdOrName, nil
}
