package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpShowsFlagCommands(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	help := out.String()
	for _, want := range []string{
		"--shell-init",
		"--list",
		"--delete",
		"--clean",
		"--clean-stale",
		"--root",
		"--version",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help output missing %q:\n%s", want, help)
		}
	}

	for _, notWant := range []string{"Available Commands:", "Use \"wtree [command] --help\""} {
		if strings.Contains(help, notWant) {
			t.Fatalf("help output unexpectedly contains %q:\n%s", notWant, help)
		}
	}
}

func TestNoArgsReturnsExitCodeOne(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil || !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help output and silent exit error, got err=%v output=%q", err, out.String())
	}
	if got, ok := err.(ExitError); !ok || got.Code != 1 {
		t.Fatalf("expected ExitError with code 1, got %T (%v)", err, err)
	}
}

func TestShellInitFlagPrintsWrapper(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--shell-init", "zsh"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(out.String(), "function wtree()") {
		t.Fatalf("shell-init output missing wrapper function:\n%s", out.String())
	}
}

func TestVersionFlagPrintsInjectedVersion(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "test-version" {
		t.Fatalf("version output = %q, want %q", got, "test-version")
	}
}

func TestDeleteFlagRequiresNoExtraPositionalArgs(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"--delete", "ABC-1234", "extra"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for extra positional args")
	}
	if !strings.Contains(err.Error(), "--delete does not accept positional arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersionFlagRejectsPositionalArgs(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"--version", "extra"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for extra positional args")
	}
	if !strings.Contains(err.Error(), "--version does not accept positional arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteFlagPassesFlagValue(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"--delete", "LM-9999"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected delete to fail outside configured context")
	}
	if strings.Contains(err.Error(), "ref is required for delete") {
		t.Fatalf("delete flag value was not passed through: %v", err)
	}
}

func TestSubcommandLikeListIsTreatedAsCreateArg(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"list", "extra", "value", "overflow"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for too many positional args")
	}
	if !strings.Contains(err.Error(), "too many positional arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newTestCommand() (*Command, *bytes.Buffer) {
	var out bytes.Buffer
	cmd := NewRootCmdWithOptions(Options{
		Stdin:   strings.NewReader(""),
		Stdout:  &out,
		Stderr:  &out,
		Version: "test-version",
	})
	return cmd, &out
}
