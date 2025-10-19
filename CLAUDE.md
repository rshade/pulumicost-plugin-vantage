# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository.

## Project Overview

**pulumicost-plugin-vantage** is a Go-based adapter that fetches normalized
cost/usage data from Vantage's REST API and maps it into PulumiCost's
internal schema with FinOps FOCUS 1.2 fields. The adapter supports
historical backfills, daily incremental syncs, tag/label dimensions,
and forecast snapshots.

**Key Reference**: `pulumi_cost_vantage_adapter_design_draft_v_0.md`
contains the complete technical design.

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
- **Error Handling**: Return errors explicitly, use `fmt.Errorf` for wrapping,
  always check context cancellation
- **Types**: Structs with `json`/`yaml` tags, implement interfaces explicitly
- **Logging**: Structured logging with fields: `adapter=vantage`, `operation`,
  `attempt`, `correlation_id`
- **Security**: Never log tokens; use environment variables for secrets;
  redact `Authorization` headers in logs

**Coverage Requirements**:

- Client package: ≥80% coverage
- Overall: ≥70% coverage

## Project Architecture

### Directory Structure

```text
cmd/pulumicost-vantage/        # CLI entry point with Cobra commands
# (pull, backfill, forecast)
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
    Forecast(ctx context.Context, reportToken string, q ForecastQuery)
        (Forecast, error)
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
- **Bookmarks**: Store `last_successful_end_date` per
  `(workspace|report_token, date_hash)`
- **Idempotency key**: Deterministic hash of `(date, dimensions, metrics)`
  —same inputs always produce same keys

### Pagination & Retries

- Use server cursors (configurable page size, default 5000)
- Exponential backoff with jitter on HTTP 429/5xx
- Honor `X-RateLimit-*` headers when present
- Configurable max retries and request timeout

### Forecast Snapshots

- Stored as separate records with `metric_type="forecast"`
- Weekly snapshots; keep last 8 for MAPE evaluation

## Configuration

See `pulumi_cost_vantage_adapter_design_draft_v_0.md` Section 4 for
full YAML config reference.

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
- `group_bys`: dimensions to group by (provider, service, account,
  project, region, resource_id, tags)
- `metrics`: cost types (cost, usage, effective_unit_price,
  amortized_cost, taxes, credits)
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

- **GET /costs**: Query params include workspace/cost_report token, date
  range, granularity, group_bys, metrics, pagination cursors
- **GET /cost_reports/{token}/forecast**: Optional forecast data
- **POST /cost_reports/{token}/exports**: Optional CSV backfill for large
  ranges (>12 months)
- **GET /workspaces**: Optional discovery

Prefer **Cost Report token** for stable/curated filters; fall back to
Workspace token when needed.

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

Set `VANTAGE_DEBUG=1` in environment (implementation detail—check
adapter/diagnostics.go)

## Security & Secrets

- **Token handling**: Read from `PULUMICOST_VANTAGE_TOKEN` env var or `.env`
  (dev only); never log
- **Redaction**: All `Authorization` headers and token values must be
  redacted from logs
- **Least privilege**: Use cost_report token (scoped to single report)
  instead of workspace token when possible
- **No secrets in commits**: `.env` and credential files are in `.gitignore`

## References

- **Design Document**: `pulumi_cost_vantage_adapter_design_draft_v_0.md`
- **AGENTS.md**: Existing guidelines (build targets, code style, interfaces,
  testing)
- **pulumicost-core**: Located at `../pulumicost-core` (local dependency)
  - FOCUS 1.2 Schema: `internal/focus/`
  - Sink Interface: `pkg/ingest/`
- **pulumicost-spec**: Located at `../pulumicost-spec` (local dependency)

## Known Constraints & Assumptions

- Assumes `pulumicost-core` is available (FOCUS types, Sink interface)
- Vantage API provides fields: `provider`, `service`, `account`, `project`,
  `region`, `resource_id`, `tags`, `cost`, `usage`, `amortized_cost`
  (when available), `currency`
- Actual field availability varies by provider and Vantage configuration
- No VQL filter support in v1; relies on curated Cost Reports

## Development Roadmap & Issue Tracking

**v0.1.0 MVP Structure**: 9 phases, 36 GitHub issues, 107-144 estimated hours

### Phases Overview

- **Phase 1**: Bootstrap ✅ (complete)
- **Phase 2**: CLI & Configuration (3 issues, #3-#5)
- **Phase 3**: REST Client (5 issues, #6-#10)
- **Phase 4**: Adapter Core (4 issues, #11-#12, #30-#31)
- **Phase 5**: Sync & Backfill (4 issues, #13-#16)
- **Phase 6**: Forecast (1 issue, #17)
- **Phase 7**: Testing (5 issues, #18-#22)
- **Phase 8**: Documentation & Release (7 issues, #23-#29)
- **Phase 9**: CI/CD & Hardening (7 issues, #33-#39)

### Issue Structure

Each GitHub issue includes:

- **Goal**: What to build
- **Acceptance Criteria**: Testable checkpoints (must all pass before closing)
- **Effort**: S(mall) 1-2h, M(edium) 3-5h, L(arge) 6-8h
- **Dependencies**: What issues must be done first
- **References**: Design sections and prompt files

**Example workflow:**

1. Pick issue from GitHub (e.g., #3: CLI Bootstrap)
2. Read acceptance criteria
3. Use corresponding prompt file (e.g., `prompts/bootstrap.md`)
4. Reference design document (e.g., Section 13)
5. Implement until all criteria checked
6. Run `make lint && make test`
7. Open PR, review, merge

### Prompt Files & Phases

```text
prompts/bootstrap.md              → Phase 2 (CLI & Config)
prompts/client.md                 → Phase 3 (REST Client)
prompts/adapter.md                → Phases 4-5 (Adapter & Sync)
prompts/tests.md                  → Phase 7 (Testing)
prompts/docs.md                   → Phase 8 (Documentation)
prompts/ci-and-repo-hardening.md  → Phase 9 (CI/CD & Hardening)
```

### Critical Path

Must follow this order (dependencies matter):

```text
Phase 2 (CLI) → Phase 3 (Client) → Phase 4 (Adapter) → Phase 5 (Sync)
                                 ↓
                            Phase 6 (Forecast)
                                 ↓
