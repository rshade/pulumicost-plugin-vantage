# PulumiCost Vantage Plugin - Development Roadmap

**Project Version Target:** v0.1.0 (MVP)
**Status:** In Development
**Last Updated:** 2025-10-16

---

## Vision & Goals

Deliver a production-ready Vantage cost adapter for PulumiCost that:
- Fetches normalized cost/usage data from Vantage's REST API
- Maps costs to FOCUS 1.2 schema for standardized analysis
- Supports incremental daily syncs and historical backfills
- Includes comprehensive error handling and observability
- Is tested, documented, and ready for enterprise use

**Target Completion:** December 2025
**Success Criteria:** All v0.1 issues closed with ≥70% overall test coverage, ≥80% client coverage

---

## Release Roadmap

### v0.1.0 (MVP) - January 2026
Core read-only adapter with incremental sync and basic forecasts.

**Key Features:**
- ✅ REST client with auth, retries, pagination
- ✅ Cost ingestion via `/costs` endpoint
- ✅ FOCUS 1.2 schema mapping
- ✅ Incremental sync with bookmarks (D-3 to D-1)
- ✅ Forecast snapshots (basic support)
- ✅ Comprehensive testing (contract tests, golden fixtures)
- ✅ User documentation

**Out of Scope:**
- CSV export for massive backfills (>12 months)
- VQL filter support (v2)
- SaaS connector enrichment
- Real-time sync

---

## Milestones

### M1: Project Bootstrap ✅ (COMPLETE)
**Target:** Complete
**Deliverables:**
- Module structure and dependencies
- Makefile, linting config
- Design documentation in prompts/
- Repository hygiene (gitignore, README)

---

### M2: CLI & Configuration (Week 1-2)
**Target:** Oct 30, 2025
**Deliverables:**
- Cobra CLI with `pull`, `backfill`, `forecast` commands
- Config types and YAML parsing
- Environment variable support
- docs/CONFIG.md

---

### M3: REST Client Implementation (Week 2-4)
**Target:** Nov 13, 2025
**Deliverables:**
- HTTP client with Bearer token auth
- Request/response models for `/costs` and `/forecast`
- Cursor-based pagination
- Exponential backoff + jitter retry logic
- X-RateLimit-* header handling
- ≥80% test coverage
- internal/vantage/client/

---

### M4: Adapter Core (Week 4-6)
**Target:** Nov 27, 2025
**Deliverables:**
- Vantage → PulumiCost schema mapping
- FOCUS 1.2 field population
- Tag normalization (lowercase kebab-case)
- Idempotency key generation
- Diagnostic tracking for missing fields
- internal/vantage/adapter/

---

### M5: Incremental & Backfill Sync (Week 6-7)
**Target:** Dec 4, 2025
**Deliverables:**
- Incremental sync logic (D-3 to D-1 window)
- Bookmark persistence via Sink interface
- Backfill chunking by month
- Sync orchestration and error recovery
- ≥70% overall test coverage

---

### M6: Forecast & Snapshots (Week 7)
**Target:** Dec 10, 2025
**Deliverables:**
- Forecast endpoint integration
- Weekly snapshot storage
- Separate metric_type handling
- Snapshot retention policy (last 8)

---

### M7: Testing & Quality (Week 7-8)
**Target:** Dec 13, 2025
**Deliverables:**
- Wiremock contract tests
- Golden fixture test data
- End-to-end integration tests
- Coverage validation (≥70% overall, ≥80% client)
- Linting passes with zero errors

---

### M8: Documentation & Release (Week 8-9)
**Target:** Dec 20, 2025
**Deliverables:**
- docs/CONFIG.md (detailed reference)
- docs/TROUBLESHOOTING.md
- docs/FORECAST.md
- CHANGELOG.md
- Example configs
- v0.1.0 tag and GitHub release

---

## Issue Phases & Tasks

### Phase 1: CLI & Configuration (M2)

