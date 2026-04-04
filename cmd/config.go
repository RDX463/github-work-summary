package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".gws")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Create empty config file if it doesn't exist
		configPath := filepath.Join(home, ".gws.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			os.Create(configPath)
		}
	}
}

func saveConfig() error {
	return viper.WriteConfig()
}
