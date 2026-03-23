package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "github-work-summary",
	Short:        "Summarize your GitHub work from the last 24 hours",
	SilenceUsage: true,
}

func Execute() error {
	// Match help/usage command name to whichever executable name was used
	// (for example, "gws" via alias/symlink).
	if len(os.Args) > 0 {
		if execName := filepath.Base(os.Args[0]); execName != "" {
			rootCmd.Use = execName
		}
	}
	return rootCmd.Execute()
}
