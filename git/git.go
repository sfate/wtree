package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Client interface {
	CurrentBranch(dir string) (string, error)
	LastCommitTime(dir, branch string) (time.Time, error)
	BaseBranch(dir string) (string, error)
	FindBranch(dir, branch string) (string, error)
	BranchExists(dir, branch string) (bool, error)
	RemoteBranchExists(dir, remote, branch string) (bool, error)
	EnsureBranch(dir, branch, baseBranch string) error
	WorktreeRegistered(projectDir, worktreeDir string) (bool, error)
	WorktreeAdd(projectDir, worktreeDir, branch string) error
	WorktreeRemove(projectDir, worktreeDir string) error
}

type CLI struct{}

func NewClient() Client {
	return CLI{}
}

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
func (CLI) CurrentBranch(dir string) (string, error) {
	return Run(dir, "branch", "--show-current")
}

// LastCommitTime returns the time of the last commit on the given branch.
func (CLI) LastCommitTime(dir, branch string) (time.Time, error) {
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
func (CLI) BaseBranch(dir string) (string, error) {
	out, err := Run(dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot determine base branch: %w", err)
	}
	parts := strings.Split(out, "/")
	return parts[len(parts)-1], nil
}

// FindBranch returns the full branch name when an exact local branch exists.
func (c CLI) FindBranch(dir, branch string) (string, error) {
	exists, err := c.BranchExists(dir, branch)
	if err != nil {
		return "", err
	}
	if exists {
		return branch, nil
	}
	return "", nil
}

// BranchExists returns true if a local branch with the exact name exists.
func (CLI) BranchExists(dir, branch string) (bool, error) {
	if branch == "" {
		return false, nil
	}
	return refExists(dir, "refs/heads/"+branch)
}

// RemoteBranchExists returns true if a remote branch with the exact name exists.
func (CLI) RemoteBranchExists(dir, remote, branch string) (bool, error) {
	if remote == "" || branch == "" {
		return false, nil
	}
	return refExists(dir, "refs/remotes/"+remote+"/"+branch)
}

func refExists(dir, ref string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)
	cmd.Dir = dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// EnsureBranch creates the branch from baseBranch if it does not exist.
func (c CLI) EnsureBranch(dir, branch, baseBranch string) error {
	exists, err := c.BranchExists(dir, branch)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = Run(dir, "branch", branch, baseBranch)
	return err
}

// WorktreeRegistered checks whether worktreeDir is already registered as a worktree.
func (CLI) WorktreeRegistered(projectDir, worktreeDir string) (bool, error) {
	out, err := Run(projectDir, "worktree", "list", "--porcelain")
	if err != nil {
		return false, err
	}
	target, err := normalizePath(worktreeDir)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		path, err := normalizePath(strings.TrimPrefix(line, "worktree "))
		if err != nil {
			return false, err
		}
		if path == target {
			return true, nil
		}
	}
	return false, nil
}

// WorktreeAdd adds a new worktree at worktreeDir for the given branch.
func (CLI) WorktreeAdd(projectDir, worktreeDir, branch string) error {
	_, err := Run(projectDir, "worktree", "add", "--force", worktreeDir, branch)
	return err
}

// WorktreeRemove removes the worktree at worktreeDir.
func (CLI) WorktreeRemove(projectDir, worktreeDir string) error {
	_, err := Run(projectDir, "worktree", "remove", worktreeDir, "--force")
	return err
}

func normalizePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if os.IsNotExist(err) {
		return filepath.Clean(abs), nil
	}
	return "", err
}
