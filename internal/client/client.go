package client

import (
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client" // Register client implementation
)

// GetClient returns an authenticated Anytype client using the stored configuration
func GetClient(cfg *config.Config) anytype.Client {
	return anytype.NewClient(
		anytype.WithBaseURL(cfg.BaseURL),
		anytype.WithAppKey(cfg.AppKey),
	)
}
