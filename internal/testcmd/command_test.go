package testcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantArgs    []string
		wantVerbose bool
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotArgs, gotVerbose, gotFlags := buildArgs(test.args)
			require.Equal(t, test.wantArgs, gotArgs)
			require.Equal(t, test.wantVerbose, gotVerbose)
			require.Equal(t, test.wantFlags, gotFlags)
		})
	}
}
