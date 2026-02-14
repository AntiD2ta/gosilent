package runner_test

import (
	"context"
	"io"
	"testing"

	"github.com/AntiD2ta/gosilent/internal/runner"
	"github.com/stretchr/testify/require"
)

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running process.
	result, err := runner.Run(ctx, "sleep", "30")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Cancel the context to kill the process.
	cancel()

	// Drain stdout before calling Wait.
	_, _ = io.ReadAll(result.Stdout)

	err = result.Wait()
	require.Error(t, err)
}

func TestRun_FailingCommand(t *testing.T) {
	ctx := context.Background()
	result, err := runner.Run(ctx, "sh", "-c", "exit 1")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Drain stdout before calling Wait (required by exec.Cmd contract).
	_, _ = io.ReadAll(result.Stdout)

	err = result.Wait()
	require.Error(t, err)
}

func TestRun_Success(t *testing.T) {
	ctx := context.Background()
	result, err := runner.Run(ctx, "echo", "hello")
	require.NoError(t, err)
	require.NotNil(t, result)

	out, err := io.ReadAll(result.Stdout)
	require.NoError(t, err)
	require.Equal(t, "hello\n", string(out))

	err = result.Wait()
	require.NoError(t, err)
}
