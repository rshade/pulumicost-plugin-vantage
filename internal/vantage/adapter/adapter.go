// Package adapter provides the Vantage adapter for PulumiCost.
package adapter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rshade/pulumicost-plugin-vantage/internal/vantage/client"
)

// CostRecord represents a cost record in PulumiCost's internal schema with FOCUS 1.2 fields.
type CostRecord struct {
	// Core dimensions
	Timestamp      time.Time         `json:"timestamp"`
	Provider       string            `json:"provider,omitempty"`
	Service        string            `json:"service,omitempty"`
	AccountID      string            `json:"account_id,omitempty"`
	SubscriptionID string            `json:"subscription_id,omitempty"`
	Project        string            `json:"project,omitempty"`
	Region         string            `json:"region,omitempty"`
	ResourceID     string            `json:"resource_id,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`

	// Usage metrics
	UsageAmount *float64 `json:"usage_amount,omitempty"`
	UsageUnit   string   `json:"usage_unit,omitempty"`

	// Cost metrics
	ListCost      *float64 `json:"list_cost,omitempty"`
	NetCost       *float64 `json:"net_cost,omitempty"`
	AmortizedCost *float64 `json:"amortized_cost,omitempty"`
	TaxCost       *float64 `json:"tax_cost,omitempty"`
	CreditAmount  *float64 `json:"credit_amount,omitempty"`
	RefundAmount  *float64 `json:"refund_amount,omitempty"`

	// Metadata
	Currency          string `json:"currency,omitempty"`
	SourceReportToken string `json:"source_report_token,omitempty"`
	QueryHash         string `json:"query_hash"`
	LineItemID        string `json:"line_item_id"` // FOCUS 1.2 idempotency key (report_token, date, dimensions, metrics hash)
	MetricType        string `json:"metric_type,omitempty"` // "cost" or "forecast"

	// Diagnostics
	Diagnostics *Diagnostics `json:"diagnostics,omitempty"`
}

// Sink defines the interface for persisting cost records.
// This interface is assumed to exist in pulumicost-core.
type Sink interface {
	// WriteRecords writes cost records to the data store
	WriteRecords(ctx context.Context, records []CostRecord) error

	// GetBookmark retrieves the last successful sync bookmark
	GetBookmark(ctx context.Context, key string) (string, error)

	// SetBookmark stores the last successful sync bookmark
	SetBookmark(ctx context.Context, key string, value string) error
}

// Adapter implements the Vantage adapter for PulumiCost.
type Adapter struct {
	client client.Client
	logger client.Logger
}

// New creates a new Vantage adapter.
func New(client client.Client, logger client.Logger) *Adapter {
	return &Adapter{
		client: client,
		logger: logger,
	}
}

// Sync performs a cost data sync operation.
func (a *Adapter) Sync(ctx context.Context, cfg Config, sink Sink) error {
	a.logger.Info(ctx, "Starting Vantage adapter sync", map[string]interface{}{
		"operation": "sync",
		"source":    "vantage",
	})

	// Determine sync mode based on configuration
	if cfg.EndDate == nil {
		// Incremental sync: D-3 to D-1
		return a.syncIncremental(ctx, cfg, sink)
	}

	// Backfill sync: specified date range
	return a.syncBackfill(ctx, cfg, sink)
}

// syncIncremental performs incremental sync with D-3 to D-1 lag window.
func (a *Adapter) syncIncremental(ctx context.Context, cfg Config, sink Sink) error {
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -3) // D-3
	endDate := now.AddDate(0, 0, -1)   // D-1

	a.logger.Info(ctx, "Performing incremental sync", map[string]interface{}{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})

	return a.syncDateRange(ctx, cfg, sink, startDate, endDate, false)
}

// syncBackfill performs backfill sync for the specified date range.
func (a *Adapter) syncBackfill(ctx context.Context, cfg Config, sink Sink) error {
	startDate := cfg.StartDate
	endDate := *cfg.EndDate

	a.logger.Info(ctx, "Performing backfill sync", map[string]interface{}{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})

	return a.syncDateRange(ctx, cfg, sink, startDate, endDate, true)
}

// syncDateRange syncs data for a specific date range.
func (a *Adapter) syncDateRange(ctx context.Context, cfg Config, sink Sink, startDate, endDate time.Time, isBackfill bool) error {
	// For backfill, chunk by month to limit payload size
	if isBackfill && endDate.Sub(startDate).Hours() > 24*30 {
		return a.syncChunked(ctx, cfg, sink, startDate, endDate)
	}

	// Single range sync
	return a.syncSingleRange(ctx, cfg, sink, startDate, endDate, isBackfill)
}

// syncChunked performs chunked sync by month for large date ranges.
func (a *Adapter) syncChunked(ctx context.Context, cfg Config, sink Sink, startDate, endDate time.Time) error {
	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for current.Before(endDate) {
		chunkEnd := time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		if chunkEnd.After(endDate) {
			chunkEnd = endDate
		}

		if err := a.syncSingleRange(ctx, cfg, sink, current, chunkEnd, true); err != nil {
			return fmt.Errorf("syncing chunk %s to %s: %w", current.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), err)
		}

		current = chunkEnd
	}

	return nil
}

