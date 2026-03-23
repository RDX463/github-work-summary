package cmd

import (
	"errors"
	"fmt"

	"github.com/RDX463/github-work-summary/internal/auth"
	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Select repositories for the work summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRepos(cmd)
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}

func runRepos(cmd *cobra.Command) error {
	store := auth.NewKeyringStore(auth.DefaultServiceName, auth.DefaultTokenAccount)
	token, err := store.GetToken()
	if err != nil {
		return fmt.Errorf("unable to read GitHub token: %w. run `github-work-summary login` first", err)
	}

	client, err := githubapi.NewClient(token)
	if err != nil {
		return err
	}

	repos, err := client.ListAccessibleRepositories(cmd.Context())
	if err != nil {
		if errors.Is(err, githubapi.ErrUnauthorized) {
			return fmt.Errorf("stored token is invalid or expired. run `github-work-summary login` again")
		}
		return err
	}
	if len(repos) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No repositories found for this account.")
		return nil
	}

	options := make([]ui.SelectOption, 0, len(repos))
	for _, repo := range repos {
		label := repo.FullName
		if repo.Private {
			label += " (private)"
		} else {
			label += " (public)"
		}
		if repo.Fork {
			label += " [fork]"
		}
		if repo.Archived {
			label += " [archived]"
		}
		options = append(options, ui.SelectOption{
			Label: label,
			Value: repo.FullName,
		})
	}

	selected, err := ui.MultiSelectCheckboxes(
		cmd.InOrStdin(),
		cmd.OutOrStdout(),
		"Select repositories to include in your summary:",
		options,
	)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			fmt.Fprintln(cmd.OutOrStdout(), "Repository selection cancelled.")
			return nil
		}
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected %d repositories:\n", len(selected))
	for _, repo := range selected {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", repo.Value)
	}
	return nil
}
