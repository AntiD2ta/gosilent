# gosilent

Context-efficient test runner for AI agents. Wraps `go test -json` and produces
ultra-compact output -- a single line for the entire passing suite, failure
details only when needed. Designed to minimize LLM context token consumption.

## Why?

`go test` output grows with every package. A 10-package run produces 10 lines
(one per package). When an AI agent runs tests in a loop, every token of that
output counts against its context window.

Common workarounds pipe through `grep` or `tail`, then re-run `go test` on
failure to get details. That doubles execution time.

gosilent solves both problems: it parses the full JSON stream in a single run,
emits **one line** for the entire suite on success, and expands only the
failures.

### Before: `go test ./...`

```
ok  	github.com/stretchr/testify/assert	0.42s
ok  	github.com/stretchr/testify/http	0.31s
ok  	github.com/stretchr/testify/mock	1.20s
ok  	github.com/stretchr/testify/require	0.38s
ok  	github.com/stretchr/testify/suite	0.81s
?   	github.com/stretchr/testify	[no test files]
?   	github.com/stretchr/testify/...	[no test files]
```

### After: `gosilent test ./...`

```
ok  all packages (739 tests, 3.12s)
```

**7 lines collapsed to 1.**

## Output Format

```
# All tests pass -- single line, any number of packages
ok  all packages (67 tests, 1.71s)

# Single package -- shows package name
ok  formatter (21 tests, 0.00s)

# Significant flags are shown
ok  all packages -race (67 tests, 7.14s)

# Test failures -- only failing packages expanded
FAIL  example.com/foo
  TestSub
    sub_test.go:15: expected 3, got 5

1 failed, 2 passed (67 tests, 2.43s)

# Build failure
FAIL  example.com/broken [build failed]
  foo.go:10:5: undefined: DoesNotExist

1 failed (0 tests, 0.00s)
```

No-test-files packages are completely omitted. Skipped tests within passing
packages are not mentioned. Per-package pass lines are eliminated.

## Installation

```bash
go install github.com/AntiD2ta/gosilent/cmd/gosilent@latest
```

Or build from source:

```bash
git clone https://github.com/AntiD2ta/gosilent.git
cd gosilent
make build
```

Requires Go 1.26+.

## Usage

```bash
# Run all tests
gosilent test ./...

# With go test flags (passed through)
gosilent test -race ./...
gosilent test -run TestFoo ./pkg/

# Per-package detail mode
gosilent test --detail ./...

# Verbose mode -- bypass JSON parsing, raw go test output
gosilent test --verbose ./...

# Print version
gosilent version
```

`gosilent test [args...]` runs `go test -json [args...]` under the hood. The
`-json` flag is injected automatically unless already present or `--verbose` is
set.

### Flags

| Flag | Description |
|------|-------------|
| `--detail` | Per-package output with test counts (e.g., `PASS pkg 42/42 3.45s`) |
| `--verbose` | Raw `go test` output, bypasses JSON parsing entirely |

`--verbose` takes precedence over `--detail` since it bypasses JSON parsing.
Everything else passes through to `go test` unchanged.

## Benchmarks

gosilent vs `go test` on popular open-source projects (all tests passing):

| Project | `go test` | `gosilent` | Ratio |
|---------|----------:|-----------:|------:|
| gorilla/mux (273 tests) | 9 tokens | 7 tokens | 1.3x |
| spf13/cobra (402 tests) | 19 tokens | 9 tokens | 2.1x |
| go-chi/chi (143 tests) | 22 tokens | 10 tokens | 2.2x |
| stretchr/testify (739 tests) | 103 tokens | 9 tokens | 11.4x |
| gosilent (91 tests, 7 pkgs) | 101 tokens | 9 tokens | 11.2x |

Reproduce with `./scripts/benchmark.sh`.

## How It Works

gosilent spawns `go test -json` as a subprocess and streams its output through
a JSON parser that builds per-package results. A compact formatter then
aggregates all packages into a single summary line on success, or expands only
the failing packages with their test names and assertion output -- stripping
Go's test boilerplate (`=== RUN`, `--- FAIL`, etc.).

## Dependencies

- [`urfave/cli/v2`](https://github.com/urfave/cli) -- CLI framework
- [`stretchr/testify`](https://github.com/stretchr/testify) -- test assertions (dev only)

## License

[Apache License 2.0](LICENSE)
