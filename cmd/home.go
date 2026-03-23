package cmd

import (
	"fmt"
	"os"

	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/RDX463/github-work-summary/internal/version"
	"github.com/spf13/cobra"
)

func runHome(cmd *cobra.Command) error {
	in := cmd.InOrStdin()
	if !ui.IsInteractiveTerminal(in) {
		return cmd.Help()
	}

	inFile, ok := in.(*os.File)
	if !ok {
		return cmd.Help()
	}

	action, err := ui.RunHomeMenu(inFile, cmd.OutOrStdout(), ui.HomeMenuOptions{
		RepositoryURL: "https://github.com/RDX463/github-work-summary",
		Tagline:       "Summarize your GitHub work from terminal.",
	})
	if err != nil {
		return err
	}

	switch action {
	case ui.HomeActionSummary:
		return runSummary(cmd)
	case ui.HomeActionRepos:
		return runRepos(cmd)
	case ui.HomeActionLogin:
		return runLogin(cmd)
	case ui.HomeActionLogout:
		return runLogout(cmd)
	case ui.HomeActionHelp:
		return cmd.Help()
	case ui.HomeActionVersion:
		fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", cmd.Use, version.Current())
		return nil
	case ui.HomeActionQuit:
		return nil
	default:
		return nil
	}
}