Phase 7 (Testing can start after Phase 3) → Phase 8 (Docs) → Phase 9 (Release)
```

### For AI-Assisted Development (Claude Code / OpenCode)

1. Load `pulumi_cost_vantage_adapter_design_draft_v_0.md` as context
2. Select phase prompt file (e.g., `prompts/client.md`)
3. Reference related GitHub issues for acceptance criteria
4. Generate code until all criteria met
5. Commit with Conventional Commits format (see below)

### Conventional Commits & Commitlint

**Format**: `type(scope): subject`

**Types**: feat, fix, docs, style, refactor, perf, test, chore, ci, revert
**Scopes**: infra, client, adapter, config, test, docs

**Examples**:

```text
feat(client): add exponential backoff retry logic
fix(adapter): correct FOCUS field mapping for amortized_cost
test(client): add pagination contract tests
chore(deps): update Go modules
```

**Validation**:

```bash
cat COMMIT_MESSAGE.md | npx commitlint
# Or use npm script:
npm run commitlint -- COMMIT_MESSAGE.md
```

Configuration in `commitlint.config.js` enforces:

- Type must be one of 10 allowed types
- Scope must be lowercase
- Subject required, no trailing period
- Header max 100 characters

### Markdown Linting

**Command**:

```bash
npm run markdownlint
```

Lints all `.md` files in the repository (excluding `node_modules`).
Uses `markdownlint-cli` for consistent markdown formatting.

### Repository Organization

- `prompts/` - OpenCode/Claude Code implementation guides (now includes
  PROJECT_SUMMARY.md and DEVELOPMENT_ROADMAP.md)
- `docs/` - User documentation (config, troubleshooting, deployment)
- `.github/` - GitHub workflows, templates, actions config
- `internal/vantage/` - Adapter code (client, adapter, contracts)
- `cmd/pulumicost-vantage/` - CLI entry point
- `test/wiremock/` - Mock server fixtures

### Success Criteria for v0.1.0

- ✓ All 36 issues closed
- ✓ ≥70% overall test coverage
- ✓ ≥80% client package coverage
- ✓ 0 linting errors (`make lint` passes)
- ✓ 0 security vulnerabilities
- ✓ Wiremock contract tests passing
- ✓ Golden fixture tests passing
- ✓ End-to-end integration test passing
- ✓ Complete documentation (CONFIG.md, TROUBLESHOOTING.md, FORECAST.md)
- ✓ GitHub Actions CI/CD passing on all workflows
- ✓ v0.1.0 GitHub release published

## Post-Implementation Cleanup

Once all 36 GitHub issues are closed and v0.1.0 is feature-complete:

### Prompts Directory

The `prompts/` directory contains implementation guides for Claude Code /
OpenCode and should be archived or removed once development is complete:

**Files to remove or archive:**

```bash
prompts/bootstrap.md                  # No longer needed
prompts/client.md                     # No longer needed
prompts/adapter.md                    # No longer needed
prompts/tests.md                      # No longer needed
prompts/docs.md                       # No longer needed
prompts/ci-and-repo-hardening.md      # No longer needed
prompts/PROJECT_SUMMARY.md            # Archive to docs/ or remove
prompts/DEVELOPMENT_ROADMAP.md        # Archive to docs/ or remove
```

#### Optional: Archive instead of delete

```bash
mkdir docs/archive-prompts/
mv prompts/* docs/archive-prompts/
rmdir prompts
```

### Keep in Repository Root

- `CLAUDE.md` - Future developer reference (this file, keep updated)
- `TODO.md` - Implementation record and history (keep for reference)
- `pulumi_cost_vantage_adapter_design_draft_v_0.md` - Technical design
  (keep indefinitely)
- `AGENTS.md` - Build/test procedures (keep updated)

### Documentation to Keep

Everything in `docs/` directory:

- `docs/CONFIG.md` - User-facing configuration reference
- `docs/TROUBLESHOOTING.md` - Support and debugging
- `docs/FORECAST.md` - Feature documentation
- `docs/DEPLOYMENT.md` - Operations guide
- `docs/examples/` - Example configurations

This keeps the repository clean while preserving historical implementation
guidance in archived form if needed for future versions.

## Troubleshooting

**Auth errors**: Verify `PULUMICOST_VANTAGE_TOKEN` is set and valid
**Rate limits (429)**: Adapter automatically backs off; check
`X-RateLimit-Reset` header
**Pagination issues**: Enable verbose logging and inspect cursor values
**Tag mapping failures**: Review `tag_prefix_filters` and allow/deny lists
in config
**Commit validation fails**: Ensure message follows Conventional Commits
format; check `commitlint.config.js`
**Issue not clear**: Read GitHub issue body for acceptance criteria;
reference corresponding prompt file; check design document section
