package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

	for {
		action, err := ui.RunHomeMenu(inFile, cmd.OutOrStdout(), ui.HomeMenuOptions{
			RepositoryURL: "https://github.com/RDX463/github-work-summary",
			Tagline:       "Summarize your GitHub work from terminal.",
		})
		if err != nil {
			return err
		}

		switch action {
		case ui.HomeActionSummary:
			if err := runSummary(cmd); err != nil {
				return err
			}
		case ui.HomeActionRepos:
			if err := runRepos(cmd); err != nil {
				return err
			}
		case ui.HomeActionLogin:
			if err := runLogin(cmd); err != nil {
				return err
			}
		case ui.HomeActionLogout:
			if err := runLogout(cmd); err != nil {
				return err
			}
		case ui.HomeActionHelp:
			if err := cmd.Help(); err != nil {
				return err
			}
		case ui.HomeActionVersion:
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s %s\n", ui.Cyan(out, fmt.Sprintf("%s version", cmd.Use)), ui.Bold(out, version.Current()))
		case ui.HomeActionQuit:
			return nil
		default:
			continue
		}

		if err := promptReturnToMenu(inFile, cmd.OutOrStdout()); err != nil {
			return err
		}
	}
}

func promptReturnToMenu(in *os.File, out io.Writer) error {
	fmt.Fprintf(out, "\n%s", ui.Gray(out, "Press Enter to return to menu..."))
	reader := bufio.NewReader(in)
	_, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	fmt.Fprintln(out)
	return nil
}
