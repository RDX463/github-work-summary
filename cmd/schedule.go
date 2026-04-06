package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/RDX463/github-work-summary/internal/schedule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage automated background work summaries",
}

var scheduleSetCmd = &cobra.Command{
	Use:   "set \"HH:MM\" or \"Day HH:MM\"",
	Short: "Set a recurring schedule for automated summaries",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schedStr := args[0]
		s, err := schedule.Parse(schedStr)
		if err != nil { return err }

		platform, _ := cmd.Flags().GetString("share")
		if platform == "" { return fmt.Errorf("must specify --share [slack|discord]") }

		// Save to config
		viper.Set("automation.schedule", schedStr)
		viper.Set("automation.share", platform)
		saveConfig()

		// OS Specific Installation
		exe, _ := os.Executable()
		home, _ := os.UserHomeDir()
		logPath := filepath.Join(home, ".gws", "automation.log")
		errPath := filepath.Join(home, ".gws", "automation.err")
		
		cfg := schedule.LaunchAgentConfig{
			Label:          "com.rdx.gws.summary",
			ExecutablePath: exe,
			SharePlatform:  platform,
			Hour:           s.Hour,
			Minute:         s.Minute,
			Day:            int(s.Day),
			LogPath:        logPath,
			ErrorPath:      errPath,
		}

		if err := schedule.InstallLaunchAgent(cfg); err != nil {
			return fmt.Errorf("failed to install background job: %w", err)
		}

		fmt.Printf("✅ Automation scheduled for %s. Reports will be sent to %s.\n", schedStr, platform)
		fmt.Printf("Logs available at: %s\n", logPath)
		return nil
	},
}

var scheduleClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all automated background jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("automation.schedule", "")
		saveConfig()
		
		_ = schedule.RemoveLaunchAgent("com.rdx.gws.summary")
		fmt.Println("🛑 Background automation cleared.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(scheduleSetCmd)
	scheduleCmd.AddCommand(scheduleClearCmd)
	
	scheduleSetCmd.Flags().String("share", "", "Platform to share to (slack, discord)")
}
