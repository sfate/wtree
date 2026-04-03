package wtree

import (
	"fmt"
	"io"
	"os"
)

// UI carries stdin/stdout/stderr for CLI-facing interactions.
type UI struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// NewUI creates a UI with defaults for nil streams.
func NewUI(stdin io.Reader, stdout, stderr io.Writer) *UI {
	if stdin == nil {
		stdin = os.Stdin
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	return &UI{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

func (u *UI) In() io.Reader {
	return u.stdin
}

func (u *UI) Out() io.Writer {
	return u.stdout
}

func (u *UI) Err() io.Writer {
	return u.stderr
}

func (u *UI) Infof(format string, args ...any) {
	_, _ = fmt.Fprintf(u.stdout, format, args...)
}

func (u *UI) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(u.stderr, format, args...)
}

// ManagerDeps holds optional runtime dependencies for Manager.
type ManagerDeps struct {
	UI *UI
}
