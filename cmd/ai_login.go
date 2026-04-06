package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/RDX463/github-work-summary/internal/auth"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
)

const (
	googleAIServiceName = "github-work-summary-google-ai"
	googleAIAccountName = "api-key"
)

var aiLoginCmd = &cobra.Command{
	Use:   "ai-login",
	Short: "Store Google Gemini API key securely in keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAILogin(cmd)
	},
}

func init() {
	rootCmd.AddCommand(aiLoginCmd)
}

func runAILogin(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	fmt.Fprintln(out, ui.Bold(out, "Step 1: Get your Google AI API Key"))
	fmt.Fprintf(out, "Go to %s and generate a free API key.\n\n", ui.Cyan(out, "https://aistudio.google.com/app/apikey"))

	fmt.Fprint(out, ui.Bold(out, "Step 2: Enter your API Key: "))
	
	reader := bufio.NewReader(in)
	key, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	key = strings.TrimSpace(key)

	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	store := auth.NewKeyringStore(googleAIServiceName, googleAIAccountName)
	if err := store.SaveToken(key); err != nil {
		return err
	}

	fmt.Fprintln(out, ui.Green(out, "\n✓ Google AI API Key stored securely in OS keychain."))
	fmt.Fprintln(out, ui.Gray(out, "You can now use the --ai flag with the summary command."))
	
	return nil
}

func getGoogleAIKey() (string, error) {
	store := auth.NewKeyringStore(googleAIServiceName, googleAIAccountName)
	return store.GetToken()
}
