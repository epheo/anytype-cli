// Package auth runs the challenge/code flow against a local Anytype app.
package auth

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/epheo/anytype-go"
)

const appName = "anytype-cli"

// Login returns the API key after the user types the code shown by the Anytype app.
func Login(ctx context.Context, baseURL string, in io.Reader, out io.Writer) (string, error) {
	client := anytype.NewClient(anytype.WithBaseURL(baseURL))

	challenge, err := client.Auth().CreateChallenge(ctx, appName)
	if err != nil {
		return "", fmt.Errorf("start challenge: %w", err)
	}

	fmt.Fprintln(out, "A verification code is now displayed in your Anytype app.")
	fmt.Fprint(out, "Enter verification code: ")
	code, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && code == "" {
		return "", fmt.Errorf("read code: %w", err)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("empty code")
	}

	key, err := client.Auth().CreateApiKey(ctx, challenge.ChallengeID, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	return key.ApiKey, nil
}
