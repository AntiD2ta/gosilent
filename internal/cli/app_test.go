package cli_test

import (
	"bytes"
	"testing"

	gosilentcli "github.com/AntiD2ta/gosilent/internal/cli"
	"github.com/stretchr/testify/require"
	ucli "github.com/urfave/cli/v2"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name   string
		verify func(t *testing.T, app *ucli.App)
	}{
		{
			name: "HasVersionCommand",
			verify: func(t *testing.T, app *ucli.App) {
				t.Helper()
				found := false
				for _, cmd := range app.Commands {
					if cmd.Name == "version" {
						found = true
						break
					}
				}
				require.True(t, found, "app should have a 'version' command")
			},
		},
		{
			name: "HasTestCommand",
			verify: func(t *testing.T, app *ucli.App) {
				t.Helper()
				found := false
				for _, cmd := range app.Commands {
					if cmd.Name == "test" {
						found = true
						break
					}
				}
				require.True(t, found, "app should have a 'test' command")
			},
		},
		{
			name: "AppName",
			verify: func(t *testing.T, app *ucli.App) {
				t.Helper()
				require.Equal(t, "gosilent", app.Name)
			},
		},
		{
			name: "AppUsage",
			verify: func(t *testing.T, app *ucli.App) {
				t.Helper()
				require.NotEmpty(t, app.Usage)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := gosilentcli.NewApp("dev")
			require.NotNil(t, app)
			test.verify(t, app)
		})
	}
}

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "DevVersion",
			version:  "dev",
			expected: "gosilent dev\n",
		},
		{
			name:     "ReleaseVersion",
			version:  "v1.2.3",
			expected: "gosilent v1.2.3\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := gosilentcli.NewApp(test.version)
			var buf bytes.Buffer
			app.Writer = &buf

			err := app.Run([]string{"gosilent", "version"})
			require.NoError(t, err)
			require.Equal(t, test.expected, buf.String())
		})
	}
}
