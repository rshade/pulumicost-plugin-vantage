// Package client provides HTTP client functionality for Vantage API.
package client

import (
	"context"
	"errors"
	"time"
)

const (
	defaultTimeout = 60 * time.Second
	defaultRetries = 5
)

// Client defines the interface for interacting with Vantage API.
type Client interface {
	// Costs fetches cost data with pagination.
	Costs(ctx context.Context, query Query) (Page, error)
	// Forecast fetches forecast data for a cost report.
	Forecast(ctx context.Context, reportToken string, query ForecastQuery) (Forecast, error)
}

// Config holds client configuration.
type Config struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	MaxRetries int
	Logger     Logger
}

// DefaultConfig returns a default client configuration.
func DefaultConfig(token string) Config {
	return Config{
		BaseURL:    "https://api.vantage.sh",
		Token:      token,
		Timeout:    defaultTimeout,
		MaxRetries: defaultRetries,
		Logger:     NewNoopLogger(),
	}
}

// client implements the Client interface.
type client struct {
	httpClient *httpClient
	logger     Logger
}

// New creates a new Vantage API client.
func New(config Config) (Client, error) {
	if config.Token == "" {
		return nil, errors.New("token is required")
	}

	// Defensive defaults in case callers don't use DefaultConfig
	if config.Logger == nil {
		config.Logger = NewNoopLogger()
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = defaultRetries
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.vantage.sh"
	}

	httpClient := newHTTPClient(config)

	return &client{
		httpClient: httpClient,
		logger:     config.Logger,
	}, nil
}

// Costs implements Client.Costs.
func (c *client) Costs(ctx context.Context, query Query) (Page, error) {
	return c.httpClient.doCostsRequest(ctx, query)
}

// Forecast implements Client.Forecast.
func (c *client) Forecast(ctx context.Context, reportToken string, query ForecastQuery) (Forecast, error) {
	return c.httpClient.doForecastRequest(ctx, reportToken, query)
}
