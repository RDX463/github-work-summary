package auth

import "errors"

var (
	ErrMissingClientID    = errors.New("missing GitHub OAuth client id")
	ErrDeviceCodeExpired  = errors.New("device code expired before authorization completed")
	ErrAccessDenied       = errors.New("authorization denied by user")
	ErrInvalidDeviceCode  = errors.New("invalid or rejected device code")
	ErrDeviceFlowDisabled = errors.New("device flow is disabled for this OAuth app")
	ErrInvalidToken       = errors.New("stored GitHub token is invalid")
)
