package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	hook "github.com/robotn/gohook"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/arnemeerbott/spy/internal/clipboard"
	"github.com/arnemeerbott/spy/internal/recorder"
	"github.com/arnemeerbott/spy/internal/transcriber"
)

var (
	cfgFile   string
	modelFlag string
	audioFlag string
)

var rootCmd = &cobra.Command{
	Use:   "spy",
	Short: "Hold ⌥ Option to record speech; release to transcribe and copy to clipboard.",
	Long: `spy — speech-to-clipboard via local whisper model.

Hold the Option (⌥) key to start recording. Release to stop,
transcribe, and copy the result to your clipboard.

Requires: sox, whisper-cli (whisper.cpp).
macOS only.`,
	RunE: runListener,
}

// Execute is the entrypoint called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/spy/config.yaml)")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "path to whisper model .bin file")
	rootCmd.Flags().StringVar(&audioFlag, "audio-file", "", "temp audio file path (default: /tmp/spy_speech.wav)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(filepath.Join(home, ".config", "spy"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetDefault("audio_file", "/tmp/spy_speech.wav")
	viper.AutomaticEnv()

	// Ignore missing config file — user may not have run `spy config set` yet.
	_ = viper.ReadInConfig()

	// CLI flags override config file values.
	if modelFlag != "" {
		viper.Set("model", modelFlag)
	}
	if audioFlag != "" {
		viper.Set("audio_file", audioFlag)
	}
}

func runListener(cmd *cobra.Command, args []string) error {
	model := viper.GetString("model")
	if model == "" {
		return fmt.Errorf("no model configured — run: spy config set model /path/to/model.bin")
	}
	if _, err := os.Stat(model); err != nil {
		return fmt.Errorf("model file not found: %s", model)
	}

	audioFile := viper.GetString("audio_file")

	fmt.Println("🕵️  spy is watching.")
	fmt.Println("   Hold ⌥ Option to record. Release to transcribe & copy.")
	fmt.Println("   Press Ctrl+C to quit.\n")

	rec := recorder.New(audioFile)
	recording := false

	hook.Register(hook.KeyDown, []string{"alt"}, func(e hook.Event) {
		if !recording {
			recording = true
			fmt.Print("🎙️  Recording... (release ⌥ to stop)")
			if err := rec.Start(); err != nil {
				fmt.Printf("\n❌ Failed to start recording: %v\n", err)
				recording = false
			}
		}
	})

	hook.Register(hook.KeyUp, []string{"alt"}, func(e hook.Event) {
		if recording {
			recording = false
			fmt.Println()
			if err := rec.Stop(); err != nil {
				fmt.Printf("❌ Failed to stop recording: %v\n", err)
				return
			}
			go processAudio(model, audioFile)
		}
	})

	// Graceful shutdown on Ctrl+C.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n👋 Goodbye.")
		if recording {
			_ = rec.Stop()
		}
		hook.End()
	}()

	s := hook.Start()
	<-hook.Process(s)
	return nil
}

// processAudio transcribes and copies to clipboard.
func processAudio(model, audioFile string) {
	fmt.Print("⏳ Processing... ")

	text, err := transcriber.Transcribe(model, audioFile)
	if err != nil {
		fmt.Printf("\n❌ Transcription error: %v\n", err)
		return
	}
	if text == "" {
		fmt.Println("\n🔇 No speech detected.")
		return
	}

	fmt.Printf("\n📝 %s\n", text)

	if err := clipboard.Copy(text); err != nil {
		fmt.Printf("❌ Clipboard error: %v\n", err)
		return
	}
	fmt.Println("📋 Copied to clipboard.")
}
