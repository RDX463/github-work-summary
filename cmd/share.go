package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/RDX463/github-work-summary/internal/auth"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/RDX463/github-work-summary/internal/notify"
	"github.com/spf13/cobra"
)

const (
	slackWebhookService   = "github-work-summary-slack-webhook"
	discordWebhookService = "github-work-summary-discord-webhook"
	webhookAccount        = "url"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share your work summary with your team",
}

var shareSetupCmd = &cobra.Command{
	Use:   "setup [slack|discord]",
	Short: "Configure Slack or Discord webhooks for sharing",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please specify a platform: `gws share setup slack` or `gws share setup discord`")
		}
		return runShareSetup(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.AddCommand(shareSetupCmd)
}

func runShareSetup(cmd *cobra.Command, platform string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	var serviceName string
	var setupURL string

	switch strings.ToLower(platform) {
	case "slack":
		serviceName = slackWebhookService
		setupURL = "https://api.slack.com/messaging/webhooks"
	case "discord":
		serviceName = discordWebhookService
		setupURL = "https://support.discord.com/hc/en-us/articles/228383668"
	default:
		return fmt.Errorf("unsupported platform: %s (use slack or discord)", platform)
	}

	fmt.Fprintf(out, "%s %s\n", ui.Bold(out, "Configuring Integration:"), ui.Cyan(out, platform))
	fmt.Fprintf(out, "To get a webhook URL, visit: %s\n\n", ui.Gray(out, setupURL))

	fmt.Fprintf(out, "%s Enter your %s Webhook URL: ", ui.Bold(out, "Setup:"), platform)
	
	reader := bufio.NewReader(in)
	url, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	url = strings.TrimSpace(url)

	if url == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("invalid URL: must start with https://")
	}

	store := auth.NewKeyringStore(serviceName, webhookAccount)
	if err := store.SaveToken(url); err != nil {
		return fmt.Errorf("failed to save webhook: %w", err)
	}

	fmt.Fprintf(out, "\n%s %s webhook stored securely in OS keychain.\n", ui.Green(out, "✓"), platform)
	fmt.Fprintf(out, "%s You can now use `--share %s` with the summary command.\n", ui.Gray(out, "Tip:"), platform)
	
	return nil
}

func getWebhook(platform string) (string, error) {
	var serviceName string
	switch strings.ToLower(platform) {
	case "slack":
		serviceName = slackWebhookService
	case "discord":
		serviceName = discordWebhookService
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}

	store := auth.NewKeyringStore(serviceName, webhookAccount)
	token, err := store.GetToken()
	if err != nil || token == "" {
		return "", fmt.Errorf("webhook not configured. Run `gws share setup %s` first", strings.ToLower(platform))
	}
	return token, nil
}

func getNotifier(platform string) (notify.Notifier, error) {
	url, err := getWebhook(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve webhook for %s: %w. run `gws share setup %s` first", platform, err, platform)
	}

	switch strings.ToLower(platform) {
	case "slack":
		return &notify.SlackNotifier{WebhookURL: url}, nil
	case "discord":
		return &notify.DiscordNotifier{WebhookURL: url}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
