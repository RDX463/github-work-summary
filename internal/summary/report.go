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
	CategoryFeatures    Category = "Features/Implementations"
	CategoryBugFixes    Category = "Bug Fixes"
	CategoryMaintenance Category = "Maintenance/Refactor"
	CategoryOther       Category = "Other"
)

// RepoSummary groups and classifies commits for one repository.
type RepoSummary struct {
	Repository  string
	Features    []githubapi.Commit
	BugFixes    []githubapi.Commit
	Maintenance []githubapi.Commit
	Other       []githubapi.Commit
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
			case CategoryMaintenance:
				repoSummary.Maintenance = append(repoSummary.Maintenance, commit)
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

	// 1. Check prefixes (Conventional Commits style)
	if strings.HasPrefix(subject, "fix:") || strings.HasPrefix(subject, "fix(") || strings.Contains(subject, "fix!") {
		return CategoryBugFixes
	}
	if strings.HasPrefix(subject, "feat:") || strings.HasPrefix(subject, "feat(") || strings.Contains(subject, "feat!") || strings.HasPrefix(subject, "perf:") || strings.HasPrefix(subject, "perf(") {
		return CategoryFeatures
	}
	maintenancePrefixes := []string{"refactor:", "chore:", "test:", "docs:", "style:", "ci:"}
	for _, p := range maintenancePrefixes {
		if strings.HasPrefix(subject, p) || strings.HasPrefix(subject, strings.TrimSuffix(p, ":")+"(") {
			return CategoryMaintenance
		}
	}

	// 2. Keyword matching
	bugFixKeywords := []string{"fix", "bug", "issue", "hotfix", "security", "patch", "resolves", "closes", "regression"}
	for _, k := range bugFixKeywords {
		if strings.Contains(subject, k) {
			return CategoryBugFixes
		}
	}

	featureKeywords := []string{"feat", "add", "new", "implement", "enhance", "perf", "feature"}
	for _, k := range featureKeywords {
		if strings.Contains(subject, k) {
			return CategoryFeatures
		}
	}

	maintenanceKeywords := []string{"refactor", "chore", "test", "docs", "clean", "style", "ci", "deps"}
	for _, k := range maintenanceKeywords {
		if strings.Contains(subject, k) {
			return CategoryMaintenance
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
		if len(repo.Maintenance) > 0 {
			sb.WriteString("### Maintenance & Refactoring\n")
			for _, c := range repo.Maintenance {
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
