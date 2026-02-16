package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// binaryPath holds the absolute path to the built gosilent binary.
var binaryPath string

func TestMain(m *testing.M) {
	// Prevent infinite recursion: when TestE2E_SelfTest runs gosilent test ./...,
	// the nested go test would re-enter this TestMain. Skip if already nested.
	if os.Getenv("GOSILENT_E2E_NESTED") == "1" {
		os.Exit(0)
	}

	// Build the binary once for all E2E tests.
	tmp, err := os.MkdirTemp("", "gosilent-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	binaryPath = filepath.Join(tmp, "gosilent")
	build := exec.Command("go", "build", "-o", binaryPath, "./cmd/gosilent/")
	build.Dir = projectRoot()
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build gosilent: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// projectRoot returns the absolute path to the gosilent project root.
func projectRoot() string {
	// e2e/ is one level below the project root.
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(dir)
}

// fixtureDir returns the absolute path to a fixture project.
func fixtureDir(name string) string {
	return filepath.Join(projectRoot(), "testdata", "projects", name)
}

// runGosilent runs the gosilent binary with the given args and working directory.
// Returns combined output and exit code.
func runGosilent(t *testing.T, dir string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	// Set GOSILENT_E2E_NESTED to prevent infinite recursion when the self-test
	// runs gosilent test ./... (which would re-enter this E2E suite).
	cmd.Env = append(os.Environ(), "GOSILENT_E2E_NESTED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), exitErr.ExitCode()
		}
		t.Fatalf("unexpected error running gosilent: %v", err)
	}
	return string(out), 0
}

func TestE2E_PassingPackage(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("passing"), "test", "./...")

	require.Equal(t, 0, exitCode, "exit code should be 0 for passing tests")
	require.Contains(t, stdout, "ok  passing")
	require.Contains(t, stdout, "2 tests")
	require.NotContains(t, stdout, "FAIL")
	// Should be a single line (compact format).
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	require.Len(t, lines, 1, "single passing package should produce exactly one line")
}

func TestE2E_FailingPackage(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("failing"), "test", "./...")

	require.Equal(t, 1, exitCode, "exit code should be 1 for failing tests")
	require.Contains(t, stdout, "FAIL  example.com/failing")
	require.Contains(t, stdout, "TestDoubleBroken")
	require.Contains(t, stdout, "Double(5) = 10, want 99", "should include assertion message")
	require.Contains(t, stdout, "1 failed")
}

func TestE2E_BuildBroken(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("buildbroken"), "test", "./...")

	require.Equal(t, 1, exitCode, "exit code should be 1 for build failures")
	require.Contains(t, stdout, "FAIL  example.com/buildbroken [build failed]")
	require.Contains(t, stdout, "undefined: DoesNotExist", "should include compiler error")
	require.Contains(t, stdout, "1 failed")
}

func TestE2E_WithSkips(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("withskips"), "test", "./...")

	require.Equal(t, 0, exitCode, "exit code should be 0 when tests pass with skips")
	require.Contains(t, stdout, "ok  withskips")
	require.Contains(t, stdout, "3 tests")
	require.NotContains(t, stdout, "skipped", "compact mode should not mention skips")
	// Should be a single line (compact format).
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	require.Len(t, lines, 1, "passing package with skips should produce exactly one line")
}

func TestE2E_VerboseMode(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("passing"), "test", "--verbose", "-v", "-count=1", "./...")

	require.Equal(t, 0, exitCode, "exit code should be 0 for passing tests in verbose mode")
	// Verbose mode passes through raw go test output, which includes test markers.
	require.Contains(t, stdout, "--- PASS:", "verbose mode should show raw go test markers")
	require.Contains(t, stdout, "=== RUN", "verbose mode should show RUN markers")
	// Should NOT contain gosilent's compact format.
	require.NotContains(t, stdout, "PASS example.com/passing 2/2")
}

func TestE2E_SelfTest(t *testing.T) {
	stdout, exitCode := runGosilent(t, projectRoot(), "test", "./...")

	require.Equal(t, 0, exitCode, "gosilent's own tests should pass")

	// Compact format: single line for the entire suite.
	require.Contains(t, stdout, "ok  all packages")
	require.Regexp(t, regexp.MustCompile(`\d+ tests`), stdout, "should show total test count")
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	require.Len(t, lines, 1, "compact output should be exactly one line")

	// No FAIL, no per-package names.
	require.NotContains(t, stdout, "FAIL")
	require.NotContains(t, stdout, "github.com/AntiD2ta/gosilent")
}

func TestE2E_DetailMode(t *testing.T) {
	stdout, exitCode := runGosilent(t, fixtureDir("passing"), "test", "--detail", "./...")

	require.Equal(t, 0, exitCode, "exit code should be 0 for passing tests in detail mode")
	// Detail mode restores the old per-package format.
	require.Contains(t, stdout, "PASS example.com/passing")
	require.Contains(t, stdout, "2/2")
	// Should NOT contain compact format markers.
	require.NotContains(t, stdout, "ok  passing")
}

func TestE2E_SelfTest_Detail(t *testing.T) {
	stdout, exitCode := runGosilent(t, projectRoot(), "test", "--detail", "./...")

	require.Equal(t, 0, exitCode, "gosilent's own tests should pass in detail mode")

	// Detail mode shows each package individually.
	knownPackages := []string{
		"github.com/AntiD2ta/gosilent/internal/cli",
		"github.com/AntiD2ta/gosilent/internal/formatter",
		"github.com/AntiD2ta/gosilent/internal/runner",
		"github.com/AntiD2ta/gosilent/internal/testcmd",
		"github.com/AntiD2ta/gosilent/internal/testjson",
		"github.com/AntiD2ta/gosilent/e2e",
	}
	for _, pkg := range knownPackages {
		require.Contains(t, stdout, pkg, "detail output should include package %s", pkg)
	}

	// Summary line with passed count (multiple packages → summary).
	require.Regexp(t, regexp.MustCompile(`\d+ passed.*\(\d+\.\d+s\)`), stdout, "should have a summary line")
}
