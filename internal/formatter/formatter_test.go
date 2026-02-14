package formatter_test

import (
	"testing"

	"github.com/AntiD2ta/gosilent/internal/formatter"
	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/stretchr/testify/require"
)

func TestFormat_SinglePass(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package: "example.com/foo",
			Elapsed: 3.46,
			Tests: []*testjson.TestResult{
				{Name: "TestOne", Status: testjson.StatusPass},
				{Name: "TestTwo", Status: testjson.StatusPass},
			},
		},
	}

	got := formatter.Format(results)
	require.Equal(t, "PASS example.com/foo 2/2 3.46s\n", got)
}

func TestFormat_PassWithSkips(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package: "example.com/foo",
			Elapsed: 3.45,
			Tests: []*testjson.TestResult{
				{Name: "TestA", Status: testjson.StatusPass},
				{Name: "TestB", Status: testjson.StatusPass},
				{Name: "TestC", Status: testjson.StatusSkip},
			},
		},
	}

	got := formatter.Format(results)
	require.Equal(t, "PASS example.com/foo 2/3 3.45s (1 skipped)\n", got)
}

func TestFormat_Failure(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package: "example.com/bar",
			Elapsed: 2.10,
			Tests: []*testjson.TestResult{
				{Name: "TestGood", Status: testjson.StatusPass},
				{
					Name:   "TestBroken",
					Status: testjson.StatusFail,
					Output: []string{
						"=== RUN   TestBroken\n",
						"    bar_test.go:15: expected 3, got 5\n",
						"--- FAIL: TestBroken (0.00s)\n",
					},
				},
			},
		},
	}

	got := formatter.Format(results)
	expected := "FAIL example.com/bar 1/2 2.10s\n" +
		"\n" +
		"  FAIL TestBroken\n" +
		"    bar_test.go:15: expected 3, got 5\n" +
		"\n"
	require.Equal(t, expected, got)
}

func TestFormat_BuildFailure(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package:     "example.com/broken",
			BuildFailed: true,
			Output: []string{
				"./foo.go:10:5: undefined: DoesNotExist\n",
			},
		},
	}

	got := formatter.Format(results)
	expected := "FAIL example.com/broken [build failed]\n" +
		"  ./foo.go:10:5: undefined: DoesNotExist\n" +
		"\n"
	require.Equal(t, expected, got)
}

func TestFormat_NoTestFiles(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package:     "example.com/notests",
			NoTestFiles: true,
		},
	}

	got := formatter.Format(results)
	require.Equal(t, "SKIP example.com/notests [no test files]\n", got)
}

func TestFormat_MultiplePackages(t *testing.T) {
	results := []*testjson.PackageResult{
		{
			Package: "example.com/a",
			Elapsed: 1.50,
			Tests: []*testjson.TestResult{
				{Name: "TestA1", Status: testjson.StatusPass},
				{Name: "TestA2", Status: testjson.StatusPass},
			},
		},
		{
			Package: "example.com/b",
			Elapsed: 2.10,
			Tests: []*testjson.TestResult{
				{Name: "TestB1", Status: testjson.StatusPass},
				{
					Name:   "TestB2",
					Status: testjson.StatusFail,
					Output: []string{
						"=== RUN   TestB2\n",
						"    b_test.go:10: wrong answer\n",
						"--- FAIL: TestB2 (0.00s)\n",
					},
				},
			},
		},
		{
			Package:     "example.com/c",
			NoTestFiles: true,
		},
		{
			Package: "example.com/d",
			Elapsed: 0.50,
			Tests: []*testjson.TestResult{
				{Name: "TestD1", Status: testjson.StatusPass},
			},
		},
	}

	got := formatter.Format(results)
	expected := "PASS example.com/a 2/2 1.50s\n" +
		"FAIL example.com/b 1/2 2.10s\n" +
		"\n" +
		"  FAIL TestB2\n" +
		"    b_test.go:10: wrong answer\n" +
		"\n" +
		"SKIP example.com/c [no test files]\n" +
		"PASS example.com/d 1/1 0.50s\n" +
		"2 passed, 1 failed, 1 skipped (4.10s)\n"
	require.Equal(t, expected, got)
}

func TestHasFailures(t *testing.T) {
	tests := []struct {
		name     string
		results  []*testjson.PackageResult
		expected bool
	}{
		{
			name: "AllPass",
			results: []*testjson.PackageResult{
				{Package: "a", Tests: []*testjson.TestResult{
					{Name: "T1", Status: testjson.StatusPass},
				}},
			},
			expected: false,
		},
		{
			name: "TestFailure",
			results: []*testjson.PackageResult{
				{Package: "a", Tests: []*testjson.TestResult{
					{Name: "T1", Status: testjson.StatusFail},
				}},
			},
			expected: true,
		},
		{
			name: "BuildFailure",
			results: []*testjson.PackageResult{
				{Package: "a", BuildFailed: true},
			},
			expected: true,
		},
		{
			name: "NoTestFiles",
			results: []*testjson.PackageResult{
				{Package: "a", NoTestFiles: true},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := formatter.HasFailures(test.results)
			require.Equal(t, test.expected, got)
		})
	}
}
