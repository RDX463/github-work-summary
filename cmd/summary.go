package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/RDX463/github-work-summary/internal/auth"
	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultSummaryWindow  = 24 * time.Hour
	fallbackSummaryWindow = 30 * 24 * time.Hour
	maxRepoConcurrency    = 6
)

type repoFetchResult struct {
	repo   string
	result githubapi.BranchCommitResult
	err    error
}

type repoBranchStatus struct {
	Scanned []string
	Active  map[string]int
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Fetch and summarize your commits from the last 24 hours",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSummary(cmd)
	},
}

var summaryBranches []string
var summaryChooseBranch bool
var summarySince string
var summaryUntil string
var summaryDuration string
var summaryPickRepos bool
var summaryOutputFile string
var summaryMarkdown bool
var summaryJSON bool

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().StringSliceVarP(&summaryBranches, "branch", "b", nil, "Branch name(s) to include")
	summaryCmd.Flags().BoolVar(&summaryChooseBranch, "choose-branch", false, "Open interactive branch selector")

	// New Flags
	summaryCmd.Flags().StringVarP(&summarySince, "since", "s", "", "Start date or relative duration (e.g., '2024-03-20' or '2d')")
	summaryCmd.Flags().StringVarP(&summaryUntil, "until", "u", "", "End date (e.g., '2024-03-21')")
	summaryCmd.Flags().StringVarP(&summaryDuration, "duration", "d", "", "Time window size (e.g., '48h', '3d', '1w')")
	summaryCmd.Flags().BoolVar(&summaryPickRepos, "pick", false, "Manually select repositories instead of using saved ones")

	summaryCmd.Flags().StringVarP(&summaryOutputFile, "output", "o", "", "File to write the summary to")
	summaryCmd.Flags().BoolVarP(&summaryMarkdown, "markdown", "m", false, "Output in Markdown format")
	summaryCmd.Flags().BoolVar(&summaryJSON, "json", false, "Output in JSON format")

}

