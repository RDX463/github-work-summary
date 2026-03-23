package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/RDX463/github-work-summary/internal/update"
	"github.com/RDX463/github-work-summary/internal/version"
	"github.com/spf13/cobra"
)

func maybeNotifyUpdate(cmd *cobra.Command) {
	// Allow disabling in automation or CI.
	if os.Getenv("GWS_NO_UPDATE_CHECK") == "1" {
		return
	}

	// Avoid noise for shell completion generation.
	if cmd != nil && cmd.Name() == "completion" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	notice, err := update.Check(ctx, version.Repo, version.Current())
	if err != nil || notice == nil {
		return
	}

	out := cmd.ErrOrStderr()
	fmt.Fprintf(out, "\nUpdate available: %s -> %s\n", notice.CurrentVersion, notice.LatestVersion)
	if len(notice.Changes) > 0 {
		fmt.Fprintln(out, "What's new:")
		for _, change := range notice.Changes {
			fmt.Fprintf(out, "- %s\n", change)
		}
	}
	if strings.TrimSpace(notice.URL) != "" {
		fmt.Fprintf(out, "Release: %s\n", notice.URL)
	}
	fmt.Fprintln(out, "Update: curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | bash")
	fmt.Fprintln(out)
}
