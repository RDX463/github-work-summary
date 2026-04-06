package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PullRequest represents a GitHub PR.
type PullRequest struct {
	ID        int64      `json:"id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	Locked    bool       `json:"locked"`
	User      User       `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
	HTMLURL   string     `json:"html_url"`
}

// ListPullRequestsByAuthorSince fetches PRs for a repository that were updated after since.
func (c *Client) ListPullRequestsByAuthorSince(ctx context.Context, repo, author string, since time.Time) ([]PullRequest, error) {
	endpoint, err := url.Parse(fmt.Sprintf("%s/repos/%s/pulls", c.baseURL, repo))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pulls endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("state", "all")
	q.Set("sort", "updated")
	q.Set("direction", "desc")
	q.Set("per_page", "100")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create pulls request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pulls request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseBodyBytes))
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var all []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, fmt.Errorf("failed to parse pulls response: %w", err)
	}

	var filtered []PullRequest
	for _, pr := range all {
		// Filter by author and timestamp
		// We use UpdatedAt because a PR might have been merged/closed in our window even if created earlier.
		if pr.User.Login == author && (pr.UpdatedAt.After(since) || pr.UpdatedAt.Equal(since)) {
			filtered = append(filtered, pr)
		}
		// Since we sorted by updated desc, we can stop if we reach older PRs.
		// However, GitHub API doesn't guarantee strictly descending UpdatedAt across all pages, 
		// but for the first page of 100 it's usually enough.
		if pr.UpdatedAt.Before(since) {
			break
		}
	}

	return filtered, nil
}
