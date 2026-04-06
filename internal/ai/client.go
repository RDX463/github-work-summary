package ai

import (
	"context"
	"fmt"
	"strings"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
	"google.golang.org/genai"
)

type Client struct {
	genaiClient *genai.Client
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &Client{genaiClient: client}, nil
}

func (c *Client) SummarizeReport(ctx context.Context, report summary.Report) (string, error) {
	prompt := c.buildReportPrompt(report)
	
	result, err := c.genaiClient.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash", // Using a stable fast model
		[]*genai.Content{{Parts: []*genai.Part{{Text: prompt}}}},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func (c *Client) buildReportPrompt(report summary.Report) string {
	var b strings.Builder
	b.WriteString("You are a professional software engineering manager summarizing a developer's daily work based on their GitHub activity.\n")
	b.WriteString("Analyze the following commit messages and pull request titles and provide a concise, high-impact summary of the day's work.\n")
	b.WriteString("Focus on the 'What' and 'Why' instead of just listing the 'How'. Use professional language.\n\n")
	
	fmt.Fprintf(&b, "Timeframe: %s to %s\n", report.WindowStart.Format("2006-01-02"), report.WindowEnd.Format("2006-01-02"))
	fmt.Fprintf(&b, "Total Commits: %d\n", report.TotalCommits)
	fmt.Fprintf(&b, "Total Pull Requests: %d\n\n", report.TotalPRs)

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

func addCommitsToPrompt(b *strings.Builder, category string, commits []githubapi.Commit) {
	if len(commits) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", category)
	for _, c := range commits {
		fmt.Fprintf(b, "- %s\n", c.Message)
	}
}
