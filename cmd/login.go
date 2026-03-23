package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/RDX463/github-work-summary/internal/auth"
	"github.com/spf13/cobra"
)

var loginClientID string
var loginClientSecret string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub using device flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin(cmd)
	},
}

func init() {
	defaultClientID := strings.TrimSpace(os.Getenv(auth.EnvClientID))
	if defaultClientID == "" {
		defaultClientID = auth.DefaultOAuthClientID
	}
	loginCmd.Flags().StringVar(&loginClientID, "client-id", defaultClientID, "GitHub OAuth App client ID")
	loginCmd.Flags().StringVar(&loginClientSecret, "client-secret", strings.TrimSpace(os.Getenv(auth.EnvClientSecret)), "GitHub OAuth App client secret (optional)")
	rootCmd.AddCommand(loginCmd)
}

func splitClientCredentials(clientID, clientSecret string) (string, string, bool) {
	id := strings.TrimSpace(clientID)
	secret := strings.TrimSpace(clientSecret)
	if !strings.Contains(id, ",") {
		return id, secret, false
	}

	parts := strings.SplitN(id, ",", 2)
	id = strings.TrimSpace(parts[0])
	if secret == "" && len(parts) == 2 {
		secret = strings.TrimSpace(parts[1])
	}
	return id, secret, true
}

func runLogin(cmd *cobra.Command) error {
	cfg := auth.DefaultConfig()

	clientID, clientSecret, hadCombinedInput := splitClientCredentials(loginClientID, loginClientSecret)
	cfg.ClientID = clientID
	cfg.ClientSecret = clientSecret
	if hadCombinedInput {
		fmt.Fprintln(cmd.ErrOrStderr(), "Detected combined client credentials input. Using first value as client ID.")
	}

	client, err := auth.NewClient(cfg, nil)
	if err != nil {
		if errors.Is(err, auth.ErrMissingClientID) {
			return fmt.Errorf("%w. set %s or use --client-id", err, auth.EnvClientID)
		}
		return err
	}

	prompt, err := client.StartDeviceFlow(cmd.Context())
	if err != nil {
		if errors.Is(err, auth.ErrDeviceFlowDisabled) {
			return fmt.Errorf(
				"%w. enable Device Flow in GitHub Developer Settings > OAuth Apps > your app",
				err,
			)
		}
		return err
	}

	fmt.Printf("Open %s and enter code: %s\n", prompt.VerificationURI, prompt.UserCode)
	if prompt.VerificationURIComplete != "" {
		fmt.Printf("Or open this one-time URL: %s\n", prompt.VerificationURIComplete)
	}

	token, err := client.PollForToken(cmd.Context(), prompt)
	if err != nil {
		return err
	}

	if err := client.SaveToken(token); err != nil {
		return err
	}

	fmt.Println("GitHub authentication complete. Token stored in OS keychain.")
	return nil
}
