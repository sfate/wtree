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
)

// Config controls the behaviour of a Manager.
type Config struct {
	// BaseDir is the root directory under which per-project worktree directories
	// are created (default: ~/.worktrees).
	BaseDir string

	// EnvFile is the name of the per-worktree environment file (default: .wtree_env).
	EnvFile string

	// TicketPrefix is the prefix used to recognise ticket references and
	// automatically derive branch names (default: "ABC-").
	TicketPrefix string

	// BranchPrefix is prepended to the numeric part of a ticket reference to
	// form the branch name (default: "ob-abc-").
	BranchPrefix string

	// PostNavigation is an optional hook called after a worktree is created or
	// switched to. It receives the ref, the project name, and the worktree
	// directory path.
	PostNavigation func(ref, projectName, worktreeDir string) error

	// PostDelete is an optional hook called after a worktree is removed.
	PostDelete func(ref string) error
}

// DefaultConfig returns a Config with sensible defaults and no hooks.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		BaseDir:      filepath.Join(home, ".worktrees"),
		EnvFile:      ".wtree_env",
		TicketPrefix: "ABC-",
		BranchPrefix: "ob-abc-",
	}
}

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
	cfg        Config
	projectDir string
}

// New creates a Manager rooted at the git project that contains the current
// working directory. It returns an error if the CWD is not inside a git repo.
func New(cfg Config) (*Manager, error) {
	dir, err := FindGitRoot()
	if err != nil {
		return nil, err
	}
	return &Manager{cfg: cfg, projectDir: dir}, nil
}

// NewAt creates a Manager with an explicit project directory, skipping git
// root detection.
func NewAt(cfg Config, projectDir string) *Manager {
	return &Manager{cfg: cfg, projectDir: projectDir}
}

// ProjectDir returns the absolute path of the git project root.
func (m *Manager) ProjectDir() string { return m.projectDir }

// ProjectName returns the base name of the project directory.
func (m *Manager) ProjectName() string { return filepath.Base(m.projectDir) }

func (m *Manager) worktreeProjectDir() string {
	return filepath.Join(m.cfg.BaseDir, m.ProjectName())
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

		branch, err := gitCurrentBranch(wtDir)
		if err != nil || branch == "" {
			continue
		}

		commitTime, _ := gitLastCommitTime(wtDir, branch)

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
	registered, err := gitWorktreeRegistered(m.projectDir, wtDir)
	if err != nil {
		return "", false, err
	}
	if registered {
		fmt.Println("Worktree already exists.")
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
		baseBranch, err = gitBaseBranch(m.projectDir)
		if err != nil {
			return "", false, err
		}
	}

	fmt.Printf("Using branch name: %s\n", branch)
	if err := gitEnsureBranch(m.projectDir, branch, baseBranch); err != nil {
		return "", false, err
	}

	if err := gitWorktreeAdd(m.projectDir, wtDir, branch); err != nil {
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

	fmt.Printf("Removing worktree: %s\n", ref)
	if err := gitWorktreeRemove(m.projectDir, wtDir); err != nil {
		return err
	}
	fmt.Println("Worktree removed.")

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
			fmt.Printf("No worktrees to clean for project: %s\n", m.ProjectName())
			return nil
		}
		return err
	}

	fmt.Printf("Cleaning all worktrees for project: %s\n", m.ProjectName())

	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}
		ref := de.Name()
		wtDir := filepath.Join(wpd, ref)
		fmt.Printf("Removing: %s\n", ref)
		_ = gitWorktreeRemove(m.projectDir, wtDir)

		if m.cfg.PostDelete != nil {
			_ = m.cfg.PostDelete(ref)
		}
	}

	// Remove empty project worktree directory.
	m.removeEmptyDir(wpd)
	fmt.Println("All worktrees cleaned.")
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
		fmt.Printf("Removing: %s\n", e.Ref)
		wtDir := m.worktreeDir(e.Ref)
		_ = gitWorktreeRemove(m.projectDir, wtDir)

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

	found, err := gitFindBranch(m.projectDir, branchName)
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
