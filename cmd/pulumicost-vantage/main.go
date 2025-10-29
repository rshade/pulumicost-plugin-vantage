// Package main provides the CLI entry point for the PulumiCost Vantage plugin.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

const (
	defaultBackfillMonths = 12
)

func buildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pulumicost-vantage",
		Short: "PulumiCost Vantage adapter for fetching cost data",
		Long: `A Go-based adapter that fetches normalized cost/usage data from Vantage's REST API
and maps it into PulumiCost's internal schema with FinOps FOCUS 1.2 fields.`,
		Version: version,
	}

	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Perform incremental cost data sync",
		Long:  `Fetch cost data incrementally using bookmarks. Defaults to D-3 to D-1 lag window.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// TODO: implement pull logic
			return errors.New("pull command not yet implemented")
		},
	}

	backfillCmd := &cobra.Command{
		Use:   "backfill",
		Short: "Backfill historical cost data",
		Long:  `Fetch historical cost data for a specified number of months.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// TODO: implement backfill logic
			return errors.New("backfill command not yet implemented")
		},
	}

	forecastCmd := &cobra.Command{
		Use:   "forecast",
		Short: "Generate forecast snapshot",
		Long:  `Fetch and store forecast data as a separate metric family.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// TODO: implement forecast logic
			return errors.New("forecast command not yet implemented")
		},
	}

	// Add common flags
	rootCmd.PersistentFlags().String("config", "", "Path to configuration file")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		panic(err)
	}

	// Add commands
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(backfillCmd)
	rootCmd.AddCommand(forecastCmd)

	// Add command-specific flags
	backfillCmd.Flags().Int("months", defaultBackfillMonths, "Number of months to backfill")

	return rootCmd
}

func main() {
	ctx := context.Background()
	rootCmd := buildRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
