package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfate/wtree/internal/config"
	wtree "github.com/sfate/wtree/internal/wtree"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	switch args[0] {
	case "--help", "-h":
		printHelp()
	case "--list":
		runList()
	case "--delete":
		if len(args) < 2 {
			fatal("--delete requires a <ref> argument")
		}
		runDelete(args[1])
	case "--clean", "--clear":
		runClean()
	case "--clean-stale":
		runCleanStale()
	case "--root":
		runRoot()
	default:
		// create mode: wtree <ref> [branch] [base_branch]
		ref := args[0]
		var branch, baseBranch string
		if len(args) > 1 {
			branch = args[1]
		}
		if len(args) > 2 {
			baseBranch = args[2]
		}
		runCreate(ref, branch, baseBranch)
	}
}

func newManager() *wtree.Manager {
	projectDir, err := wtree.FindGitRoot()
	if err != nil {
		fatal(err.Error())
	}

	fileCfg, err := config.Load()
	if err != nil {
		fatal(fmt.Sprintf("config: %v", err))
	}

	projectName := filepath.Base(projectDir)
	projCfg, created := fileCfg.FindOrCreate(projectName, projectDir)
	if created {
		if err := fileCfg.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
		}
		fatal(fmt.Sprintf("added project %q to config — please fill in the required fields and re-run", projCfg.Name))
	}
	cfg := wtree.DefaultConfig()
	if fileCfg.BaseDir != "" {
		cfg.BaseDir = fileCfg.ExpandedBaseDir()
	}
	if projCfg.TicketPrefix != "" {
		cfg.TicketPrefix = projCfg.TicketPrefix
		cfg.BranchPrefix = projCfg.BranchPrefix + strings.ToLower(projCfg.TicketPrefix)
	}

	if script := projCfg.Hooks.ExpandedPostNavigation(); script != "" {
		cfg.PostNavigation = func(ref, projectName, worktreeDir string) error {
			return runHook(script, ref, projectName, worktreeDir)
		}
	}
	if script := projCfg.Hooks.ExpandedPostDelete(); script != "" {
		cfg.PostDelete = func(ref string) error {
			return runHook(script, ref)
		}
	}

	return wtree.NewAt(cfg, projectDir)
}

func runList() {
	m := newManager()
	entries, err := m.List()
	if err != nil {
		fatal(err.Error())
	}
	if len(entries) == 0 {
		fmt.Printf("No worktrees found for project: %s\n", m.ProjectName())
		return
	}
	fmt.Printf("Worktrees for project: %s\n", m.ProjectName())
	wtree.PrintTable(entries)
}

func runCreate(ref, branch, baseBranch string) {
	m := newManager()
	dir, _, err := m.Create(ref, branch, baseBranch)
	if err != nil {
		fatal(err.Error())
	}
	navigate(dir)
}

func runDelete(ref string) {
	m := newManager()
	if err := m.Delete(ref); err != nil {
		fatal(err.Error())
	}
}

func runClean() {
	m := newManager()
	if err := m.Clean(); err != nil {
		fatal(err.Error())
	}
}

func runCleanStale() {
	m := newManager()
	stale, err := m.CleanStale(14 * 24 * time.Hour)
	if err != nil {
		fatal(err.Error())
	}
	if len(stale) == 0 {
		fmt.Printf("No stale worktrees found for project: %s\n", m.ProjectName())
		return
	}

	fmt.Printf("Stale worktrees for project: %s (no activity in 2+ weeks)\n", m.ProjectName())
	wtree.PrintTable(stale)

	fmt.Printf("\nRemove %d worktree(s)? [y/N] ", len(stale))
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("Aborted.")
		return
	}

	if err := m.DeleteEntries(stale); err != nil {
		fatal(err.Error())
	}
	fmt.Println("Done.")
}

func runRoot() {
	m := newManager()
	navigate(m.ProjectDir())
}

func runHook(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func navigate(dir string) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	fmt.Printf("Switching to: %s\n", dir)
	cmd := exec.Command(shell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal(err.Error())
	}
}

func printHelp() {
	help := `Git Worktree Helper

Usage:
  wtree <ref> [branch] [base_branch]   Create/switch to worktree
  wtree --list                         List worktrees for project
  wtree --delete <ref>                 Delete worktree by <ref>
  wtree --clean/--clear                Remove all project worktrees
  wtree --clean-stale                  Remove worktrees inactive for 2+ weeks
  wtree --root                         Navigate to project root
  wtree --help/-h                      Show this help

Examples:
  wtree ABC-1234                        Auto-detect branch for ticket
  wtree feature-x my-branch            Create worktree with custom branch
  wtree --list                         Show all worktrees
  wtree --delete ABC-1234               Remove specific worktree`
	fmt.Println(help)
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
