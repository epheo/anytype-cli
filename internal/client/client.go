// Package client builds SDK clients from CLI config.
package client

import (
	"github.com/epheo/anytype-cli/internal/config"
	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client" // registers the HTTP implementation
)

func New(cfg *config.Config) anytype.Client {
	return anytype.NewClient(
		anytype.WithBaseURL(cfg.BaseURL),
		anytype.WithAppKey(cfg.AppKey),
	)
}
