# Forecast Snapshots

This document explains how the PulumiCost Vantage plugin handles forecast
data and snapshots.

## Overview

The plugin supports optional forecast functionality that retrieves cost
projections from Vantage's `/cost_reports/{token}/forecast` endpoint. Forecast
data is stored as a separate metric family to distinguish it from actual
historical cost data.

## How Forecast Works

### Forecast API

The plugin calls Vantage's forecast endpoint to retrieve projected cost data:

```http
GET /cost_reports/{token}/forecast
Authorization: Bearer <token>
```

### Data Structure

Forecast records are stored with these characteristics:

- **`metric_type`**: Set to `"forecast"` to distinguish from actual costs
- **Time range**: Covers future periods (typically 3-12 months ahead)
- **Dimensions**: Same grouping dimensions as cost data
- **Metrics**: Projected costs, usage, and other metrics
- **Source identification**: Links back to the original cost report

### Storage Format

Forecast data follows the same FOCUS 1.2 schema as regular cost data but includes:

```json
{
  "timestamp": "2025-01-01T00:00:00Z",
  "metric_type": "forecast",
  "provider": "aws",
  "service": "ec2",
  "account_id": "123456789",
  "net_cost": 1500.00,
  "usage_amount": 720.0,
  "source_report_token": "cr_xxx",
  "forecast_generated_at": "2024-10-01T00:00:00Z"
}
```

## Snapshot Schedule

### Frequency

- **Default**: Weekly snapshots
- **Configurable**: Can be adjusted based on requirements
- **Trigger**: Manual via CLI or automated scheduling

### Retention

- **Default**: Keep last 8 snapshots
- **Purpose**: Enable MAPE (Mean Absolute Percentage Error) evaluation
- **Cleanup**: Automatic removal of older snapshots

## Usage

### CLI Command

Generate a forecast snapshot:

```bash
./bin/pulumicost-vantage forecast --config ./config.yaml --out ./data/forecast.json
```

### Configuration

Enable forecast in your configuration:

```yaml
params:
  include_forecast: true
  cost_report_token: "cr_xxx"  # Required for forecast
  forecast_snapshot_frequency: "weekly"  # Optional
  forecast_retention_count: 8  # Optional
```

### Integration with Sync

Forecast can be included in regular sync operations:

```yaml
params:
  include_forecast: true  # Include forecast in pull operations
```

## Output Locations

### File Output

When using the `forecast` command with `--out`:

```text
./data/forecast.json  # JSON format
./data/forecast.parquet  # Parquet format (if supported)
```

### Database Storage

Forecast records are stored in the same data store as regular cost data but
tagged with `metric_type = "forecast"`.

### Sink Integration

Forecast data is written through the same Sink interface as cost data,
allowing flexible storage options (database, parquet files, etc.).

## Analysis and Reporting

### MAPE Evaluation

Compare forecast accuracy using the retained snapshots:

- **Current forecast** vs **actual costs** when they become available
- **Historical forecasts** vs **realized costs**
- **Trend analysis** for forecast quality improvement

### Reporting

Forecast data enables:

- **Budget planning** with projected costs
- **Cost trend analysis** including future projections
- **Anomaly detection** comparing forecast vs actual
- **Scenario planning** for cost optimization decisions

## Limitations

### Data Availability

- Forecast requires a Cost Report token (not available with Workspace tokens)
- Forecast data may not be available for all Vantage configurations
- Forecast accuracy depends on historical data quality and patterns

### Time Horizons

- Typical forecast horizon: 3-12 months
- Granularity: Usually monthly for longer forecasts
- Confidence intervals: May not always be provided

### Performance Considerations

- Forecast API calls may be slower than regular cost queries
- Large forecast datasets may require pagination
- Additional storage space for multiple snapshots

## Troubleshooting

### Common Issues

#### "Forecast not available"

- Verify Cost Report token is configured
- Check if forecast is enabled in Vantage workspace
- Confirm sufficient historical data exists

#### "Empty forecast results"

- Check date ranges (forecast covers future periods)
- Verify report configuration includes forecast-enabled services
- Review Vantage documentation for forecast requirements

#### "Forecast accuracy issues"

- Ensure sufficient historical data (minimum 3-6 months recommended)
- Check for significant cost pattern changes
- Consider adjusting forecast parameters in Vantage

### Debug Logging

Enable verbose logging to inspect forecast API calls:

```bash
export PULUMICOST_VANTAGE_VERBOSE=true
./bin/pulumicost-vantage forecast --config config.yaml
```

Look for log entries with `operation=forecast` for troubleshooting.

## Best Practices

### Configuration Best Practices

1. **Enable selectively**: Only enable forecast when needed to reduce
   API load
2. **Schedule appropriately**: Weekly snapshots typically sufficient for
   most use cases
3. **Monitor accuracy**: Regularly review forecast vs actual cost comparisons

### Operations

1. **Separate storage**: Consider isolating forecast data for analysis
2. **Version control**: Track forecast snapshot versions for audit trails
3. **Alerting**: Set up alerts for forecast generation failures

### Analysis

1. **Compare regularly**: Use MAPE calculations to assess forecast quality
2. **Adjust parameters**: Refine Vantage forecast settings based on accuracy metrics
3. **Document assumptions**: Record any manual adjustments or overrides
