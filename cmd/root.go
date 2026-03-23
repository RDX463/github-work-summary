package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "github-work-summary",
	Short:        "Summarize your GitHub work from the last 24 hours",
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}
