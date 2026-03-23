package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
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
	Branches   []string
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

type repoBranchListItem struct {
	Name string `json:"name"`
}

// BranchCommitResult carries commits and branch-selection metadata.
type BranchCommitResult struct {
	Commits         []Commit
	ScannedBranches []string
	MissingBranches []string
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

	return c.listCommitsByAuthorSinceRef(ctx, owner, repo, author, since, "")
}

// ListCommitsByAuthorSinceAcrossBranches returns commits in a repository authored by `author`
// since `since` across all branches, deduplicated by commit SHA.
func (c *Client) ListCommitsByAuthorSinceAcrossBranches(ctx context.Context, repoFullName, author string, since time.Time) ([]Commit, error) {
	result, err := c.ListCommitsByAuthorSinceByBranches(ctx, repoFullName, author, since, nil)
	if err != nil {
		return nil, err
	}
	return result.Commits, nil
}

// ListCommitsByAuthorSinceByBranches returns commits for requested branches.
// If requestedBranches is empty, all repository branches are scanned.
func (c *Client) ListCommitsByAuthorSinceByBranches(
	ctx context.Context,
	repoFullName, author string,
	since time.Time,
	requestedBranches []string,
) (BranchCommitResult, error) {
	owner, repo, err := splitRepoFullName(repoFullName)
	if err != nil {
		return BranchCommitResult{}, err
	}

	branches, err := c.listBranchNames(ctx, owner, repo)
	if err != nil {
		return BranchCommitResult{}, err
	}
	if len(branches) == 0 {
		return BranchCommitResult{
			Commits:         []Commit{},
			ScannedBranches: []string{},
			MissingBranches: dedupeBranchNames(requestedBranches),
		}, nil
	}

	available := make(map[string]struct{}, len(branches))
	for _, branch := range branches {
		available[branch] = struct{}{}
	}

	targetBranches := make([]string, 0)
	missing := make([]string, 0)
	if len(requestedBranches) == 0 {
		targetBranches = append(targetBranches, branches...)
	} else {
		for _, branch := range dedupeBranchNames(requestedBranches) {
			if _, ok := available[branch]; ok {
				targetBranches = append(targetBranches, branch)
			} else {
				missing = append(missing, branch)
			}
		}
	}

	seen := make(map[string]Commit)
	for _, branch := range targetBranches {
		commits, err := c.listCommitsByAuthorSinceRef(ctx, owner, repo, author, since, branch)
		if err != nil {
			return BranchCommitResult{}, fmt.Errorf("failed to list commits for branch %q in %s/%s: %w", branch, owner, repo, err)
		}
		for _, commit := range commits {
			existing, exists := seen[commit.SHA]
			if exists {
				existing.Branches = append(existing.Branches, branch)
				existing.Branches = dedupeBranchNames(existing.Branches)
				seen[commit.SHA] = existing
				continue
			}
			commit.Branches = []string{branch}
			seen[commit.SHA] = commit
		}
	}

	all := make([]Commit, 0, len(seen))
	for _, commit := range seen {
		all = append(all, commit)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].AuthoredAt.After(all[j].AuthoredAt)
	})
	return BranchCommitResult{
		Commits:         all,
		ScannedBranches: targetBranches,
		MissingBranches: missing,
	}, nil
}

func (c *Client) listCommitsByAuthorSinceRef(
	ctx context.Context,
	owner, repo, author string,
	since time.Time,
	ref string,
) ([]Commit, error) {
	var all []Commit
	for page := 1; ; page++ {
		pageItems, err := c.fetchCommitPage(ctx, owner, repo, author, since, ref, page)
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

func (c *Client) fetchCommitPage(ctx context.Context, owner, repo, author string, since time.Time, ref string, page int) ([]userCommitListItem, error) {
	endpoint, err := url.Parse(c.baseURL + "/repos/" + owner + "/" + repo + "/commits")
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("author", author)
	q.Set("since", since.UTC().Format(time.RFC3339))
	q.Set("per_page", strconv.Itoa(repoPageSize))
	q.Set("page", strconv.Itoa(page))
	if trimmedRef := strings.TrimSpace(ref); trimmedRef != "" {
		q.Set("sha", trimmedRef)
	}
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

func (c *Client) listBranchNames(ctx context.Context, owner, repo string) ([]string, error) {
	names := make([]string, 0)
	seen := make(map[string]struct{})

	for page := 1; ; page++ {
		items, err := c.fetchBranchPage(ctx, owner, repo, page)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			name := strings.TrimSpace(item.Name)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
		if len(items) < repoPageSize {
			break
		}
	}
	return names, nil
}

func (c *Client) fetchBranchPage(ctx context.Context, owner, repo string, page int) ([]repoBranchListItem, error) {
	endpoint, err := url.Parse(c.baseURL + "/repos/" + owner + "/" + repo + "/branches")
	if err != nil {
		return nil, fmt.Errorf("failed to parse branch endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("per_page", strconv.Itoa(repoPageSize))
	q.Set("page", strconv.Itoa(page))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch request for %s/%s: %w", owner, repo, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("github branch request timed out for %s/%s: %w", owner, repo, err)
		}
		return nil, fmt.Errorf("github branch request failed for %s/%s: %w", owner, repo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read branch response for %s/%s: %w", owner, repo, err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusConflict {
		// Empty repositories can return 409 for branch listing.
		return []repoBranchListItem{}, nil
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var payload []repoBranchListItem
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse branch response for %s/%s: %w", owner, repo, err)
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

func dedupeBranchNames(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
