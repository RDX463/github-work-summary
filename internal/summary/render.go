package summary

import (
	"fmt"
	"io"
	"strings"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

const (
	colorReset  = "\x1b[0m"
	colorBold   = "\x1b[1m"
	colorCyan   = "\x1b[36m"
	colorGreen  = "\x1b[32m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorGray   = "\x1b[90m"
)

// Render prints a colorized work report to the terminal.
func Render(w io.Writer, report Report) {
	fmt.Fprintf(w, "%s%sGitHub Work Summary%s\n", colorBold, colorCyan, colorReset)
	fmt.Fprintf(w, "%sWindow:%s %s -> %s\n", colorGray, colorReset, report.WindowStart.Format(time.RFC1123), report.WindowEnd.Format(time.RFC1123))
	fmt.Fprintf(w, "%sTotal Activity:%s %d Commits, %d PRs\n", colorGray, colorReset, report.TotalCommits, report.TotalPRs)

	if report.AISummary != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s%sAI IMPACT SUMMARY%s\n", colorBold, colorGreen, colorReset)
		fmt.Fprintf(w, "%s%s\n", colorReset, report.AISummary)
	}

	if len(report.TicketInfo) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s%sRELATED TICKETS%s\n", colorBold, colorCyan, colorReset)
		for _, t := range report.TicketInfo {
			fmt.Fprintf(w, "  • %s: %s (%s%s%s)\n", colorBold+t.ID+colorReset, t.Title, colorGray, t.Status, colorReset)
		}
	}

	fmt.Fprintln(w)

	if len(report.Repositories) == 0 {
		fmt.Fprintln(w, "No repositories selected.")
		return
	}

	if report.TotalCommits == 0 {
		fmt.Fprintf(
			w,
			"No commits in the last %s across %d selected repositories.\n",
			formatSummaryWindow(report.WindowEnd.Sub(report.WindowStart)),
			len(report.Repositories),
		)
		return
	}

	for _, repo := range report.Repositories {
		repoTotalCommits := len(repo.Features) + len(repo.BugFixes) + len(repo.Maintenance) + len(repo.Other)
		fmt.Fprintf(w, "%s%s%s (%d commits, %d PRs)\n", colorBold, repo.Repository, colorReset, repoTotalCommits, len(repo.PullRequests))
		renderPullRequests(w, repo.PullRequests)
		renderCategory(w, CategoryFeature, repo.Features, colorGreen)
		renderCategory(w, CategoryBugFix, repo.BugFixes, colorRed)
		renderCategory(w, CategoryMaintenance, repo.Maintenance, colorCyan)
		renderCategory(w, CategoryOther, repo.Other, colorYellow)
		fmt.Fprintln(w)
	}
}

func renderPullRequests(w io.Writer, pulls []githubapi.PullRequest) {
	fmt.Fprintf(w, "  %sPull Requests%s (%d)\n", colorCyan, colorReset, len(pulls))
	if len(pulls) == 0 {
		fmt.Fprintln(w, "    - none")
		return
	}

	for _, pr := range pulls {
		status := pr.State
		statusColor := colorGray
		if pr.MergedAt != nil {
			status = "merged"
			statusColor = colorCyan
		} else if pr.State == "open" {
			statusColor = colorGreen
		} else if pr.State == "closed" {
			statusColor = colorRed
		}

		fmt.Fprintf(
			w,
			"    - %s [#%d] %s%s%s\n",
			pr.Title,
			pr.Number,
			statusColor,
			status,
			colorReset,
		)
		if pr.HTMLURL != "" {
			fmt.Fprintf(w, "      %s%s%s\n", colorGray, pr.HTMLURL, colorReset)
		}
	}
}

func renderCategory(w io.Writer, category Category, commits []githubapi.Commit, color string) {
	fmt.Fprintf(w, "  %s%s%s (%d)\n", color, category, colorReset, len(commits))
	if len(commits) == 0 {
		fmt.Fprintln(w, "    - none")
		return
	}

	for _, commit := range commits {
		subject := ShortSubject(commit.Message)
		shortSHA := shortenSHA(commit.SHA)
		branchInfo := formatCommitBranchInfo(commit.Branches)
		fmt.Fprintf(
			w,
			"    - %s [%s] %s%s\n",
			subject,
			shortSHA,
			commit.AuthoredAt.Local().Format("2006-01-02 15:04"),
			branchInfo,
		)
		if commit.HTMLURL != "" {
			fmt.Fprintf(w, "      %s%s%s\n", colorGray, commit.HTMLURL, colorReset)
		}
	}
}

func shortenSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

func formatSummaryWindow(d time.Duration) string {
	if d <= 0 {
		return "selected window"
	}
	hours := d.Hours()
	if hours < 48 {
		return fmt.Sprintf("%d hours", int(hours+0.5))
	}
	return fmt.Sprintf("%d days", int(hours/24+0.5))
}

func formatCommitBranchInfo(branches []string) string {
	cleaned := SanitizeAndSortBranches(branches)
	if len(cleaned) == 0 {
		return ""
	}
	if len(cleaned) <= 2 {
		return " | branch: " + strings.Join(cleaned, ", ")
	}
	return fmt.Sprintf(" | branches: %s (+%d more)", strings.Join(cleaned[:2], ", "), len(cleaned)-2)
}