#### Issue 1.1: Create Cobra CLI Bootstrap
**Goal:** Scaffold the main CLI entry point with command structure
**Acceptance Criteria:**
- [ ] `cmd/pulumicost-vantage/main.go` compiles and runs
- [ ] Commands: `pull`, `backfill`, `forecast` are stubbed
- [ ] --config and --help flags work
- [ ] Version flag displays correctly
- [ ] CLI returns non-zero on errors

**Effort:** L (Large - 4-6 hours)
**Dependencies:** None

---

#### Issue 1.2: Implement Config Types & YAML Parsing
**Goal:** Define Config struct and load from YAML files
**Acceptance Criteria:**
- [ ] `internal/vantage/adapter/config.go` defines Config struct (Section 13 of design)
- [ ] All fields have yaml/json tags and validation
- [ ] Config loads from file via viper
- [ ] Environment variables override config (PULUMICOST_VANTAGE_TOKEN)
- [ ] Invalid configs return clear error messages
- [ ] Unit tests cover happy path + 3 error cases

**Effort:** M (Medium - 2-3 hours)
**Dependencies:** Issue 1.1

---

#### Issue 1.3: Create CONFIG.md Documentation
**Goal:** Document all configuration options with examples
**Acceptance Criteria:**
- [ ] docs/CONFIG.md covers all Config struct fields
- [ ] Includes YAML example from Section 4 of design
- [ ] Environment variable reference
- [ ] Common configuration patterns
- [ ] Security best practices (token handling)

**Effort:** S (Small - 1-2 hours)
**Dependencies:** Issue 1.2

---

### Phase 2: REST Client Implementation (M3)

#### Issue 2.1: Create HTTP Client with Auth & Models
**Goal:** Implement core HTTP client with Bearer token authentication
**Acceptance Criteria:**
- [ ] `internal/vantage/client/client.go` exports Client interface
- [ ] `internal/vantage/client/models.go` defines request/response structs
- [ ] Authorization header is properly set and redacted in logs
- [ ] Context timeouts are respected
- [ ] HTTP errors are wrapped with context
- [ ] Unit tests with httptest mock server
- [ ] ≥80% code coverage for client package

**Effort:** L (Large - 5-7 hours)
**Dependencies:** None

---

#### Issue 2.2: Implement Pagination & Cursor Handling
**Goal:** Add cursor-based pagination for /costs endpoint
**Acceptance Criteria:**
- [ ] `internal/vantage/client/pager.go` handles cursor pagination
- [ ] Costs(ctx, query) returns paginated results
- [ ] Cursor is properly passed between requests
- [ ] Stop condition (no next cursor) is detected
- [ ] Unit tests cover multi-page scenarios
- [ ] ≥80% test coverage for pager logic

**Effort:** M (Medium - 3-4 hours)
**Dependencies:** Issue 2.1

---

#### Issue 2.3: Add Retry Logic & Rate Limit Handling
**Goal:** Implement exponential backoff with jitter for resilience
**Acceptance Criteria:**
- [ ] HTTP 429 (rate limit) triggers exponential backoff with jitter
- [ ] HTTP 5xx errors retry up to max_retries
- [ ] X-RateLimit-Reset header is respected (sleep until reset)
- [ ] Backoff jitter prevents thundering herd
- [ ] Context cancellation stops retries
- [ ] Unit tests cover all retry scenarios
- [ ] ≥80% test coverage

**Effort:** M (Medium - 3-4 hours)
**Dependencies:** Issue 2.1

---

#### Issue 2.4: Implement Forecast Endpoint
**Goal:** Add forecast API integration
**Acceptance Criteria:**
- [ ] Forecast(ctx, reportToken, query) method implemented
- [ ] Returns forecast data structure
- [ ] Handles missing/empty forecasts gracefully
- [ ] Unit tests with mock responses
- [ ] ≥80% test coverage

**Effort:** S (Small - 2-3 hours)
**Dependencies:** Issue 2.1

---

