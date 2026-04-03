// Package wtree provides a library for managing git worktrees organised by ticket
// or arbitrary reference names. It keeps worktrees in a central directory and
// exposes lifecycle hooks so embedders can run project-specific setup/teardown.
package wtree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sfate/wtree/config"
	gitpkg "github.com/sfate/wtree/git"
)

// WorktreeEntry describes a single worktree tracked by the manager.
type WorktreeEntry struct {
	Ref          string
	Branch       string
	LastCommitAt time.Time
}

// RelativeAge returns a human-readable string such as "3 days ago".
func (e WorktreeEntry) RelativeAge() string {
	return relativeTime(e.LastCommitAt)
}

// Manager provides operations over git worktrees for a single project.
type Manager struct {
	cfg        config.ProjectConfig
	projectDir string
	reporter   Reporter
}

// New creates a Manager rooted at the git project that contains the current
// working directory. It returns an error if the CWD is not inside a git repo.
func New(cfg config.ProjectConfig) (*Manager, error) {
	dir, err := gitpkg.FindRoot()
	if err != nil {
		return nil, err
	}
	return NewManager(dir, cfg, ManagerDeps{}), nil
}

// NewAt creates a Manager with an explicit project directory, skipping git
// root detection.
func NewAt(cfg config.ProjectConfig, projectDir string) *Manager {
	return NewManager(projectDir, cfg, ManagerDeps{})
}

// NewManager creates a Manager with explicit project directory and dependencies.
func NewManager(projectDir string, cfg config.ProjectConfig, deps ManagerDeps) *Manager {
	reporter := deps.Reporter
	if reporter == nil {
		reporter = noopReporter{}
	}
	return &Manager{
		cfg:        cfg,
		projectDir: projectDir,
		reporter:   reporter,
	}
}

// ProjectDir returns the absolute path of the git project root.
func (m *Manager) ProjectDir() string { return m.projectDir }

// ProjectName returns the base name of the project directory.
func (m *Manager) ProjectName() string { return filepath.Base(m.projectDir) }

func (m *Manager) worktreeProjectDir() string {
	return m.cfg.BaseDir
}

func (m *Manager) worktreeDir(ref string) string {
	return filepath.Join(m.worktreeProjectDir(), ref)
}

// List returns all worktree entries for the current project, sorted by
// LastCommitAt descending.
func (m *Manager) List() ([]WorktreeEntry, error) {
	wpd := m.worktreeProjectDir()
	dirEntries, err := os.ReadDir(wpd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []WorktreeEntry
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		ref := de.Name()
		wtDir := filepath.Join(wpd, ref)

		branch, err := gitpkg.CurrentBranch(wtDir)
		if err != nil || branch == "" {
			continue
		}

		commitTime, _ := gitpkg.LastCommitTime(wtDir, branch)

		entries = append(entries, WorktreeEntry{
			Ref:          ref,
			Branch:       branch,
			LastCommitAt: commitTime,
		})
	}

	sortEntriesDesc(entries)
	return entries, nil
}

// Create creates (or navigates to an existing) worktree for the given ref.
//
// If branch is empty, the manager attempts to derive one from the ref using
// TicketPrefix/BranchPrefix. If baseBranch is empty, it is resolved from
// origin/HEAD.
//
// Returns the worktree directory path, whether the worktree already existed,
// and any error.
func (m *Manager) Create(ref, branch, baseBranch string) (dir string, existed bool, err error) {
	if ref == "" {
		return "", false, fmt.Errorf("ref is required")
	}

	wtDir := m.worktreeDir(ref)

	// Ensure parent directory exists.
	if err := os.MkdirAll(wtDir, 0o755); err != nil {
		return "", false, err
	}

	// Check if worktree is already registered.
	registered, err := gitpkg.WorktreeRegistered(m.projectDir, wtDir)
	if err != nil {
		return "", false, err
	}
	if registered {
		m.reporter.WorktreeExists()
		if m.cfg.PostNavigation != nil {
			if err := m.cfg.PostNavigation(ref, m.ProjectName(), wtDir); err != nil {
				return wtDir, true, err
			}
		}
		return wtDir, true, nil
	}

	// Derive branch if not supplied.
	if branch == "" {
		branch, err = m.deriveBranch(ref)
		if err != nil {
			return "", false, err
		}
	}

	// Derive base branch if not supplied.
	if baseBranch == "" {
		baseBranch, err = gitpkg.BaseBranch(m.projectDir)
		if err != nil {
			return "", false, err
		}
	}

	m.reporter.UsingBranchName(branch)
	if err := gitpkg.EnsureBranch(m.projectDir, branch, baseBranch); err != nil {
		return "", false, err
	}

	if err := gitpkg.WorktreeAdd(m.projectDir, wtDir, branch); err != nil {
		return "", false, err
	}

	if m.cfg.PostNavigation != nil {
		if err := m.cfg.PostNavigation(ref, m.ProjectName(), wtDir); err != nil {
			return wtDir, false, err
		}
	}
	return wtDir, false, nil
}

