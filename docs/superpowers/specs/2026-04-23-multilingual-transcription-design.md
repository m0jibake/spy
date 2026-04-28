# Multilingual Transcription Design

**Date:** 2026-04-23  
**Status:** Approved

## Goal

Add German transcription support to `spy` without changing existing English behavior. Language is selected per-recording via keybind.

## Keybinds

| Keybind | Language | whisper-cli flag |
|---|---|---|
| `Option` (existing) | English | `-l en` |
| `Option + 1` (new) | German | `-l de` |

## Components

### `internal/transcriber/transcriber.go`

Add a `language string` parameter to `Transcribe`:

```go
func Transcribe(modelPath, audioFile, language string) (string, error)
```

Pass it to `whisper-cli` as `-l <language>`:

```go
exec.Command("whisper-cli", "-m", modelPath, "-f", audioFile, "-l", language)
```

### `cmd/root.go`

- Register a second keybind pair for `alt+1` (German) alongside the existing `alt` (English) pair.
- Thread the language string through `processAudio(model, audioFile, language string)`.
- Show language in the recording prompt: `🎙️  Recording [DE]... (release ⌥1 to stop)`
- On startup, warn if the configured model path contains `.en.` — English-only models cannot transcribe German.

## Error Handling

- Startup warning (not fatal) if model path contains `.en.`: `⚠️  Model appears to be English-only. German transcription requires a multilingual model.`
- No other new error paths — existing error handling covers transcription failures.

## Known macOS behaviour

`Option + 1` normally produces `¡` on a US keyboard. Since gohook observes but does not suppress key events on macOS, this character may be typed into the focused app while holding the keybind. In practice this is harmless (the key is held, not tapped repeatedly), but worth noting.

## No config changes

Language is fully determined by keybind. No new config keys.

## Prerequisites for the user

- Must use a multilingual whisper model (e.g. `ggml-small.bin`, `ggml-medium.bin` — no `.en.` in filename).
- `whisper-cli` must be installed (already required).
