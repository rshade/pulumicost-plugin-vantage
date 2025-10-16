// Package client provides HTTP client functionality for Vantage API
package client

import (
	"context"
	"fmt"
	"time"
)

// Client defines the interface for interacting with Vantage API
type Client interface {
	// Costs fetches cost data with pagination
	Costs(ctx context.Context, query Query) (Page, error)
	// Forecast fetches forecast data for a cost report
	Forecast(ctx context.Context, reportToken string, query ForecastQuery) (Forecast, error)
}

// Config holds client configuration
type Config struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	MaxRetries int
	Logger     Logger
}

// defaultLogger is the default no-op logger instance
var defaultLogger = &noopLogger{}

// DefaultConfig returns a default client configuration
func DefaultConfig(token string) Config {
	return Config{
		BaseURL:    "https://api.vantage.sh",
		Token:      token,
		Timeout:    60 * time.Second,
		MaxRetries: 5,
		Logger:     defaultLogger,
	}
}

// client implements the Client interface
type client struct {
	httpClient *httpClient
	logger     Logger
}

// New creates a new Vantage API client
func New(config Config) (Client, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	httpClient, err := newHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client: %w", err)
	}

	return &client{
		httpClient: httpClient,
		logger:     config.Logger,
	}, nil
}

// Costs implements Client.Costs
func (c *client) Costs(ctx context.Context, query Query) (Page, error) {
	return c.httpClient.doCostsRequest(ctx, query)
}

// Forecast implements Client.Forecast
func (c *client) Forecast(ctx context.Context, reportToken string, query ForecastQuery) (Forecast, error) {
	return c.httpClient.doForecastRequest(ctx, reportToken, query)
}
