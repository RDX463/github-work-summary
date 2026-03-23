package auth

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

// TokenStore abstracts secure credential storage.
type TokenStore interface {
	SaveToken(token string) error
	GetToken() (string, error)
	DeleteToken() error
}

// KeyringStore stores the OAuth token in the OS credential store.
type KeyringStore struct {
	serviceName string
	accountName string
}

func NewKeyringStore(serviceName, accountName string) *KeyringStore {
	return &KeyringStore{
		serviceName: serviceName,
		accountName: accountName,
	}
}

func (k *KeyringStore) SaveToken(token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return fmt.Errorf("token is empty")
	}
	if err := keyring.Set(k.serviceName, k.accountName, trimmed); err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return nil
}

func (k *KeyringStore) GetToken() (string, error) {
	token, err := keyring.Get(k.serviceName, k.accountName)
	if err != nil {
		return "", fmt.Errorf("failed to read token from keyring: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return "", ErrInvalidToken
	}
	return token, nil
}

func (k *KeyringStore) DeleteToken() error {
	if err := keyring.Delete(k.serviceName, k.accountName); err != nil {
		return fmt.Errorf("failed to remove token from keyring: %w", err)
	}
	return nil
}
