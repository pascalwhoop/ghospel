package main

import (
	"log"
	"os"

	"github.com/pascalwhoop/ghospel/internal/cli"
)

// Version information injected at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Version = version
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
