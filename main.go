package main

import (
	"os"

	"github.com/rohan/github-work-summary/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
