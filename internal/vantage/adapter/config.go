// Package adapter provides configuration and mapping logic for the Vantage adapter.
package adapter

import (
	"time"
)

// Config holds the configuration for the Vantage adapter.
type Config struct {
	Token           string        `yaml:"token" json:"token"`
	WorkspaceToken  string        `yaml:"workspace_token,omitempty" json:"workspace_token,omitempty"`
	CostReportToken string        `yaml:"cost_report_token,omitempty" json:"cost_report_token,omitempty"`
	StartDate       time.Time     `yaml:"start_date" json:"start_date"`
	EndDate         *time.Time    `yaml:"end_date,omitempty" json:"end_date,omitempty"`
	Granularity     string        `yaml:"granularity" json:"granularity"`
	GroupBys        []string      `yaml:"group_bys" json:"group_bys"`
	Metrics         []string      `yaml:"metrics" json:"metrics"`
	IncludeForecast bool          `yaml:"include_forecast" json:"include_forecast"`
	PageSize        int           `yaml:"page_size" json:"page_size"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
}