#### Issue 2.5: Create Logger Interface & Redaction
**Goal:** Add structured logging with token redaction
**Acceptance Criteria:**
- [ ] `internal/vantage/client/logger.go` defines Logger interface
- [ ] Default no-op logger provided
- [ ] Authorization headers are redacted in logs
- [ ] Token values never appear in output
- [ ] Structured fields: adapter=vantage, operation, attempt
- [ ] Unit tests verify redaction works
- [ ] ≥80% test coverage

**Effort:** S (Small - 2 hours)
**Dependencies:** Issue 2.1

---

### Phase 3: Adapter Core Mapping (M4)

#### Issue 3.1: Implement Vantage → FOCUS 1.2 Schema Mapping
**Goal:** Map Vantage cost fields to PulumiCost CostRecord (Section 6 of design)
**Acceptance Criteria:**
- [ ] `internal/vantage/adapter/mapping.go` implements row-to-record conversion
- [ ] All 15+ fields from design Section 6 are mapped correctly
- [ ] Missing fields are populated with nil and diagnostic note
- [ ] timestamp → usage_start_time, cost → net_cost, etc.
- [ ] Unit tests with golden fixtures for each field combination
- [ ] ≥90% test coverage for mapping logic
- [ ] Deterministic output (no randomization)

**Effort:** L (Large - 6-8 hours)
**Dependencies:** Issue 2.1

---

#### Issue 3.2: Implement Tag Normalization & Filtering
**Goal:** Normalize tags to lowercase kebab-case with allow/deny lists
**Acceptance Criteria:**
- [ ] `internal/vantage/adapter/normalize.go` normalizes tag keys
- [ ] Converts `CamelCase_Key` → `camel-case-key`
- [ ] Supports allow/deny regex lists
- [ ] Preserves original values in labels_raw
- [ ] Drops high-cardinality keys (configurable)
- [ ] Merges provider native tags + Kubernetes labels
- [ ] Unit tests cover all normalization rules
- [ ] ≥85% test coverage

**Effort:** M (Medium - 4-5 hours)
**Dependencies:** Issue 3.1

---

#### Issue 3.3: Generate Idempotency Keys
**Goal:** Create deterministic idempotency keys for deduplication
**Acceptance Criteria:**
- [ ] Key is hash of (report_token, date, dimensions, metrics)
- [ ] Same inputs always produce same key
- [ ] Used for line_item_id in FOCUS 1.2
- [ ] Unit tests verify determinism across 100+ variations
- [ ] ≥95% test coverage (simple hash function)

**Effort:** S (Small - 1-2 hours)
**Dependencies:** Issue 3.1

---

#### Issue 3.4: Add Diagnostics & Missing Field Tracking
**Goal:** Track and report issues with missing or invalid fields
**Acceptance Criteria:**
- [ ] `internal/vantage/adapter/diagnostics.go` tracks data quality
- [ ] Missing fields are logged with field name + reason
- [ ] Diagnostic summary available after sync
- [ ] Unit tests for 5+ diagnostic scenarios
- [ ] ≥80% test coverage

**Effort:** S (Small - 2 hours)
**Dependencies:** Issue 3.1

---

### Phase 4: Incremental & Backfill Sync (M5)

#### Issue 4.1: Implement Incremental Sync Logic
**Goal:** Daily sync with D-3 to D-1 lag window for late postings
**Acceptance Criteria:**
- [ ] `internal/vantage/adapter/adapter.go` implements Sync(ctx, cfg, sink)
- [ ] Reads last_successful_end_date from bookmarks
- [ ] Queries D-3 to D-1 (configurable lag)
- [ ] Handles time zone correctly
- [ ] Recovers from partial failures (idempotent)
- [ ] Unit tests cover edge cases (first run, gaps, overlaps)
- [ ] ≥75% test coverage

**Effort:** L (Large - 6-8 hours)
**Dependencies:** Issue 3.1, Issue 2.1

---

#### Issue 4.2: Implement Backfill Logic
**Goal:** Chunk backfill requests by month to avoid timeouts
**Acceptance Criteria:**
- [ ] Backfill chunked by month (30-day chunks)
- [ ] Supports start_date to end_date range
- [ ] Falls back gracefully on large ranges
- [ ] Respects rate limits during backfill
- [ ] Unit tests for 6-month, 12-month backfills
- [ ] ≥75% test coverage

