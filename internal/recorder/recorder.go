package recorder

import (
	"os/exec"
	"syscall"
)

// Recorder captures audio from the default input device using sox.
type Recorder struct {
	audioFile string
	proc      *exec.Cmd
}

func New(audioFile string) *Recorder {
	return &Recorder{audioFile: audioFile}
}

// Start begins recording audio to the configured file.
func (r *Recorder) Start() error {
	r.proc = exec.Command("sox", "-d", r.audioFile)
	r.proc.Stdout = nil
	r.proc.Stderr = nil
	return r.proc.Start()
}

// Stop sends SIGTERM to sox so it can finalize the WAV header before exiting,
// then waits for the process to exit cleanly.
func (r *Recorder) Stop() error {
	if r.proc == nil {
		return nil
	}
	// SIGTERM (not SIGKILL) lets sox flush and write the correct WAV file length.
	_ = r.proc.Process.Signal(syscall.SIGTERM)
	_ = r.proc.Wait()
	r.proc = nil
	return nil
}
