package cli

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name      string
		injected  string
		buildInfo *debug.BuildInfo
		ok        bool
		expected  string
	}{
		{
			name:     "InjectedVersionWins",
			injected: "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:      "FallsBackToModuleVersion",
			injected:  "dev",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}},
			ok:        true,
			expected:  "v1.0.0",
		},
		{
			name:      "DevelBuildStaysDev",
			injected:  "dev",
			buildInfo: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:        true,
			expected:  "dev",
		},
		{
			name:     "NoBuildInfoStaysDev",
			injected: "dev",
			ok:       false,
			expected: "dev",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readBuildInfo := func() (*debug.BuildInfo, bool) {
				return test.buildInfo, test.ok
			}
			require.Equal(t, test.expected, resolveVersion(test.injected, readBuildInfo))
		})
	}
}