func runSummary(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
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

	// 1. Try loading from config first
	selectedRepos := viper.GetStringSlice("repositories")

	// 2. Fallback to interactive selection if nothing is saved or --pick is used
	if len(selectedRepos) == 0 || summaryPickRepos {
		var err error
		selectedRepos, err = selectRepositories(cmd, repos)
		if err != nil {
			if errors.Is(err, ui.ErrSelectionCancelled) {
				fmt.Fprintln(out, ui.Yellow(out, "Repository selection cancelled."))
				return nil
			}
			return err
		}

		// 3. Save the selection automatically for next time
		viper.Set("repositories", selectedRepos)
		saveConfig()
	} else {
		fmt.Fprintf(out, "%s %s\n\n", ui.Bold(out, "Using saved repositories:"), ui.Gray(out, strings.Join(selectedRepos, ", ")))
	}

	resolvedBranches, branchWarnings, err := resolveSummaryBranches(cmd, client, selectedRepos)
	if err != nil {
		if errors.Is(err, ui.ErrSelectionCancelled) {
			fmt.Fprintln(out, ui.Yellow(out, "Branch selection cancelled."))
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
	if summaryUntil != "" {
		parsedUntil, err := parseFlexibleTime(summaryUntil, time.Now())
		if err != nil {
			return err
		}
		windowEnd = parsedUntil
	}
	// Default to 24h
	windowStart := windowEnd.Add(-defaultSummaryWindow)
	if summarySince != "" {
		parsedSince, err := parseFlexibleTime(summarySince, windowEnd)
		if err != nil {
			return err
		}
		windowStart = parsedSince
	} else if summaryDuration != "" {
		// Use duration relative to windowEnd
		d, err := parseFlexibleTime(summaryDuration, windowEnd) // returns reference.Add(-d)
		if err != nil {
			return err
		}
		windowStart = d
	}

	repoCommits, branchStatus, warnings, err := fetchCommitsAcrossRepos(cmd.Context(), client, selectedRepos, user.Login, windowStart, resolvedBranches)
	if err != nil {
		if errors.Is(err, githubapi.ErrUnauthorized) {
			return fmt.Errorf("stored token is invalid or expired. run `github-work-summary login` again")
		}
		return err
	}

	report := summary.BuildReport(windowStart, windowEnd, repoCommits)
	allWarnings := append([]string(nil), branchWarnings...)
	allWarnings = append(allWarnings, warnings...)

	if report.TotalCommits == 0 {
		fallbackStart := windowEnd.Add(-fallbackSummaryWindow)
		fallbackCommits, fallbackBranchStatus, fallbackWarnings, err := fetchCommitsAcrossRepos(cmd.Context(), client, selectedRepos, user.Login, fallbackStart, resolvedBranches)
		if err != nil {
			if errors.Is(err, githubapi.ErrUnauthorized) {
				return fmt.Errorf("stored token is invalid or expired. run `github-work-summary login` again")
			}
			return err
		}

		allWarnings = append(allWarnings, prefixWarnings("fallback", fallbackWarnings)...)
		fallbackReport := summary.BuildReport(fallbackStart, windowEnd, fallbackCommits)
		if fallbackReport.TotalCommits > 0 {
			fmt.Fprintf(
				out,
				"%s\n\n",
				ui.Yellow(
					out,
					fmt.Sprintf(
						"No commits found in the last 24 hours. Showing recent commits from the last %d days instead.",
						int(fallbackSummaryWindow.Hours()/24),
					),
				),
			)
			report = fallbackReport
			branchStatus = fallbackBranchStatus
		} else {
			fmt.Fprintf(
				out,
				"%s\n\n",
				ui.Yellow(out, fmt.Sprintf("No commits found in the last 24 hours or in the last %d days.", int(fallbackSummaryWindow.Hours()/24))),
			)
		}
	}

	renderBranchStatus(out, branchStatus)
	renderBranchFilter(out, resolvedBranches)

	var output string
	var errExport error

	if summaryJSON {
		output, errExport = report.ToJSON()
	} else if summaryMarkdown {
		output = report.ToMarkdown()
	}

	if errExport != nil {
		return fmt.Errorf("failed to generate export: %w", errExport)
	}

	// If no export format selected, just render to terminal as usual
	if output == "" {
		summary.Render(out, report)
	} else {
		// Handle file output or stdout
		if summaryOutputFile != "" {
			if err := os.WriteFile(summaryOutputFile, []byte(output), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Fprintf(out, "\n%s %s\n", ui.Green(out, "Successfully exported summary to:"), ui.Bold(out, summaryOutputFile))
		} else {
			fmt.Fprintln(out, output)
		}
	}

	if len(allWarnings) > 0 {
		fmt.Fprintln(out, ui.Bold(out, ui.Yellow(out, "Warnings:")))
		for _, warning := range allWarnings {
			fmt.Fprintf(out, "%s %s\n", ui.Red(out, "•"), warning)
		}
	}
	return nil
}

func resolveSummaryBranches(cmd *cobra.Command, client githubapi.GitHubClient, selectedRepos []string) ([]string, []string, error) {
	// 1. Try flags first, then fallback to config
	branches := sanitizeBranches(summaryBranches)
	if len(branches) == 0 {
		branches = viper.GetStringSlice("branches")
	}

	if len(branches) > 0 {
		return branches, nil, nil
	}
	in := cmd.InOrStdin()
	interactive := ui.IsInteractiveTerminal(in)
	if !interactive {
		if summaryChooseBranch {
			return nil, nil, fmt.Errorf("--choose-branch requires an interactive terminal")
		}
		return nil, nil, nil
	}
	if !summaryChooseBranch {
		choose, err := askWhetherChooseBranch(cmd)
		if err != nil {
			return nil, nil, err
		}
		if !choose {
			return nil, nil, nil
		}
	}

	branchRepoCount, warnings, err := fetchBranchesAcrossRepos(cmd.Context(), client, selectedRepos)
	if err != nil {
		return nil, nil, err
	}
	if len(branchRepoCount) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), ui.Yellow(cmd.OutOrStdout(), "No branches found from selected repositories. Continuing with all branches."))
		return nil, warnings, nil
	}

	selected, err := selectBranches(cmd, branchRepoCount)
	if err != nil {
		return nil, nil, err
	}

	// 2. Save the selection for next time
	viper.Set("branches", selected)
	saveConfig()

	return selected, warnings, nil
}

