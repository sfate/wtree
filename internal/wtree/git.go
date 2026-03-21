package wtree

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// gitRun executes a git command in the given directory and returns trimmed stdout.
func gitRun(dir string, args ...string) (string, error) {
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

// FindGitRoot returns the root directory of the git project containing the CWD.
func FindGitRoot() (string, error) {
	out, err := gitRun(".", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return filepath.Dir(out), nil
}

// gitCurrentBranch returns the current branch name for the repo at dir.
func gitCurrentBranch(dir string) (string, error) {
	return gitRun(dir, "branch", "--show-current")
}

// gitLastCommitTime returns the time of the last commit on the given branch.
func gitLastCommitTime(dir, branch string) (time.Time, error) {
	out, err := gitRun(dir, "log", "-1", "--format=%ct", branch)
	if err != nil {
		return time.Time{}, err
	}
	ts, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse commit timestamp: %w", err)
	}
	return time.Unix(ts, 0), nil
}

// gitBaseBranch returns the default branch name (e.g. main, master) from origin/HEAD.
func gitBaseBranch(dir string) (string, error) {
	out, err := gitRun(dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot determine base branch: %w", err)
	}
	parts := strings.Split(out, "/")
	return parts[len(parts)-1], nil
}

// gitFindBranch searches local branch refs for the first branch matching pattern.
// Returns the full branch name or empty string if not found.
func gitFindBranch(dir, pattern string) (string, error) {
	out, err := gitRun(dir, "show-ref", "--heads")
	if err != nil {
		// show-ref exits 1 when no refs found
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

// gitBranchExists returns true if a local branch matching the name exists.
func gitBranchExists(dir, branch string) bool {
	out, _ := gitRun(dir, "show-ref", "--heads")
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

// gitEnsureBranch creates the branch from baseBranch if it does not exist.
func gitEnsureBranch(dir, branch, baseBranch string) error {
	if gitBranchExists(dir, branch) {
		return nil
	}
	fmt.Printf("Branch does not exist.. creating from: %s.\n", baseBranch)
	_, err := gitRun(dir, "branch", branch, baseBranch)
	return err
}

// gitWorktreeRegistered checks whether worktreeDir is already registered as a worktree.
func gitWorktreeRegistered(projectDir, worktreeDir string) (bool, error) {
	out, err := gitRun(projectDir, "worktree", "list")
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

// gitWorktreeAdd adds a new worktree at worktreeDir for the given branch.
func gitWorktreeAdd(projectDir, worktreeDir, branch string) error {
	_, err := gitRun(projectDir, "worktree", "add", "--force", worktreeDir, branch)
	return err
}

// gitWorktreeRemove removes the worktree at worktreeDir.
func gitWorktreeRemove(projectDir, worktreeDir string) error {
	_, err := gitRun(projectDir, "worktree", "remove", worktreeDir, "--force")
	return err
}
