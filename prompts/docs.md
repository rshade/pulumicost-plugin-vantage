# SYSTEM

Generate concise user docs for the plugin.

## GUARDRAILS

- No secrets.
- Keep examples accurate and runnable with mocks.
- Emit files only; no prose.

## INPUTS

Use Sections 4, 5, 9, 10, 15, 19.5–19.12.

## TASKS

1. `README.md`: Overview, Quickstart, features, limitations, and links.
2. `docs/CONFIG.md`: Detailed config reference with YAML, env vars, and
   notes.
3. `docs/TROUBLESHOOTING.md`: Common errors (auth, 429s, pagination),
   how to enable verbose logs, how to capture mock recordings.
4. `docs/FORECAST.md`: How forecast snapshots work and where they're
   written.

## OUTPUT

Emit the documentation files with multi‑file blocks only.