// syncSingleRange syncs a single date range.
func (a *Adapter) syncSingleRange(ctx context.Context, cfg Config, sink Sink, startDate, endDate time.Time, isBackfill bool) error {
	query := client.Query{
		WorkspaceToken:  cfg.WorkspaceToken,
		CostReportToken: cfg.CostReportToken,
		StartAt:         startDate,
		EndAt:           endDate,
		Granularity:     cfg.Granularity,
		GroupBys:        cfg.GroupBys,
		Metrics:         cfg.Metrics,
		PageSize:        cfg.PageSize,
	}

	// Generate idempotency key
	queryHash := a.generateQueryHash(query)

	// Check bookmark for incremental sync
	bookmarkKey := fmt.Sprintf("vantage_%s", queryHash)
	if !isBackfill {
		lastEndDate, err := sink.GetBookmark(ctx, bookmarkKey)
		if err == nil && lastEndDate != "" {
			// Resume from bookmark
			if parsed, err := time.Parse(time.RFC3339, lastEndDate); err == nil {
				query.StartAt = parsed
				a.logger.Info(ctx, "Resuming from bookmark", map[string]interface{}{
					"bookmark": lastEndDate,
				})
			}
		}
	}

	pager := client.NewPager(a.client, query, a.logger)

	var allRecords []CostRecord
	pageCount := 0

	for pager.HasMore() || pageCount == 0 {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("fetching page: %w", err)
		}

		// Convert Vantage rows to CostRecords
		for _, row := range page.Data {
			record := a.mapVantageRowToCostRecord(row, query, queryHash, "cost")
			allRecords = append(allRecords, record)
		}

		pageCount++
		if !page.HasMore {
			break
		}
	}

	a.logger.Info(ctx, "Fetched cost data", map[string]interface{}{
		"pages":      pageCount,
		"records":    len(allRecords),
		"query_hash": queryHash,
	})

	// Write records
	if err := sink.WriteRecords(ctx, allRecords); err != nil {
		return fmt.Errorf("writing records: %w", err)
	}

	// Update bookmark
	if !isBackfill {
		bookmarkValue := endDate.Format(time.RFC3339)
		if err := sink.SetBookmark(ctx, bookmarkKey, bookmarkValue); err != nil {
			a.logger.Warn(ctx, "Failed to update bookmark", map[string]interface{}{
				"error": err,
			})
		}
	}

	// Handle forecast if enabled
	if cfg.IncludeForecast && cfg.CostReportToken != "" {
		if err := a.syncForecast(ctx, cfg, sink, startDate, endDate, queryHash); err != nil {
			a.logger.Warn(ctx, "Forecast sync failed", map[string]interface{}{
				"error": err,
			})
		}
	}

	return nil
}

// syncForecast syncs forecast data for the given date range.
func (a *Adapter) syncForecast(ctx context.Context, cfg Config, sink Sink, startDate, endDate time.Time, queryHash string) error {
	forecastQuery := client.ForecastQuery{
		StartAt:     startDate,
		EndAt:       endDate,
		Granularity: cfg.Granularity,
	}

	forecast, err := a.client.Forecast(ctx, cfg.CostReportToken, forecastQuery)
	if err != nil {
		return fmt.Errorf("fetching forecast: %w", err)
	}

	var forecastRecords []CostRecord
	for _, row := range forecast.Data {
		record := a.mapVantageRowToCostRecord(client.CostRow{
			BucketStart: row.BucketStart,
			BucketEnd:   row.BucketEnd,
			Cost:        row.Cost,
			Currency:    row.Currency,
		}, client.Query{
			CostReportToken: cfg.CostReportToken,
			Granularity:     cfg.Granularity,
		}, queryHash, "forecast")
		forecastRecords = append(forecastRecords, record)
	}

	a.logger.Info(ctx, "Fetched forecast data", map[string]interface{}{
		"records":    len(forecastRecords),
		"query_hash": queryHash,
	})

	return sink.WriteRecords(ctx, forecastRecords)
}

// generateQueryHash creates a stable hash for idempotency.
func (a *Adapter) generateQueryHash(query client.Query) string {
	// Create a stable string representation
	parts := []string{
		query.WorkspaceToken,
		query.CostReportToken,
		query.StartAt.Format(time.RFC3339),
		query.EndAt.Format(time.RFC3339),
		query.Granularity,
	}

	// Sort groupbys and metrics for consistency
	groupBys := make([]string, len(query.GroupBys))
	copy(groupBys, query.GroupBys)
	sort.Strings(groupBys)
	parts = append(parts, strings.Join(groupBys, ","))

	metrics := make([]string, len(query.Metrics))
	copy(metrics, query.Metrics)
	sort.Strings(metrics)
	parts = append(parts, strings.Join(metrics, ","))

	// Generate hash
	hash := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", hash[:16]) // First 32 hex chars
}