**Effort:** M (Medium - 4-5 hours)
**Dependencies:** Issue 4.1

---

#### Issue 4.3: Implement Bookmark Persistence
**Goal:** Store & retrieve last sync state via Sink interface
**Acceptance Criteria:**
- [ ] Bookmarks store key: (report_token, date_hash)
- [ ] Bookmarks store value: (last_end_date, cursor, attempts)
- [ ] UpdateBookmark called after successful batch
- [ ] Recovered on restart (idempotent)
- [ ] Unit tests mock Sink interface
- [ ] ≥80% test coverage

**Effort:** M (Medium - 3-4 hours)
**Dependencies:** Issue 4.1

---

#### Issue 4.4: Add Sync Error Recovery & Retries
**Goal:** Handle transient failures and implement retry strategies
**Acceptance Criteria:**
- [ ] Transient errors (429, 5xx) are retried with backoff
- [ ] Permanent errors (401, 404) fail fast with clear message
- [ ] Partial success doesn't lose data (idempotent)
- [ ] Circuit breaker after N consecutive failures
- [ ] Unit tests for 10+ failure scenarios
- [ ] ≥80% test coverage

**Effort:** M (Medium - 4 hours)
**Dependencies:** Issue 4.1, Issue 2.3

---

### Phase 5: Forecast & Snapshots (M6)

#### Issue 5.1: Implement Forecast Snapshot Storage
**Goal:** Store forecast results as separate metric_type records
**Acceptance Criteria:**
- [ ] Forecast records stored with metric_type="forecast"
- [ ] Separate from cost records (cost data not duplicated)
- [ ] Weekly snapshot logic implemented
- [ ] Keeps last 8 snapshots (configurable)
- [ ] Unit tests verify storage and retention
- [ ] ≥75% test coverage

**Effort:** M (Medium - 4 hours)
**Dependencies:** Issue 2.4, Issue 4.1

---

### Phase 6: Testing & Quality (M7)

#### Issue 6.1: Create Wiremock Contract Test Setup
**Goal:** Mock Vantage API for contract testing
**Acceptance Criteria:**
- [ ] Wiremock running in docker-compose
- [ ] Mappings created for /costs (3 pages) and /forecast
- [ ] Test fixtures in internal/vantage/contracts/
- [ ] make wiremock-up/down targets work
- [ ] Tests run against mock without external calls
- [ ] ≥5 golden fixtures for different scenarios

**Effort:** M (Medium - 3-4 hours)
**Dependencies:** None

---

#### Issue 6.2: Implement Contract Tests
**Goal:** Test client against Wiremock
**Acceptance Criteria:**
- [ ] Client tests hit mock /costs endpoint (multi-page)
- [ ] Client tests hit mock /forecast endpoint
- [ ] Tests verify pagination, retries, errors
- [ ] All tests pass with ≥80% coverage
- [ ] No external network calls

**Effort:** M (Medium - 4 hours)
**Dependencies:** Issue 6.1, Issue 2.1

---

#### Issue 6.3: Implement Golden Fixture Tests for Mapping
**Goal:** Validate adapter output against expected records
**Acceptance Criteria:**
- [ ] Golden JSON input fixtures for each dimension combo
- [ ] Expected CostRecord output fixtures
- [ ] Mapping tests compare actual vs golden
- [ ] Tests verify idempotency (same input = same output)
- [ ] ≥10 golden fixtures covering edge cases
- [ ] Tests pass with ≥90% accuracy

**Effort:** L (Large - 6 hours)
**Dependencies:** Issue 3.1, Issue 6.1

---

#### Issue 6.4: Validate Coverage Requirements
**Goal:** Ensure coverage meets acceptance criteria
**Acceptance Criteria:**
- [ ] make test-coverage runs all tests
- [ ] Client package: ≥80% coverage
- [ ] Overall: ≥70% coverage
- [ ] Coverage report generated and documented
- [ ] All critical paths covered (no skipped tests)

