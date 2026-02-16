# gosilent

Context-efficient test runner for AI agents. Wraps `go test -json` and produces
ultra-compact output -- one line per passing package, failure details only when
needed. Designed to minimize LLM context token consumption.

## Why?

Raw `go test` output is verbose. A 50-package run with one failure produces
hundreds of lines, most of which are noise. When an AI agent runs tests in a
loop, every token of that output counts against its context window.

Common workarounds pipe through `grep` or `tail`, then re-run `go test` on
failure to get details. That doubles execution time.

gosilent solves both problems: it parses the full JSON stream in a single run,
emits one line per passing package, and expands only the failures.

### Before: raw `go test`

```
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestSub
    sub_test.go:15: expected 3, got 5
--- FAIL: TestSub (0.00s)
=== RUN   TestMul
--- PASS: TestMul (0.00s)
FAIL
FAIL    example.com/foo   0.42s
ok      example.com/bar   1.20s
ok      example.com/baz   0.81s
```

### After: gosilent

```
PASS example.com/bar 12/12 1.20s
PASS example.com/baz 8/8 0.81s
FAIL example.com/foo 2/3 0.42s

  FAIL TestSub
    sub_test.go:15: expected 3, got 5

3 passed, 1 failed (2.43s)
```

Passing packages collapse to a single line. Failures show only the relevant
test name and assertion output -- no `=== RUN`, no `--- FAIL`, no boilerplate.

## Output Format

```
# Success -- one line per package
PASS example.com/foo 42/42 3.45s

# Success with skipped tests
PASS example.com/foo 40/42 3.45s (2 skipped)

# No test files
SKIP example.com/notests [no test files]

# Failure -- compact header + only failure details (boilerplate stripped)
FAIL example.com/foo 41/42 2.10s

  FAIL TestSub
    sub_test.go:15: expected 3, got 5

# Build failure
FAIL example.com/broken [build failed]
  foo.go:10:5: undefined: DoesNotExist

# Summary line (only when multiple packages)
3 passed, 1 failed, 1 skipped (6.23s)
```

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
gosilent test -race -count=1 ./...
gosilent test -run TestFoo ./pkg/

# Verbose mode -- bypass JSON parsing, raw go test output
gosilent test --verbose ./...

# Print version
gosilent version
```

`gosilent test [args...]` runs `go test -json [args...]` under the hood. The
`-json` flag is injected automatically unless it is already present or
`--verbose` is set.

`--verbose` is the only gosilent-specific flag. Everything else passes through
to `go test` unchanged.

## How It Works

gosilent spawns `go test -json` as a subprocess and streams its output through
a JSON parser that builds per-package results. A compact formatter then reduces
each passing package to a single status line and expands only the failing tests,
stripping Go's test boilerplate (`=== RUN`, `--- FAIL`, etc.) to leave just the
assertion messages.

## Dependencies

- [`urfave/cli/v2`](https://github.com/urfave/cli) -- CLI framework
- [`stretchr/testify`](https://github.com/stretchr/testify) -- test assertions (dev only)

## License

[Apache License 2.0](LICENSE)
