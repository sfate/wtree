package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpShowsFlagCommands(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--help"})

	require.NoError(t, cmd.Execute())

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
		require.Contains(t, help, want)
	}

	for _, notWant := range []string{"Available Commands:", "Use \"wtree [command] --help\""} {
		require.NotContains(t, help, notWant)
	}
}

func TestNoArgsReturnsExitCodeOne(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	got, ok := err.(ExitError)
	require.True(t, ok)
	require.Equal(t, 1, got.Code)
	require.Contains(t, out.String(), "Usage:")
}

func TestShellInitFlagPrintsWrapper(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--shell-init", "zsh"})

	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "function wtree()")
}

func TestVersionFlagPrintsInjectedVersion(t *testing.T) {
	cmd, out := newTestCommand()
	cmd.SetArgs([]string{"--version"})

	require.NoError(t, cmd.Execute())
	require.Equal(t, "test-version", strings.TrimSpace(out.String()))
}

func TestVersionFlagAllowsExtraArgs(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"--version", "extra"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestDeleteFlagPassesFlagValue(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"--delete", "LM-9999"})

	err := cmd.Execute()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "ref is required for delete")
}

func TestSubcommandLikeListIsTreatedAsCreateArg(t *testing.T) {
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"list", "extra", "value", "overflow"})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "too many positional arguments")
}

func newTestCommand() (*Command, *bytes.Buffer) {
	var out bytes.Buffer
	cmd := NewRootCmd(Options{
		Stdin:   strings.NewReader(""),
		Stdout:  &out,
		Stderr:  &out,
		Version: "test-version",
	})
	return cmd, &out
}
