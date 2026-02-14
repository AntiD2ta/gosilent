package testcmd

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/AntiD2ta/gosilent/internal/formatter"
	"github.com/AntiD2ta/gosilent/internal/runner"
	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/urfave/cli/v2"
)

// buildArgs constructs the arguments for 'go test' from the raw args
// passed after 'gosilent test'. It returns the full go args slice and
// whether --verbose mode was requested.
func buildArgs(args []string) (goArgs []string, verbose bool) {
	// Scan args for --verbose (strip it) and -json (detect it).
	var filtered []string
	hasJSON := false
	for _, arg := range args {
		if arg == "--verbose" {
			verbose = true
			continue
		}
		if arg == "-json" {
			hasJSON = true
		}
		filtered = append(filtered, arg)
	}

	goArgs = []string{"test"}
	if !verbose && !hasJSON {
		goArgs = append(goArgs, "-json")
	}
	goArgs = append(goArgs, filtered...)
	return goArgs, verbose
}

// Command returns the "test" subcommand for the gosilent CLI.
func Command() *cli.Command {
	return &cli.Command{
		Name:            "test",
		Usage:           "Run go test with compact output",
		SkipFlagParsing: true,
		Action: func(c *cli.Context) error {
			goArgs, verbose := buildArgs(c.Args().Slice())

			result, err := runner.Run(c.Context, "go", goArgs...)
			if err != nil {
				return fmt.Errorf("failed to start go test: %w", err)
			}

			if verbose {
				return runVerbose(c.App.Writer, result)
			}
			return runJSON(c.App.Writer, result)
		},
	}
}

// runVerbose passes go test output through without JSON parsing.
func runVerbose(w io.Writer, result *runner.Result) error {
	io.Copy(w, result.Stdout)
	waitErr := result.Wait()
	// Propagate exit code from go test.
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return cli.Exit("", exitErr.ExitCode())
	}
	return waitErr
}

// runJSON parses JSON output and formats compact results.
func runJSON(w io.Writer, result *runner.Result) error {
	results, parseErr := testjson.Parse(result.Stdout)
	waitErr := result.Wait()

	if parseErr != nil {
		return fmt.Errorf("failed to parse test output: %w", parseErr)
	}

	fmt.Fprint(w, formatter.Format(results))

	if formatter.HasFailures(results) {
		return cli.Exit("", 1)
	}

	// Ignore waitErr when we have valid parsed results — go test returns
	// non-zero on test failures, but we determine failure from parsed output.
	_ = waitErr
	return nil
}