**Effort:** S (Small - 1-2 hours)
**Dependencies:** All other issues

---

#### Issue 6.5: Lint & Code Quality
**Goal:** Ensure code passes all linting rules
**Acceptance Criteria:**
- [ ] make lint passes with zero errors
- [ ] golangci-lint configured in .golangci.yml
- [ ] All imports organized (standard → third-party → internal)
- [ ] No unused variables or functions
- [ ] Code follows Go style guide

**Effort:** S (Small - 1 hour)
**Dependencies:** All code phases

---

### Phase 7: Documentation (M8)

#### Issue 7.1: Create TROUBLESHOOTING.md
**Goal:** Document common errors and solutions
**Acceptance Criteria:**
- [ ] Covers auth failures (401, invalid token)
- [ ] Covers rate limits (429, retry backoff explanation)
- [ ] Covers pagination errors
- [ ] Covers tag/field mapping issues
- [ ] How to enable verbose logging
- [ ] How to capture Wiremock recordings
- [ ] ≥10 common issues documented

**Effort:** S (Small - 2 hours)
**Dependencies:** All implementation phases

---

#### Issue 7.2: Create FORECAST.md
**Goal:** Document forecast snapshot feature
**Acceptance Criteria:**
- [ ] Explains forecast data flow
- [ ] Documents snapshot retention policy
- [ ] Shows how to query forecast records
- [ ] Explains metric_type="forecast" handling
- [ ] ≥3 examples of forecast usage

**Effort:** S (Small - 1-2 hours)
**Dependencies:** Issue 5.1

---

#### Issue 7.3: Create Example Configs
**Goal:** Provide copy-paste ready configs
**Acceptance Criteria:**
- [ ] docs/examples/basic.yaml (minimal config)
- [ ] docs/examples/full.yaml (all options)
- [ ] docs/examples/kubernetes.yaml (with K8s tags)
- [ ] docs/examples/multi-cloud.yaml (AWS+GCP)
- [ ] Each example commented and runnable

**Effort:** S (Small - 1-2 hours)
**Dependencies:** Issue 1.2

---

#### Issue 7.4: Create CHANGELOG.md
**Goal:** Document v0.1.0 release notes
**Acceptance Criteria:**
- [ ] New Features section
- [ ] Known Limitations section
- [ ] Breaking Changes (if any)
- [ ] Upgrading guide
- [ ] Contributors acknowledged

**Effort:** S (Small - 1 hour)
**Dependencies:** All issues

---

### Phase 8: Release & Polish (M8-9)

#### Issue 8.1: Create GitHub Release
**Goal:** Tag and release v0.1.0
**Acceptance Criteria:**
- [ ] Git tag v0.1.0 created
- [ ] GitHub Release created with release notes
- [ ] Binary artifacts attached
- [ ] SBOM generated (optional)
- [ ] Release published and visible

**Effort:** S (Small - 30 min)
**Dependencies:** All issues

---

#### Issue 8.2: Verify End-to-End with Mocks
**Goal:** Final integration test before release
**Acceptance Criteria:**
- [ ] make wiremock-up starts server
- [ ] ./bin/pulumicost-vantage pull --config docs/examples/basic.yaml succeeds
- [ ] Records written to output
- [ ] No errors or warnings
- [ ] Forecast records present if enabled

**Effort:** S (Small - 1 hour)
**Dependencies:** All implementation phases

---

#### Issue 8.3: Document Deployment & Operations
**Goal:** Create ops guide for running in production
**Acceptance Criteria:**
- [ ] docs/DEPLOYMENT.md created
- [ ] Docker image Dockerfile specified
- [ ] Kubernetes example manifests
- [ ] Environment setup instructions
- [ ] Monitoring/observability guidance

**Effort:** M (Medium - 2-3 hours)
**Dependencies:** All documentation

---

---

## Issue Effort Estimates

| Size | Hours | Example |
|------|-------|---------|
| S (Small) | 1-2 | Documentation, simple configs, small features |
| M (Medium) | 3-5 | Single package implementation, moderate complexity |
| L (Large) | 6-8 | Multi-file packages, complex logic, integration |

