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
		{"/models/ggml-small.bin", "/tmp/spy.wav", "fr", []string{"-m", "/models/ggml-small.bin", "-f", "/tmp/spy.wav", "-l", "fr"}},
	}
	for _, tt := range tests {
		got := buildWhisperArgs(tt.modelPath, tt.audioFile, tt.language)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("buildWhisperArgs(%q, %q, %q) = %v, want %v", tt.modelPath, tt.audioFile, tt.language, got, tt.want)
		}
	}
}
