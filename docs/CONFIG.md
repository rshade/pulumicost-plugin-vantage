# Configuration Reference

This document describes the configuration options for the PulumiCost Vantage plugin.

## Configuration File Format

The plugin accepts configuration via YAML files. Environment variable
substitution is supported using `${VAR_NAME}` syntax.

## Example Configuration

```yaml
version: 0.1
source: vantage
credentials:
  token: ${PULUMICOST_VANTAGE_TOKEN}
params:
  workspace_token: "ws_..."         # optional if using cost_report_token
  cost_report_token: "cr_..."       # preferred for stable queries
  start_date: "2024-01-01"          # ISO date, default: 12 months back
  end_date: null                     # default: today
  granularity: "day"                # day|month
  group_bys: ["provider","service","account","project","region","resource_id","tags"]
  metrics: ["cost","usage","effective_unit_price"]
  include_forecast: true
  tag_prefix_filters: ["user:","kubernetes.io/"]
  request_timeout_seconds: 60
  page_size: 5000
  max_retries: 5
```

## Configuration Parameters

### credentials.token

- **Type**: `string`
- **Required**: Yes
- **Description**: Vantage API token. Can be a service token or user token.
- **Environment Variable**: `PULUMICOST_VANTAGE_TOKEN`
- **Security**: Never logged or printed in output.

### params.workspace_token

- **Type**: `string`
- **Required**: No (required if `cost_report_token` not provided)
- **Description**: Workspace token for accessing Vantage workspace-level APIs.

### params.cost_report_token

- **Type**: `string`
- **Required**: No (required if `workspace_token` not provided)
- **Description**: Cost Report token for stable, curated cost queries.
  Preferred over workspace token.

### params.start_date

- **Type**: `string` (ISO date format: YYYY-MM-DD)
- **Required**: No
- **Default**: 12 months ago
- **Description**: Start date for cost data retrieval.

### params.end_date

- **Type**: `string` (ISO date format: YYYY-MM-DD) or `null`
- **Required**: No
- **Default**: Today
- **Description**: End date for cost data retrieval. Set to `null` for current date.

### params.granularity

- **Type**: `string`
- **Required**: No
- **Default**: `"day"`
- **Allowed Values**: `"day"`, `"month"`
- **Description**: Time granularity for cost aggregation.

### params.group_bys

- **Type**: `array` of `string`
- **Required**: No
- **Default**: `["provider","service","account","project","region","resource_id","tags"]`
- **Description**: Dimensions to group cost data by. Available dimensions
  depend on Vantage configuration.

### params.metrics

- **Type**: `array` of `string`
- **Required**: No
- **Default**: `["cost","usage","effective_unit_price"]`
- **Description**: Metrics to retrieve. Common values: `cost`, `usage`,
  `effective_unit_price`, `amortized_cost`, `taxes`, `credits`.

### params.include_forecast

- **Type**: `boolean`
- **Required**: No
- **Default**: `true`
- **Description**: Whether to include forecast snapshots in sync operations.

### params.tag_prefix_filters

- **Type**: `array` of `string`
- **Required**: No
- **Default**: `["user:","kubernetes.io/"]`
- **Description**: Tag prefixes to include in processing. Used for filtering
  high-cardinality tags.

### params.request_timeout_seconds

- **Type**: `integer`
- **Required**: No
- **Default**: `60`
- **Description**: HTTP request timeout in seconds.

### params.page_size

- **Type**: `integer`
- **Required**: No
- **Default**: `5000`
- **Description**: Number of records per API page request.

### params.max_retries

- **Type**: `integer`
- **Required**: No
- **Default**: `5`
- **Description**: Maximum number of retry attempts for failed API requests.

## Authentication

The plugin supports authentication via API tokens provided through the `token`
configuration field or `PULUMICOST_VANTAGE_TOKEN` environment variable.

- **Cost Report Token**: Preferred for stable, curated queries with predefined
  filters
- **Workspace Token**: Required for broader workspace access when Cost Report
  tokens are not available

## Environment Variables

All configuration values can be overridden using environment variables:

- `PULUMICOST_VANTAGE_TOKEN`: API token
- `PULUMICOST_VANTAGE_WORKSPACE_TOKEN`: Workspace token
- `PULUMICOST_VANTAGE_COST_REPORT_TOKEN`: Cost Report token
- `PULUMICOST_VANTAGE_START_DATE`: Start date
- `PULUMICOST_VANTAGE_END_DATE`: End date
- `PULUMICOST_VANTAGE_GRANULARITY`: Granularity
- `PULUMICOST_VANTAGE_TIMEOUT`: Request timeout in seconds
- `PULUMICOST_VANTAGE_PAGE_SIZE`: Page size
- `PULUMICOST_VANTAGE_MAX_RETRIES`: Max retries

## Notes

- Token values are never logged or exposed in error messages
- The adapter prefers Cost Report tokens for stable queries
- Missing fields in Vantage responses are handled gracefully with null values
- Tag normalization converts keys to lowercase kebab-case
- Incremental sync uses bookmarks to track last successful sync dates
