package summary

import (
	"encoding/json"
	"fmt"
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

func (r Report) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r Report) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Work Summary: %s to %s\n\n", r.WindowStart.Format("2006-01-02"), r.WindowEnd.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Total Commits: %d\n\n", r.TotalCommits))

	for _, repo := range r.Repositories {
		if len(repo.Features) == 0 && len(repo.BugFixes) == 0 && len(repo.Other) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n", repo.Repository))

		if len(repo.Features) > 0 {
			sb.WriteString("### Features\n")
			for _, c := range repo.Features {
				sb.WriteString(fmt.Sprintf("- %s ([link](%s))\n", firstLine(c.Message), c.HTMLURL))
			}
		}
		if len(repo.BugFixes) > 0 {
			sb.WriteString("### Bug Fixes\n")
			for _, c := range repo.BugFixes {
				sb.WriteString(fmt.Sprintf("- %s ([link](%s))\n", firstLine(c.Message), c.HTMLURL))
			}
		}
		if len(repo.Other) > 0 {
			sb.WriteString("### Other\n")
			for _, c := range repo.Other {
				sb.WriteString(fmt.Sprintf("- %s ([link](%s))\n", firstLine(c.Message), c.HTMLURL))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
