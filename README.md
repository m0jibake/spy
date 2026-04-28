# spy

Speech-to-clipboard via a local whisper model. Hold a hotkey to record, release to transcribe — result is copied to your clipboard.

macOS only.

## Requirements

- [sox](https://sox.sourceforge.net/) — audio recording
- [whisper-cli](https://github.com/ggerganov/whisper.cpp) — transcription
- A whisper model file (multilingual, e.g. `ggml-small.bin`)

Install dependencies via Homebrew:

```
brew install sox
brew install whisper-cpp
```

## Configuration

Set your model path before first use:

```
spy config set model /path/to/ggml-small.bin
```

> Use a multilingual model (without `.en.` in the filename) if you want German transcription.

## Usage

| Action | Result |
|--------|--------|
| Hold `⌥` (Option) | Record in English |
| Hold `⌥⇧` (Option + Shift) | Record in German |
| Release key(s) | Stop, transcribe, copy to clipboard |
| `Ctrl+C` | Quit |

## Reinstalling after source changes

If you installed `spy` via Homebrew, the Homebrew binary takes priority in your PATH. Build and overwrite it directly:

```
cd /path/to/spy
go build -o spy . && sudo mv spy /opt/homebrew/bin/spy
```

To confirm the right binary is being used after reinstalling:

```
which spy
```

This should point to `/opt/homebrew/bin/spy`.
