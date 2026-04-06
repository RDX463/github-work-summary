package summary

import (
	"testing"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

func TestBuildReport(t *testing.T) {
	windowEnd := time.Now()
	windowStart := windowEnd.Add(-24 * time.Hour)

	commits := []githubapi.Commit{
		{Message: "feat: feature 1", AuthoredAt: windowStart.Add(1 * time.Hour), RepoName: "owner/repo1"},
		{Message: "fix: bug 1", AuthoredAt: windowStart.Add(2 * time.Hour), RepoName: "owner/repo1"},
		{Message: "chore: cleanup", AuthoredAt: windowStart.Add(2 * time.Hour + 30*time.Minute), RepoName: "owner/repo1"},
		{Message: "other: miscellaneous", AuthoredAt: windowStart.Add(3 * time.Hour), RepoName: "owner/repo2"},
	}
	
	pulls := []githubapi.PullRequest{
		{Title: "pr 1", Number: 1, State: "open", UpdatedAt: windowStart.Add(4 * time.Hour), RepoName: "owner/repo1"},
	}

	report := BuildReport(commits, pulls, windowStart, windowEnd)

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
