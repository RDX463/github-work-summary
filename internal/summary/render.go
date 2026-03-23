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
	fmt.Fprintf(w, "%sTotal Commits:%s %d\n", colorGray, colorReset, report.TotalCommits)
	fmt.Fprintln(w)

	if len(report.Repositories) == 0 {
		fmt.Fprintln(w, "No repositories selected.")
		return
	}

	if report.TotalCommits == 0 {
		fmt.Fprintf(w, "No commits in the last 24 hours across %d selected repositories.\n", len(report.Repositories))
		return
	}

	for _, repo := range report.Repositories {
		repoTotal := len(repo.Features) + len(repo.BugFixes) + len(repo.Other)

		fmt.Fprintf(w, "%s%s%s (%d)\n", colorBold, repo.Repository, colorReset, repoTotal)
		renderCategory(w, CategoryFeatures, repo.Features, colorGreen)
		renderCategory(w, CategoryBugFixes, repo.BugFixes, colorRed)
		renderCategory(w, CategoryOther, repo.Other, colorYellow)
		fmt.Fprintln(w)
	}
}

func renderCategory(w io.Writer, category Category, commits []githubapi.Commit, color string) {
	fmt.Fprintf(w, "  %s%s%s (%d)\n", color, category, colorReset, len(commits))
	if len(commits) == 0 {
		fmt.Fprintln(w, "    - none")
		return
	}

	for _, commit := range commits {
		subject := shortSubject(commit.Message)
		shortSHA := shortenSHA(commit.SHA)
		fmt.Fprintf(w, "    - %s [%s] %s\n", subject, shortSHA, commit.AuthoredAt.Local().Format("2006-01-02 15:04"))
		if commit.HTMLURL != "" {
			fmt.Fprintf(w, "      %s%s%s\n", colorGray, commit.HTMLURL, colorReset)
		}
	}
}

func shortSubject(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "(no commit message)"
	}
	line := trimmed
	if idx := strings.IndexByte(trimmed, '\n'); idx >= 0 {
		line = strings.TrimSpace(trimmed[:idx])
	}
	if len(line) > 90 {
		return line[:87] + "..."
	}
	return line
}

func shortenSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}
