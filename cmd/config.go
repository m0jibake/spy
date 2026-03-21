package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify spy configuration",
	Long:  "Print the current spy configuration or set individual values.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Config file: %s\n\n", viper.ConfigFileUsed())
		fmt.Printf("model:      %s\n", viper.GetString("model"))
		fmt.Printf("audio_file: %s\n", viper.GetString("audio_file"))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Example: `  spy config set model /path/to/ggml-tiny.en.bin
  spy config set audio_file /tmp/spy_speech.wav`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		allowed := map[string]bool{"model": true, "audio_file": true}
		if !allowed[key] {
			return fmt.Errorf("unknown config key %q (allowed: model, audio_file)", key)
		}

		viper.Set(key, value)

		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			cfgFile = filepath.Join(home, ".config", "spy", "config.yaml")
		}

		if err := os.MkdirAll(filepath.Dir(cfgFile), 0o755); err != nil {
			return err
		}

		if err := viper.WriteConfigAs(cfgFile); err != nil {
			return err
		}

		fmt.Printf("✅ Set %s = %s  (saved to %s)\n", key, value, cfgFile)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
