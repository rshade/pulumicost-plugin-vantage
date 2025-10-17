// Package adapter provides the Vantage adapter for PulumiCost.
package adapter

import (
	"regexp"
	"strings"
)

// normalizeTags normalizes tag keys and applies filtering.
func (a *Adapter) normalizeTags(tags map[string]string) map[string]string {
	if tags == nil {
		return nil
	}

	normalized := make(map[string]string)

	for key, value := range tags {
		// Normalize key to lower-kebab-case
		normalizedKey := a.normalizeTagKey(key)

		// Apply filters
		if a.shouldIncludeTag(normalizedKey, value) {
			normalized[normalizedKey] = value
		}
	}

	return normalized
}

// normalizeTagKey converts tag keys to lower-kebab-case.
func (a *Adapter) normalizeTagKey(key string) string {
	// Convert to lowercase
	key = strings.ToLower(key)

	// Replace underscores and spaces with hyphens
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.ReplaceAll(key, " ", "-")

	// Remove consecutive hyphens
	for strings.Contains(key, "--") {
		key = strings.ReplaceAll(key, "--", "-")
	}

	// Trim hyphens from start and end
	key = strings.Trim(key, "-")

	return key
}

// shouldIncludeTag determines if a tag should be included based on filters.
func (a *Adapter) shouldIncludeTag(key, _ string) bool {
	// Denylist high-cardinality patterns first
	denyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`.*pod.*uid.*`),      // Pod UIDs
		regexp.MustCompile(`.*container.*id.*`), // Container IDs
		regexp.MustCompile(`.*node.*name.*`),    // Node names (often high cardinality)
	}

	for _, pattern := range denyPatterns {
		if pattern.MatchString(key) {
			return false
		}
	}

	// Default allowlist (can be made configurable)
	allowPrefixes := []string{"user:", "kubernetes.io/"}

	// Check allowlist
	for _, prefix := range allowPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Allow other tags by default
	return true
}