func askWhetherChooseBranch(cmd *cobra.Command) (bool, error) {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	fmt.Fprintln(out, ui.Gray(out, "Branch filter: press Enter for all branches, or type 'b' then Enter to choose branch(es)."))
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read branch filter choice: %w", err)
	}
	choice := strings.TrimSpace(strings.ToLower(line))
	switch choice {
	case "", "a", "all":
		return false, nil
	case "b", "branch", "branches", "s", "select":
		return true, nil
	default:
		fmt.Fprintln(out, ui.Yellow(out, "Unknown choice. Using all branches."))
		return false, nil
	}
}

func loadGitHubClientFromKeychain() (githubapi.GitHubClient, error) {
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
	client githubapi.GitHubClient,
	selectedRepos []string,
	author string,
	since time.Time,
	branches []string,
) (map[string][]githubapi.Commit, map[string]repoBranchStatus, []string, error) {
	repoCommits := make(map[string][]githubapi.Commit, len(selectedRepos))
	statusByRepo := make(map[string]repoBranchStatus, len(selectedRepos))
	for _, repo := range selectedRepos {
		repoCommits[repo] = []githubapi.Commit{}
		statusByRepo[repo] = repoBranchStatus{
			Scanned: []string{},
			Active:  map[string]int{},
		}
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

			result, err := client.ListCommitsByAuthorSinceByBranches(ctx, repoName, author, since, branches)
			results <- repoFetchResult{
				repo:   repoName,
				result: result,
				err:    err,
			}
		}()
	}

	wg.Wait()
	close(results)

	warnings := make([]string, 0)
	for res := range results {
		if res.err != nil {
			if errors.Is(res.err, githubapi.ErrUnauthorized) {
				return nil, nil, nil, res.err
			}
			warnings = append(warnings, fmt.Sprintf("%s: %v", res.repo, res.err))
			continue
		}
		repoCommits[res.repo] = res.result.Commits

		active := make(map[string]int)
		for _, commit := range res.result.Commits {
			for _, branch := range commit.Branches {
				active[branch]++
			}
		}
		statusByRepo[res.repo] = repoBranchStatus{
			Scanned: append([]string(nil), res.result.ScannedBranches...),
			Active:  active,
		}
		if len(res.result.MissingBranches) > 0 {
			warnings = append(
				warnings,
				fmt.Sprintf("%s: branch(es) not found: %s", res.repo, strings.Join(res.result.MissingBranches, ", ")),
			)
		}
	}

	return repoCommits, statusByRepo, warnings, nil
}

func fetchBranchesAcrossRepos(
	ctx context.Context,
	client githubapi.GitHubClient,
	selectedRepos []string,
) (map[string]int, []string, error) {
	branchRepoCount := make(map[string]int)
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

			branches, err := client.ListRepositoryBranches(ctx, repoName)
			res := repoFetchResult{repo: repoName}
			if err != nil {
				res.err = err
				results <- res
				return
			}
			res.result = githubapi.BranchCommitResult{ScannedBranches: branches}
			results <- res
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
		for _, branch := range res.result.ScannedBranches {
			branchRepoCount[branch]++
		}
	}

	return branchRepoCount, warnings, nil
}

func prefixWarnings(prefix string, warnings []string) []string {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		out = append(out, fmt.Sprintf("[%s] %s", prefix, warning))
	}
	return out
}

