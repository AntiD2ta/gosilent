package runner

import (
	"context"
	"io"
	"os/exec"
)

// Result holds the output and exit status of a command execution.
type Result struct {
	Stdout io.Reader
	cmd    *exec.Cmd
}

// Wait blocks until the command finishes and returns any error.
func (r *Result) Wait() error {
	return r.cmd.Wait()
}

// Run starts the named command with the given arguments and returns a Result
// whose Stdout can be consumed while the process is still running.
func Run(ctx context.Context, name string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- gosilent intentionally executes subprocess (go test)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Result{
		Stdout: stdout,
		cmd:    cmd,
	}, nil
}
