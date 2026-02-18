#!/usr/bin/env bash
#
# gosilent benchmark script
#
# Compares token efficiency of gosilent vs go test
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
    [prometheus/prometheus]="v3.9.1"
    [etcd-io/etcd]="v3.6.8"
    [cockroachdb/pebble]="v2.1.4"
    [kubernetes/client-go]="v0.35.1"
    [kubernetes/api]="v0.35.1"
)

# Order matters for consistent output
PROJECT_ORDER=(
    "gorilla/mux"
    "spf13/cobra"
    "tidwall/gjson"
    "go-chi/chi"
    "stretchr/testify"
    "prometheus/prometheus"
    "etcd-io/etcd"
    "cockroachdb/pebble"
    "kubernetes/client-go"
    "kubernetes/api"
)

BENCH_DIR="${BENCH_DIR:-/tmp/gosilent-bench}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKIP_CLONE=false
CHARS_PER_TOKEN=4  # Conservative estimate for LLM token counting
GO_TEST_TIMEOUT="10m"  # go test -timeout value

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
    (cd "$dir" && go test -short -timeout "$GO_TEST_TIMEOUT" ./... > "${prefix}-gotest.txt" 2>&1) || true

    # gosilent test
    (cd "$dir" && "$BENCH_DIR/gosilent" test -short -timeout "$GO_TEST_TIMEOUT" ./... > "${prefix}-gosilent.txt" 2>&1) || true
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

# Record commit hashes for reproducibility
log "Recording commit hashes"
: > "$BENCH_DIR/commits.txt"
for repo in "${PROJECT_ORDER[@]}"; do
    name="$(basename "$repo")"
    dir="$BENCH_DIR/$name"
    if [ -d "$dir/.git" ]; then
        hash=$(git -C "$dir" rev-parse HEAD 2>/dev/null)
        echo "$repo ${PROJECTS[$repo]} $hash" >> "$BENCH_DIR/commits.txt"
    fi
done
# Record gosilent's own commit
gosilent_hash=$(git -C "$PROJECT_ROOT" rev-parse HEAD 2>/dev/null)
echo "gosilent (self) dev $gosilent_hash" >> "$BENCH_DIR/commits.txt"

# Also benchmark gosilent itself
log "Benchmarking gosilent (self)"
(cd "$PROJECT_ROOT" && go test -short -timeout "$GO_TEST_TIMEOUT" ./... > "$BENCH_DIR/gosilent-self-gotest.txt" 2>&1) || true
(cd "$PROJECT_ROOT" && "$BENCH_DIR/gosilent" test -short -timeout "$GO_TEST_TIMEOUT" ./... > "$BENCH_DIR/gosilent-self-gosilent.txt" 2>&1) || true

# Run benchmarks
for repo in "${PROJECT_ORDER[@]}"; do
    benchmark_project "$repo"
done

# ─── Report ─────────────────────────────────────────────────────────

log "Generating report"
echo ""

python3 << 'PYEOF'
import math, os

BENCH_DIR = os.environ.get("BENCH_DIR", "/tmp/gosilent-bench")
CHARS_PER_TOKEN = 4

