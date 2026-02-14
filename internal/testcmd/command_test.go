package testcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantArgs     []string
		wantVerbose  bool
	}{
		{
			name:     "BasicPackagePattern",
			args:     []string{"./..."},
			wantArgs: []string{"test", "-json", "./..."},
		},
		{
			name:     "WithRaceFlag",
			args:     []string{"-race", "./pkg/"},
			wantArgs: []string{"test", "-json", "-race", "./pkg/"},
		},
		{
			name:     "WithMultipleFlags",
			args:     []string{"-race", "-count=1", "./..."},
			wantArgs: []string{"test", "-json", "-race", "-count=1", "./..."},
		},
		{
			name:     "WithRunFlag",
			args:     []string{"-run", "TestFoo", "./pkg/"},
			wantArgs: []string{"test", "-json", "-run", "TestFoo", "./pkg/"},
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
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotArgs, gotVerbose := buildArgs(test.args)
			require.Equal(t, test.wantArgs, gotArgs)
			require.Equal(t, test.wantVerbose, gotVerbose)
		})
	}
}
