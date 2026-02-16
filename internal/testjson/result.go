package testjson

import "strings"

// TestStatus represents the outcome of a test.
type TestStatus int

const (
	StatusPass TestStatus = iota
	StatusFail
	StatusSkip
)

// TestResult holds the outcome and output of a single test function.
type TestResult struct {
	Name   string
	Status TestStatus
	Output []string
}

// PackageResult holds the aggregated results for a single package.
type PackageResult struct {
	Package       string
	Elapsed       float64
	Tests         []*TestResult
	Output        []string // package-level output (build errors, etc.)
	NoTestFiles   bool
	BuildFailed   bool
	PackageAction Action // package-level verdict from go test -json (pass/fail/skip)
}

// LeafTests returns only the leaf tests (tests that are not parents of subtests).
// A test is a parent if another test's name starts with its name followed by "/".
func (p *PackageResult) LeafTests() []*TestResult {
	if len(p.Tests) == 0 {
		return nil
	}

	parents := make(map[string]bool)
	for _, t := range p.Tests {
		for _, other := range p.Tests {
			if other.Name != t.Name && strings.HasPrefix(other.Name, t.Name+"/") {
				parents[t.Name] = true
				break
			}
		}
	}

	var leaves []*TestResult
	for _, t := range p.Tests {
		if !parents[t.Name] {
			leaves = append(leaves, t)
		}
	}
	return leaves
}

// PassCount returns the number of leaf tests that passed.
func (p *PackageResult) PassCount() int {
	return p.countByStatus(StatusPass)
}

// FailCount returns the number of leaf tests that failed.
func (p *PackageResult) FailCount() int {
	return p.countByStatus(StatusFail)
}

// SkipCount returns the number of leaf tests that were skipped.
func (p *PackageResult) SkipCount() int {
	return p.countByStatus(StatusSkip)
}

// TotalCount returns the total number of leaf tests.
func (p *PackageResult) TotalCount() int {
	return len(p.LeafTests())
}

// FailedTests returns the leaf tests that failed.
func (p *PackageResult) FailedTests() []*TestResult {
	var failed []*TestResult
	for _, t := range p.LeafTests() {
		if t.Status == StatusFail {
			failed = append(failed, t)
		}
	}
	return failed
}

func (p *PackageResult) countByStatus(status TestStatus) int {
	count := 0
	for _, t := range p.LeafTests() {
		if t.Status == status {
			count++
		}
	}
	return count
}
