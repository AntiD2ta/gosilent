package formatter

import (
	"fmt"
	"strings"

	"github.com/AntiD2ta/gosilent/internal/testjson"
)

// Format renders package results as compact text output.
// When there are multiple packages, a summary line is appended.
func Format(results []*testjson.PackageResult) string {
	var b strings.Builder
	for _, pkg := range results {
		formatPackage(&b, pkg)
	}
	if len(results) > 1 {
		formatSummary(&b, results)
	}
	return b.String()
}

func formatSummary(b *strings.Builder, results []*testjson.PackageResult) {
	var passed, failed, skipped int
	var totalElapsed float64
	for _, pkg := range results {
		totalElapsed += pkg.Elapsed
		switch {
		case pkg.NoTestFiles:
			skipped++
		case pkg.BuildFailed || pkg.FailCount() > 0:
			failed++
		default:
			passed++
		}
	}

	var parts []string
	if passed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", passed))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	fmt.Fprintf(b, "%s (%.2fs)\n", strings.Join(parts, ", "), totalElapsed)
}

// HasFailures reports whether any package has test failures or build failures.
func HasFailures(results []*testjson.PackageResult) bool {
	for _, pkg := range results {
		if pkg.BuildFailed || pkg.FailCount() > 0 {
			return true
		}
	}
	return false
}

// boilerplatePrefixes are go test output lines that are stripped from failure details.
var boilerplatePrefixes = []string{
	"=== RUN",
	"=== PAUSE",
	"=== CONT",
	"--- FAIL:",
	"--- PASS:",
	"--- SKIP:",
	"PASS",
	"FAIL",
}

func isBoilerplate(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range boilerplatePrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return trimmed == ""
}

func formatPackage(b *strings.Builder, pkg *testjson.PackageResult) {
	if pkg.NoTestFiles {
		fmt.Fprintf(b, "SKIP %s [no test files]\n", pkg.Package)
		return
	}

	if pkg.BuildFailed {
		fmt.Fprintf(b, "FAIL %s [build failed]\n", pkg.Package)
		for _, line := range pkg.Output {
			fmt.Fprintf(b, "  %s", line)
		}
		b.WriteByte('\n')
		return
	}

	failed := pkg.FailedTests()
	if len(failed) > 0 {
		fmt.Fprintf(b, "FAIL %s %d/%d %.2fs\n",
			pkg.Package, pkg.PassCount(), pkg.TotalCount(), pkg.Elapsed)
		for _, t := range failed {
			b.WriteByte('\n')
			fmt.Fprintf(b, "  FAIL %s\n", t.Name)
			for _, line := range t.Output {
				if !isBoilerplate(line) {
					b.WriteString(line)
				}
			}
		}
		b.WriteByte('\n')
		return
	}

	fmt.Fprintf(b, "PASS %s %d/%d %.2fs",
		pkg.Package, pkg.PassCount(), pkg.TotalCount(), pkg.Elapsed)
	if skipped := pkg.SkipCount(); skipped > 0 {
		fmt.Fprintf(b, " (%d skipped)", skipped)
	}
	b.WriteByte('\n')
}
