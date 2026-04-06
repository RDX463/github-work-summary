package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

// Category represents the type of work performed.
type Category string

const (
	CategoryFeature     Category = "Features"
	CategoryBugFix      Category = "Bug Fixes"
	CategoryMaintenance Category = "Maintenance"
	CategoryOther       Category = "Other"
)

// Report contains everything needed to render a summary.
type Report struct {
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	TotalCommits int       `json:"total_commits"`
	TotalPRs     int       `json:"total_prs"`

	Repositories []RepoSummary `json:"repositories"`
	AISummary    string        `json:"ai_summary"`

	Tickets      map[string]string `json:"tickets"`     // Ticket ID -> Title
	TicketInfo   []Ticket          `json:"ticket_info"` // Full details
}

// RepoSummary compiles activity for a specific repository.
type RepoSummary struct {
	Repository   string                  `json:"repository"`
	Features     []githubapi.Commit      `json:"features"`
	BugFixes     []githubapi.Commit      `json:"bug_fixes"`
	Maintenance  []githubapi.Commit      `json:"maintenance"`
	Other        []githubapi.Commit      `json:"other"`
	PullRequests []githubapi.PullRequest `json:"pull_requests"`
}

// Ticket represents fetched metadata from Jira or Linear.
type Ticket struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

// BuildReport calculates a summary from a slice of commits and PRs.
func BuildReport(commits []githubapi.Commit, pulls []githubapi.PullRequest, start, end time.Time) Report {
	repoSummaries := make(map[string]*RepoSummary)

	for _, commit := range commits {
		repo := commit.RepoName
		if repo == "" { repo = "unknown" }
		if _, exists := repoSummaries[repo]; !exists {
			repoSummaries[repo] = &RepoSummary{Repository: repo}
		}

		s := repoSummaries[repo]
		msg := strings.ToLower(commit.Message)
		if strings.HasPrefix(msg, "feat") || strings.Contains(msg, "feature") {
			s.Features = append(s.Features, commit)
		} else if strings.HasPrefix(msg, "fix") || strings.Contains(msg, "bug") {
			s.BugFixes = append(s.BugFixes, commit)
		} else if strings.HasPrefix(msg, "chore") || strings.HasPrefix(msg, "refactor") || strings.Contains(msg, "maintenance") {
			s.Maintenance = append(s.Maintenance, commit)
		} else {
			s.Other = append(s.Other, commit)
		}
	}

	for _, pr := range pulls {
		repo := pr.RepoName
		if repo == "" { repo = "unknown" }
		if _, exists := repoSummaries[repo]; !exists {
			repoSummaries[repo] = &RepoSummary{Repository: repo}
		}
		repoSummaries[repo].PullRequests = append(repoSummaries[repo].PullRequests, pr)
	}

	sortedRepos := make([]RepoSummary, 0, len(repoSummaries))
	for _, s := range repoSummaries {
		sortedRepos = append(sortedRepos, *s)
	}
	sort.Slice(sortedRepos, func(i, j int) bool {
		return sortedRepos[i].Repository < sortedRepos[j].Repository
	})

	return Report{
		WindowStart:  start,
		WindowEnd:    end,
		TotalCommits: len(commits),
		TotalPRs:     len(pulls),
		Repositories: sortedRepos,
		Tickets:      make(map[string]string),
	}
}

// ToMarkdown generates a Markdown version of the report.
func (r *Report) ToMarkdown() string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Work Summary (%s - %s)\n\n",
		r.WindowStart.Format("Jan 02, 15:04"),
		r.WindowEnd.Format("Jan 02, 15:04"))

	if r.AISummary != "" {
		b.WriteString("## AI Impact Summary\n")
		b.WriteString(r.AISummary)
		b.WriteString("\n\n")
	}

	if len(r.TicketInfo) > 0 {
		b.WriteString("## Related Tickets\n")
		for _, t := range r.TicketInfo {
			fmt.Fprintf(&b, "- [%s](%s): %s (%s)\n", t.ID, t.URL, t.Title, t.Status)
		}
		b.WriteString("\n")
	}

	for _, repo := range r.Repositories {
		fmt.Fprintf(&b, "### %s\n\n", repo.Repository)
		
		if len(repo.Features) > 0 {
			b.WriteString("#### Features\n")
			for _, c := range repo.Features { fmt.Fprintf(&b, "- %s ([%s](%s))\n", ShortSubject(c.Message), c.SHA[:7], c.HTMLURL) }
			b.WriteString("\n")
		}
		if len(repo.BugFixes) > 0 {
			b.WriteString("#### Bug Fixes\n")
			for _, c := range repo.BugFixes { fmt.Fprintf(&b, "- %s ([%s](%s))\n", ShortSubject(c.Message), c.SHA[:7], c.HTMLURL) }
			b.WriteString("\n")
		}
		if len(repo.PullRequests) > 0 {
			b.WriteString("#### Pull Requests\n")
			for _, p := range repo.PullRequests { fmt.Fprintf(&b, "- #%d: %s ([view](%s))\n", p.Number, p.Title, p.HTMLURL) }
			b.WriteString("\n")
		}
	}

	return b.String()
}
