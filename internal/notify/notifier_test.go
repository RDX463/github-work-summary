package notify

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
	"github.com/RDX463/github-work-summary/internal/summary"
)

func TestSlackPayload(t *testing.T) {
	report := summary.Report{
		TotalCommits: 2,
		TotalPRs:     1,
		WindowStart:  time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
		WindowEnd:    time.Date(2024, 3, 20, 23, 59, 59, 0, time.UTC),
		AISummary:    "Successfully implemented new user authentication flow.",
		Repositories: []summary.RepoSummary{
			{
				Repository: "owner/repo",
				Features:   []githubapi.Commit{{Message: "feat 1"}, {Message: "feat 2"}},
			},
		},
	}

	n := &SlackNotifier{}
	payload := n.buildSlackPayload(report)

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "GitHub Work Summary") {
		t.Error("missing header")
	}
	if !strings.Contains(raw, "AI IMPACT SUMMARY") {
		t.Error("missing AI section")
	}
	if !strings.Contains(raw, "Successfully implemented") {
		t.Error("missing AI content")
	}
	if !strings.Contains(raw, "owner/repo") {
		t.Error("missing repo name")
	}
}

func TestDiscordPayload(t *testing.T) {
	report := summary.Report{
		TotalCommits: 5,
		TotalPRs:     2,
		AISummary:    "Fixed critical memory leak in background worker.",
		Repositories: []summary.RepoSummary{
			{Repository: "owner/repo"},
		},
	}

	n := &DiscordNotifier{}
	payload := n.buildDiscordPayload(report)

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, "Fixed critical memory leak") {
		t.Error("missing AI summary in embed")
	}
	if !strings.Contains(raw, "owner/repo") {
		t.Error("missing repo name in embed")
	}
}
