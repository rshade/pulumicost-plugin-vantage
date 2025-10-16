SYSTEM
Create contract tests using Wiremock and golden fixtures.

GUARDRAILS
- No external network calls; use Wiremock.
- Put fixtures under `internal/vantage/contracts/`.
- Deterministic assertions (no time.Now() in golden content).
- Emit files only; no prose.

INPUTS
Use design Sections 5, 6, 8, 9, 10, 14, and 19.5.

TASKS
1) Add `test/wiremock/` with Docker compose or Make targets to run `wiremock/wiremock:3`.
2) Create Wiremock mappings for `/costs` (2‑3 pages) and `/forecast`.
3) Create golden JSON inputs and expected `CostRecord` arrays.
4) Write Go tests under `internal/vantage/adapter/` and `internal/vantage/client/` that:
   - spin up Wiremock (via Make),
   - hit mock endpoints,
   - compare adapter output to goldens,
   - assert idempotency (same inputs → same keys).
5) Add `make wiremock-up`/`wiremock-down` and `make demo` targets if missing.

OUTPUT
Emit files (mappings, fixtures, tests) with multi‑file blocks only.
