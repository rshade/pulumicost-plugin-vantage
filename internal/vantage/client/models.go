// Package client provides HTTP client functionality for Vantage API
package client

import (
	"time"
)

// Query represents parameters for the /costs endpoint
type Query struct {
	WorkspaceToken  string    `json:"workspace_token,omitempty"`
	CostReportToken string    `json:"cost_report_token,omitempty"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Granularity     string    `json:"granularity"` // "day" or "month"
	GroupBys        []string  `json:"group_bys"`
	Metrics         []string  `json:"metrics"`
	PageSize        int       `json:"page_size,omitempty"`
	Cursor          string    `json:"cursor,omitempty"`
}

// ForecastQuery represents parameters for the /forecast endpoint
type ForecastQuery struct {
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Granularity string    `json:"granularity"` // "day" or "month"
}

// CostRow represents a single cost data row from Vantage
type CostRow struct {
	Provider           string            `json:"provider,omitempty"`
	Service            string            `json:"service,omitempty"`
	Account            string            `json:"account,omitempty"`
	Project            string            `json:"project,omitempty"`
	Region             string            `json:"region,omitempty"`
	ResourceID         string            `json:"resource_id,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
	Cost               float64           `json:"cost,omitempty"`
	UsageQuantity      float64           `json:"usage_quantity,omitempty"`
	UsageUnit          string            `json:"usage_unit,omitempty"`
	EffectiveUnitPrice float64           `json:"effective_unit_price,omitempty"`
	ListCost           float64           `json:"list_cost,omitempty"`
	AmortizedCost      float64           `json:"amortized_cost,omitempty"`
	Tax                float64           `json:"tax,omitempty"`
	Credit             float64           `json:"credit,omitempty"`
	Refund             float64           `json:"refund,omitempty"`
	Currency           string            `json:"currency,omitempty"`
	BucketStart        time.Time         `json:"bucket_start"`
	BucketEnd          time.Time         `json:"bucket_end"`
}

// CostsResponse represents the response from /costs endpoint
type CostsResponse struct {
	Data       []CostRow `json:"data"`
	NextCursor string    `json:"next_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
}

// ForecastRow represents a single forecast data row
type ForecastRow struct {
	BucketStart time.Time `json:"bucket_start"`
	BucketEnd   time.Time `json:"bucket_end"`
	Cost        float64   `json:"cost"`
	Currency    string    `json:"currency,omitempty"`
}

// ForecastResponse represents the response from /forecast endpoint
type ForecastResponse struct {
	Data []ForecastRow `json:"data"`
}

// Page represents a page of cost data with pagination info
type Page struct {
	Data       []CostRow
	NextCursor string
	HasMore    bool
}

// Forecast represents forecast data
type Forecast struct {
	Data []ForecastRow
}
