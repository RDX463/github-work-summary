package ai

import (
	"fmt"
	"strings"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
)

// BuildReportPrompt constructs the base instructions and data for any LLM to summarize the work.
func BuildReportPrompt(report summary.Report) string {
	var b strings.Builder
	b.WriteString("You are a professional software engineering manager summarizing a developer's daily work based on their GitHub activity.\n")
	b.WriteString("Analyze the following commit messages and pull request titles and provide a concise, high-impact summary of the day's work.\n")
	b.WriteString("Focus on the 'What' and 'Why' instead of just listing the 'How'. Use professional language.\n\n")
	
	fmt.Fprintf(&b, "Timeframe: %s to %s\n", report.WindowStart.Format("2006-01-02"), report.WindowEnd.Format("2006-01-02"))
	fmt.Fprintf(&b, "Total Commits: %d\n", report.TotalCommits)
	fmt.Fprintf(&b, "Total Pull Requests: %d\n\n", report.TotalPRs)

	if len(report.TicketInfo) > 0 {
		b.WriteString("Related Tickets (Business Context):\n")
		for _, t := range report.TicketInfo {
			fmt.Fprintf(&b, "- [%s] %s (Status: %s)\n", t.ID, t.Title, t.Status)
		}
		b.WriteString("\n")
	}

	for _, repo := range report.Repositories {
		fmt.Fprintf(&b, "### Repository: %s\n", repo.Repository)
		
		addCommitsToPrompt(&b, "Features", repo.Features)
		addCommitsToPrompt(&b, "Bug Fixes", repo.BugFixes)
		addCommitsToPrompt(&b, "Maintenance", repo.Maintenance)
		addCommitsToPrompt(&b, "Other", repo.Other)

		if len(repo.PullRequests) > 0 {
			b.WriteString("Pull Requests:\n")
			for _, pr := range repo.PullRequests {
				fmt.Fprintf(&b, "- %s (#%d)\n", pr.Title, pr.Number)
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\nSummary requirements:\n")
	b.WriteString("- Keep it under 200 words.\n")
	b.WriteString("- Use 2-3 short paragraphs or bullet points.\n")
	b.WriteString("- Do not use markdown headers (level 1-3) inside your response.\n")
	b.WriteString("- Highlight the key focus of the day (e.g., 'Primary focus was stabilizing the payment integration...').\n")

	return b.String()
}

// BuildPRPrompt constructs the instructions for an LLM to write a professional PR description.
func BuildPRPrompt(branchName string, commits []githubapi.Commit) string {
	var b strings.Builder
	b.WriteString("You are a professional software engineer drafting a high-impact Pull Request (PR) description.\n")
	b.WriteString(fmt.Sprintf("Branch Name: %s\n\n", branchName))
	b.WriteString("Analyze the following commit messages to understand the purpose and impact of the changes:\n")
	for _, c := range commits {
		fmt.Fprintf(&b, "- %s\n", c.Message)
		for _, id := range c.Tickets {
			fmt.Fprintf(&b, "  [Linked Ticket: %s]\n", id)
		}
	}

	b.WriteString("\nGenerate a professional PR description in Markdown format with the following sections:\n")
	b.WriteString("1. **Context/Purpose**: Why are these changes being made?\n")
	b.WriteString("2. **Key Changes**: Boldly highlight the most important technical changes.\n")
	b.WriteString("3. **Side Effects**: Mention any potential risks or areas to watch out for.\n")
	b.WriteString("4. **Testing Status**: A placeholder or AI-suggested testing steps.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Keep the tone professional but high-energy.\n")
	b.WriteString("- Focus on impact and clarify for reviewers.\n")
	b.WriteString("- Do not use level 1 headers (#). Use level 2 (##) or bold text.\n")
	b.WriteString("- Maximum length: 300 words.\n")

	return b.String()
}

// BuildPRTitlePrompt constructs the instructions for an LLM to write a professional PR title.
func BuildPRTitlePrompt(branchName string, commits []githubapi.Commit) string {
	var b strings.Builder
	b.WriteString("Generate a concise, professional Pull Request title (max 60 characters) based on these commits:\n")
	for _, c := range commits {
		fmt.Fprintf(&b, "- %s\n", c.Message)
	}
	b.WriteString("\nFollow the 'feat(scope): description' or 'fix(scope): description' conventional commits format if applicable.\n")
	b.WriteString("Return ONLY the title string, no markdown headers or conversational text.")

	return b.String()
}

func addCommitsToPrompt(b *strings.Builder, category string, commits []githubapi.Commit) {
	if len(commits) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", category)
	for _, c := range commits {
		fmt.Fprintf(b, "- %s\n", c.Message)
	}
}
