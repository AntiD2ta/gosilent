package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// NewApp creates the gosilent CLI application.
func NewApp(version string) *cli.App {
	return &cli.App{
		Name:  "gosilent",
		Usage: "Context-efficient test runner for AI agents",
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print the version",
				Action: func(c *cli.Context) error {
					fmt.Fprintf(c.App.Writer, "gosilent %s\n", version)
					return nil
				},
			},
		},
	}
}
