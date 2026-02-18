package testcmd

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantArgs    []string
		wantVerbose bool
		wantDetail  bool
		wantFlags   []string
	}{
		{
			name:     "BasicPackagePattern",
			args:     []string{"./..."},
			wantArgs: []string{"test", "-json", "./..."},
		},
		{
			name:      "WithRaceFlag",
			args:      []string{"-race", "./pkg/"},
			wantArgs:  []string{"test", "-json", "-race", "./pkg/"},
			wantFlags: []string{"-race"},
		},
		{
			name:      "WithMultipleFlags",
			args:      []string{"-race", "-count=1", "./..."},
			wantArgs:  []string{"test", "-json", "-race", "-count=1", "./..."},
			wantFlags: []string{"-race"},
		},
		{
			name:      "WithRunFlag",
			args:      []string{"-run", "TestFoo", "./pkg/"},
			wantArgs:  []string{"test", "-json", "-run", "TestFoo", "./pkg/"},
			wantFlags: []string{"-run TestFoo"},
		},
		{
			name:     "NoArgs",
			args:     []string{},
			wantArgs: []string{"test", "-json"},
		},
		{
			name:        "VerboseMode",
			args:        []string{"--verbose", "./..."},
			wantArgs:    []string{"test", "./..."},
			wantVerbose: true,
		},
		{
			name:        "VerboseModeWithFlags",
			args:        []string{"-race", "--verbose", "./pkg/"},
			wantArgs:    []string{"test", "-race", "./pkg/"},
			wantVerbose: true,
			wantFlags:   []string{"-race"},
		},
		{
			name:        "VerboseModeAtEnd",
			args:        []string{"./...", "--verbose"},
			wantArgs:    []string{"test", "./..."},
			wantVerbose: true,
		},
		{
			name:     "JsonAlreadyPresent",
			args:     []string{"-json", "./..."},
			wantArgs: []string{"test", "-json", "./..."},
		},
		{
			name:     "JsonAlreadyPresentWithOtherFlags",
			args:     []string{"-race", "-json", "./pkg/"},
			wantArgs: []string{"test", "-race", "-json", "./pkg/"},
			wantFlags: []string{"-race"},
		},
		{
			name:      "FlagRace",
			args:      []string{"-race", "./..."},
			wantArgs:  []string{"test", "-json", "-race", "./..."},
			wantFlags: []string{"-race"},
		},
		{
			name:      "FlagShort",
			args:      []string{"-short", "./..."},
			wantArgs:  []string{"test", "-json", "-short", "./..."},
			wantFlags: []string{"-short"},
		},
		{
			name:      "FlagTagsEquals",
			args:      []string{"-tags=integration", "./..."},
			wantArgs:  []string{"test", "-json", "-tags=integration", "./..."},
			wantFlags: []string{"-tags integration"},
		},
		{
			name:      "FlagTagsSpace",
			args:      []string{"-tags", "integration", "./..."},
			wantArgs:  []string{"test", "-json", "-tags", "integration", "./..."},
			wantFlags: []string{"-tags integration"},
		},
		{
			name:      "FlagCountNonDefault",
			args:      []string{"-count=5", "./..."},
			wantArgs:  []string{"test", "-json", "-count=5", "./..."},
			wantFlags: []string{"-count 5"},
		},
		{
			name:     "FlagCountDefault",
			args:     []string{"-count=1", "./..."},
			wantArgs: []string{"test", "-json", "-count=1", "./..."},
		},
		{
			name:      "FlagRunEquals",
			args:      []string{"-run=TestFoo", "./..."},
			wantArgs:  []string{"test", "-json", "-run=TestFoo", "./..."},
			wantFlags: []string{"-run TestFoo"},
		},
		{
			name:      "MultipleSignificantFlags",
			args:      []string{"-race", "-short", "-tags=integration", "./..."},
			wantArgs:  []string{"test", "-json", "-race", "-short", "-tags=integration", "./..."},
			wantFlags: []string{"-race", "-short", "-tags integration"},
		},
		{
			name:      "FlagCountSpaceNonDefault",
			args:      []string{"-count", "5", "./..."},
			wantArgs:  []string{"test", "-json", "-count", "5", "./..."},
			wantFlags: []string{"-count 5"},
		},
		{
			name:     "FlagCountSpaceDefault",
			args:     []string{"-count", "1", "./..."},
			wantArgs: []string{"test", "-json", "-count", "1", "./..."},
		},
		{
			name:       "DetailMode",
			args:       []string{"--detail", "./..."},
			wantArgs:   []string{"test", "-json", "./..."},
			wantDetail: true,
		},
		{
			name:       "DetailModeWithFlags",
			args:       []string{"-race", "--detail", "./pkg/"},
			wantArgs:   []string{"test", "-json", "-race", "./pkg/"},
			wantDetail: true,
			wantFlags:  []string{"-race"},
		},
		{
			name:        "DetailAndVerbose",
			args:        []string{"--detail", "--verbose", "./..."},
			wantArgs:    []string{"test", "./..."},
			wantVerbose: true,
			wantDetail:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotArgs, gotVerbose, gotDetail, gotFlags := buildArgs(test.args)
			require.Equal(t, test.wantArgs, gotArgs)
			require.Equal(t, test.wantVerbose, gotVerbose)
			require.Equal(t, test.wantDetail, gotDetail)
			require.Equal(t, test.wantFlags, gotFlags)
		})
	}
}

// makeExitError runs a command that exits with the given code, returning the *exec.ExitError.
func makeExitError(code int) error {
	return exec.Command("sh", "-c", fmt.Sprintf("exit %d", code)).Run()
}

func TestResolveExit(t *testing.T) {
	tests := []struct {
		name         string
		hasFailures  bool
		waitErr      error
		wantCode     int  // expected exit code; -1 means nil error expected
		wantOtherErr bool // true if expecting a non-ExitCoder error
	}{
		{
			name:        "FailuresDetected_NilWaitErr",
			hasFailures: true,
			waitErr:     nil,
			wantCode:    1,
		},
		{
			name:        "FailuresDetected_WithWaitErr",
			hasFailures: true,
			waitErr:     makeExitError(1),
			wantCode:    1,
		},
		{
			name:     "NoFailures_NilWaitErr",
			waitErr:  nil,
			wantCode: -1,
		},
		{
			name:     "NoFailures_WaitExitError",
			waitErr:  makeExitError(2),
			wantCode: 2,
		},
		{
			name:         "NoFailures_WaitOtherError",
			waitErr:      errors.New("pipe broken"),
			wantOtherErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := resolveExit(test.hasFailures, test.waitErr)

			if test.wantCode == -1 {
				require.NoError(t, err)
				return
			}
			if test.wantOtherErr {
				require.Error(t, err)
				var exitCoder cli.ExitCoder
				require.False(t, errors.As(err, &exitCoder), "expected non-ExitCoder error")
				return
			}

			require.Error(t, err)
			var exitCoder cli.ExitCoder
			require.True(t, errors.As(err, &exitCoder), "expected ExitCoder")
			require.Equal(t, test.wantCode, exitCoder.ExitCode())
		})
	}
}
