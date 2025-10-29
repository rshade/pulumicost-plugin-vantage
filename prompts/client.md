# SYSTEM

You are implementing a resilient HTTP client for Vantage's REST API in
Go.

## GUARDRAILS

- No secrets in code or logs; redact `Authorization` header.
- Respect context timeouts and cancellations.
- Retries with exponential backoff + jitter on 429/5xx.
- Honor `X-RateLimit-*` headers when present.
- ≥80% coverage for the `client` package.
- Emit files using the multi‑file format; no prose.

## INPUTS

Use the design's Section 5 (Endpoint Usage), Section 8
(Pagination/Retry), Section 11 (Observability), Section 13 (Interfaces),
and 19.5 (AC).

## TASKS

1. Create package `internal/vantage/client` with:
   - `client.go`: public `Client` interface and constructor `New`.
   - `http.go`: low‑level HTTP, backoff, rate‑limit handling, redact
     logs.
   - `models.go`: request/response structs for `/costs` and `/forecast`.
   - `pager.go`: cursor pagination helpers.
   - `logger.go`: minimal interface used by adapter; default no‑op.
2. Implement `Costs(ctx, q)` (cursor pagination) and `Forecast(ctx,
   token, q)`.
3. Unit tests under `internal/vantage/client/` using `httptest`.
4. Add small examples in `_test.go` demonstrating usage.

## OUTPUT

Emit all files and tests with multi‑file blocks only.
