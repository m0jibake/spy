# Multilingual Transcription Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add German transcription support via a second keybind (`Option+1`) while keeping `Option` for English.

**Architecture:** Add a `language` parameter to `transcriber.Transcribe` that passes `-l <lang>` to `whisper-cli`. In `cmd/root.go`, register a second keybind pair for `alt+1`; when both `alt` and `1` are detected, the language is upgraded from `en` to `de` mid-press. A startup warning is printed if the configured model appears to be English-only.

**Tech Stack:** Go 1.24, `github.com/robotn/gohook`, `whisper-cli` (whisper.cpp CLI)

---

## File Map

| File | Change |
|---|---|
| `internal/transcriber/transcriber.go` | Add `language string` param to `Transcribe`; extract `buildWhisperArgs` helper |
| `internal/transcriber/transcriber_test.go` | Create — tests for `buildWhisperArgs` |
| `cmd/root.go` | Add German keybind pair; thread `language` through `processAudio`; add model warning |

---

### Task 1: Add language param to `Transcribe` and extract testable helper

**Files:**
- Modify: `internal/transcriber/transcriber.go`
- Create: `internal/transcriber/transcriber_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/transcriber/transcriber_test.go`:

```go
package transcriber

import (
	"reflect"
	"testing"
)

func TestBuildWhisperArgs(t *testing.T) {
	tests := []struct {
		modelPath string
		audioFile string
		language  string
		want      []string
	}{
		{"/models/ggml-small.bin", "/tmp/spy.wav", "en", []string{"-m", "/models/ggml-small.bin", "-f", "/tmp/spy.wav", "-l", "en"}},
		{"/models/ggml-small.bin", "/tmp/spy.wav", "de", []string{"-m", "/models/ggml-small.bin", "-f", "/tmp/spy.wav", "-l", "de"}},
	}
	for _, tt := range tests {
		got := buildWhisperArgs(tt.modelPath, tt.audioFile, tt.language)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("buildWhisperArgs(%q, %q, %q) = %v, want %v", tt.modelPath, tt.audioFile, tt.language, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
go test ./internal/transcriber/
```

Expected: `FAIL — undefined: buildWhisperArgs`

- [ ] **Step 3: Update `internal/transcriber/transcriber.go`**

Replace the file content with:

```go
package transcriber

import (
	"os/exec"
	"regexp"
	"strings"
)

var (
	reBrackets  = regexp.MustCompile(`\[.*?\]`)
	reTimestamp = regexp.MustCompile(`\d+:\d+:\d+\.\d+\s*-->\s*\d+:\d+:\d+\.\d+`)
)

func buildWhisperArgs(modelPath, audioFile, language string) []string {
	return []string{"-m", modelPath, "-f", audioFile, "-l", language}
}

// Transcribe runs whisper-cli on audioFile using modelPath and returns
// the cleaned transcript text. language is a BCP-47 tag (e.g. "en", "de").
func Transcribe(modelPath, audioFile, language string) (string, error) {
	out, err := exec.Command("whisper-cli", buildWhisperArgs(modelPath, audioFile, language)...).Output()
	if err != nil {
		return "", err
	}

	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "-->") {
			continue
		}
		cleaned := reBrackets.ReplaceAllString(line, "")
		cleaned = reTimestamp.ReplaceAllString(cleaned, "")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			lines = append(lines, cleaned)
		}
	}

	return strings.Join(lines, " "), nil
}
```

- [ ] **Step 4: Run test to confirm it passes**

```bash
go test ./internal/transcriber/
```

Expected: `ok  github.com/arnemeerbott/spy/internal/transcriber`

- [ ] **Step 5: Commit**

```bash
git add internal/transcriber/transcriber.go internal/transcriber/transcriber_test.go
git commit -m "feat: add language param to transcriber.Transcribe"
```

---

### Task 2: Update `cmd/root.go` — German keybind, language threading, model warning

**Files:**
- Modify: `cmd/root.go`

**How the dual-trigger works:**
Pressing `Option+1` fires the `alt` KeyDown first (starting English recording), then immediately fires the `alt+1` KeyDown which upgrades `recordingLang` to `"de"`. Releasing `1` fires `alt+1` KeyUp which stops and processes as German. The `alt` KeyUp handler checks `recordingLang == "en"` and skips, so there's no double-processing.

- [ ] **Step 1: Write the failing build check**

```bash
go build ./...
```

Expected: `FAIL — too many arguments in call to transcriber.Transcribe` (because `root.go` still calls the old 2-arg signature)

- [ ] **Step 2: Replace `cmd/root.go` with the updated version**

```go
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

Hold the Option (⌥) key to start recording in English. Release to stop,
transcribe, and copy the result to your clipboard.

For German: hold ⌥ then also hold 1 — release 1 to transcribe as German.

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
	fmt.Println("   Hold ⌥ to record in English. Release to transcribe & copy.")
	fmt.Println("   Hold ⌥ then 1 to record in German. Release 1 to transcribe & copy.")
	fmt.Println("   Press Ctrl+C to quit.\n")

	if strings.Contains(model, ".en.") {
		fmt.Println("⚠️  Model appears to be English-only. German transcription requires a multilingual model (e.g. ggml-small.bin).\n")
	}

	rec := recorder.New(audioFile)
	recording := false
	recordingLang := ""

	hook.Register(hook.KeyDown, []string{"alt"}, func(e hook.Event) {
		if !recording {
			recording = true
			recordingLang = "en"
			fmt.Print("🎙️  Recording [EN]... (release ⌥ to stop)")
			if err := rec.Start(); err != nil {
				fmt.Printf("\n❌ Failed to start recording: %v\n", err)
				recording = false
			}
		}
	})

	// alt+1 fires after alt alone when both are held — upgrade language to German.
	hook.Register(hook.KeyDown, []string{"alt", "1"}, func(e hook.Event) {
		if recording && recordingLang == "en" {
			recordingLang = "de"
			fmt.Print("\r🎙️  Recording [DE]... (release 1 to stop)     ")
		}
	})

	hook.Register(hook.KeyUp, []string{"alt", "1"}, func(e hook.Event) {
		if recording && recordingLang == "de" {
			recording = false
			fmt.Println()
			if err := rec.Stop(); err != nil {
				fmt.Printf("❌ Failed to stop recording: %v\n", err)
				return
			}
			go processAudio(model, audioFile, "de")
		}
	})

	hook.Register(hook.KeyUp, []string{"alt"}, func(e hook.Event) {
		if recording && recordingLang == "en" {
			recording = false
			fmt.Println()
			if err := rec.Stop(); err != nil {
				fmt.Printf("❌ Failed to stop recording: %v\n", err)
				return
			}
			go processAudio(model, audioFile, "en")
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
func processAudio(model, audioFile, language string) {
	fmt.Print("⏳ Processing... ")

	text, err := transcriber.Transcribe(model, audioFile, language)
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
```

- [ ] **Step 3: Build to confirm it compiles**

```bash
go build ./...
```

Expected: no output (success)

- [ ] **Step 4: Run tests to confirm nothing is broken**

```bash
go test ./...
```

Expected: `ok  github.com/arnemeerbott/spy/internal/transcriber`

- [ ] **Step 5: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add German keybind (⌥1) and model warning"
```

---

## Post-implementation user instructions

To use German transcription, the user must have a **multilingual** whisper model (no `.en.` in filename). Download one:

```bash
# example — small multilingual model (~466 MB)
curl -L -o ~/models/ggml-small.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin

spy config set model ~/models/ggml-small.bin
```

Then:
- Hold `⌥` → English
- Hold `⌥`, then also hold `1` → German (release `1` to stop)
