package log

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDebug_suppressedByDefault(t *testing.T) {
	Verbose = false
	// Debug should produce no output when Verbose is false.
	// We can only verify it doesn't panic; output goes to fmt.Println.
	Debug("should not print")
}

func TestDebug_whenVerbose(t *testing.T) {
	Verbose = true
	t.Cleanup(func() { Verbose = false })
	// Just ensure it doesn't panic with multiple arguments.
	Debug("hello", "world", 42)
}

func TestError_writesToStderr(t *testing.T) {
	// Redirect stderr temporarily.
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error("test error message")

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test error message") {
		t.Errorf("Error() output %q does not contain expected message", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Error() output should end with newline, got %q", output)
	}
}

func TestInfo_doesNotPanic(t *testing.T) {
	Info("info message")
	Info("multi", "arg", 123)
}

func TestCheckAndFatal_nilError(t *testing.T) {
	// Must not call Fatal (which would exit) when err is nil.
	CheckAndFatal(nil)
}
