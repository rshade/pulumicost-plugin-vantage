package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiagnostics_AddMissingField(t *testing.T) {
	diag := NewDiagnostics()

	// Test adding missing fields.
	diag.AddMissingField("provider", "required field is empty")
	diag.AddMissingField("service", "required field is empty")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.MissingFields, 2)
	assert.Equal(t, "required field is empty", diag.MissingFields["provider"])
	assert.Equal(t, "required field is empty", diag.MissingFields["service"])
}

func TestDiagnostics_AddWarning(t *testing.T) {
	diag := NewDiagnostics()

	// Test adding warnings.
	diag.AddWarning("negative_cost")
	diag.AddWarning("missing_unit")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.Warnings, 2)
	assert.Contains(t, diag.Warnings, "negative_cost")
	assert.Contains(t, diag.Warnings, "missing_unit")
}

func TestDiagnostics_SetSourceInfo(t *testing.T) {
	diag := NewDiagnostics()

	// Test setting source info.
	diag.SetSourceInfo("api_version", "v1")
	diag.SetSourceInfo("record_count", 100)

	assert.Equal(t, "v1", diag.SourceInfo["api_version"])
	assert.Equal(t, 100, diag.SourceInfo["record_count"])
}

func TestDiagnostics_HasIssues(t *testing.T) {
	diag := NewDiagnostics()

	// Initially no issues.
	assert.False(t, diag.HasIssues())

	// Add missing field.
	diag.AddMissingField("provider", "required field is empty")
	assert.True(t, diag.HasIssues())

	// Reset and test with warning.
	diag = NewDiagnostics()
	diag.AddWarning("negative_cost")
	assert.True(t, diag.HasIssues())

	// Reset and test with both.
	diag = NewDiagnostics()
	diag.AddMissingField("provider", "required field is empty")
	diag.AddWarning("negative_cost")
	assert.True(t, diag.HasIssues())
}

func TestDiagnosticsSummary_AddRecordDiagnostics(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Test with nil diagnostics.
	summary.AddRecordDiagnostics(nil)
	assert.Equal(t, 1, summary.TotalRecords)
	assert.Equal(t, 0, summary.RecordsWithIssues)

	// Test with diagnostics that have issues.
	diag1 := NewDiagnostics()
	diag1.AddMissingField("provider", "required field is empty")
	diag1.AddWarning("negative_cost")

	summary.AddRecordDiagnostics(diag1)
	assert.Equal(t, 2, summary.TotalRecords)
	assert.Equal(t, 1, summary.RecordsWithIssues)
	assert.Equal(t, 1, summary.MissingFields["provider"])
	assert.Equal(t, 1, summary.Warnings["negative_cost"])

	// Test with another record with same issues.
	diag2 := NewDiagnostics()
	diag2.AddMissingField("provider", "required field is empty")
	diag2.AddWarning("negative_cost")

	summary.AddRecordDiagnostics(diag2)
	assert.Equal(t, 3, summary.TotalRecords)
	assert.Equal(t, 2, summary.RecordsWithIssues)
	assert.Equal(t, 2, summary.MissingFields["provider"])
	assert.Equal(t, 2, summary.Warnings["negative_cost"])
}

func TestDiagnosticsSummary_AddRecordDiagnostics_DifferentIssues(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Record 1: missing provider.
	diag1 := NewDiagnostics()
	diag1.AddMissingField("provider", "required field is empty")
	summary.AddRecordDiagnostics(diag1)

	// Record 2: missing service.
	diag2 := NewDiagnostics()
	diag2.AddMissingField("service", "required field is empty")
	summary.AddRecordDiagnostics(diag2)

	// Record 3: warning only.
	diag3 := NewDiagnostics()
	diag3.AddWarning("negative_cost")
	summary.AddRecordDiagnostics(diag3)

	// Record 4: no issues.
	diag4 := NewDiagnostics()
	summary.AddRecordDiagnostics(diag4)

	assert.Equal(t, 4, summary.TotalRecords)
	assert.Equal(t, 3, summary.RecordsWithIssues)
	assert.Equal(t, 1, summary.MissingFields["provider"])
	assert.Equal(t, 1, summary.MissingFields["service"])
	assert.Equal(t, 1, summary.Warnings["negative_cost"])
}

func TestDiagnosticsSummary_HasIssues(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Initially no issues.
	assert.False(t, summary.HasIssues())

	// Add record with no issues.
	summary.AddRecordDiagnostics(NewDiagnostics())
	assert.False(t, summary.HasIssues())

	// Add record with issues.
	diag := NewDiagnostics()
	diag.AddMissingField("provider", "required field is empty")
	summary.AddRecordDiagnostics(diag)
	assert.True(t, summary.HasIssues())
}

func TestDiagnosticsSummary_SourceInfoMerging(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Record 1 with source info.
	diag1 := NewDiagnostics()
	diag1.SetSourceInfo("api_version", "v1")
	diag1.SetSourceInfo("source", "api")
	summary.AddRecordDiagnostics(diag1)

	assert.Equal(t, "v1", summary.SourceInfo["api_version"])
	assert.Equal(t, "api", summary.SourceInfo["source"])

	// Record 2 with different source info (should overwrite).
	diag2 := NewDiagnostics()
	diag2.SetSourceInfo("api_version", "v2")
	diag2.SetSourceInfo("region", "us-east-1")
	summary.AddRecordDiagnostics(diag2)

	assert.Equal(t, "v2", summary.SourceInfo["api_version"])   // overwritten
	assert.Equal(t, "api", summary.SourceInfo["source"])       // preserved
	assert.Equal(t, "us-east-1", summary.SourceInfo["region"]) // added
}

func TestDiagnostics_ComplexScenario(t *testing.T) {
	// Test a complex scenario with multiple types of issues.
	diag := NewDiagnostics()

	// Add multiple missing fields.
	diag.AddMissingField("provider", "required field is empty")
	diag.AddMissingField("service", "required field is empty")
	diag.AddMissingField("net_cost", "required field is nil or zero")

	// Add multiple warnings.
	diag.AddWarning("negative_cost")
	diag.AddWarning("usage_amount_present_but_unit_missing")
	diag.AddWarning("unusual_cost_pattern")

	// Add source info.
	diag.SetSourceInfo("api_version", "v1.2.3")
	diag.SetSourceInfo("data_quality_score", 85)

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.MissingFields, 3)
	assert.Len(t, diag.Warnings, 3)
	assert.Len(t, diag.SourceInfo, 2)

	// Verify specific content.
	assert.Equal(t, "required field is empty", diag.MissingFields["provider"])
	assert.Equal(t, "required field is nil or zero", diag.MissingFields["net_cost"])
	assert.Contains(t, diag.Warnings, "negative_cost")
	assert.Contains(t, diag.Warnings, "usage_amount_present_but_unit_missing")
	assert.Equal(t, "v1.2.3", diag.SourceInfo["api_version"])
	assert.Equal(t, 85, diag.SourceInfo["data_quality_score"])
}