# Project data: (display_name, file_prefix)
projects = [
    ("gorilla/mux",         "mux"),
    ("spf13/cobra",         "cobra"),
    ("tidwall/gjson",       "gjson"),
    ("go-chi/chi",          "chi"),
    ("stretchr/testify",    "testify"),
    ("prometheus/prometheus","prometheus"),
    ("etcd-io/etcd",        "etcd"),
    ("cockroachdb/pebble",  "pebble"),
    ("kubernetes/client-go","client-go"),
    ("kubernetes/api",      "api"),
    ("gosilent (self)",     "gosilent-self"),
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
    gs = file_chars(f"{BENCH_DIR}/{prefix}-gosilent.txt")
    gt_lines = file_lines(f"{BENCH_DIR}/{prefix}-gotest.txt")
    gs_lines = file_lines(f"{BENCH_DIR}/{prefix}-gosilent.txt")
    data.append((name, gt, gs, gt_lines, gs_lines))

# Header
print("=" * 80)
print("GOSILENT BENCHMARK REPORT")
print("=" * 80)
print()
print(f"Token estimate: ~{CHARS_PER_TOKEN} chars/token (conservative for code output)")
go_ver = os.popen("go version").read().strip()
print(f"Go: {go_ver}")
print()

# Token table
print("TOKEN COMPARISON: gosilent vs go test")
print("-" * 80)
hdr = f"{'Project':<26} │ {'go test':>10} {'gosilent':>10} │ {'Saved':>10} {'Ratio':>8}"
print(hdr)
print("─" * 26 + "─┼─" + "─" * 10 + "─" * 11 + "─┼─" + "─" * 10 + "─" * 9)

total_gt = total_gs = 0

for name, gt, gs, gt_l, gs_l in data:
    if gt == 0 and gs == 0:
        continue
    gt_t = tokens(gt); gs_t = tokens(gs)
    total_gt += gt_t; total_gs += gs_t

    saved = gt_t - gs_t
    saved_str = f"-{saved}" if saved > 0 else f"+{-saved}"
    ratio = f"{gt_t/gs_t:.1f}x" if gs_t > 0 else "N/A"
    print(f"{name:<26} │ {gt_t:>10,} {gs_t:>10,} │ {saved_str:>10} {ratio:>8}")

print("─" * 26 + "─┼─" + "─" * 10 + "─" * 11 + "─┼─" + "─" * 10 + "─" * 9)
total_saved = total_gt - total_gs
ratio = f"{total_gt/total_gs:.1f}x" if total_gs > 0 else "N/A"
print(f"{'TOTAL':<26} │ {total_gt:>10,} {total_gs:>10,} │ {f'-{total_saved}':>10} {ratio:>8}")
print()

# Lines table
print("OUTPUT LINES COMPARISON")
print("-" * 80)
hdr = f"{'Project':<26} │ {'go test':>10} {'gosilent':>10} │ {'Reduction':>12}"
print(hdr)
print("─" * 26 + "─┼─" + "─" * 10 + "─" * 11 + "─┼─" + "─" * 12)

for name, gt, gs, gt_l, gs_l in data:
    if gt == 0 and gs == 0:
        continue
    if gt_l > 0:
        pct = f"{(1 - gs_l/gt_l)*100:.0f}% fewer" if gs_l < gt_l else "same"
    else:
        pct = "N/A"
    print(f"{name:<26} │ {gt_l:>10,} {gs_l:>10,} │ {pct:>12}")

print()

# Summary
print("SUMMARY")
print("-" * 80)
if total_gs > 0:
    pct = (1 - total_gs/total_gt)*100
    print(f"gosilent uses {pct:.1f}% fewer tokens than go test ({total_gt:,} → {total_gs:,})")
print()

# Raw chars for reference
print("RAW CHARACTER COUNTS")
print("-" * 80)
hdr = f"{'Project':<26} │ {'go test':>12} {'gosilent':>12}"
print(hdr)
print("─" * 26 + "─┼─" + "─" * 12 + "─" * 13)
for name, gt, gs, *_ in data:
    if gt == 0 and gs == 0:
        continue
    print(f"{name:<26} │ {gt:>12,} {gs:>12,}")

# Commit hashes
commits_path = f"{BENCH_DIR}/commits.txt"
if os.path.exists(commits_path):
    print()
    print("COMMIT REFERENCES")
    print("-" * 80)
    with open(commits_path) as f:
        for line in f:
            parts = line.strip().split(" ", 2)
            if len(parts) == 3:
                repo, version, sha = parts
                print(f"{repo:<26} {version:<14} {sha}")

print()
print("=" * 80)
print(f"Raw data saved to: {BENCH_DIR}/")
print("=" * 80)
PYEOF
