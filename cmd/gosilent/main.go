package main

import (
	"log"
	"os"

	"github.com/AntiD2ta/gosilent/internal/cli"
)

// version is set via -ldflags at build time.
var version = "dev"

func main() {
	app := cli.NewApp(version)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
