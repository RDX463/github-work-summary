package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/RDX463/github-work-summary/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles (work, personal, etc.)",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		active := getActiveProfileName()
		profiles := getProfileNames()

		fmt.Fprintln(out, ui.Bold(out, ui.Cyan(out, "Profiles:")))
		for _, name := range profiles {
			prefix := "  "
			if name == active {
				prefix = "➤ "
				fmt.Fprintf(out, "%s %s %s\n", ui.Green(out, prefix), ui.Bold(out, ui.Green(out, name)), ui.Gray(out, "(active)"))
			} else {
				fmt.Fprintf(out, "%s %s\n", prefix, name)
			}
		}
	},
}

var profileSwitchCmd = &cobra.Command{
	Use:   "switch [name]",
	Short: "Change the default active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.ToLower(strings.TrimSpace(args[0]))
		profiles := getProfileNames()
		found := false
		for _, p := range profiles {
			if p == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("profile %q does not exist. run `gws profile add %s` first", name, name)
		}

		viper.Set(keyActiveProfile, name)
		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s Switched active profile to %s\n", ui.Green(cmd.OutOrStdout(), "✓"), ui.Bold(cmd.OutOrStdout(), name))
		return nil
	},
}

var profileAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Create a new configuration profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.ToLower(strings.TrimSpace(args[0]))
		profiles := getProfileNames()
		for _, p := range profiles {
			if p == name {
				return fmt.Errorf("profile %q already exists", name)
			}
		}

		// Initializing the profile with empty settings
		viper.Set(getProfileKey(name, "repositories"), []string{})
		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s Added new profile %s\n", ui.Green(cmd.OutOrStdout(), "✓"), ui.Bold(cmd.OutOrStdout(), name))
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Remove a configuration profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.ToLower(strings.TrimSpace(args[0]))
		if name == defaultProfile {
			return fmt.Errorf("cannot delete the default profile")
		}

		active := getActiveProfileName()
		if active == name {
			return fmt.Errorf("cannot delete an active profile. switch to another profile first")
		}

		profiles := viper.GetStringMap("profiles")
		delete(profiles, name)
		viper.Set("profiles", profiles)

		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s Deleted profile %s\n", ui.Green(cmd.OutOrStdout(), "✓"), ui.Bold(cmd.OutOrStdout(), name))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileDeleteCmd)
}

func getProfileNames() []string {
	profilesMap := viper.GetStringMap("profiles")
	names := make([]string, 0, len(profilesMap)+1)
	for name := range profilesMap {
		names = append(names, name)
	}

	// Ensure "default" is always present in the list even if map is empty
	hasDefault := false
	for _, n := range names {
		if n == defaultProfile {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		names = append(names, defaultProfile)
	}

	sort.Strings(names)
	return names
}
