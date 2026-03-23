package summary

import (
	"sort"
	"strings"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

// Category is a commit classification bucket.
type Category string

const (
	CategoryFeatures Category = "Features/Implementations"
	CategoryBugFixes Category = "Bug Fixes"
	CategoryOther    Category = "Other"
)

// RepoSummary groups and classifies commits for one repository.
type RepoSummary struct {
	Repository string
	Features   []githubapi.Commit
	BugFixes   []githubapi.Commit
	Other      []githubapi.Commit
}

// Report is the terminal output model for a 24-hour work summary.
type Report struct {
	WindowStart  time.Time
	WindowEnd    time.Time
	Repositories []RepoSummary
	TotalCommits int
}

// BuildReport creates a classified report from raw commits keyed by repository full name.
func BuildReport(windowStart, windowEnd time.Time, repoCommits map[string][]githubapi.Commit) Report {
	repoNames := make([]string, 0, len(repoCommits))
	for repo := range repoCommits {
		repoNames = append(repoNames, repo)
	}
	sort.Strings(repoNames)

	repositories := make([]RepoSummary, 0, len(repoNames))
	total := 0
	for _, repo := range repoNames {
		commits := append([]githubapi.Commit(nil), repoCommits[repo]...)
		sort.Slice(commits, func(i, j int) bool {
			return commits[i].AuthoredAt.After(commits[j].AuthoredAt)
		})

		repoSummary := RepoSummary{Repository: repo}
		for _, commit := range commits {
			switch CategorizeMessage(commit.Message) {
			case CategoryFeatures:
				repoSummary.Features = append(repoSummary.Features, commit)
			case CategoryBugFixes:
				repoSummary.BugFixes = append(repoSummary.BugFixes, commit)
			default:
				repoSummary.Other = append(repoSummary.Other, commit)
			}
		}

		total += len(commits)
		repositories = append(repositories, repoSummary)
	}

	return Report{
		WindowStart:  windowStart,
		WindowEnd:    windowEnd,
		Repositories: repositories,
		TotalCommits: total,
	}
}

// CategorizeMessage uses lightweight keyword matching to bucket commit intent.
func CategorizeMessage(message string) Category {
	subject := strings.ToLower(firstLine(message))
	subject = strings.TrimSpace(subject)

	bugFixPrefixes := []string{"fix", "bugfix", "hotfix"}
	for _, prefix := range bugFixPrefixes {
		if strings.HasPrefix(subject, prefix+":") || strings.HasPrefix(subject, prefix+"(") || strings.HasPrefix(subject, prefix+" ") {
			return CategoryBugFixes
		}
	}

	featurePrefixes := []string{"feat", "feature", "build"}
	for _, prefix := range featurePrefixes {
		if strings.HasPrefix(subject, prefix+":") || strings.HasPrefix(subject, prefix+"(") || strings.HasPrefix(subject, prefix+" ") {
			return CategoryFeatures
		}
	}

	bugFixKeywords := []string{
		"fix", "fixed", "fixes", "bug", "hotfix", "issue", "resolve", "resolved",
		"patch", "regression", "correct", "repair",
	}
	for _, keyword := range bugFixKeywords {
		if strings.Contains(subject, keyword) {
			return CategoryBugFixes
		}
	}

	featureKeywords := []string{
		"feat", "feature", "add", "added", "implement", "implemented", "create", "created",
		"introduce", "support", "enable", "new", "refactor", "improve", "enhance",
	}
	for _, keyword := range featureKeywords {
		if strings.Contains(subject, keyword) {
			return CategoryFeatures
		}
	}

	return CategoryOther
}

func firstLine(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return ""
	}
	if idx := strings.IndexByte(trimmed, '\n'); idx >= 0 {
		return strings.TrimSpace(trimmed[:idx])
	}
	return trimmed
}
