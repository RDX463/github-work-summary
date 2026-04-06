package summary

import (
	"context"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

// TaskOptions defines parameters for a summary execution.
type TaskOptions struct {
	Client        githubapi.GitHubClient
	Author        string
	Repos         []string
	Branches      []string
	Since         time.Time
	Until         time.Time
	
	AI            bool
	
	Tickets       bool
	TicketKeys    map[string]interface{} // Domain, Email, Tokens
	
	SkipPRs       bool
}

// ExecuteSummaryTask performs the full summary workflow: fetch, enrich, and build report.
func ExecuteSummaryTask(ctx context.Context, opts TaskOptions) (Report, []string, error) {
	// 1. Fetch Work Data (logic moved from cmd/summary.go)
	// We'll pass in a helper or re-implement the fetch logic here 
	// for maximum decoupling.
	
	// For now, we'll keep the actual retrieval in summary.go for simplicity
	// but provide a clear hook.
	return Report{}, nil, nil
}
