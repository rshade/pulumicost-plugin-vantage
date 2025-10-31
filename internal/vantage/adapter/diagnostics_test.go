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

func TestDiagnostics_MissingAccountID(t *testing.T) {
	diag := NewDiagnostics()

	// Test adding missing account_id field.
	diag.AddMissingField("account_id", "FOCUS 1.2 field billing_account_id is empty")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.MissingFields, 1)
	assert.Equal(t, "FOCUS 1.2 field billing_account_id is empty", diag.MissingFields["account_id"])
}

func TestDiagnostics_MissingRegion(t *testing.T) {
	diag := NewDiagnostics()

	// Test adding missing region field.
	diag.AddMissingField("region", "FOCUS 1.2 field region is empty")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.MissingFields, 1)
	assert.Equal(t, "FOCUS 1.2 field region is empty", diag.MissingFields["region"])
}

func TestDiagnostics_MissingCurrency(t *testing.T) {
	diag := NewDiagnostics()

	// Test adding missing currency field.
	diag.AddMissingField("currency", "FOCUS 1.2 field billing_currency is empty")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.MissingFields, 1)
	assert.Equal(t, "FOCUS 1.2 field billing_currency is empty", diag.MissingFields["currency"])
}

func TestDiagnostics_UsageUnitWithoutAmount(t *testing.T) {
	diag := NewDiagnostics()

	// Test warning for usage_unit without usage_amount.
	diag.AddWarning("usage_unit_present_but_amount_missing")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.Warnings, 1)
	assert.Contains(t, diag.Warnings, "usage_unit_present_but_amount_missing")
}

func TestDiagnostics_ListCostLessThanNetCost(t *testing.T) {
	diag := NewDiagnostics()

	// Test warning for list_cost less than net_cost.
	diag.AddWarning("list_cost_less_than_net_cost")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.Warnings, 1)
	assert.Contains(t, diag.Warnings, "list_cost_less_than_net_cost")
}

func TestDiagnostics_MissingResourceID(t *testing.T) {
	diag := NewDiagnostics()

	// Test warning for missing resource_id.
	diag.AddWarning("missing_resource_id")

	assert.True(t, diag.HasIssues())
	assert.Len(t, diag.Warnings, 1)
	assert.Contains(t, diag.Warnings, "missing_resource_id")
}

func TestDiagnosticsSummary_MultipleMissingFieldTypes(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Record 1: missing provider and account_id.
	diag1 := NewDiagnostics()
	diag1.AddMissingField("provider", "required FOCUS 1.2 field cloud_provider is empty")
	diag1.AddMissingField("account_id", "FOCUS 1.2 field billing_account_id is empty")
	summary.AddRecordDiagnostics(diag1)

	// Record 2: missing region and currency.
	diag2 := NewDiagnostics()
	diag2.AddMissingField("region", "FOCUS 1.2 field region is empty")
	diag2.AddMissingField("currency", "FOCUS 1.2 field billing_currency is empty")
	summary.AddRecordDiagnostics(diag2)

	// Record 3: multiple warnings.
	diag3 := NewDiagnostics()
	diag3.AddWarning("negative_net_cost")
	diag3.AddWarning("list_cost_less_than_net_cost")
	diag3.AddWarning("missing_resource_id")
	summary.AddRecordDiagnostics(diag3)

	// Verify summary aggregation.
	assert.Equal(t, 3, summary.TotalRecords)
	assert.Equal(t, 3, summary.RecordsWithIssues)
	assert.Equal(t, 1, summary.MissingFields["provider"])
	assert.Equal(t, 1, summary.MissingFields["account_id"])
	assert.Equal(t, 1, summary.MissingFields["region"])
	assert.Equal(t, 1, summary.MissingFields["currency"])
	assert.Equal(t, 1, summary.Warnings["negative_net_cost"])
	assert.Equal(t, 1, summary.Warnings["list_cost_less_than_net_cost"])
	assert.Equal(t, 1, summary.Warnings["missing_resource_id"])
}

func TestDiagnosticsSummary_LargeScaleAggregation(t *testing.T) {
	summary := NewDiagnosticsSummary()

	// Simulate processing 100 records with various issues.
	for i := range 100 {
		diag := NewDiagnostics()

		// Every 10th record has missing provider.
		if i%10 == 0 {
			diag.AddMissingField("provider", "required FOCUS 1.2 field cloud_provider is empty")
		}

		// Every 5th record has missing currency.
		if i%5 == 0 {
			diag.AddMissingField("currency", "FOCUS 1.2 field billing_currency is empty")
		}

		// Every 7th record has negative cost warning.
		if i%7 == 0 {
			diag.AddWarning("negative_net_cost")
		}

		summary.AddRecordDiagnostics(diag)
	}

	// Verify aggregation counts.
	assert.Equal(t, 100, summary.TotalRecords)
	assert.Positive(t, summary.RecordsWithIssues)
	assert.Equal(t, 10, summary.MissingFields["provider"])
	assert.Equal(t, 20, summary.MissingFields["currency"])
	assert.Equal(t, 15, summary.Warnings["negative_net_cost"])
}
