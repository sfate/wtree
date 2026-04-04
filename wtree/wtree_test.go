package wtree

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfate/wtree/config"
	"github.com/stretchr/testify/require"
)

type fakeGitClient struct {
	branches        map[string]bool
	findBranch      map[string]string
	removeErrors    map[string]error
	removedPaths    []string
	baseBranch      string
	baseBranchErr   error
	ensureBranchErr error
}

func (f *fakeGitClient) CurrentBranch(dir string) (string, error) {
	return "", nil
}

func (f *fakeGitClient) LastCommitTime(dir, branch string) (time.Time, error) {
	return time.Time{}, nil
}

func (f *fakeGitClient) BaseBranch(dir string) (string, error) {
	if f.baseBranchErr != nil {
		return "", f.baseBranchErr
	}
	if f.baseBranch != "" {
		return f.baseBranch, nil
	}
	return "main", nil
}

func (f *fakeGitClient) FindBranch(dir, branch string) (string, error) {
	if found, ok := f.findBranch[branch]; ok {
		return found, nil
	}
	return "", nil
}

func (f *fakeGitClient) BranchExists(dir, branch string) (bool, error) {
	return f.branches[branch], nil
}

func (f *fakeGitClient) EnsureBranch(dir, branch, baseBranch string) error {
	return f.ensureBranchErr
}

func (f *fakeGitClient) WorktreeRegistered(projectDir, worktreeDir string) (bool, error) {
	return false, nil
}

func (f *fakeGitClient) WorktreeAdd(projectDir, worktreeDir, branch string) error {
	return nil
}

func (f *fakeGitClient) WorktreeRemove(projectDir, worktreeDir string) error {
	f.removedPaths = append(f.removedPaths, worktreeDir)
	if err, ok := f.removeErrors[worktreeDir]; ok {
		return err
	}
	return nil
}

func TestSortEntriesDesc(t *testing.T) {
	now := time.Now()
	entries := []WorktreeEntry{
		{Ref: "old", LastCommitAt: now.Add(-3 * 24 * time.Hour)},
		{Ref: "newest", LastCommitAt: now.Add(-1 * time.Hour)},
		{Ref: "middle", LastCommitAt: now.Add(-1 * 24 * time.Hour)},
	}

	sortEntriesDesc(entries)

	require.Equal(t, "newest", entries[0].Ref)
	require.Equal(t, "middle", entries[1].Ref)
	require.Equal(t, "old", entries[2].Ref)
}

func TestWorktreeEntryRelativeAge(t *testing.T) {
	e := WorktreeEntry{
		Ref:          "test",
		LastCommitAt: time.Now().Add(-2 * time.Hour),
	}
	require.Equal(t, "2 hours ago", e.RelativeAge())
}

func TestDefaultProjectConfig(t *testing.T) {
	cfg := config.DefaultProjectConfig("myproject", "/code/myproject")

	require.Equal(t, "myproject", cfg.Name)
	require.Equal(t, "/code/myproject", cfg.Path)
	require.Equal(t, "/code/myproject/.wtree", cfg.BaseDir)
	require.Equal(t, "ABC-", cfg.TicketPrefix)
	require.Equal(t, "ob-", cfg.BranchPrefix)
}

