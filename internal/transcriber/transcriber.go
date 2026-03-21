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

// Transcribe runs whisper-cli on audioFile using modelPath and returns
// the cleaned transcript text.
func Transcribe(modelPath, audioFile string) (string, error) {
	out, err := exec.Command("whisper-cli", "-m", modelPath, "-f", audioFile).Output()
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
