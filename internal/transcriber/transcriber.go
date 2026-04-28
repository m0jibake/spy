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

// buildWhisperArgs returns the whisper-cli argument slice for the given model, audio file, and BCP-47 language tag.
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
