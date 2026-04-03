package wtree

import (
	"fmt"
	"io"
)

// Reporter receives user-visible status events from Manager operations.
type Reporter interface {
	WorktreeExists()
	UsingBranchName(branch string)
	RemovingWorktree(ref string)
	WorktreeRemoved()
	NoWorktreesToClean(project string)
	CleaningAllWorktrees(project string)
	Removing(ref string)
	AllWorktreesCleaned()
}

// ManagerDeps holds optional runtime dependencies for Manager.
type ManagerDeps struct {
	Reporter Reporter
}

type noopReporter struct{}

func (noopReporter) WorktreeExists()             {}
func (noopReporter) UsingBranchName(string)      {}
func (noopReporter) RemovingWorktree(string)     {}
func (noopReporter) WorktreeRemoved()            {}
func (noopReporter) NoWorktreesToClean(string)   {}
func (noopReporter) CleaningAllWorktrees(string) {}
func (noopReporter) Removing(string)             {}
func (noopReporter) AllWorktreesCleaned()        {}

// CLIReporter writes status messages to stdout/stderr for the CLI.
type CLIReporter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewCLIReporter returns a Reporter that preserves the existing CLI output split.
func NewCLIReporter(stdout, stderr io.Writer) Reporter {
	return &CLIReporter{stdout: stdout, stderr: stderr}
}

func (r *CLIReporter) WorktreeExists() {
	_, _ = fmt.Fprintln(r.stderr, "Worktree already exists.")
}

func (r *CLIReporter) UsingBranchName(branch string) {
	_, _ = fmt.Fprintf(r.stderr, "Using branch name: %s\n", branch)
}

func (r *CLIReporter) RemovingWorktree(ref string) {
	_, _ = fmt.Fprintf(r.stdout, "Removing worktree: %s\n", ref)
}

func (r *CLIReporter) WorktreeRemoved() {
	_, _ = fmt.Fprintln(r.stdout, "Worktree removed.")
}

func (r *CLIReporter) NoWorktreesToClean(project string) {
	_, _ = fmt.Fprintf(r.stdout, "No worktrees to clean for project: %s\n", project)
}

func (r *CLIReporter) CleaningAllWorktrees(project string) {
	_, _ = fmt.Fprintf(r.stdout, "Cleaning all worktrees for project: %s\n", project)
}

func (r *CLIReporter) Removing(ref string) {
	_, _ = fmt.Fprintf(r.stdout, "Removing: %s\n", ref)
}

func (r *CLIReporter) AllWorktreesCleaned() {
	_, _ = fmt.Fprintln(r.stdout, "All worktrees cleaned.")
}
