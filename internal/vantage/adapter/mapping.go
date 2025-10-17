// Package adapter provides the Vantage adapter for PulumiCost.
package adapter

import (
	"github.com/rshade/pulumicost-plugin-vantage/internal/vantage/client"
)

// mapVantageRowToCostRecord converts a Vantage CostRow to a PulumiCost CostRecord.
func (a *Adapter) mapVantageRowToCostRecord(row client.CostRow, query client.Query, queryHash, metricType string) CostRecord {
	record := CostRecord{
		Timestamp:         row.BucketStart,
		Provider:          row.Provider,
		Service:           row.Service,
		AccountID:         row.Account,
		Project:           row.Project,
		Region:            row.Region,
		ResourceID:        row.ResourceID,
		Currency:          row.Currency,
		SourceReportToken: query.CostReportToken,
		QueryHash:         queryHash,
		MetricType:        metricType,
		Diagnostics:       &Diagnostics{},
	}

	// Map usage metrics
	if row.UsageQuantity != 0 {
		record.UsageAmount = &row.UsageQuantity
	}
	record.UsageUnit = row.UsageUnit

	// Map cost metrics
	if row.ListCost != 0 {
		record.ListCost = &row.ListCost
	}
	if row.Cost != 0 {
		record.NetCost = &row.Cost
	}
	if row.AmortizedCost != 0 {
		record.AmortizedCost = &row.AmortizedCost
	}
	if row.Tax != 0 {
		record.TaxCost = &row.Tax
	}
	if row.Credit != 0 {
		record.CreditAmount = &row.Credit
	}
	if row.Refund != 0 {
		record.RefundAmount = &row.Refund
	}

	// Normalize and map tags
	record.Labels = a.normalizeTags(row.Tags)

	// Add diagnostics for missing fields
	a.addDiagnostics(&record, row)

	return record
}

// addDiagnostics adds diagnostic information for missing or problematic fields.
func (a *Adapter) addDiagnostics(record *CostRecord, _ client.CostRow) {
	diag := record.Diagnostics

	// Check for missing required fields
	if record.Provider == "" {
		diag.MissingFields = append(diag.MissingFields, "provider")
	}
	if record.Service == "" {
		diag.MissingFields = append(diag.MissingFields, "service")
	}
	if record.NetCost == nil || *record.NetCost == 0 {
		diag.MissingFields = append(diag.MissingFields, "net_cost")
	}

	// Check for zero values that might indicate data issues
	if record.UsageAmount != nil && *record.UsageAmount == 0 && record.UsageUnit == "" {
		diag.Warnings = append(diag.Warnings, "usage_amount_present_but_unit_missing")
	}

	// Check for unusual cost values
	if record.NetCost != nil && *record.NetCost < 0 {
		diag.Warnings = append(diag.Warnings, "negative_net_cost")
	}

	// If no diagnostics were added, set to nil
	if len(diag.MissingFields) == 0 && len(diag.Warnings) == 0 {
		record.Diagnostics = nil
	}
}
