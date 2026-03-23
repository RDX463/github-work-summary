package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBodyBytes = 1 << 20 // 1 MiB

type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error"`
	ErrorDescription        string `json:"error_description"`
}

type accessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// DeviceCodePrompt contains user-facing instructions for completing device flow.
type DeviceCodePrompt struct {
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresAt               time.Time
	PollInterval            time.Duration

	// DeviceCode is kept internal to the CLI flow and should not be displayed.
	DeviceCode string
}

// Client handles GitHub OAuth device flow and token storage.
type Client struct {
	cfg        Config
	httpClient *http.Client
	store      TokenStore
}

func NewClient(cfg Config, store TokenStore) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if store == nil {
		store = NewKeyringStore(cfg.ServiceName, DefaultTokenAccount)
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		store: store,
	}, nil
}

// StartDeviceFlow requests a user code and verification URL from GitHub.
func (c *Client) StartDeviceFlow(ctx context.Context) (DeviceCodePrompt, error) {
	form := url.Values{}
	form.Set("client_id", c.cfg.ClientID)
	form.Set("scope", c.cfg.Scope)

	body, statusCode, err := c.postForm(ctx, c.cfg.DeviceCodeURL, form)
	if err != nil {
		return DeviceCodePrompt{}, err
	}

	var response deviceCodeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return DeviceCodePrompt{}, fmt.Errorf("failed to parse device code response: %w", err)
	}

	if response.Error != "" {
		return DeviceCodePrompt{}, c.githubError(response.Error, response.ErrorDescription)
	}
	if statusCode >= 400 {
		return DeviceCodePrompt{}, fmt.Errorf("device code request failed with status %d", statusCode)
	}
	if strings.TrimSpace(response.DeviceCode) == "" || strings.TrimSpace(response.UserCode) == "" || strings.TrimSpace(response.VerificationURI) == "" {
		return DeviceCodePrompt{}, errors.New("device code response missing required fields")
	}

	pollInterval := time.Duration(response.Interval) * time.Second
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	expiresAfter := time.Duration(response.ExpiresIn) * time.Second
	if expiresAfter <= 0 {
		expiresAfter = 15 * time.Minute
	}

	return DeviceCodePrompt{
		UserCode:                response.UserCode,
		VerificationURI:         response.VerificationURI,
		VerificationURIComplete: response.VerificationURIComplete,
		ExpiresAt:               time.Now().Add(expiresAfter),
		PollInterval:            pollInterval,
		DeviceCode:              response.DeviceCode,
	}, nil
}

// PollForToken polls GitHub until the user authorizes the device or the code expires.
func (c *Client) PollForToken(ctx context.Context, prompt DeviceCodePrompt) (string, error) {
	if strings.TrimSpace(prompt.DeviceCode) == "" {
		return "", ErrInvalidDeviceCode
	}

	pollInterval := prompt.PollInterval
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	for {
		if !prompt.ExpiresAt.IsZero() && time.Now().After(prompt.ExpiresAt) {
			return "", ErrDeviceCodeExpired
		}

		tokenResp, statusCode, err := c.requestAccessToken(ctx, prompt.DeviceCode)
		if err != nil {
			if isTimeoutError(err) {
				// Network timeout should not break the full login flow; we retry.
				if err := sleepWithContext(ctx, pollInterval); err != nil {
					return "", err
				}
				continue
			}
			return "", err
		}

		if statusCode == http.StatusUnauthorized {
			return "", ErrInvalidToken
		}
		if statusCode >= 500 {
			if err := sleepWithContext(ctx, pollInterval); err != nil {
				return "", err
			}
			continue
		}

		if strings.TrimSpace(tokenResp.AccessToken) != "" {
			return tokenResp.AccessToken, nil
		}

		switch tokenResp.Error {
		case "authorization_pending":
			// User has not approved yet; keep polling.
		case "slow_down":
			pollInterval += 5 * time.Second
		case "expired_token":
			return "", ErrDeviceCodeExpired
		case "access_denied":
			return "", ErrAccessDenied
		case "incorrect_device_code", "invalid_grant":
			return "", ErrInvalidDeviceCode
		case "invalid_client":
			return "", ErrMissingClientID
		case "":
			return "", errors.New("token response did not include an access token")
		default:
			return "", c.githubError(tokenResp.Error, tokenResp.ErrorDescription)
		}

		if err := sleepWithContext(ctx, pollInterval); err != nil {
			return "", err
		}
	}
}

// SaveToken stores the access token in the OS keychain.
func (c *Client) SaveToken(token string) error {
	return c.store.SaveToken(token)
}

// GetToken reads the stored access token from the OS keychain.
func (c *Client) GetToken() (string, error) {
	return c.store.GetToken()
}

// ValidateToken checks whether a token is still accepted by GitHub.
func (c *Client) ValidateToken(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.TokenValidationURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build token validation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return fmt.Errorf("token validation request timed out: %w", err)
		}
		return fmt.Errorf("token validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBodyBytes)); err != nil {
		return fmt.Errorf("failed to read token validation response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrInvalidToken
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) requestAccessToken(ctx context.Context, deviceCode string) (accessTokenResponse, int, error) {
	form := url.Values{}
	form.Set("client_id", c.cfg.ClientID)
	if strings.TrimSpace(c.cfg.ClientSecret) != "" {
		form.Set("client_secret", c.cfg.ClientSecret)
	}
	form.Set("device_code", deviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	body, statusCode, err := c.postForm(ctx, c.cfg.AccessTokenURL, form)
	if err != nil {
		return accessTokenResponse{}, 0, err
	}

	var response accessTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return accessTokenResponse{}, 0, fmt.Errorf("failed to parse access token response: %w", err)
	}
	return response, statusCode, nil
}

func (c *Client) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, 0, fmt.Errorf("request timed out: %w", err)
		}
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, resp.StatusCode, nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func (c *Client) githubError(code, description string) error {
	switch code {
	case "device_flow_disabled":
		if strings.TrimSpace(description) != "" {
			return fmt.Errorf("%w: %s", ErrDeviceFlowDisabled, description)
		}
		return ErrDeviceFlowDisabled
	}

	if strings.TrimSpace(description) != "" {
		return fmt.Errorf("github oauth error %q: %s", code, description)
	}
	return fmt.Errorf("github oauth error %q", code)
}
