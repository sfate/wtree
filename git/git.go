package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Run executes a git command in the given directory and returns trimmed stdout.
func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// FindRoot returns the root directory of the git project containing the CWD.
func FindRoot() (string, error) {
	out, err := Run(".", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return filepath.Dir(out), nil
}

// CurrentBranch returns the current branch name for the repo at dir.
func CurrentBranch(dir string) (string, error) {
	return Run(dir, "branch", "--show-current")
}

// LastCommitTime returns the time of the last commit on the given branch.
func LastCommitTime(dir, branch string) (time.Time, error) {
	out, err := Run(dir, "log", "-1", "--format=%ct", branch)
	if err != nil {
		return time.Time{}, err
	}
	ts, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse commit timestamp: %w", err)
	}
	return time.Unix(ts, 0), nil
}

// BaseBranch returns the default branch name (e.g. main, master) from origin/HEAD.
func BaseBranch(dir string) (string, error) {
	out, err := Run(dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot determine base branch: %w", err)
	}
	parts := strings.Split(out, "/")
	return parts[len(parts)-1], nil
}

// FindBranch searches local branch refs for the first branch matching pattern.
// Returns the full branch name or empty string if not found.
func FindBranch(dir, pattern string) (string, error) {
	out, err := Run(dir, "show-ref", "--heads")
	if err != nil {
		return "", nil
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := strings.TrimPrefix(fields[1], "refs/heads/")
		if strings.Contains(ref, pattern) {
			return ref, nil
		}
	}
	return "", nil
}

// BranchExists returns true if a local branch matching the name exists.
func BranchExists(dir, branch string) bool {
	out, _ := Run(dir, "show-ref", "--heads")
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 {
			ref := strings.TrimPrefix(fields[1], "refs/heads/")
			if strings.Contains(ref, branch) {
				return true
			}
		}
	}
	return false
}

// EnsureBranch creates the branch from baseBranch if it does not exist.
func EnsureBranch(dir, branch, baseBranch string) error {
	if BranchExists(dir, branch) {
		return nil
	}
	_, err := Run(dir, "branch", branch, baseBranch)
	return err
}

// WorktreeRegistered checks whether worktreeDir is already registered as a worktree.
func WorktreeRegistered(projectDir, worktreeDir string) (bool, error) {
	out, err := Run(projectDir, "worktree", "list")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, worktreeDir) {
			return true, nil
		}
	}
	return false, nil
}

// WorktreeAdd adds a new worktree at worktreeDir for the given branch.
func WorktreeAdd(projectDir, worktreeDir, branch string) error {
	_, err := Run(projectDir, "worktree", "add", "--force", worktreeDir, branch)
	return err
}

// WorktreeRemove removes the worktree at worktreeDir.
func WorktreeRemove(projectDir, worktreeDir string) error {
	_, err := Run(projectDir, "worktree", "remove", worktreeDir, "--force")
	return err
}
