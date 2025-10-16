# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**pulumicost-plugin-vantage** is a Go-based adapter that fetches normalized cost/usage data from Vantage's REST API and maps it into PulumiCost's internal schema with FinOps FOCUS 1.2 fields. The adapter supports historical backfills, daily incremental syncs, tag/label dimensions, and forecast snapshots.

**Key Reference**: `pulumi_cost_vantage_adapter_design_draft_v_0.md` contains the complete technical design.

## Build & Development Commands

```bash
make build          # Build the binary
make test           # Run all tests
make lint           # Run golangci-lint
make fmt            # Format code with gofmt/goimports
make wiremock-up    # Start mock server (for contract tests)
make wiremock-down  # Stop mock server
make demo           # Run pull against mocks and print records
go test ./... -v    # Run tests with verbose output
go test -run TestName  # Run single test
```

## Code Style & Standards

- **Language**: Go 1.22+
- **Imports**: Standard library → third-party → internal packages
- **Naming**: camelCase for variables/functions, PascalCase for exported types
- **Error Handling**: Return errors explicitly, use `fmt.Errorf` for wrapping, always check context cancellation
- **Types**: Structs with `json`/`yaml` tags, implement interfaces explicitly
- **Logging**: Structured logging with fields: `adapter=vantage`, `operation`, `attempt`, `correlation_id`
- **Security**: Never log tokens; use environment variables for secrets; redact `Authorization` headers in logs

**Coverage Requirements**:
- Client package: ≥80% coverage
- Overall: ≥70% coverage

## Project Architecture

### Directory Structure

```
cmd/pulumicost-vantage/        # CLI entry point with Cobra commands (pull, backfill, forecast)
internal/vantage/
  ├── client/                   # REST client with retry/backoff logic
  ├── adapter/                  # Mapping and sync logic for FOCUS 1.2
  │   ├── config.go             # Config struct with yaml/json tags
  │   ├── adapter.go            # Sync orchestration
  │   ├── mapping.go            # Vantage row → CostRecord + FOCUS 1.2
  │   └── normalize.go           # Tag normalization, allow/deny filters
  └── contracts/                # Golden test fixtures
test/wiremock/                  # Mock server configurations
docs/                           # User documentation
```

### Core Interfaces

```go
// Client interface (internal/vantage/client)
type Client interface {
    Costs(ctx context.Context, q Query) (Page, error)
    Forecast(ctx context.Context, reportToken string, q ForecastQuery) (Forecast, error)
}

// Adapter interface (internal/vantage/adapter)
type Adapter interface {
    Sync(ctx context.Context, cfg Config, sink Sink) error
}

// Sink (from pulumicost-core, persists cost records)
type Sink interface {
    Write(ctx context.Context, records []CostRecord) error
    UpdateBookmark(ctx context.Context, key string, value interface{}) error
}
```

## Key Features & Implementation Notes

### Schema Mapping (Vantage → PulumiCost → FOCUS 1.2)

Critical fields:
- `timestamp` (Vantage bucket start) → `usage_start_time`
- `provider` → `cloud_provider`
- `service` → `service_name`
- `account` → `billing_account_id`
- `region` → `region`
- `resource_id` → `resource_id`
- `cost` (net) → `net_cost`
- `amortized_cost` (if present) → `net_amortized_cost`
- `tags` → `labels` (normalized to lowercase kebab-case)

Missing fields become `nil` with a diagnostic note.

### Tag Normalization Strategy

- Normalize tag keys to lowercase kebab-case (configurable)
- Support allow/deny lists via regex
- Preserve original values in `labels_raw`
- Drop high-cardinality keys (e.g., pod UID)

### Incremental Sync & Idempotency

- **Incremental window**: D-3 to D-1 (configurable lag for late postings)
- **Bookmarks**: Store `last_successful_end_date` per `(workspace|report_token, date_hash)`
- **Idempotency key**: Deterministic hash of `(date, dimensions, metrics)`—same inputs always produce same keys

### Pagination & Retries

- Use server cursors (configurable page size, default 5000)
- Exponential backoff with jitter on HTTP 429/5xx
- Honor `X-RateLimit-*` headers when present
- Configurable max retries and request timeout

