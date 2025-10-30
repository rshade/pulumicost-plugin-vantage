# PulumiCost Vantage Plugin

A Go-based adapter that fetches normalized cost/usage data from Vantage's
REST API and maps it into PulumiCost's internal schema with FinOps FOCUS
1.2 fields.

## Features

- Fetch costs via `/costs` endpoint using Cost Report tokens or Workspace
  tokens
- Support for daily granularity with common dimension grouping (provider,
  service, account, project, region, resource_id, tags)
- Capture list, net, and amortized costs with taxes, credits, and refunds
- Incremental sync with bookmarks and rate limit backoff
- Forecast snapshot support
- FOCUS 1.2 compatible records
- Comprehensive error handling and observability

## Limitations

- Read-only adapter (cannot create or modify Vantage resources)
- No direct cost optimization recommendations (handled by PulumiCost analyzers)
- Forecast functionality requires Cost Report tokens
- Rate limiting may affect large data syncs
- Tag cardinality limits may require filtering for performance

## Quick Start

### Prerequisites

- Go 1.24.7+
- `make`
- Docker (for running mock tests)
- `golangci-lint` (for linting)

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Lint Code

```bash
make lint
```

### Format Code

```bash
make fmt
```

## Configuration

See [docs/CONFIG.md](docs/CONFIG.md) for detailed configuration options.

### Basic Example

```yaml
version: 0.1
source: vantage
credentials:
  token: ${PULUMICOST_VANTAGE_TOKEN}
params:
  cost_report_token: "cr_..."
  start_date: "2024-01-01"
  granularity: "day"
  group_bys: ["provider", "service", "account", "region"]
  metrics: ["cost", "usage"]
  include_forecast: true
```

## CLI Commands

```bash
# Backfill last 12 months
./bin/pulumicost-vantage backfill --config ./config.yaml --months 12

# Daily incremental sync
./bin/pulumicost-vantage pull --config ./config.yaml

# Forecast snapshot
./bin/pulumicost-vantage forecast --config ./config.yaml --out ./data/forecast.json
```

## Testing with Mock Server

```bash
# Start Wiremock mock server
make wiremock-up

# Run tests against mock
make demo

# Stop mock server
make wiremock-down
```

## Documentation

- [Configuration Reference](docs/CONFIG.md)
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- [Forecast Snapshots](docs/FORECAST.md)
- [Design Document](pulumi_cost_vantage_adapter_design_draft_v_0.md)

## Development

### Project Structure

```text
cmd/pulumicost-vantage/        # CLI entry point
internal/vantage/
  ├── client/                  # REST client
  ├── adapter/                 # Mapping and sync logic
  └── contracts/               # Test fixtures
test/wiremock/                 # Mock server configs
docs/                          # Documentation
prompts/                       # OpenCode prompts
```

### Code Style

- Go conventions with 1.24.7+
- Structured logging with `adapter=vantage`, `operation`, `attempt` fields
- Never log tokens; use environment variables for secrets
- Comprehensive error handling

## Security

- Token provided via `PULUMICOST_VANTAGE_TOKEN` environment variable
- Tokens never logged or printed
- Least-privilege: prefer cost_report token over workspace token

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
