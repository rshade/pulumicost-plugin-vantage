package adapter

// Diagnostics provides diagnostic information about data quality and mapping issues.
type Diagnostics struct {
	// MissingFields maps field names to reasons why they are missing or invalid.
	MissingFields map[string]string `json:"missing_fields,omitempty"`

	// Warnings lists non-critical issues or unusual data patterns.
	Warnings []string `json:"warnings,omitempty"`

	// SourceInfo provides information about the data source.
	SourceInfo map[string]interface{} `json:"source_info,omitempty"`
}

// NewDiagnostics creates a new diagnostics instance.
func NewDiagnostics() *Diagnostics {
	return &Diagnostics{
		MissingFields: make(map[string]string),
		Warnings:      make([]string, 0),
		SourceInfo:    make(map[string]interface{}),
	}
}

// AddMissingField adds a missing field with a reason to the diagnostics.
func (d *Diagnostics) AddMissingField(field, reason string) {
	if d.MissingFields == nil {
		d.MissingFields = make(map[string]string)
	}
	d.MissingFields[field] = reason
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

// DiagnosticsSummary aggregates diagnostic information across multiple records.
type DiagnosticsSummary struct {
	// TotalRecords is the total number of records processed.
	TotalRecords int `json:"total_records"`

	// RecordsWithIssues is the number of records that had diagnostic issues.
	RecordsWithIssues int `json:"records_with_issues"`

	// MissingFields maps field names to the number of records missing that field.
	MissingFields map[string]int `json:"missing_fields,omitempty"`

	// Warnings maps warning types to the number of occurrences.
	Warnings map[string]int `json:"warnings,omitempty"`

	// SourceInfo provides aggregated information about data sources.
	SourceInfo map[string]interface{} `json:"source_info,omitempty"`
}

// NewDiagnosticsSummary creates a new diagnostics summary.
func NewDiagnosticsSummary() *DiagnosticsSummary {
	return &DiagnosticsSummary{
		MissingFields: make(map[string]int),
		Warnings:      make(map[string]int),
		SourceInfo:    make(map[string]interface{}),
	}
}

// AddRecordDiagnostics adds diagnostics from a single record to the summary.
func (ds *DiagnosticsSummary) AddRecordDiagnostics(diag *Diagnostics) {
	ds.TotalRecords++

	if diag != nil {
		if diag.HasIssues() {
			ds.RecordsWithIssues++
		}

		for field := range diag.MissingFields {
			ds.MissingFields[field]++
		}

		for _, warning := range diag.Warnings {
			ds.Warnings[warning]++
		}

		// Merge source info (last one wins for now).
		for key, value := range diag.SourceInfo {
			ds.SourceInfo[key] = value
		}
	}
}

// HasIssues returns true if any records had issues.
func (ds *DiagnosticsSummary) HasIssues() bool {
	return ds.RecordsWithIssues > 0
}
