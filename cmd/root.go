package cmd

import (
	"os"
	"path/filepath"

	"github.com/RDX463/github-work-summary/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "github-work-summary",
	Short:             "Summarize your GitHub work from the last 24 hours",
	SilenceUsage:      true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		maybeNotifyUpdate(cmd)
	},
}

func Execute() error {
	// Match help/usage command name to whichever executable name was used
	// (for example, "gws" via alias/symlink).
	if len(os.Args) > 0 {
		if execName := filepath.Base(os.Args[0]); execName != "" {
			rootCmd.Use = execName
		}
	}
	rootCmd.Version = version.Current()
	return rootCmd.Execute()
}
