package clipboard

import (
	"bytes"
	"os/exec"
)

// Copy sends text to the macOS clipboard via pbcopy.
func Copy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}
