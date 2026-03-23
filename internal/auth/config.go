package auth

import (
	"errors"
	"strings"
	"time"
)

const (
	// EnvClientID is the env var used by the CLI to load the OAuth app client id.
	EnvClientID = "GITHUB_CLIENT_ID"
	// EnvClientSecret is optional and used by some OAuth app setups.
	EnvClientSecret = "GITHUB_CLIENT_SECRET"

	// DefaultOAuthClientID is the default client id used by this CLI.
	DefaultOAuthClientID = "Ov23ligntwBLeqDJi7Xg"

	DefaultServiceName        = "github-work-summary"
	DefaultTokenAccount       = "access_token"
	defaultScope              = "repo read:user"
	defaultHTTPTimeout        = 12 * time.Second
	defaultDeviceCodeURL      = "https://github.com/login/device/code"
	defaultAccessTokenURL     = "https://github.com/login/oauth/access_token"
	defaultTokenValidationURL = "https://api.github.com/user"
)

// Config contains runtime settings for GitHub device-flow authentication.
type Config struct {
	ClientID           string
	ClientSecret       string
	Scope              string
	ServiceName        string
	DeviceCodeURL      string
	AccessTokenURL     string
	TokenValidationURL string
	HTTPTimeout        time.Duration
}

// DefaultConfig returns safe defaults for GitHub's OAuth device flow.
func DefaultConfig() Config {
	return Config{
		ClientID:           DefaultOAuthClientID,
		Scope:              defaultScope,
		ServiceName:        DefaultServiceName,
		DeviceCodeURL:      defaultDeviceCodeURL,
		AccessTokenURL:     defaultAccessTokenURL,
		TokenValidationURL: defaultTokenValidationURL,
		HTTPTimeout:        defaultHTTPTimeout,
	}
}

func (c Config) validate() error {
	if strings.TrimSpace(c.ClientID) == "" {
		return ErrMissingClientID
	}
	if strings.TrimSpace(c.Scope) == "" {
		return errors.New("scope cannot be empty")
	}
	if strings.TrimSpace(c.ServiceName) == "" {
		return errors.New("service name cannot be empty")
	}
	if strings.TrimSpace(c.DeviceCodeURL) == "" {
		return errors.New("device code URL cannot be empty")
	}
	if strings.TrimSpace(c.AccessTokenURL) == "" {
		return errors.New("access token URL cannot be empty")
	}
	if strings.TrimSpace(c.TokenValidationURL) == "" {
		return errors.New("token validation URL cannot be empty")
	}
	if c.HTTPTimeout <= 0 {
		return errors.New("HTTP timeout must be greater than zero")
	}
	return nil
}
