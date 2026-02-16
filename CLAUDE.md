# CLAUDE.md

Instructions for AI agents (Claude Code, etc.) working on this codebase.

## Project

- **Name**: gosilent
- **Module**: `github.com/AntiD2ta/gosilent`
- **Purpose**: Context-efficient test runner for AI agents. Wraps `go test -json` to produce compact output.

## Build and Test

```bash
# Build the binary
go build ./cmd/gosilent/

# Run all tests (preferred -- dogfoods gosilent itself)
./gosilent test ./...

# Run specific package
./gosilent test ./internal/testjson/...

# With race detector
./gosilent test -race ./...

# Verbose when full output needed
./gosilent test --verbose ./...

# Quality checks (all at once)
make quality

# Individual quality checks
go vet ./...
make lint
make fix-check
make gosec

# Unit tests only
make test-unit

# E2E tests only
make test-e2e
```

## Project Structure

```
cmd/gosilent/main.go          -- Entry point
internal/cli/app.go           -- CLI app construction + command registration
internal/testcmd/command.go   -- "test" command: wires runner -> parser -> formatter
internal/runner/runner.go     -- Process execution + stdout streaming
internal/testjson/event.go    -- TestEvent struct + action constants
internal/testjson/result.go   -- PackageResult, TestResult, TestStatus
internal/testjson/parser.go   -- JSON stream parser
internal/formatter/formatter.go -- Compact output renderer
e2e/e2e_test.go              -- End-to-end tests
testdata/*.jsonl              -- JSON fixtures for parser tests
testdata/projects/            -- Fixture Go projects for E2E tests
```

## Quality Gates

- `make quality` — runs build, vet, lint, fix-check, gosec in sequence
- `make lint` — golangci-lint with standard linters (errcheck, govet, staticcheck, etc.)
- `make fix-check` — fails if `go fix` detects un-applied modernizations
- `make gosec` — security scanner (G204 suppressed in runner.go for intentional subprocess)
- `make test-unit` — unit tests with race detector (`./internal/...`)
- `make test-e2e` — E2E tests with race detector (`./e2e/...`)
- CI: quality.yml (static analysis) and test.yml (unit → e2e) run as separate GitHub workflows

## Key Conventions

- CLI framework: urfave/cli/v2
- Test framework: testify/require (table-driven tests, require for fail-fast)
- TDD: Write failing test first, watch it fail, write minimal code to pass
- Fixtures: testdata/*.jsonl are realistic `go test -json` output streams
- Version injection: `go build -ldflags "-X main.version=..."` via Makefile

## Pipeline

The test command flows: runner -> parser -> formatter -> exit code

- Runner executes `go test -json` and streams stdout
- Parser reads the JSON stream into []*PackageResult
- Formatter renders results as compact text
- HasFailures() determines exit code

## Important Notes

- The E2E test suite has a recursion guard (GOSILENT_E2E_NESTED env var) to prevent infinite loops when gosilent tests itself.
- Parser handles both pre-Go 1.24 and Go 1.24+ build failure formats (build-output events with ImportPath).
- `--verbose` flag bypasses JSON parsing entirely, passing raw go test output through.
- `SkipFlagParsing = true` on the test command -- all flags except --verbose pass through to go test.