### Forecast Snapshots

- Stored as separate records with `metric_type="forecast"`
- Weekly snapshots; keep last 8 for MAPE evaluation

## Configuration

See `pulumi_cost_vantage_adapter_design_draft_v_0.md` Section 4 for full YAML config reference.

**Required**:
```yaml
credentials:
  token: ${PULUMICOST_VANTAGE_TOKEN}  # Via env var, never logged
params:
  cost_report_token: "cr_..."         # Preferred over workspace_token
```

**Optional**:
- `start_date`, `end_date`: ISO dates (default: 12 months back → today)
- `granularity`: `"day"` or `"month"` (default: `"day"`)
- `group_bys`: dimensions to group by (provider, service, account, project, region, resource_id, tags)
- `metrics`: cost types (cost, usage, effective_unit_price, amortized_cost, taxes, credits)
- `include_forecast`: enable forecast snapshots
- `page_size`: API page size (default 5000)
- `request_timeout_seconds`, `max_retries`: resilience config

## Testing Strategy

### Unit Tests
- Parameter building, pagination logic, tag normalization
- Golden JSON → CostRecord mapping assertions
- All client operations with `httptest`

### Contract Tests
- Run Wiremock server via `make wiremock-up`
- Mock `/costs` endpoint (multi-page pagination)
- Mock `/forecast` endpoint
- Assert adapter produces correct records

### Fixtures Location
- Golden samples: `internal/vantage/contracts/`
- Wiremock mappings: `test/wiremock/`

## API Endpoints Used

- **GET /costs**: Query params include workspace/cost_report token, date range, granularity, group_bys, metrics, pagination cursors
- **GET /cost_reports/{token}/forecast**: Optional forecast data
- **POST /cost_reports/{token}/exports**: Optional CSV backfill for large ranges (>12 months)
- **GET /workspaces**: Optional discovery

Prefer **Cost Report token** for stable/curated filters; fall back to Workspace token when needed.

## Common Development Workflows

### Run tests before commit
```bash
make lint && make test
```

### Debug a single test
```bash
go test -run TestCostsMapping -v
```

### Start mock server for manual testing
```bash
make wiremock-up
# Then run: ./pulumicost-vantage pull --config test/config.yaml
make wiremock-down
```

### Add verbose logging for troubleshooting
Set `VANTAGE_DEBUG=1` in environment (implementation detail—check adapter/diagnostics.go)

## Security & Secrets

- **Token handling**: Read from `PULUMICOST_VANTAGE_TOKEN` env var or `.env` (dev only); never log
- **Redaction**: All `Authorization` headers and token values must be redacted from logs
- **Least privilege**: Use cost_report token (scoped to single report) instead of workspace token when possible
- **No secrets in commits**: `.env` and credential files are in `.gitignore`

## References

- **Design Document**: `pulumi_cost_vantage_adapter_design_draft_v_0.md`
- **AGENTS.md**: Existing guidelines (build targets, code style, interfaces, testing)
- **pulumicost-core**: Located at `../pulumicost-core` (local dependency)
  - FOCUS 1.2 Schema: `internal/focus/`
  - Sink Interface: `pkg/ingest/`
- **pulumicost-spec**: Located at `../pulumicost-spec` (local dependency)

## Known Constraints & Assumptions

- Assumes `pulumicost-core` is available (FOCUS types, Sink interface)
- Vantage API provides fields: `provider`, `service`, `account`, `project`, `region`, `resource_id`, `tags`, `cost`, `usage`, `amortized_cost` (when available), `currency`
- Actual field availability varies by provider and Vantage configuration
- No VQL filter support in v1; relies on curated Cost Reports

## Troubleshooting

**Auth errors**: Verify `PULUMICOST_VANTAGE_TOKEN` is set and valid
**Rate limits (429)**: Adapter automatically backs off; check `X-RateLimit-Reset` header
**Pagination issues**: Enable verbose logging and inspect cursor values
**Tag mapping failures**: Review `tag_prefix_filters` and allow/deny lists in config
