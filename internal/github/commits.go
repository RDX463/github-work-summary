package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// User is the authenticated GitHub user.
type User struct {
	Login string `json:"login"`
}

// Commit is a minimal commit payload needed by the summary command.
type Commit struct {
	SHA        string
	Message    string
	HTMLURL    string
	AuthoredAt time.Time
}

type userCommitListItem struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Message string `json:"message"`
		Author  struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// GetAuthenticatedUser fetches the current user using the stored access token.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (User, error) {
	endpoint := c.baseURL + "/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return User{}, fmt.Errorf("github user request timed out: %w", err)
		}
		return User{}, fmt.Errorf("github user request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseBodyBytes))
	if err != nil {
		return User{}, fmt.Errorf("failed to read user response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return User{}, ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return User{}, parseAPIError(resp.StatusCode, body)
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, fmt.Errorf("failed to parse user response: %w", err)
	}
	if strings.TrimSpace(user.Login) == "" {
		return User{}, fmt.Errorf("github user response missing login")
	}
	return user, nil
}

// ListCommitsByAuthorSince returns commits in a repository authored by `author` since `since`.
func (c *Client) ListCommitsByAuthorSince(ctx context.Context, repoFullName, author string, since time.Time) ([]Commit, error) {
	owner, repo, err := splitRepoFullName(repoFullName)
	if err != nil {
		return nil, err
	}

	var all []Commit
	for page := 1; ; page++ {
		pageItems, err := c.fetchCommitPage(ctx, owner, repo, author, since, page)
		if err != nil {
			return nil, err
		}

		for _, item := range pageItems {
			authoredAt, err := parseGitHubTimestamp(item.Commit.Author.Date)
			if err != nil {
				continue
			}
			all = append(all, Commit{
				SHA:        item.SHA,
				Message:    item.Commit.Message,
				HTMLURL:    item.HTMLURL,
				AuthoredAt: authoredAt,
			})
		}
		if len(pageItems) < repoPageSize {
			break
		}
	}
	return all, nil
}

func (c *Client) fetchCommitPage(ctx context.Context, owner, repo, author string, since time.Time, page int) ([]userCommitListItem, error) {
	endpoint, err := url.Parse(c.baseURL + "/repos/" + owner + "/" + repo + "/commits")
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("author", author)
	q.Set("since", since.UTC().Format(time.RFC3339))
	q.Set("per_page", strconv.Itoa(repoPageSize))
	q.Set("page", strconv.Itoa(page))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit request for %s/%s: %w", owner, repo, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("github commit request timed out for %s/%s: %w", owner, repo, err)
		}
		return nil, fmt.Errorf("github commit request failed for %s/%s: %w", owner, repo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read commit response for %s/%s: %w", owner, repo, err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusConflict {
		// Empty repositories can return 409 for commit listing.
		return []userCommitListItem{}, nil
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var payload []userCommitListItem
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse commit response for %s/%s: %w", owner, repo, err)
	}
	return payload, nil
}

func splitRepoFullName(fullName string) (string, string, error) {
	trimmed := strings.TrimSpace(fullName)
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository name %q, expected owner/repo", fullName)
	}
	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository name %q, expected owner/repo", fullName)
	}
	return owner, repo, nil
}

func parseGitHubTimestamp(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseAPIError(statusCode int, body []byte) error {
	var apiErr errorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Message) != "" {
		return fmt.Errorf("github API error (%d): %s", statusCode, apiErr.Message)
	}
	return fmt.Errorf("github API error (%d)", statusCode)
}
