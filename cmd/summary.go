package cmd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RDX463/github-work-summary/internal/auth"
	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
)

const (
	defaultSummaryWindow = 24 * time.Hour
	maxRepoConcurrency   = 6
)

type repoFetchResult struct {
	repo    string
	commits []githubapi.Commit
	err     error
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Fetch and summarize your commits from the last 24 hours",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSummary(cmd)
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command) error {
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
		fmt.Fprintln(cmd.OutOrStdout(), "No repositories found for this account.")
		return nil
	}

	selectedRepos, err := selectRepositories(cmd, repos)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			fmt.Fprintln(cmd.OutOrStdout(), "Repository selection cancelled.")
			return nil
		}
		return err
	}

	user, err := client.GetAuthenticatedUser(cmd.Context())
	if err != nil {
		if errors.Is(err, githubapi.ErrUnauthorized) {
			return fmt.Errorf("stored token is invalid or expired. run `github-work-summary login` again")
		}
		return err
	}

	windowEnd := time.Now()
	windowStart := windowEnd.Add(-defaultSummaryWindow)

	repoCommits, warnings, err := fetchCommitsAcrossRepos(cmd.Context(), client, selectedRepos, user.Login, windowStart)
	if err != nil {
		if errors.Is(err, githubapi.ErrUnauthorized) {
			return fmt.Errorf("stored token is invalid or expired. run `github-work-summary login` again")
		}
		return err
	}

	report := summary.BuildReport(windowStart, windowEnd, repoCommits)
	summary.Render(cmd.OutOrStdout(), report)

	if len(warnings) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Warnings:")
		for _, warning := range warnings {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", warning)
		}
	}
	return nil
}

func loadGitHubClientFromKeychain() (*githubapi.Client, error) {
	store := auth.NewKeyringStore(auth.DefaultServiceName, auth.DefaultTokenAccount)
	token, err := store.GetToken()
	if err != nil {
		return nil, fmt.Errorf("unable to read GitHub token: %w. run `github-work-summary login` first", err)
	}

	client, err := githubapi.NewClient(token)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func selectRepositories(cmd *cobra.Command, repos []githubapi.Repository) ([]string, error) {
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
		"Select repositories to include in your work summary:",
		options,
	)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(selected))
	for _, item := range selected {
		result = append(result, item.Value)
	}
	return result, nil
}

func fetchCommitsAcrossRepos(
	ctx context.Context,
	client *githubapi.Client,
	selectedRepos []string,
	author string,
	since time.Time,
) (map[string][]githubapi.Commit, []string, error) {
	repoCommits := make(map[string][]githubapi.Commit, len(selectedRepos))
	for _, repo := range selectedRepos {
		repoCommits[repo] = []githubapi.Commit{}
	}

	sem := make(chan struct{}, maxRepoConcurrency)
	results := make(chan repoFetchResult, len(selectedRepos))
	var wg sync.WaitGroup

	for _, repo := range selectedRepos {
		repoName := repo
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			commits, err := client.ListCommitsByAuthorSince(ctx, repoName, author, since)
			results <- repoFetchResult{
				repo:    repoName,
				commits: commits,
				err:     err,
			}
		}()
	}

	wg.Wait()
	close(results)

	warnings := make([]string, 0)
	for res := range results {
		if res.err != nil {
			if errors.Is(res.err, githubapi.ErrUnauthorized) {
				return nil, nil, res.err
			}
			warnings = append(warnings, fmt.Sprintf("%s: %v", res.repo, res.err))
			continue
		}
		repoCommits[res.repo] = res.commits
	}

	return repoCommits, warnings, nil
}
