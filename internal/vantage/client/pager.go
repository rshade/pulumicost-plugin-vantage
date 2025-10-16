// Package client provides HTTP client functionality for Vantage API
package client

import (
	"context"
	"fmt"
)

// Pager provides cursor-based pagination for cost queries
type Pager struct {
	client Client
	query  Query
	logger Logger
}

// NewPager creates a new pager for the given query
func NewPager(client Client, query Query, logger Logger) *Pager {
	return &Pager{
		client: client,
		query:  query,
		logger: logger,
	}
}

// NextPage fetches the next page of cost data
func (p *Pager) NextPage(ctx context.Context) (Page, error) {
	currentQuery := p.query
	if p.query.Cursor != "" {
		currentQuery.Cursor = p.query.Cursor
	}

	page, err := p.client.Costs(ctx, currentQuery)
	if err != nil {
		p.logger.Error(ctx, "Failed to fetch costs page", map[string]interface{}{
			"error":  err,
			"cursor": currentQuery.Cursor,
		})
		return Page{}, fmt.Errorf("fetching costs page: %w", err)
	}

	// Update cursor for next page
	p.query.Cursor = page.NextCursor

	p.logger.Debug(ctx, "Fetched costs page", map[string]interface{}{
		"rows":        len(page.Data),
		"next_cursor": page.NextCursor,
		"has_more":    page.HasMore,
	})

	return page, nil
}

// HasMore returns true if there are more pages to fetch
func (p *Pager) HasMore() bool {
	return p.query.Cursor != ""
}

// AllPages fetches all pages and returns them as a single slice
// Note: This can be memory-intensive for large datasets
func (p *Pager) AllPages(ctx context.Context) ([]CostRow, error) {
	var allRows []CostRow

	for p.HasMore() || p.query.Cursor == "" {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		allRows = append(allRows, page.Data...)

		if !page.HasMore {
			break
		}
	}

	p.logger.Info(ctx, "Fetched all cost pages", map[string]interface{}{
		"total_rows": len(allRows),
	})
	return allRows, nil
}
