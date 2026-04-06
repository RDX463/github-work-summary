package githubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL          = "https://api.github.com"
	githubAPIVersion        = "2022-11-28"
	defaultRequestTimeout   = 15 * time.Second
	maxAPIResponseBodyBytes = 2 << 20 // 2 MiB
	repoPageSize            = 100
)

var (
	ErrUnauthorized = errors.New("github API unauthorized")
)

// type Repository is now in types.go

type errorResponse struct {
	Message string `json:"message"`
}

// Client performs authenticated requests to GitHub's REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(token string) (*Client, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil, fmt.Errorf("github token cannot be empty")
	}

	return &Client{
		baseURL: defaultBaseURL,
		token:   trimmed,
		httpClient: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}, nil
}

// ListAccessibleRepositories fetches all repositories the authenticated user can access.
func (c *Client) ListAccessibleRepositories(ctx context.Context) ([]Repository, error) {
	var all []Repository
	seen := make(map[string]struct{})

	for page := 1; ; page++ {
		repos, err := c.fetchRepoPage(ctx, page)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if _, exists := seen[repo.FullName]; exists {
				continue
			}
			seen[repo.FullName] = struct{}{}
			all = append(all, repo)
		}
		if len(repos) < repoPageSize {
			break
		}
	}

	return all, nil
}

func (c *Client) fetchRepoPage(ctx context.Context, page int) ([]Repository, error) {
	endpoint, err := url.Parse(c.baseURL + "/user/repos")
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitHub endpoint: %w", err)
	}

	q := endpoint.Query()
	q.Set("per_page", strconv.Itoa(repoPageSize))
	q.Set("page", strconv.Itoa(page))
	q.Set("sort", "full_name")
	q.Set("direction", "asc")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, fmt.Errorf("github repository request timed out: %w", err)
		}
		return nil, fmt.Errorf("github repository request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read repository response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var repos []Repository
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repository response: %w", err)
	}

	return repos, nil
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
