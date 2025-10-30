# AGENTS.md - PulumiCost Vantage Plugin

## Build/Lint/Test Commands

- `make build` - Build binary
- `make test` - Run all tests with race detection
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run golangci-lint with strict checks
- `make fmt` - Format with gofmt/goimports
- `make clean` - Remove built artifacts
- `go test -run TestName ./... -v` - Run single test
- `golangci-lint run` - Direct lint command
- `npm run commitlint -- COMMIT_MESSAGE.md` - Validate commit message format
- `npm run markdownlint` - Lint markdown files

## Code Style Guidelines

- **Language**: Go 1.24.7+
- **Imports**: Standard library → third-party → internal packages
- **Naming**: camelCase variables/functions, PascalCase exported types
- **Error Handling**: Return errors, fmt.Errorf wrapping, context cancellation checks
- **Types**: Structs with json/yaml tags, explicit interface implementation
- **Logging**: Structured with adapter=vantage, operation, attempt fields
- **Security**: Never log tokens, use env vars for secrets

## Project Structure

- `cmd/pulumicost-vantage/` - CLI with Cobra
- `internal/vantage/client/` - REST client with retry/backoff
- `internal/vantage/adapter/` - FOCUS 1.2 mapping/sync logic
- `test/wiremock/` - Mock server configs

## Testing Requirements

- ≥80% client coverage, ≥70% overall
- Contract tests with Wiremock, golden file validation
- `make wiremock-up/down` for mock server
