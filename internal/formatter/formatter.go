package formatter

import (
	"fmt"
	"path"
	"strings"

	"github.com/AntiD2ta/gosilent/internal/testjson"
)

// compactStats holds aggregated statistics for the compact formatter.
type compactStats struct {
	totalTests   int
	totalElapsed float64
	failedPkgs   []*testjson.PackageResult
}

func aggregateStats(results []*testjson.PackageResult) compactStats {
	var s compactStats
	for _, pkg := range results {
		if pkg.NoTestFiles {
			continue
		}
		s.totalTests += pkg.TotalCount()
		s.totalElapsed += pkg.Elapsed
		if pkg.PackageAction == testjson.ActionFail {
			s.failedPkgs = append(s.failedPkgs, pkg)
		}
	}
	return s
}

// packageIdentifier returns the display name for the test run.
// Single non-NoTestFiles package → last path segment; otherwise "all packages".
func packageIdentifier(results []*testjson.PackageResult) string {
	var activePkg string
	activeCount := 0
	for _, pkg := range results {
		if !pkg.NoTestFiles {
			activePkg = pkg.Package
			activeCount++
		}
	}
	if activeCount == 1 {
		return path.Base(activePkg)
	}
	return "all packages"
}

// Format renders results as ultra-compact output: a single line on success,
// failure details only on failure. The flags parameter contains significant
// test flags (e.g., "-race", "-tags integration") to display in the output.
func Format(results []*testjson.PackageResult, flags []string) string {
	s := aggregateStats(results)
	identifier := packageIdentifier(results)

	var flagsSuffix string
	if len(flags) > 0 {
		flagsSuffix = " " + strings.Join(flags, " ")
	}

	if len(s.failedPkgs) == 0 {
		return fmt.Sprintf("ok  %s%s (%d tests, %.2fs)\n", identifier, flagsSuffix, s.totalTests, s.totalElapsed)
	}

	var b strings.Builder
	for _, pkg := range s.failedPkgs {
		formatFailedPackageCompact(&b, pkg)
	}

	passedPkgs := len(results) - len(s.failedPkgs)
	// Subtract NoTestFiles packages from passed count.
	for _, pkg := range results {
		if pkg.NoTestFiles {
			passedPkgs--
		}
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("%d failed", len(s.failedPkgs)))
	if passedPkgs > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", passedPkgs))
	}
	_, _ = fmt.Fprintf(&b, "%s%s (%d tests, %.2fs)\n",
		strings.Join(parts, ", "), flagsSuffix, s.totalTests, s.totalElapsed)

	return b.String()
}

func formatFailedPackageCompact(b *strings.Builder, pkg *testjson.PackageResult) {
	if pkg.BuildFailed {
		_, _ = fmt.Fprintf(b, "FAIL  %s [build failed]\n", pkg.Package)
		for _, line := range pkg.Output {
			_, _ = fmt.Fprintf(b, "  %s", line)
		}
		b.WriteByte('\n')
		return
	}

	_, _ = fmt.Fprintf(b, "FAIL  %s\n", pkg.Package)
	for _, t := range pkg.FailedTests() {
		_, _ = fmt.Fprintf(b, "  %s\n", t.Name)
		for _, line := range t.Output {
			if !isBoilerplate(line) {
				b.WriteString(line)
			}
		}
	}
	b.WriteByte('\n')
}

// FormatDetail renders package results with per-package detail lines.
// When there are multiple packages, a summary line is appended.
func FormatDetail(results []*testjson.PackageResult) string {
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

// HasFailures reports whether any package failed (test failures or build failures).
func HasFailures(results []*testjson.PackageResult) bool {
	for _, pkg := range results {
		if pkg.PackageAction == testjson.ActionFail {
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
