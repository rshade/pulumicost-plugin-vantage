// Package client provides HTTP client functionality for Vantage API
package client

import (
	"context"
)

// Logger defines the minimal logging interface used by the client.
// Implementations should be safe for concurrent use.
type Logger interface {
	// Debug logs debug-level messages with structured fields
	Debug(ctx context.Context, msg string, fields map[string]interface{})

	// Info logs info-level messages with structured fields
	Info(ctx context.Context, msg string, fields map[string]interface{})

	// Warn logs warning-level messages with structured fields
	Warn(ctx context.Context, msg string, fields map[string]interface{})

	// Error logs error-level messages with structured fields
	Error(ctx context.Context, msg string, fields map[string]interface{})
}

// noopLogger provides a no-op implementation of Logger
type noopLogger struct{}

func (n *noopLogger) Debug(_ context.Context, _ string, _ map[string]interface{}) {}
func (n *noopLogger) Info(_ context.Context, _ string, _ map[string]interface{})  {}
func (n *noopLogger) Warn(_ context.Context, _ string, _ map[string]interface{})  {}
func (n *noopLogger) Error(_ context.Context, _ string, _ map[string]interface{}) {}

// NewNoopLogger returns a logger that discards all messages
func NewNoopLogger() Logger {
	return &noopLogger{}
}
