// Package adapter provides the Vantage adapter for PulumiCost.
package adapter

// Diagnostics provides diagnostic information about data quality and mapping issues.
type Diagnostics struct {
	// MissingFields lists fields that were expected but not present in the source data
	MissingFields []string `json:"missing_fields,omitempty"`

	// Warnings lists non-critical issues or unusual data patterns
	Warnings []string `json:"warnings,omitempty"`

	// SourceInfo provides information about the data source
	SourceInfo map[string]interface{} `json:"source_info,omitempty"`
}

// NewDiagnostics creates a new diagnostics instance.
func NewDiagnostics() *Diagnostics {
	return &Diagnostics{
		MissingFields: make([]string, 0),
		Warnings:      make([]string, 0),
		SourceInfo:    make(map[string]interface{}),
	}
}

// AddMissingField adds a missing field to the diagnostics.
func (d *Diagnostics) AddMissingField(field string) {
	d.MissingFields = append(d.MissingFields, field)
}

// AddWarning adds a warning to the diagnostics.
func (d *Diagnostics) AddWarning(warning string) {
	d.Warnings = append(d.Warnings, warning)
}

// SetSourceInfo sets source information.
func (d *Diagnostics) SetSourceInfo(key string, value interface{}) {
	if d.SourceInfo == nil {
		d.SourceInfo = make(map[string]interface{})
	}
	d.SourceInfo[key] = value
}

// HasIssues returns true if there are any missing fields or warnings.
func (d *Diagnostics) HasIssues() bool {
	return len(d.MissingFields) > 0 || len(d.Warnings) > 0
}
