#!/usr/bin/env bash
#
# gosilent benchmark script
#
# Compares token efficiency of gosilent vs go test and go test -v
# against well-known Go projects at pinned versions.
#
# Usage:
#   ./scripts/benchmark.sh              # Full benchmark (clone + test + report)
#   ./scripts/benchmark.sh --skip-clone # Skip cloning, reuse existing repos
#
# Requirements:
#   - go (matching go.mod version)
#   - git
#   - python3 (for token analysis)
#
# Output:
#   Prints a formatted benchmark report to stdout.
#   Raw data files are saved to $BENCH_DIR for further analysis.

set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────
# Pin project versions for deterministic results.
# Update these when re-baselining benchmarks.
declare -A PROJECTS=(
    [gorilla/mux]="v1.8.1"
    [spf13/cobra]="v1.10.2"
    [tidwall/gjson]="v1.18.0"
    [go-chi/chi]="v5.2.5"
    [stretchr/testify]="v1.11.1"
)

# Order matters for consistent output
PROJECT_ORDER=(
    "gorilla/mux"
    "spf13/cobra"
    "tidwall/gjson"
    "go-chi/chi"
    "stretchr/testify"
)

BENCH_DIR="${BENCH_DIR:-/tmp/gosilent-bench}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKIP_CLONE=false
CHARS_PER_TOKEN=4  # Conservative estimate for LLM token counting

# ─── Argument parsing ───────────────────────────────────────────────
for arg in "$@"; do
    case "$arg" in
        --skip-clone) SKIP_CLONE=true ;;
        --help|-h)
            echo "Usage: $0 [--skip-clone]"
            echo ""
            echo "  --skip-clone  Reuse previously cloned repos in $BENCH_DIR"
            exit 0
            ;;
        *)
            echo "Unknown argument: $arg" >&2
            exit 1
            ;;
    esac
done

# ─── Functions ──────────────────────────────────────────────────────

log() {
    echo ">>> $*" >&2
}

# Clone a project at a pinned version
clone_project() {
    local repo="$1"
    local version="$2"
    local name
    name="$(basename "$repo")"
    local dest="$BENCH_DIR/$name"

    if [ -d "$dest" ]; then
        rm -rf "$dest"
    fi

    log "Cloning $repo@$version"
    git clone --depth 1 --branch "$version" "https://github.com/$repo.git" "$dest" 2>/dev/null
    (cd "$dest" && go mod download 2>/dev/null)
}

# Run go test and gosilent for a project, save outputs to files
benchmark_project() {
    local repo="$1"
    local name
    name="$(basename "$repo")"
    local dir="$BENCH_DIR/$name"
    local prefix="$BENCH_DIR/$name"

    log "Benchmarking $repo"

    # go test (non-verbose)
    (cd "$dir" && go test ./... > "${prefix}-gotest.txt" 2>&1) || true

    # go test -v (verbose)
    (cd "$dir" && go test -v ./... > "${prefix}-gotest-v.txt" 2>&1) || true

    # gosilent test
    (cd "$dir" && "$BENCH_DIR/gosilent" test ./... > "${prefix}-gosilent.txt" 2>&1) || true
}

# Get character count for a file
chars() {
    wc -c < "$1" | tr -d ' '
}

# Get line count for a file
lines() {
    wc -l < "$1" | tr -d ' '
}

# ─── Main ───────────────────────────────────────────────────────────

mkdir -p "$BENCH_DIR"

# Build gosilent
log "Building gosilent"
(cd "$PROJECT_ROOT" && go build -o "$BENCH_DIR/gosilent" ./cmd/gosilent/)

# Clone projects (unless --skip-clone)
if [ "$SKIP_CLONE" = false ]; then
    for repo in "${PROJECT_ORDER[@]}"; do
        clone_project "$repo" "${PROJECTS[$repo]}"
    done
else
    log "Skipping clone (--skip-clone)"
fi

# Also benchmark gosilent itself
log "Benchmarking gosilent (self)"
(cd "$PROJECT_ROOT" && go test ./... > "$BENCH_DIR/gosilent-self-gotest.txt" 2>&1) || true
(cd "$PROJECT_ROOT" && go test -v ./... > "$BENCH_DIR/gosilent-self-gotest-v.txt" 2>&1) || true
(cd "$PROJECT_ROOT" && "$BENCH_DIR/gosilent" test ./... > "$BENCH_DIR/gosilent-self-gosilent.txt" 2>&1) || true

# Run benchmarks
for repo in "${PROJECT_ORDER[@]}"; do
    benchmark_project "$repo"
done

# ─── Report ─────────────────────────────────────────────────────────

log "Generating report"
echo ""

python3 << 'PYEOF'
import math, os, sys

BENCH_DIR = os.environ.get("BENCH_DIR", "/tmp/gosilent-bench")
CHARS_PER_TOKEN = 4

# Project data: (display_name, file_prefix)
projects = [
    ("gorilla/mux",    "mux"),
    ("spf13/cobra",    "cobra"),
    ("tidwall/gjson",  "gjson"),
    ("go-chi/chi",     "chi"),
    ("stretchr/testify", "testify"),
    ("gosilent (self)", "gosilent-self"),
]