func TestDeriveBranchErrorPath(t *testing.T) {
	projectDir := t.TempDir()
	m := NewManager(projectDir, config.DefaultProjectConfig("repo", projectDir), ManagerDeps{
		Git: &fakeGitClient{},
	})

	// Ref without TicketPrefix should return an error.
	_, err := m.deriveBranch("no-prefix-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "[branch] is required")
}

func TestDeriveBranchHappyPath(t *testing.T) {
	projectDir := t.TempDir()
	m := NewManager(projectDir, config.DefaultProjectConfig("repo", projectDir), ManagerDeps{
		Git: &fakeGitClient{},
	})

	// Use a ticket ref that is unlikely to match an existing branch.
	branch, err := m.deriveBranch("ABC-99999")
	require.NoError(t, err)
	require.Equal(t, "ob-abc-99999", branch)
}

func TestDeriveBranchUsesExactBranchMatch(t *testing.T) {
	projectDir := t.TempDir()
	m := NewManager(projectDir, config.DefaultProjectConfig("repo", projectDir), ManagerDeps{
		Git: &fakeGitClient{
			findBranch: map[string]string{
				"ob-abc-1234": "ob-abc-1234",
			},
		},
	})

	branch, err := m.deriveBranch("ABC-1234")
	require.NoError(t, err)
	require.Equal(t, "ob-abc-1234", branch)
}

func TestCleanReturnsJoinedErrors(t *testing.T) {
	projectDir := t.TempDir()
	baseDir := filepath.Join(projectDir, ".wtree")
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "ref-a"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "ref-b"), 0o755))

	var out bytes.Buffer
	gitClient := &fakeGitClient{
		removeErrors: map[string]error{
			filepath.Join(baseDir, "ref-b"): errors.New("remove failed"),
		},
	}
	m := NewManager(projectDir, config.ProjectConfig{
		Name:    "repo",
		Path:    projectDir,
		BaseDir: baseDir,
		PostDelete: func(ref string) error {
			if ref == "ref-a" {
				return errors.New("hook failed")
			}
			return nil
		},
	}, ManagerDeps{
		Git: gitClient,
		UI:  NewUI(nil, &out, &out),
	})

	err := m.Clean()
	require.Error(t, err)
	require.ErrorContains(t, err, "remove ref-b")
	require.ErrorContains(t, err, "post-delete ref-a")
}

func TestDeleteEntriesReturnsJoinedErrors(t *testing.T) {
	projectDir := t.TempDir()
	baseDir := filepath.Join(projectDir, ".wtree")
	var out bytes.Buffer
	gitClient := &fakeGitClient{
		removeErrors: map[string]error{
			filepath.Join(baseDir, "ref-b"): errors.New("remove failed"),
		},
	}
	m := NewManager(projectDir, config.ProjectConfig{
		Name:    "repo",
		Path:    projectDir,
		BaseDir: baseDir,
		PostDelete: func(ref string) error {
			if ref == "ref-a" {
				return errors.New("hook failed")
			}
			return nil
		},
	}, ManagerDeps{
		Git: gitClient,
		UI:  NewUI(nil, &out, &out),
	})

	err := m.DeleteEntries([]WorktreeEntry{
		{Ref: "ref-a"},
		{Ref: "ref-b"},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "post-delete ref-a")
	require.ErrorContains(t, err, "remove ref-b")
}

func TestCreatePropagatesBranchExistsError(t *testing.T) {
	projectDir := t.TempDir()
	baseDir := filepath.Join(projectDir, ".wtree")
	gitClient := &branchExistsErrorGit{err: errors.New("branch lookup failed")}
	m := NewManager(projectDir, config.ProjectConfig{
		Name:         "repo",
		Path:         projectDir,
		BaseDir:      baseDir,
		TicketPrefix: "ABC-",
		BranchPrefix: "ob-",
	}, ManagerDeps{
		Git: gitClient,
		UI:  NewUI(nil, &bytes.Buffer{}, &bytes.Buffer{}),
	})

	_, _, err := m.Create("ABC-1234", "", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "branch lookup failed")
}

type branchExistsErrorGit struct {
	fakeGitClient
	err error
}

func (f *branchExistsErrorGit) BranchExists(dir, branch string) (bool, error) {
	return false, f.err
}

func (f *branchExistsErrorGit) BaseBranch(dir string) (string, error) {
	return "main", nil
}

func (f *branchExistsErrorGit) WorktreeRegistered(projectDir, worktreeDir string) (bool, error) {
	return false, nil
}

func (f *branchExistsErrorGit) EnsureBranch(dir, branch, baseBranch string) error {
	return fmt.Errorf("unexpected ensure call")
}
