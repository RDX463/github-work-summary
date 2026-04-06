package ai

import (
	"context"
	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
)

// Provider defines the interface for AI summarization engines.
type Provider interface {
	// Summarize generates a high-impact summary of the given work report.
	Summarize(ctx context.Context, report summary.Report) (string, error)

	// GeneratePRDescription creates a professional pull request description based on the branch and its commits.
	GeneratePRDescription(ctx context.Context, branchName string, commits []githubapi.Commit) (string, error)

	// GeneratePRTitle creates a concise, high-impact title for a pull request.
	GeneratePRTitle(ctx context.Context, branchName string, commits []githubapi.Commit) (string, error)

	// Name returns the provider's identifier.
	Name() string
}
