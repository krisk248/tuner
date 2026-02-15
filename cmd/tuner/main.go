package main

import (
	"os"

	"github.com/krisk248/tuner/internal/cli"
)

var Version = "dev"

func main() {
	cli.SetVersion(Version)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
