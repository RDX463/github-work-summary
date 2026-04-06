package ai

import (
	"context"
	"github.com/RDX463/github-work-summary/internal/summary"
)

// Provider defines the interface for AI summarization engines.
type Provider interface {
	// Summarize generates a high-impact summary of the given work report.
	Summarize(ctx context.Context, report summary.Report) (string, error)

	// Name returns the provider's identifier.
	Name() string
}