def file_chars(path):
    try:
        return os.path.getsize(path)
    except FileNotFoundError:
        return 0

def file_lines(path):
    try:
        with open(path) as f:
            return sum(1 for _ in f)
    except FileNotFoundError:
        return 0

def tokens(chars):
    return math.ceil(chars / CHARS_PER_TOKEN)

# Collect data
data = []
for name, prefix in projects:
    gt = file_chars(f"{BENCH_DIR}/{prefix}-gotest.txt")
    gtv = file_chars(f"{BENCH_DIR}/{prefix}-gotest-v.txt")
    gs = file_chars(f"{BENCH_DIR}/{prefix}-gosilent.txt")
    gt_lines = file_lines(f"{BENCH_DIR}/{prefix}-gotest.txt")
    gtv_lines = file_lines(f"{BENCH_DIR}/{prefix}-gotest-v.txt")
    gs_lines = file_lines(f"{BENCH_DIR}/{prefix}-gosilent.txt")
    data.append((name, gt, gtv, gs, gt_lines, gtv_lines, gs_lines))

# Header
print("=" * 96)
print("GOSILENT BENCHMARK REPORT")
print("=" * 96)
print()
print(f"Token estimate: ~{CHARS_PER_TOKEN} chars/token (conservative for code output)")
go_ver = os.popen("go version").read().strip()
print(f"Go: {go_ver}")
print()

# Token table
print("TOKEN COMPARISON")
print("-" * 96)
hdr = f"{'Project':<22} │ {'go test':>10} {'go test -v':>12} {'gosilent':>10} │ {'vs -v':>8} {'Ratio':>8}"
print(hdr)
print("─" * 22 + "─┼─" + "─" * 10 + "─" * 13 + "─" * 11 + "─┼─" + "─" * 8 + "─" * 9)

total_gt = total_gtv = total_gs = 0
pass_gtv = pass_gs = 0

for name, gt, gtv, gs, *_ in data:
    gt_t = tokens(gt); gtv_t = tokens(gtv); gs_t = tokens(gs)
    total_gt += gt_t; total_gtv += gtv_t; total_gs += gs_t

    if "testify" not in name:
        pass_gtv += gtv_t; pass_gs += gs_t

    pct = f"{(1 - gs_t/gtv_t)*100:.1f}%" if gtv_t > 0 else "N/A"
    ratio = f"{gtv_t/gs_t:.0f}x" if gs_t > 0 else "N/A"
    print(f"{name:<22} │ {gt_t:>10,} {gtv_t:>12,} {gs_t:>10,} │ {pct:>8} {ratio:>8}")

print("─" * 22 + "─┼─" + "─" * 10 + "─" * 13 + "─" * 11 + "─┼─" + "─" * 8 + "─" * 9)
pct = f"{(1 - total_gs/total_gtv)*100:.1f}%" if total_gtv > 0 else "N/A"
ratio = f"{total_gtv/total_gs:.0f}x" if total_gs > 0 else "N/A"
print(f"{'TOTAL':<22} │ {total_gt:>10,} {total_gtv:>12,} {total_gs:>10,} │ {pct:>8} {ratio:>8}")
print()

# Lines table
print("OUTPUT LINES COMPARISON")
print("-" * 96)
hdr = f"{'Project':<22} │ {'go test':>10} {'go test -v':>12} {'gosilent':>10} │ {'vs -v':>12}"
print(hdr)
print("─" * 22 + "─┼─" + "─" * 10 + "─" * 13 + "─" * 11 + "─┼─" + "─" * 12)

for name, gt, gtv, gs, gt_l, gtv_l, gs_l in data:
    pct = f"{(1 - gs_l/gtv_l)*100:.0f}% fewer" if gtv_l > 0 else "N/A"
    print(f"{name:<22} │ {gt_l:>10,} {gtv_l:>12,} {gs_l:>10,} │ {pct:>12}")

print()

# Summary
print("SUMMARY")
print("-" * 96)
if pass_gs > 0:
    print(f"Passing suites:  gosilent uses {(1 - pass_gs/pass_gtv)*100:.1f}% fewer tokens than go test -v ({pass_gtv:,} → {pass_gs:,})")
if total_gs > 0:
    print(f"All projects:    gosilent uses {(1 - total_gs/total_gtv)*100:.1f}% fewer tokens than go test -v ({total_gtv:,} → {total_gs:,})")
print()

# Raw chars for reference
print("RAW CHARACTER COUNTS")
print("-" * 96)
hdr = f"{'Project':<22} │ {'go test':>12} {'go test -v':>14} {'gosilent':>12}"
print(hdr)
print("─" * 22 + "─┼─" + "─" * 12 + "─" * 15 + "─" * 13)
for name, gt, gtv, gs, *_ in data:
    print(f"{name:<22} │ {gt:>12,} {gtv:>14,} {gs:>12,}")

print()
print("=" * 96)
print(f"Raw data saved to: {BENCH_DIR}/")
print("=" * 96)
PYEOF
