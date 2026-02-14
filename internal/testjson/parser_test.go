package testjson_test

import (
	"os"
	"strings"
	"testing"

	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/stretchr/testify/require"
)

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("../../testdata/" + name)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}

func TestParse_AllPass(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "all_pass.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/foo", pkg.Package)
	require.Equal(t, 3.456, pkg.Elapsed)
	require.False(t, pkg.BuildFailed)
	require.False(t, pkg.NoTestFiles)
	require.Equal(t, 2, pkg.TotalCount())
	require.Equal(t, 2, pkg.PassCount())
	require.Equal(t, 0, pkg.FailCount())
}

func TestParse_OneFailure(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "one_failure.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/bar", pkg.Package)
	require.Equal(t, 2.1, pkg.Elapsed)
	require.False(t, pkg.BuildFailed)
	require.Equal(t, 2, pkg.TotalCount())
	require.Equal(t, 1, pkg.PassCount())
	require.Equal(t, 1, pkg.FailCount())

	failed := pkg.FailedTests()
	require.Len(t, failed, 1)
	require.Equal(t, "TestBroken", failed[0].Name)
	require.Contains(t, failed[0].Output, "    bar_test.go:15: expected 3, got 5\n")
}

func TestParse_WithSubtests(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "with_subtests.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/sub", pkg.Package)
	require.Equal(t, 1.2, pkg.Elapsed)

	// LeafTests should be the 2 subtests, not the parent TestMath.
	require.Equal(t, 2, pkg.TotalCount())
	require.Equal(t, 1, pkg.PassCount())
	require.Equal(t, 1, pkg.FailCount())

	failed := pkg.FailedTests()
	require.Len(t, failed, 1)
	require.Equal(t, "TestMath/Div", failed[0].Name)
	require.Contains(t, failed[0].Output, "    math_test.go:25: division by zero\n")
}

func TestParse_BuildError(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "build_error.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/broken", pkg.Package)
	require.True(t, pkg.BuildFailed)
	require.Equal(t, 0, pkg.TotalCount())

	// Build error output should be captured in package-level output.
	require.Contains(t, pkg.Output, "./foo.go:10:5: undefined: DoesNotExist\n")
}

func TestParse_NoTestFiles(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "no_test_files.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/notests", pkg.Package)
	require.True(t, pkg.NoTestFiles)
	require.Equal(t, 0, pkg.TotalCount())
}

func TestParse_MultiplePackages(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "multiple_packages.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Results should appear in finalization order.
	// beta finishes before alpha in the fixture.
	pkgNames := []string{results[0].Package, results[1].Package}
	require.Contains(t, pkgNames, "example.com/alpha")
	require.Contains(t, pkgNames, "example.com/beta")

	// Find each package.
	var alpha, beta *testjson.PackageResult
	for _, r := range results {
		switch r.Package {
		case "example.com/alpha":
			alpha = r
		case "example.com/beta":
			beta = r
		}
	}

	require.Equal(t, 1.5, alpha.Elapsed)
	require.Equal(t, 2, alpha.TotalCount())
	require.Equal(t, 2, alpha.PassCount())

	require.Equal(t, 0.8, beta.Elapsed)
	require.Equal(t, 1, beta.TotalCount())
	require.Equal(t, 1, beta.PassCount())
}

func TestParse_SkippedTests(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "skipped_tests.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/skippy", pkg.Package)
	require.Equal(t, 3, pkg.TotalCount())
	require.Equal(t, 2, pkg.PassCount())
	require.Equal(t, 1, pkg.SkipCount())
	require.Equal(t, 0, pkg.FailCount())
}

func TestParse_Cached(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "cached.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/fast", pkg.Package)
	require.Equal(t, 1, pkg.TotalCount())
	require.Equal(t, 1, pkg.PassCount())
}

func TestParse_Panic(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "panic.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/panicky", pkg.Package)
	require.Equal(t, 1, pkg.TotalCount())
	require.Equal(t, 1, pkg.FailCount())

	failed := pkg.FailedTests()
	require.Len(t, failed, 1)
	require.Equal(t, "TestPanic", failed[0].Name)

	// Panic output should be captured.
	output := strings.Join(failed[0].Output, "")
	require.Contains(t, output, "panic: runtime error")
}

func TestParse_DataRace(t *testing.T) {
	results, err := testjson.Parse(openFixture(t, "data_race.jsonl"))
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/racy", pkg.Package)
	require.Equal(t, 1, pkg.TotalCount())
	require.Equal(t, 1, pkg.FailCount())

	failed := pkg.FailedTests()
	require.Len(t, failed, 1)
	output := strings.Join(failed[0].Output, "")
	require.Contains(t, output, "DATA RACE")
}

func TestParse_NonJSONLines(t *testing.T) {
	input := strings.NewReader(`not json at all
{"Action":"start","Package":"example.com/x"}
also not json
{"Action":"run","Package":"example.com/x","Test":"TestX"}
{"Action":"output","Package":"example.com/x","Test":"TestX","Output":"=== RUN   TestX\n"}
{"Action":"output","Package":"example.com/x","Test":"TestX","Output":"--- PASS: TestX (0.00s)\n"}
{"Action":"pass","Package":"example.com/x","Test":"TestX","Elapsed":0}
{"Action":"output","Package":"example.com/x","Output":"PASS\n"}
{"Action":"pass","Package":"example.com/x","Elapsed":0.1}
`)

	results, err := testjson.Parse(input)
	require.NoError(t, err)
	require.Len(t, results, 1)

	pkg := results[0]
	require.Equal(t, "example.com/x", pkg.Package)
	require.Equal(t, 1, pkg.TotalCount())
	require.Equal(t, 1, pkg.PassCount())
}