func renderBranchStatus(out io.Writer, statusByRepo map[string]repoBranchStatus) {
	if len(statusByRepo) == 0 {
		return
	}

	repos := make([]string, 0, len(statusByRepo))
	for repo := range statusByRepo {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	fmt.Fprintln(out, ui.Bold(out, ui.Cyan(out, "Branch Activity:")))
	for _, repo := range repos {
		status := statusByRepo[repo]
		if len(status.Scanned) == 0 {
			fmt.Fprintf(out, "%s %s: %s\n", ui.Yellow(out, "•"), ui.Bold(out, repo), ui.Gray(out, "no branches scanned"))
			continue
		}

		activeParts := make([]string, 0)
		inactive := make([]string, 0)
		for _, branch := range status.Scanned {
			count := status.Active[branch]
			if count > 0 {
				activeParts = append(activeParts, fmt.Sprintf("%s(%d)", branch, count))
			} else {
				inactive = append(inactive, branch)
			}
		}

		if len(activeParts) == 0 {
			fmt.Fprintf(
				out,
				"%s %s: %s (%s)\n",
				ui.Yellow(out, "•"),
				ui.Bold(out, repo),
				ui.Yellow(out, "no recent commits on checked branches"),
				ui.Gray(out, joinBranchNamesWithLimit(status.Scanned, 6)),
			)
			continue
		}

		fmt.Fprintf(
			out,
			"%s %s: %s %s",
			ui.Green(out, "•"),
			ui.Bold(out, repo),
			ui.Green(out, "recent ->"),
			ui.Green(out, strings.Join(activeParts, ", ")),
		)
		if len(inactive) > 0 {
			fmt.Fprintf(out, " | %s %s", ui.Yellow(out, "no recent ->"), ui.Gray(out, joinBranchNamesWithLimit(inactive, 6)))
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out)
}

func renderBranchFilter(out io.Writer, branches []string) {
	if len(branches) == 0 {
		fmt.Fprintf(out, "%s %s\n\n", ui.Bold(out, "Branch Filter:"), ui.Gray(out, "all branches"))
		return
	}
	fmt.Fprintf(
		out,
		"%s %s\n\n",
		ui.Bold(out, "Branch Filter:"),
		ui.Cyan(out, joinBranchNamesWithLimit(branches, 8)),
	)
}

func joinBranchNamesWithLimit(branches []string, limit int) string {
	if len(branches) == 0 {
		return "none"
	}
	if limit <= 0 || len(branches) <= limit {
		return strings.Join(branches, ", ")
	}
	return fmt.Sprintf("%s (+%d more)", strings.Join(branches[:limit], ", "), len(branches)-limit)
}

func sanitizeBranches(branches []string) []string {
	if len(branches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(branches))
	out := make([]string, 0, len(branches))
	for _, branch := range branches {
		name := strings.TrimSpace(branch)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func selectBranches(cmd *cobra.Command, branchRepoCount map[string]int) ([]string, error) {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()
	if !ui.IsInteractiveTerminal(in) {
		return nil, fmt.Errorf("--choose-branch requires an interactive terminal")
	}

	branchNames := make([]string, 0, len(branchRepoCount))
	for branch := range branchRepoCount {
		branchNames = append(branchNames, branch)
	}
	sort.Strings(branchNames)

	options := make([]ui.SelectOption, 0, len(branchNames))
	for _, branch := range branchNames {
		count := branchRepoCount[branch]
		label := fmt.Sprintf("%s (%d repos)", branch, count)
		options = append(options, ui.SelectOption{Label: label, Value: branch})
	}

	selected, err := ui.MultiSelectCheckboxes(
		in,
		out,
		"Select branch(es) to include in summary:",
		options,
	)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(selected))
	for _, item := range selected {
		result = append(result, item.Value)
	}
	sort.Strings(result)
	return result, nil
}

func parseFlexibleTime(input string, reference time.Time) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, nil
	}

	// 1. Try absolute date (2006-01-02)
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t, nil
	}

	// 2. Try RFC3339
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	// 3. Try relative duration (supports d, w)
	durationStr := input
	multiplier := 1.0
	if strings.HasSuffix(input, "d") {
		durationStr = strings.TrimSuffix(input, "d") + "h"
		multiplier = 24.0
	} else if strings.HasSuffix(input, "w") {
		durationStr = strings.TrimSuffix(input, "w") + "h"
		multiplier = 168.0
	}

	d, err := time.ParseDuration(durationStr)
	if err == nil {
		// Apply multiplier for d/w units
		return reference.Add(-time.Duration(float64(d) * multiplier)), nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %q", input)
}
