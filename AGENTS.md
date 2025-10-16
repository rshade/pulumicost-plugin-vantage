# AGENTS.md - PulumiCost Vantage Plugin

## Build/Lint/Test Commands
- `make build` - Build the binary
- `make test` - Run all tests
- `make lint` - Run golangci-lint
- `make fmt` - Format code with gofmt/goimports
- `go test ./... -v` - Run tests with verbose output
- `go test -run TestName` - Run single test
- `golangci-lint run` - Lint code

## Code Style Guidelines
- **Language**: Go 1.22+
- **Imports**: Standard library first, then third-party, then internal packages
- **Naming**: camelCase for variables/functions, PascalCase for exported types
- **Error Handling**: Return errors, use fmt.Errorf for wrapping, check context cancellation
- **Types**: Use structs with json/yaml tags, implement interfaces explicitly
- **Logging**: Structured logging with adapter=vantage, operation, attempt fields
- **Security**: Never log tokens, use environment variables for secrets

## Project Structure
- `cmd/pulumicost-vantage/` - CLI entry point with Cobra commands
- `internal/vantage/client/` - REST client with retry/backoff logic
- `internal/vantage/adapter/` - Mapping and sync logic for FOCUS 1.2
- `internal/vantage/contracts/` - Golden test fixtures
- `test/wiremock/` - Mock server configurations

## Key Interfaces
- `Client{Costs(ctx, q), Forecast(ctx, token, q)}`
- `Adapter{Sync(ctx, cfg, sink)}`
- `Sink` - For persisting cost records (from pulumicost-core)

## Testing
- Unit tests with ≥80% coverage for client, ≥70% overall
- Contract tests using Wiremock for API mocking
- Golden file tests for mapping validation
- `make wiremock-up/down` for mock server management