package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

// Category represents a bucket for work classification.
type Category string

const (
	CategoryFeature     Category = "Features/Implementations"
	CategoryBugFix      Category = "Bug Fixes"
	CategoryMaintenance Category = "Maintenance/Refactor"
	CategoryOther       Category = "Other"
)

// RepoSummary groups and classifies commits for one repository.
type RepoSummary struct {
	Repository   string
	Features     []githubapi.Commit
	BugFixes     []githubapi.Commit
	Maintenance  []githubapi.Commit
	Other        []githubapi.Commit
	PullRequests []githubapi.PullRequest
	AISummary    string
}

// Report is the terminal output model for a 24-hour work summary.
type Report struct {
	WindowStart  time.Time
	WindowEnd    time.Time
	Repositories []RepoSummary
	TotalCommits int
	TotalPRs     int
	AISummary    string
}

// BuildReport creates a classified report from raw commits keyed by repository full name.
func BuildReport(commits []githubapi.Commit, prs []githubapi.PullRequest, start, end time.Time) Report {
	repoMap := make(map[string]*RepoSummary)

	// Process Commits
	for _, commit := range commits {
		repoName := commit.RepoName
		if _, ok := repoMap[repoName]; !ok {
			repoMap[repoName] = &RepoSummary{Repository: repoName}
		}

		repo := repoMap[repoName]
		msg := strings.ToLower(commit.Message)

		switch {
		case strings.HasPrefix(msg, "feat") || strings.Contains(msg, "implement"):
			repo.Features = append(repo.Features, commit)
		case strings.HasPrefix(msg, "fix") || strings.Contains(msg, "bug"):
			repo.BugFixes = append(repo.BugFixes, commit)
		case strings.HasPrefix(msg, "maint") || strings.HasPrefix(msg, "refactor") || strings.HasPrefix(msg, "chore"):
			repo.Maintenance = append(repo.Maintenance, commit)
		default:
			repo.Other = append(repo.Other, commit)
		}
	}

	// Process Pull Requests
	for _, pr := range prs {
		repoName := pr.RepoName
		if _, ok := repoMap[repoName]; !ok {
			repoMap[repoName] = &RepoSummary{Repository: repoName}
		}
		repo := repoMap[repoName]
		repo.PullRequests = append(repo.PullRequests, pr)
	}

	// Convert map to slice and sort by repo name
	repos := make([]RepoSummary, 0, len(repoMap))
	totalCommits := 0
	totalPRs := 0
	for _, repo := range repoMap {
		repos = append(repos, *repo)
		totalCommits += len(repo.Features) + len(repo.BugFixes) + len(repo.Maintenance) + len(repo.Other)
		totalPRs += len(repo.PullRequests)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repository < repos[j].Repository
	})

	return Report{
		WindowStart:  start,
		WindowEnd:    end,
		Repositories: repos,
		TotalCommits: totalCommits,
		TotalPRs:     totalPRs,
	}
}

// ToMarkdown converts the report to a formatted Markdown string.
func (r Report) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# GitHub Work Summary\n\n"))
	sb.WriteString(fmt.Sprintf("**Window:** %s to %s\n\n", r.WindowStart.Format(time.RFC822), r.WindowEnd.Format(time.RFC822)))
	
	if r.AISummary != "" {
		sb.WriteString(fmt.Sprintf("## AI Impact Summary\n%s\n\n", r.AISummary))
	}

	sb.WriteString(fmt.Sprintf("## Stats\n- **Total Commits:** %d\n- **Total Pull Requests:** %d\n\n", r.TotalCommits, r.TotalPRs))

	for _, repo := range r.Repositories {
		sb.WriteString(fmt.Sprintf("### %s (%d commits, %d PRs)\n", repo.Repository, 
			len(repo.Features)+len(repo.BugFixes)+len(repo.Maintenance)+len(repo.Other),
			len(repo.PullRequests)))
		
		if repo.AISummary != "" {
			sb.WriteString(fmt.Sprintf("\n*AI Summary:* %s\n", repo.AISummary))
		}

		renderMarkdownCategory(&sb, string(CategoryFeature), repo.Features)
		renderMarkdownCategory(&sb, string(CategoryBugFix), repo.BugFixes)
		renderMarkdownCategory(&sb, string(CategoryMaintenance), repo.Maintenance)
		renderMarkdownCategory(&sb, string(CategoryOther), repo.Other)

		if len(repo.PullRequests) > 0 {
			sb.WriteString("\n#### Pull Requests\n")
			for _, pr := range repo.PullRequests {
				status := pr.State
				if pr.MergedAt != nil {
					status = "merged"
				}
				sb.WriteString(fmt.Sprintf("- [%s] %s (#%d) [%s]\n", strings.ToUpper(status), pr.Title, pr.Number, pr.HTMLURL))
			}
		}
		sb.WriteString("\n---\n")
	}

	return sb.String()
}

func renderMarkdownCategory(sb *strings.Builder, title string, commits []githubapi.Commit) {
	if len(commits) == 0 {
		return
	}
	sb.WriteString(fmt.Sprintf("\n#### %s\n", title))
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("- %s ([%s](%s))\n", c.Message, c.SHA[:7], c.HTMLURL))
	}
}