**Total Estimated Effort:** ~90-110 hours (11-14 developer-weeks)

---

## Dependencies & Critical Path

```
M1: Bootstrap ✅
  ↓
M2: CLI & Config
  ├→ Issue 1.1 (CLI)
  ├→ Issue 1.2 (Config types)
  └→ Issue 1.3 (Docs)
  ↓
M3: Client
  ├→ Issue 2.1 (Core HTTP)
  ├→ Issue 2.2 (Pagination) → depends on 2.1
  ├→ Issue 2.3 (Retries) → depends on 2.1
  ├→ Issue 2.4 (Forecast) → depends on 2.1
  └→ Issue 2.5 (Logging) → depends on 2.1
  ↓
M4: Adapter
  ├→ Issue 3.1 (Mapping) → depends on 2.1
  ├→ Issue 3.2 (Tags) → depends on 3.1
  ├→ Issue 3.3 (Idempotency) → depends on 3.1
  └→ Issue 3.4 (Diagnostics) → depends on 3.1
  ↓
M5: Sync
  ├→ Issue 4.1 (Incremental) → depends on 3.1, 2.1
  ├→ Issue 4.2 (Backfill) → depends on 4.1
  ├→ Issue 4.3 (Bookmarks) → depends on 4.1
  └→ Issue 4.4 (Error Recovery) → depends on 4.1, 2.3
  ↓
M6: Forecast
  └→ Issue 5.1 (Snapshots) → depends on 2.4, 4.1
  ↓
M7: Testing
  ├→ Issue 6.1 (Wiremock) → independent
  ├→ Issue 6.2 (Client Tests) → depends on 6.1, 2.1
  ├→ Issue 6.3 (Golden Tests) → depends on 6.1, 3.1
  ├→ Issue 6.4 (Coverage) → depends on all
  └→ Issue 6.5 (Lint) → depends on all
  ↓
M8: Docs & Release
  ├→ Issue 7.1-4 (Docs) → depends on respective features
  └→ Issue 8.1-3 (Release) → depends on all
```

---

## Success Metrics (v0.1.0)

- [ ] All 31 issues closed
- [ ] ≥70% overall test coverage
- [ ] ≥80% client package coverage
- [ ] 0 linting errors
- [ ] 0 security vulnerabilities (no hardcoded tokens)
- [ ] Wiremock contract tests passing
- [ ] Golden fixture tests passing
- [ ] End-to-end smoke test passing
- [ ] All documentation complete and reviewed
- [ ] GitHub release published

---

## Future Roadmap

### v0.2.0 (Q1 2026)
- CSV export for massive backfills (>12 months)
- VQL filter support
- Additional provider-specific fields
- Performance optimizations

### v0.3.0 (Q2 2026)
- SaaS connector enrichment (Fastly, Databricks, etc.)
- Real-time sync support
- Advanced tag filtering

### v1.0.0 (Q2 2026)
- Full feature parity with design spec
- Enterprise features (multi-workspace support)
- Kubernetes allocation heuristics

---

## Notes for Implementation Teams

### Code Style
- Go 1.24.7+
- Structured logging with fields: adapter=vantage, operation, attempt
- Never log tokens or sensitive data
- Use fmt.Errorf for error wrapping
- Explicit interface implementation

### Testing Philosophy
- Unit tests for all business logic
- Contract tests with Wiremock for API interactions
- Golden fixture tests for data transformations
- Deterministic output (no randomization in tests)
- No external network calls in CI

### Security Checklist
- [ ] Token never logged
- [ ] Auth header redacted in logs
- [ ] PULUMICOST_VANTAGE_TOKEN env var used
- [ ] No hardcoded example tokens
- [ ] YAML configs don't contain tokens
- [ ] Secrets not stored in git

### Documentation Checklist
- [ ] Every command documented with examples
- [ ] Every config option explained
- [ ] Common errors documented with solutions
- [ ] Architecture diagram provided (optional for v0.1)
