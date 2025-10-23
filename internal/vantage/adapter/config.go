// Package adapter provides configuration and mapping logic for the Vantage adapter.
package adapter

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
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

// rawConfig is an intermediate struct for unmarshaling YAML with flexible types
type rawConfig struct {
	Credentials map[string]interface{} `yaml:"credentials"`
	Params      map[string]interface{} `yaml:"params"`
}

// LoadConfig loads and parses the config from a YAML file, applying environment variable overrides.
func LoadConfig(filePath string) (*Config, error) {
	if filePath == "" {
		return nil, fmt.Errorf("config file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("config file not found: %s", filePath)
	}

	// Parse YAML file
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into intermediate struct
	var raw rawConfig
	if err := v.Unmarshal(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Extract credentials
	var token string
	if raw.Credentials != nil {
		if t, ok := raw.Credentials["token"].(string); ok {
			token = t
		}
	}

	// Apply environment variable override for token
	if envToken := os.Getenv("PULUMICOST_VANTAGE_TOKEN"); envToken != "" {
		token = envToken
	}

	// Extract params
	var workspaceToken, costReportToken, granularityStr, startDateStr, endDateStr string
	var groupBys, metrics []string
	var includeForecast bool
	var pageSize, requestTimeoutSeconds, maxRetries int

	if raw.Params != nil {
		if ws, ok := raw.Params["workspace_token"].(string); ok {
			workspaceToken = ws
		}
		if cr, ok := raw.Params["cost_report_token"].(string); ok {
			costReportToken = cr
		}
		if g, ok := raw.Params["granularity"].(string); ok {
			granularityStr = g
		}
		if s, ok := raw.Params["start_date"].(string); ok {
			startDateStr = s
		}
		if e, ok := raw.Params["end_date"].(string); ok {
			endDateStr = e
		}
		if gb, ok := raw.Params["group_bys"].([]interface{}); ok {
			for _, v := range gb {
				if str, ok := v.(string); ok {
					groupBys = append(groupBys, str)
				}
			}
		}
		if m, ok := raw.Params["metrics"].([]interface{}); ok {
			for _, v := range m {
				if str, ok := v.(string); ok {
					metrics = append(metrics, str)
				}
			}
		}
		if inc, ok := raw.Params["include_forecast"].(bool); ok {
			includeForecast = inc
		}
		if ps, ok := raw.Params["page_size"].(int); ok {
			pageSize = ps
		}
		if rts, ok := raw.Params["request_timeout_seconds"].(int); ok {
			requestTimeoutSeconds = rts
		}
		if mr, ok := raw.Params["max_retries"].(int); ok {
			maxRetries = mr
		}
	}

	// Parse dates with environment overrides
	var startDate time.Time
	if envStartDate := os.Getenv("PULUMICOST_VANTAGE_START_DATE"); envStartDate != "" {
		startDateStr = envStartDate
	}
	if startDateStr == "" {
		// Default: 12 months ago
		startDate = time.Now().AddDate(-1, 0, 0)
	} else {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD): %s", startDateStr)
		}
	}

	var endDate *time.Time
	if envEndDate := os.Getenv("PULUMICOST_VANTAGE_END_DATE"); envEndDate != "" {
		endDateStr = envEndDate
	}
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format (expected YYYY-MM-DD): %s", endDateStr)
		}
		endDate = &parsed
	}

	// Build Config struct
	cfg := &Config{
		Token:           token,
		WorkspaceToken:  workspaceToken,
		CostReportToken: costReportToken,
		StartDate:       startDate,
		EndDate:         endDate,
		Granularity:     granularityStr,
		GroupBys:        groupBys,
		Metrics:         metrics,
		IncludeForecast: includeForecast,
		PageSize:        pageSize,
		MaxRetries:      maxRetries,
	}

	// Set timeout (convert seconds to duration)
	if requestTimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(requestTimeoutSeconds) * time.Second
	} else {
		cfg.Timeout = 60 * time.Second // default
	}

	// Set page size default
	if cfg.PageSize <= 0 {
		cfg.PageSize = 5000
	}

	// Set max retries default
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}

	// Validate the config
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateConfig validates all configuration fields and returns clear error messages.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Token validation
	if cfg.Token == "" {
		return fmt.Errorf("credentials.token is required (set via YAML or PULUMICOST_VANTAGE_TOKEN environment variable)")
	}

	// At least one token type must be provided
	if cfg.WorkspaceToken == "" && cfg.CostReportToken == "" {
		return fmt.Errorf("either workspace_token or cost_report_token must be specified in params")
	}

	// Granularity validation
	if cfg.Granularity == "" {
		return fmt.Errorf("granularity must be specified in params")
	}
	if cfg.Granularity != "day" && cfg.Granularity != "month" {
		return fmt.Errorf("granularity must be 'day' or 'month', got: %s", cfg.Granularity)
	}

	// Start date validation
	if cfg.StartDate.IsZero() {
		return fmt.Errorf("start_date must be a valid ISO date (YYYY-MM-DD)")
	}

	// End date vs start date validation
	if cfg.EndDate != nil && cfg.EndDate.Before(cfg.StartDate) {
		return fmt.Errorf("end_date must not be before start_date")
	}

	// Page size validation
	if cfg.PageSize < 1 {
		return fmt.Errorf("page_size must be at least 1")
	}
	if cfg.PageSize > 10000 {
		return fmt.Errorf("page_size cannot exceed 10000")
	}

	// Timeout validation
	if cfg.Timeout < 1*time.Second {
		return fmt.Errorf("timeout must be at least 1 second")
	}

	// Max retries validation
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	// Group bys validation (should not be empty if specified)
	// Empty list is allowed (will use defaults), but if present should have valid values
	validGroupBys := map[string]bool{
		"provider":    true,
		"service":     true,
		"account":     true,
		"project":     true,
		"region":      true,
		"resource_id": true,
		"tags":        true,
	}
	for _, gb := range cfg.GroupBys {
		if !validGroupBys[gb] {
			return fmt.Errorf("invalid group_by value: %s (valid: provider, service, account, project, region, resource_id, tags)", gb)
		}
	}

	// Metrics validation
	validMetrics := map[string]bool{
		"cost":                  true,
		"usage":                 true,
		"effective_unit_price":  true,
		"amortized_cost":        true,
		"taxes":                 true,
		"credits":               true,
		"refunds":               true,
	}
	for _, m := range cfg.Metrics {
		if !validMetrics[m] {
			return fmt.Errorf("invalid metric value: %s (valid: cost, usage, effective_unit_price, amortized_cost, taxes, credits, refunds)", m)
		}
	}

	return nil
}

