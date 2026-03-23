package main

import (
	"os"

	"github.com/RDX463/github-work-summary/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
