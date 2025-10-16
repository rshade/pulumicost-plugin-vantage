// Package client provides HTTP client functionality for Vantage API
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL:    "https://api.vantage.sh",
				Token:      "test-token",
				Timeout:    time.Second * 30,
				MaxRetries: 3,
				Logger:     NewNoopLogger(),
			},
			wantErr: false,
		},
		{
			name: "missing token",
			config: Config{
				BaseURL: "https://api.vantage.sh",
				Logger:  NewNoopLogger(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_Costs(t *testing.T) {
	// Mock server response
	mockResponse := CostsResponse{
		Data: []CostRow{
			{
				Provider:    "aws",
				Service:     "EC2",
				Cost:        100.50,
				Currency:    "USD",
				BucketStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				BucketEnd:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		NextCursor: "next-page-cursor",
		HasMore:    true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// Check query parameters
		assert.Equal(t, "test-workspace", r.URL.Query().Get("workspace_token"))
		assert.Equal(t, "day", r.URL.Query().Get("granularity"))
		assert.Contains(t, r.URL.Query()["group_bys[]"], "provider")
		assert.Contains(t, r.URL.Query()["metrics[]"], "cost")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	query := Query{
		WorkspaceToken: "test-workspace",
		StartAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:          time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		Granularity:    "day",
		GroupBys:       []string{"provider", "service"},
		Metrics:        []string{"cost", "usage"},
		PageSize:       100,
	}

	page, err := client.Costs(context.Background(), query)
	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, "aws", page.Data[0].Provider)
	assert.Equal(t, "EC2", page.Data[0].Service)
	assert.Equal(t, 100.50, page.Data[0].Cost)
	assert.Equal(t, "next-page-cursor", page.NextCursor)
	assert.True(t, page.HasMore)
}

func TestClient_Forecast(t *testing.T) {
	// Mock server response
	mockResponse := ForecastResponse{
		Data: []ForecastRow{
			{
				BucketStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				BucketEnd:   time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
				Cost:        150.75,
				Currency:    "USD",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/cost_reports/test-report-token/forecast", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	query := ForecastQuery{
		StartAt:     time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		EndAt:       time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
		Granularity: "day",
	}

	forecast, err := client.Forecast(context.Background(), "test-report-token", query)
	require.NoError(t, err)

	assert.Len(t, forecast.Data, 1)
	assert.Equal(t, 150.75, forecast.Data[0].Cost)
	assert.Equal(t, "USD", forecast.Data[0].Currency)
}

func TestClient_RetryOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call fails with 503
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Second call succeeds
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CostsResponse{Data: []CostRow{}})
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 1,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	query := Query{
		WorkspaceToken: "test-workspace",
		StartAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Granularity:    "day",
	}

	_, err = client.Costs(context.Background(), query)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // Should have retried once
}

func TestClient_RateLimitHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", "60") // Reset in 60 seconds
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0, // Don't retry to avoid waiting
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	query := Query{
		WorkspaceToken: "test-workspace",
		StartAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Granularity:    "day",
	}

	_, err = client.Costs(context.Background(), query)
	assert.Error(t, err)
	// Should be a rate limit error
	assert.Contains(t, err.Error(), "rate limited")
}

func TestPager_NextPage(t *testing.T) {
	// First page response
	firstResponse := CostsResponse{
		Data: []CostRow{
			{Provider: "aws", Cost: 100},
		},
		NextCursor: "cursor-2",
		HasMore:    true,
	}

	// Second page response
	secondResponse := CostsResponse{
		Data: []CostRow{
			{Provider: "gcp", Cost: 200},
		},
		NextCursor: "",
		HasMore:    false,
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			json.NewEncoder(w).Encode(firstResponse)
		} else {
			json.NewEncoder(w).Encode(secondResponse)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	pager := NewPager(client, Query{
		WorkspaceToken: "test-workspace",
		StartAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Granularity:    "day",
	}, NewNoopLogger())

	// First page
	page1, err := pager.NextPage(context.Background())
	require.NoError(t, err)
	assert.Len(t, page1.Data, 1)
	assert.Equal(t, "aws", page1.Data[0].Provider)
	assert.Equal(t, "cursor-2", page1.NextCursor)
	assert.True(t, page1.HasMore)

	// Second page
	page2, err := pager.NextPage(context.Background())
	require.NoError(t, err)
	assert.Len(t, page2.Data, 1)
	assert.Equal(t, "gcp", page2.Data[0].Provider)
	assert.Equal(t, "", page2.NextCursor)
	assert.False(t, page2.HasMore)

	assert.Equal(t, 2, callCount)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-token")

	assert.Equal(t, "https://api.vantage.sh", config.BaseURL)
	assert.Equal(t, "test-token", config.Token)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.NotNil(t, config.Logger)
}

func TestPager_HasMore(t *testing.T) {
	client, err := New(Config{
		BaseURL:    "https://api.vantage.sh",
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	// Test with empty cursor (should return false initially)
	pager := NewPager(client, Query{}, NewNoopLogger())
	assert.False(t, pager.HasMore())

	// Test with cursor (should return true)
	pager.query.Cursor = "test-cursor"
	assert.True(t, pager.HasMore())
}

func TestPager_AllPages(t *testing.T) {
	// Mock server response with multiple pages
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// First page
			resp := CostsResponse{
				Data: []CostRow{
					{Provider: "aws", Cost: 100},
				},
				NextCursor: "cursor-2",
				HasMore:    true,
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second page (final)
			resp := CostsResponse{
				Data: []CostRow{
					{Provider: "gcp", Cost: 200},
				},
				NextCursor: "",
				HasMore:    false,
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 0,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	pager := NewPager(client, Query{
		WorkspaceToken: "test-workspace",
		StartAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndAt:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Granularity:    "day",
	}, NewNoopLogger())

	rows, err := pager.AllPages(context.Background())
	require.NoError(t, err)

	assert.Len(t, rows, 2)
	assert.Equal(t, "aws", rows[0].Provider)
	assert.Equal(t, 100.0, rows[0].Cost)
	assert.Equal(t, "gcp", rows[1].Provider)
	assert.Equal(t, 200.0, rows[1].Cost)
	assert.Equal(t, 2, callCount)
}

func TestClient_ForecastRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call fails with 503
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Second call succeeds
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ForecastResponse{Data: []ForecastRow{}})
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:    server.URL,
		Token:      "test-token",
		Timeout:    time.Second * 5,
		MaxRetries: 1,
		Logger:     NewNoopLogger(),
	})
	require.NoError(t, err)

	query := ForecastQuery{
		StartAt:     time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		EndAt:       time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
		Granularity: "day",
	}

	_, err = client.Forecast(context.Background(), "test-report-token", query)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // Should have retried once
}

func TestRedactURL(t *testing.T) {
	// Create a test HTTP client directly
	httpClient := &httpClient{
		token:  "secret-token",
		logger: NewNoopLogger(),
	}

	// Test URL without sensitive data
	originalURL := "https://api.vantage.sh/costs?param=value"
	redacted := httpClient.redactURL(originalURL)
	assert.Equal(t, originalURL, redacted)

	// Test URL with Authorization parameter
	urlWithAuth := "https://api.vantage.sh/costs?Authorization=Bearer%20secret-token&param=value"
	redacted = httpClient.redactURL(urlWithAuth)
	expected := "https://api.vantage.sh/costs?Authorization=Bearer%20****&param=value"
	assert.Equal(t, expected, redacted)
}

// Example usage demonstration
func ExampleNew() {
	// Create a client with default configuration
	client, err := New(DefaultConfig("your-api-token"))
	if err != nil {
		panic(err)
	}

	// Use the client to fetch costs
	query := Query{
		CostReportToken: "cr_your_report_token",
		StartAt:         time.Now().AddDate(0, -1, 0), // 1 month ago
		EndAt:           time.Now(),
		Granularity:     "day",
		GroupBys:        []string{"provider", "service"},
		Metrics:         []string{"cost", "usage"},
	}

	page, err := client.Costs(context.Background(), query)
	if err != nil {
		panic(err)
	}

	// Process the results
	for _, row := range page.Data {
		fmt.Printf("Provider: %s, Cost: %.2f\n", row.Provider, row.Cost)
	}
}

func ExampleClient_Forecast() {
	client, err := New(DefaultConfig("your-api-token"))
	if err != nil {
		panic(err)
	}

	// Fetch forecast data
	query := ForecastQuery{
		StartAt:     time.Now(),
		EndAt:       time.Now().AddDate(0, 3, 0), // 3 months ahead
		Granularity: "month",
	}

	forecast, err := client.Forecast(context.Background(), "cr_your_report_token", query)
	if err != nil {
		panic(err)
	}

	// Process forecast results
	for _, row := range forecast.Data {
		fmt.Printf("Forecast cost: %.2f at %s\n", row.Cost, row.BucketStart.Format("2006-01-02"))
	}
}
