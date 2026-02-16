package testcmd

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/AntiD2ta/gosilent/internal/formatter"
	"github.com/AntiD2ta/gosilent/internal/runner"
	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/urfave/cli/v2"
)

// significantBoolFlags are flags that change test behavior and are shown in output.
var significantBoolFlags = map[string]bool{
	"-race":  true,
	"-short": true,
}

// significantValueFlags are flags with a value that are shown in output.
var significantValueFlags = map[string]bool{
	"-tags":  true,
	"-count": true,
	"-run":   true,
}

// buildArgs constructs the arguments for 'go test' from the raw args
// passed after 'gosilent test'. It returns the full go args slice,
// whether --verbose mode was requested, whether --detail mode was
// requested, and any significant flags extracted for display in the output.
func buildArgs(args []string) (goArgs []string, verbose bool, detail bool, flags []string) {
	// Scan args for --verbose, --detail (strip them), -json (detect it), and significant flags.
	var filtered []string
	hasJSON := false
	for i := range args {
		arg := args[i]
		if arg == "--verbose" {
			verbose = true
			continue
		}
		if arg == "--detail" {
			detail = true
			continue
		}
		if arg == "-json" {
			hasJSON = true
		}
		filtered = append(filtered, arg)

		// Check for significant boolean flags.
		if significantBoolFlags[arg] {
			flags = append(flags, arg)
			continue
		}

		// Check for significant value flags (=form or space-separated).
		flagName, value, hasEquals := strings.Cut(arg, "=")
		if significantValueFlags[flagName] {
			if !hasEquals && i+1 < len(args) {
				value = args[i+1]
				// Don't consume next arg from filtered — it'll be added normally.
			}
			// Suppress -count=1 (default value).
			if flagName == "-count" && value == "1" {
				continue
			}
			flags = append(flags, flagName+" "+value)
		}
	}

	goArgs = []string{"test"}
	if !verbose && !hasJSON {
		goArgs = append(goArgs, "-json")
	}
	goArgs = append(goArgs, filtered...)
	return goArgs, verbose, detail, flags
}

// Command returns the "test" subcommand for the gosilent CLI.
func Command() *cli.Command {
	return &cli.Command{
		Name:            "test",
		Usage:           "Run go test with compact output",
		SkipFlagParsing: true,
		Action: func(c *cli.Context) error {
			goArgs, verbose, detail, flags := buildArgs(c.Args().Slice())

			result, err := runner.Run(c.Context, "go", goArgs...)
			if err != nil {
				return fmt.Errorf("failed to start go test: %w", err)
			}

			if verbose {
				return runVerbose(c.App.Writer, result)
			}
			return runJSON(c.App.Writer, result, detail, flags)
		},
	}
}

// runVerbose passes go test output through without JSON parsing.
func runVerbose(w io.Writer, result *runner.Result) error {
	_, _ = io.Copy(w, result.Stdout)
	waitErr := result.Wait()
	// Propagate exit code from go test.
	if exitErr, ok := errors.AsType[*exec.ExitError](waitErr); ok {
		return cli.Exit("", exitErr.ExitCode())
	}
	return waitErr
}

// runJSON parses JSON output and formats results.
// When detail is true, uses the per-package FormatDetail output;
// otherwise uses the compact Format output.
func runJSON(w io.Writer, result *runner.Result, detail bool, flags []string) error {
	results, parseErr := testjson.Parse(result.Stdout)
	waitErr := result.Wait()

	if parseErr != nil {
		return fmt.Errorf("failed to parse test output: %w", parseErr)
	}

	if detail {
		_, _ = fmt.Fprint(w, formatter.FormatDetail(results))
	} else {
		_, _ = fmt.Fprint(w, formatter.Format(results, flags))
	}

	if formatter.HasFailures(results) {
		return cli.Exit("", 1)
	}

	// Ignore waitErr when we have valid parsed results — go test returns
	// non-zero on test failures, but we determine failure from parsed output.
	_ = waitErr
	return nil
}
