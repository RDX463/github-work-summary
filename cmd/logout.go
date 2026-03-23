package cmd

import (
	"fmt"

	"github.com/RDX463/github-work-summary/internal/auth"
	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored GitHub credentials from keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogout(cmd)
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	store := auth.NewKeyringStore(auth.DefaultServiceName, auth.DefaultTokenAccount)
	err := store.DeleteToken()
	if err != nil {
		if auth.IsTokenNotFoundError(err) {
			fmt.Fprintln(out, ui.Yellow(out, "No stored GitHub token found. Already logged out."))
			return nil
		}
		return err
	}

	fmt.Fprintln(out, ui.Green(out, "Logged out. GitHub token removed from OS keychain."))
	return nil
}
