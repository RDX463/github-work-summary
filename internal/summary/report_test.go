package summary

import (
	"testing"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

func TestCategorizeMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected Category
	}{
		// Bug Fixes
		{"fix: crash on start", CategoryBugFixes},
		{"bugfix: memory leak", CategoryBugFixes},
		{"hotfix: critical security patch", CategoryBugFixes},
		{"fixed the broken layout", CategoryBugFixes},
		{"resolve issue #123", CategoryBugFixes},
		{"regression in login flow", CategoryBugFixes},
		{"patch: update dependencies", CategoryBugFixes},

		// Features
		{"feat: add dark mode", CategoryFeatures},
		{"feature(ui): new landing page", CategoryFeatures},
		{"added new export feature", CategoryFeatures},
		{"implement oauth2 flow", CategoryFeatures},
		{"introduce new billing system", CategoryFeatures},
		{"perf: faster report generation", CategoryFeatures},

		// Maintenance/Refactor
		{"refactor: code for clarity", CategoryMaintenance},
		{"chore: cleanup", CategoryMaintenance},
		{"test: add unit tests", CategoryMaintenance},
		{"docs: update readme", CategoryMaintenance},
		{"style: fix formatting", CategoryMaintenance},
		{"ci: update github actions", CategoryMaintenance},

		// Others
		{"initial commit", CategoryOther},
		{"something else", CategoryOther},
		{"", CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			got := CategorizeMessage(tt.message)
			if got != tt.expected {
				t.Errorf("CategorizeMessage(%q) = %v; want %v", tt.message, got, tt.expected)
			}
		})
	}
}

func TestBuildReport(t *testing.T) {
	windowEnd := time.Now()
	windowStart := windowEnd.Add(-24 * time.Hour)

	repoCommits := map[string][]githubapi.Commit{
		"owner/repo1": {
			{Message: "feat: feature 1", AuthoredAt: windowStart.Add(1 * time.Hour)},
			{Message: "fix: bug 1", AuthoredAt: windowStart.Add(2 * time.Hour)},
			{Message: "chore: cleanup", AuthoredAt: windowStart.Add(2 * time.Hour + 30*time.Minute)},
		},
		"owner/repo2": {
			{Message: "other: miscellaneous", AuthoredAt: windowStart.Add(3 * time.Hour)},
		},
	}
	repoPulls := map[string][]githubapi.PullRequest{
		"owner/repo1": {
			{Title: "pr 1", Number: 1, State: "open", UpdatedAt: windowStart.Add(4 * time.Hour)},
		},
	}

	report := BuildReport(windowStart, windowEnd, repoCommits, repoPulls)

	if report.TotalCommits != 4 {
		t.Errorf("expected 4 total commits, got %d", report.TotalCommits)
	}

	if len(report.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(report.Repositories))
	}

	// Verify repo1 categorization
	var repo1 RepoSummary
	for _, r := range report.Repositories {
		if r.Repository == "owner/repo1" {
			repo1 = r
			break
		}
	}

	if len(repo1.Features) != 1 || repo1.Features[0].Message != "feat: feature 1" {
		t.Errorf("repo1 features not correctly categorized: %+v", repo1.Features)
	}
	if len(repo1.BugFixes) != 1 || repo1.BugFixes[0].Message != "fix: bug 1" {
		t.Errorf("repo1 bug fixes not correctly categorized: %+v", repo1.BugFixes)
	}
	if len(repo1.Maintenance) != 1 || repo1.Maintenance[0].Message != "chore: cleanup" {
		t.Errorf("repo1 maintenance not correctly categorized: %+v", repo1.Maintenance)
	}
	if len(repo1.PullRequests) != 1 || repo1.PullRequests[0].Title != "pr 1" {
		t.Errorf("repo1 PRs not correctly included: %+v", repo1.PullRequests)
	}

	// Verify repo2 categorization
	var repo2 RepoSummary
	for _, r := range report.Repositories {
		if r.Repository == "owner/repo2" {
			repo2 = r
			break
		}
	}
	if len(repo2.Other) != 1 || repo2.Other[0].Message != "other: miscellaneous" {
		t.Errorf("repo2 other not correctly categorized: %+v", repo2.Other)
	}
}
