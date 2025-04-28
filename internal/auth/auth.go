package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-go"
)

// RunAuthentication performs the interactive authentication flow
func RunAuthentication(baseURL string) (*config.Config, error) {
	// Create unauthenticated client
	client := anytype.NewClient(
		anytype.WithBaseURL(baseURL),
	)

	// Create a context with timeout for authentication
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Step 1: Initiate auth flow and get challenge ID
	fmt.Println("Starting authentication with Anytype...")
	authResponse, err := client.Auth().DisplayCode(ctx, "anytype-cli")
	if err != nil {
		return nil, fmt.Errorf("failed to initiate authentication: %w", err)
	}
	challengeID := authResponse.ChallengeID

	// Step 2: Prompt user to enter verification code
	fmt.Println("\nPlease check your Anytype app and enter the displayed verification code.")
	fmt.Println("The code should be visible in your Anytype app's authentication screen.")
	fmt.Println("If no code appears, make sure Anytype is running and try again.")

	var code string
	prompt := &survey.Input{
		Message: "Enter verification code:",
	}
	if err := survey.AskOne(prompt, &code); err != nil {
		return nil, fmt.Errorf("authentication canceled: %w", err)
	}

	// Step 3: Complete auth by providing the code
	tokenResponse, err := client.Auth().GetToken(ctx, challengeID, code)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Save tokens to config
	cfg := &config.Config{
		AppKey:       tokenResponse.AppKey,
		SessionToken: tokenResponse.SessionToken,
		BaseURL:      baseURL,
	}

	fmt.Println("Authentication successful!")
	return cfg, nil
}

// IsAuthenticated checks if the configuration has authentication credentials
func IsAuthenticated(cfg *config.Config) bool {
	return cfg != nil && cfg.AppKey != "" && cfg.SessionToken != ""
}
