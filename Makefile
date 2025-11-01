.PHONY: build test test-coverage lint fmt vet tidy verify clean wiremock-up wiremock-down demo help

# Variables
BINARY_NAME=pulumicost-vantage
MAIN_PACKAGE=./cmd/$(BINARY_NAME)
GO_VERSION=1.24.7
COVERAGE_THRESHOLD=70
CLIENT_COVERAGE_THRESHOLD=80
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo v0.1.0-dev)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
help:
	@echo "PulumiCost Vantage Plugin - Available targets:"
	@echo "  make build              - Build the binary"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests and generate coverage report"
	@echo "  make lint               - Run golangci-lint"
	@echo "  make fmt                - Format code with gofmt and goimports"
	@echo "  make vet                - Run go vet"
	@echo "  make tidy               - Run go mod tidy and verify no changes"
	@echo "  make verify             - Run all verification checks (fmt, vet, tidy)"
	@echo "  make clean              - Remove built artifacts"
	@echo "  make wiremock-up        - Start Wiremock mock server"
	@echo "  make wiremock-down      - Stop Wiremock mock server"
	@echo "  make demo               - Run demo against mock server"
	@echo "  make help               - Show this help message"

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Binary built: bin/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	@go test ./... -v -race -timeout 5m

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -v -race -timeout 5m -coverprofile=coverage.out -covermode=atomic
	@echo "Overall coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"
	@echo "Coverage report generated: coverage.out"

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./... --timeout=5m --allow-parallel-runners

vet:
	@echo "Running go vet..."
	@go vet ./...

tidy:
	@echo "Checking go mod tidy..."
	@go mod tidy
	@if [ -n "$$(git diff --name-only go.mod go.sum)" ]; then \
		echo "ERROR: go mod tidy resulted in changes:"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi
	@echo "✅ go mod tidy check passed"

verify: fmt vet tidy
	@echo "✅ All verification checks passed"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@command -v goimports > /dev/null 2>&1 && goimports -w . || echo "⚠️  goimports not found, skipping (golangci-lint will check)"

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out
	@go clean

wiremock-up:
	@echo "Starting Wiremock mock server..."
	@docker-compose -f test/wiremock/docker-compose.yml up -d
	@echo "Wiremock server started on http://localhost:8080"

wiremock-down:
	@echo "Stopping Wiremock mock server..."
	@docker-compose -f test/wiremock/docker-compose.yml down
	@echo "Wiremock server stopped"

demo: wiremock-up
	@echo "Running demo against mock server..."
	@sleep 2
	@go run $(MAIN_PACKAGE) pull --config ./test/config-mock.yaml || true
	@make wiremock-down
