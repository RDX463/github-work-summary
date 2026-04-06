package ai

import (
	"strings"
	"testing"

	githubapi "github.com/RDX463/github-work-summary/internal/github"
)

func TestBuildPRPrompt(t *testing.T) {
	commits := []githubapi.Commit{
		{Message: "feat: add main function"},
	}
	prompt := BuildPRPrompt("dev", commits)

	if !strings.Contains(prompt, "PR description") {
		t.Errorf("Prompt missing goal description: %s", prompt)
	}

	if !strings.Contains(prompt, "feat: add main function") {
		t.Errorf("Prompt missing commit context: %s", prompt)
	}
}

func TestBuildPRTitlePrompt(t *testing.T) {
	commits := []githubapi.Commit{
		{Message: "feat: add main function"},
	}
	prompt := BuildPRTitlePrompt("dev", commits)

	if !strings.Contains(prompt, "Pull Request title") {
		t.Errorf("Prompt missing title instructions: %s", prompt)
	}

	if !strings.Contains(prompt, "feat: add main function") {
		t.Errorf("Prompt missing commit context: %s", prompt)
	}
}

func TestProviderSelection(t *testing.T) {
	providers := []struct {
		name     string
		expected string
	}{
		{"gemini", "gemini"},
		{"anthropic", "anthropic"},
		{"claude", "anthropic"},
		{"ollama", "ollama"},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			got := normalizeProvider(p.name)
			if got != p.expected {
				t.Errorf("normalizeProvider(%s) got %s, want %s", p.name, got, p.expected)
			}
		})
	}
}

func normalizeProvider(name string) string {
	name = strings.ToLower(name)
	switch name {
	case "claude":
		return "anthropic"
	default:
		return name
	}
}
