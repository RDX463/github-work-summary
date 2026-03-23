package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Notice struct {
	CurrentVersion string
	LatestVersion  string
	URL            string
	Changes        []string
}

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

type compareResponse struct {
	Commits []struct {
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
	} `json:"commits"`
}

// Check compares currentVersion against the latest GitHub release tag.
// It returns nil when no update is needed or version cannot be compared safely.
func Check(ctx context.Context, repo, currentVersion string) (*Notice, error) {
	current := strings.TrimSpace(currentVersion)
	if current == "" {
		return nil, nil
	}

	currentSemver, currentComparable := parseSemver(current)
	if !currentComparable && current != "dev" {
		return nil, nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "github-work-summary")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("release lookup failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var release latestReleaseResponse
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	latestSemver, ok := parseSemver(release.TagName)
	if !ok {
		return nil, nil
	}

	if currentComparable && compareSemver(latestSemver, currentSemver) <= 0 {
		return nil, nil
	}

	currentDisplay := current
	if currentComparable {
		currentDisplay = canonicalVersion(currentSemver)
	}

	changes := extractChanges(release.Body, 4)
	if len(changes) == 0 && currentComparable {
		changes = fetchCompareChanges(ctx, repo, currentDisplay, canonicalVersion(latestSemver), 4)
	}

	return &Notice{
		CurrentVersion: currentDisplay,
		LatestVersion:  canonicalVersion(latestSemver),
		URL:            strings.TrimSpace(release.HTMLURL),
		Changes:        changes,
	}, nil
}

type semver struct {
	major int
	minor int
	patch int
}

func parseSemver(raw string) (semver, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return semver{}, false
	}
	s = strings.TrimPrefix(s, "v")
	if cut := strings.IndexAny(s, "-+"); cut >= 0 {
		s = s[:cut]
	}

	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		return semver{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, false
	}

	return semver{major: major, minor: minor, patch: patch}, true
}

func compareSemver(a, b semver) int {
	switch {
	case a.major != b.major:
		if a.major > b.major {
			return 1
		}
		return -1
	case a.minor != b.minor:
		if a.minor > b.minor {
			return 1
		}
		return -1
	case a.patch != b.patch:
		if a.patch > b.patch {
			return 1
		}
		return -1
	default:
		return 0
	}
}

func canonicalVersion(v semver) string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func extractChanges(body string, max int) []string {
	if max <= 0 {
		return nil
	}

	lines := strings.Split(body, "\n")
	changes := make([]string, 0, max)
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		if strings.HasPrefix(text, "- ") || strings.HasPrefix(text, "* ") || strings.HasPrefix(text, "+ ") {
			text = strings.TrimSpace(text[2:])
		} else {
			text = trimNumericListPrefix(text)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		changes = append(changes, text)
		if len(changes) >= max {
			break
		}
	}
	return changes
}

func trimNumericListPrefix(text string) string {
	// Handles list-like prefixes: "1. item", "2) item"
	if len(text) < 3 {
		return text
	}

	i := 0
	for i < len(text) && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(text)-1 {
		return text
	}
	if text[i] != '.' && text[i] != ')' {
		return text
	}
	if text[i+1] != ' ' {
		return text
	}
	return text[i+2:]
}

func fetchCompareChanges(ctx context.Context, repo, fromTag, toTag string, max int) []string {
	if max <= 0 {
		return nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/compare/%s...%s", repo, fromTag, toTag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "github-work-summary")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil
	}

	var payload compareResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if len(payload.Commits) == 0 {
		return nil
	}

	changes := make([]string, 0, max)
	for _, item := range payload.Commits {
		subject := strings.TrimSpace(firstLine(item.Commit.Message))
		if subject == "" {
			continue
		}
		changes = append(changes, subject)
		if len(changes) >= max {
			break
		}
	}
	return changes
}

func firstLine(message string) string {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return ""
	}
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		return strings.TrimSpace(msg[:i])
	}
	return msg
}
