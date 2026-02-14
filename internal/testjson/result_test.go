package testjson_test

import (
	"testing"

	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/stretchr/testify/require"
)

func TestPackageResult_LeafTests(t *testing.T) {
	tests := []struct {
		name     string
		tests    []*testjson.TestResult
		expected []string // expected leaf test names
	}{
		{
			name:     "NoTests",
			tests:    nil,
			expected: nil,
		},
		{
			name: "AllLeafTests",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusFail},
			},
			expected: []string{"TestA", "TestB"},
		},
		{
			name: "WithSubtests",
			tests: []*testjson.TestResult{
				{Name: "TestFoo", Status: testjson.StatusPass},
				{Name: "TestFoo/A", Status: testjson.StatusPass},
				{Name: "TestFoo/B", Status: testjson.StatusFail},
			},
			expected: []string{"TestFoo/A", "TestFoo/B"},
		},
		{
			name: "NestedSubtests",
			tests: []*testjson.TestResult{
				{Name: "TestFoo", Status: testjson.StatusPass},
				{Name: "TestFoo/A", Status: testjson.StatusPass},
				{Name: "TestFoo/A/X", Status: testjson.StatusPass},
				{Name: "TestFoo/A/Y", Status: testjson.StatusPass},
				{Name: "TestFoo/B", Status: testjson.StatusPass},
			},
			expected: []string{"TestFoo/A/X", "TestFoo/A/Y", "TestFoo/B"},
		},
		{
			name: "MixedLeafAndParent",
			tests: []*testjson.TestResult{
				{Name: "TestAlone", Status: testjson.StatusPass},
				{Name: "TestParent", Status: testjson.StatusPass},
				{Name: "TestParent/Sub", Status: testjson.StatusPass},
			},
			expected: []string{"TestAlone", "TestParent/Sub"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := &testjson.PackageResult{Tests: test.tests}
			leaves := pkg.LeafTests()
			var names []string
			for _, leaf := range leaves {
				names = append(names, leaf.Name)
			}
			require.Equal(t, test.expected, names)
		})
	}
}

func TestPackageResult_Counts(t *testing.T) {
	tests := []struct {
		name      string
		tests     []*testjson.TestResult
		passCount int
		failCount int
		skipCount int
		total     int
	}{
		{
			name:      "NoTests",
			tests:     nil,
			passCount: 0,
			failCount: 0,
			skipCount: 0,
			total:     0,
		},
		{
			name: "AllPass",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusPass},
			},
			passCount: 2,
			failCount: 0,
			skipCount: 0,
			total:     2,
		},
		{
			name: "Mixed",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusFail},
				{Name: "TestC", Status: testjson.StatusSkip},
			},
			passCount: 1,
			failCount: 1,
			skipCount: 1,
			total:     3,
		},
		{
			name: "SubtestsExcludeParents",
			tests: []*testjson.TestResult{
				{Name: "TestFoo", Status: testjson.StatusPass},
				{Name: "TestFoo/A", Status: testjson.StatusPass},
				{Name: "TestFoo/B", Status: testjson.StatusFail},
				{Name: "TestBar", Status: testjson.StatusSkip},
			},
			passCount: 1,
			failCount: 1,
			skipCount: 1,
			total:     3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := &testjson.PackageResult{Tests: test.tests}
			require.Equal(t, test.passCount, pkg.PassCount(), "PassCount")
			require.Equal(t, test.failCount, pkg.FailCount(), "FailCount")
			require.Equal(t, test.skipCount, pkg.SkipCount(), "SkipCount")
			require.Equal(t, test.total, pkg.TotalCount(), "TotalCount")
		})
	}
}

func TestPackageResult_FailedTests(t *testing.T) {
	tests := []struct {
		name     string
		tests    []*testjson.TestResult
		expected []string // expected failed test names
	}{
		{
			name:     "NoTests",
			tests:    nil,
			expected: nil,
		},
		{
			name: "AllPass",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusPass},
			},
			expected: nil,
		},
		{
			name: "OneFailing",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusFail, Output: []string{"expected 3, got 5"}},
			},
			expected: []string{"TestB"},
		},
		{
			name: "MultipleFailures",
			tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusFail, Output: []string{"err A"}},
				{Name: "TestB", Status: testjson.StatusPass},
				{Name: "TestC", Status: testjson.StatusFail, Output: []string{"err C"}},
			},
			expected: []string{"TestA", "TestC"},
		},
		{
			name: "SubtestFailure",
			tests: []*testjson.TestResult{
				{Name: "TestFoo", Status: testjson.StatusFail},
				{Name: "TestFoo/A", Status: testjson.StatusPass},
				{Name: "TestFoo/B", Status: testjson.StatusFail, Output: []string{"sub failed"}},
			},
			expected: []string{"TestFoo/B"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := &testjson.PackageResult{Tests: test.tests}
			failed := pkg.FailedTests()
			var names []string
			for _, f := range failed {
				names = append(names, f.Name)
			}
			require.Equal(t, test.expected, names)
		})
	}
}