// Delete removes the worktree identified by ref.
func (m *Manager) Delete(ref string) error {
	if ref == "" {
		return fmt.Errorf("ref is required for delete")
	}

	wtDir := m.worktreeDir(ref)
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		return fmt.Errorf("worktree not found: %s", ref)
	}

	m.reporter.RemovingWorktree(ref)
	if err := gitpkg.WorktreeRemove(m.projectDir, wtDir); err != nil {
		return err
	}
	m.reporter.WorktreeRemoved()

	if m.cfg.PostDelete != nil {
		return m.cfg.PostDelete(ref)
	}
	return nil
}

// Clean removes all worktrees for the current project.
func (m *Manager) Clean() error {
	wpd := m.worktreeProjectDir()
	dirEntries, err := os.ReadDir(wpd)
	if err != nil {
		if os.IsNotExist(err) {
			m.reporter.NoWorktreesToClean(m.ProjectName())
			return nil
		}
		return err
	}

	m.reporter.CleaningAllWorktrees(m.ProjectName())

	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		ref := de.Name()
		wtDir := filepath.Join(wpd, ref)
		m.reporter.Removing(ref)
		_ = gitpkg.WorktreeRemove(m.projectDir, wtDir)

		if m.cfg.PostDelete != nil {
			_ = m.cfg.PostDelete(ref)
		}
	}

	// Remove empty project worktree directory.
	m.removeEmptyDir(wpd)
	m.reporter.AllWorktreesCleaned()
	return nil
}

// CleanStale returns worktree entries with no commit activity within the given
// age duration. Results are sorted by LastCommitAt descending.
func (m *Manager) CleanStale(age time.Duration) ([]WorktreeEntry, error) {
	entries, err := m.List()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-age)
	var stale []WorktreeEntry
	for _, e := range entries {
		if !e.LastCommitAt.IsZero() && e.LastCommitAt.Before(cutoff) {
			stale = append(stale, e)
		}
	}
	return stale, nil
}

// DeleteEntries removes all worktrees in the provided slice.
func (m *Manager) DeleteEntries(entries []WorktreeEntry) error {
	for _, e := range entries {
		m.reporter.Removing(e.Ref)
		wtDir := m.worktreeDir(e.Ref)
		_ = gitpkg.WorktreeRemove(m.projectDir, wtDir)

		if m.cfg.PostDelete != nil {
			_ = m.cfg.PostDelete(e.Ref)
		}
	}

	m.removeEmptyDir(m.worktreeProjectDir())
	return nil
}

// deriveBranch converts a ticket-style ref (e.g. "ABC-1234") into a branch name.
func (m *Manager) deriveBranch(ref string) (string, error) {
	if m.cfg.TicketPrefix == "" || !strings.HasPrefix(ref, m.cfg.TicketPrefix) {
		return "", fmt.Errorf("[branch] is required")
	}
	num := strings.TrimPrefix(ref, m.cfg.TicketPrefix)
	branchName := m.cfg.BranchPrefix + num

	found, err := gitpkg.FindBranch(m.projectDir, branchName)
	if err != nil {
		return branchName, nil
	}
	if found != "" {
		return found, nil
	}
	return branchName, nil
}

func (m *Manager) removeEmptyDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err == nil && len(entries) == 0 {
		_ = os.Remove(dir)
	}
}

func sortEntriesDesc(entries []WorktreeEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastCommitAt.After(entries[j].LastCommitAt)
	})
}
