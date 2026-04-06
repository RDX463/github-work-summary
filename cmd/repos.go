package cmd

import (
	"errors"
	"fmt"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	out := cmd.OutOrStdout()
	profileName := getActiveProfileName()
	fmt.Fprintf(out, "%s %s\n", ui.Bold(out, "Profile:"), ui.Cyan(out, profileName))

	client, err := loadGitHubClientFromKeychain()
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
		fmt.Fprintln(out, ui.Yellow(out, "No repositories found for this account."))
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
		out,
		"Select repositories to include in your summary:",
		options,
	)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			fmt.Fprintln(out, ui.Yellow(out, "Repository selection cancelled."))
			return nil
		}
		return err
	}

	fmt.Fprintf(out, "\n%s\n", ui.Bold(out, ui.Cyan(out, fmt.Sprintf("Selected %d repositories:", len(selected)))))
	// 1. Create a slice to store just the repository names
	repoNames := make([]string, 0, len(selected))
	for _, repo := range selected {
		fmt.Fprintf(out, "%s %s\n", ui.Green(out, "•"), ui.Bold(out, repo.Value))
		repoNames = append(repoNames, repo.Value)
	}

	// 2. Save to config using the profile-specific key
	viper.Set(getProfileKey(profileName, "repositories"), repoNames)
	if err := saveConfig(); err != nil {
		fmt.Fprintf(out, "\n%s %v\n", ui.Red(out, "Error saving config:"), err)
	}

	return nil
}
