package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/sfate/wtree/config"
	wtree "github.com/sfate/wtree/wtree"
)

type rootArgs struct {
	ShellInit  string `arg:"--shell-init" placeholder:"SHELL" help:"Print shell integration for a shell (zsh, bash)"`
	List       bool   `arg:"--list" help:"List worktrees for the current project"`
	DeleteRef  string `arg:"--delete" placeholder:"REF" help:"Delete a worktree by ref"`
	Clean      bool   `arg:"--clean,-c" help:"Remove all worktrees for the current project"`
	CleanStale bool   `arg:"--clean-stale" help:"Remove worktrees inactive for 2+ weeks"`
	Root       bool   `arg:"--root" help:"Print the current project root"`
	Version    bool   `arg:"--version,-v" help:"Print wtree version"`

	Ref        string `arg:"positional" placeholder:"REF"`
	Branch     string `arg:"positional" placeholder:"BRANCH"`
	BaseBranch string `arg:"positional" placeholder:"BASE_BRANCH"`
}

func (a *rootArgs) Description() string {
	return "A CLI for managing git worktrees organised by ticket reference or arbitrary name."
}

func (a *rootArgs) Epilog() string {
	return "Examples:\n  wtree ABC-1234\n  wtree feature-x my-branch develop\n  wtree --list\n  wtree --delete ABC-1234"
}

type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return ""
}

type Options struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
}

type App struct {
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	version        string
	managerFactory func() (*wtree.Manager, error)
	hookRunner     func(script string, args ...string) error
}

type Command struct {
	app  *App
	args []string
}

func NewRootCmd() *Command {
	return NewRootCmdWithOptions(Options{})
}

func NewRootCmdWithOptions(opts Options) *Command {
	app := NewApp(opts)
	return app.NewRootCmd()
}

func NewApp(opts Options) *App {
	stdin := opts.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	app := &App{
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		version: opts.Version,
	}
	app.managerFactory = app.newManager
	app.hookRunner = app.runHook
	return app
}

func (a *App) NewRootCmd() *Command {
	return &Command{app: a}
}

func (c *Command) SetArgs(args []string) {
	c.args = make([]string, len(args))
	copy(c.args, args)
}

func (c *Command) Execute() error {
	argv := c.args
	if argv == nil {
		argv = os.Args[1:]
	}

	args := &rootArgs{}
	parser, err := arg.NewParser(arg.Config{
		Program: "wtree",
		Out:     c.app.stdout,
	}, args)
	if err != nil {
		return err
	}

	if len(argv) == 0 {
		parser.WriteHelp(c.app.stdout)
		return ExitError{Code: 1}
	}

	err = parser.Parse(argv)
	switch {
	case err == nil:
		return c.app.runParsedArgs(args)
	case errors.Is(err, arg.ErrHelp):
		parser.WriteHelp(c.app.stdout)
		return nil
	case errors.Is(err, arg.ErrVersion):
		// Keep version output off stdout so stale shell wrappers do not treat it
		// as a navigation target and attempt to cd into it.
		_, _ = fmt.Fprintln(c.app.stderr, c.app.version)
		return nil
	default:
		return err
	}
}

func (a *App) runParsedArgs(args *rootArgs) error {
	modes := 0
	if args.ShellInit != "" {
		modes++
	}
	if args.List {
		modes++
	}
	if args.DeleteRef != "" {
		modes++
	}
	if args.Clean {
		modes++
	}
	if args.CleanStale {
		modes++
	}
	if args.Root {
		modes++
	}
	if args.Version {
		modes++
	}
	if modes > 1 {
		return errors.New("options --shell-init, --list, --delete, --clean/--clear, and --clean-stale/--root are mutually exclusive")
	}

	hasPositionals := args.Ref != "" || args.Branch != "" || args.BaseBranch != ""

	switch {
	case args.ShellInit != "":
		if hasPositionals {
			return errors.New("--shell-init does not accept positional arguments")
		}
		return a.runShellInit(args.ShellInit)
	case args.List:
		if hasPositionals {
			return errors.New("--list does not accept positional arguments")
		}
		return a.runList()
	case args.DeleteRef != "":
		if hasPositionals {
			return errors.New("--delete does not accept positional arguments")
		}
		return a.runDelete(args.DeleteRef)
	case args.Clean:
		if hasPositionals {
			return errors.New("--clean/--clear does not accept positional arguments")
		}
		return a.runClean()
	case args.CleanStale:
		if hasPositionals {
			return errors.New("--clean-stale does not accept positional arguments")
		}
		return a.runCleanStale()
	case args.Root:
		if hasPositionals {
			return errors.New("--root does not accept positional arguments")
		}
		return a.runRoot()
	case args.Version:
		if hasPositionals {
			return errors.New("--version does not accept positional arguments")
		}
		return a.runVersion()
	default:
		return a.runCreate(args.Ref, args.Branch, args.BaseBranch)
	}
}

func (a *App) newManager() (*wtree.Manager, error) {
	projectDir, err := wtree.FindGitRoot()
	if err != nil {
		return nil, err
	}

	fileCfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	projectName := filepath.Base(projectDir)
	projCfg, created := fileCfg.FindOrCreate(projectName, projectDir)
	if created {
		if err := fileCfg.Save(); err != nil {
			_, _ = fmt.Fprintf(a.stderr, "Warning: could not save config: %v\n", err)
		}
		return nil, fmt.Errorf("added project %q to config — please fill in the required fields and re-run", projCfg.Name)
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
			return a.hookRunner(script, ref, projectName, worktreeDir)
		}
	}
	if script := projCfg.Hooks.ExpandedPostDelete(); script != "" {
		cfg.PostDelete = func(ref string) error {
			return a.hookRunner(script, ref)
		}
	}

	return wtree.NewAt(cfg, projectDir), nil
}

func (a *App) runList() error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	entries, err := m.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		_, _ = fmt.Fprintf(a.stdout, "No worktrees found for project: %s\n", m.ProjectName())
		return nil
	}
	_, _ = fmt.Fprintf(a.stdout, "Worktrees for project: %s\n", m.ProjectName())
	wtree.PrintTableTo(a.stdout, entries)
	return nil
}

func (a *App) runCreate(ref, branch, baseBranch string) error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	dir, _, err := m.Create(ref, branch, baseBranch)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(a.stdout, dir)
	return nil
}

func (a *App) runDelete(ref string) error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	return m.Delete(ref)
}

func (a *App) runClean() error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	return m.Clean()
}

func (a *App) runCleanStale() error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	stale, err := m.CleanStale(14 * 24 * time.Hour)
	if err != nil {
		return err
	}
	if len(stale) == 0 {
		_, _ = fmt.Fprintf(a.stdout, "No stale worktrees found for project: %s\n", m.ProjectName())
		return nil
	}

	_, _ = fmt.Fprintf(a.stdout, "Stale worktrees for project: %s (no activity in 2+ weeks)\n", m.ProjectName())
	wtree.PrintTableTo(a.stdout, stale)

	_, _ = fmt.Fprintf(a.stdout, "\nRemove %d worktree(s)? [y/N] ", len(stale))
	reader := bufio.NewReader(a.stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" {
		_, _ = fmt.Fprintln(a.stdout, "Aborted.")
		return nil
	}

	if err := m.DeleteEntries(stale); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(a.stdout, "Done.")
	return nil
}

func (a *App) runRoot() error {
	m, err := a.managerFactory()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(a.stdout, m.ProjectDir())
	return nil
}

func (a *App) runShellInit(shell string) error {
	switch shell {
	case "zsh":
		_, _ = io.WriteString(a.stdout, shellInitZsh)
		return nil
	case "bash":
		_, _ = io.WriteString(a.stdout, shellInitBash)
		return nil
	default:
		return fmt.Errorf("unsupported shell %q — supported: zsh, bash", shell)
	}
}

const shellInitZsh = `
function wtree() {
  case "$1" in
    --shell-init|-h|--help|--list|--delete|--clean|--clear|--clean-stale|--version)
      command wtree "$@"
      ;;
    *)
      local output exit_code dir shell_code
      output=$(command wtree "$@")
      exit_code=$?
      dir=$(printf '%s' "$output" | head -1)
      shell_code=$(printf '%s' "$output" | tail -n +2)
      if [[ -n "$dir" ]]; then
        cd "$dir"
        [[ -n "$shell_code" ]] && eval "$shell_code"
      fi
      return $exit_code
      ;;
  esac
}
`

const shellInitBash = `
function wtree() {
  case "$1" in
    --shell-init|-h|--help|--list|--delete|--clean|--clear|--clean-stale|--version)
      command wtree "$@"
      ;;
    *)
      local output exit_code dir shell_code
      output=$(command wtree "$@")
      exit_code=$?
      dir=$(printf '%s' "$output" | head -1)
      shell_code=$(printf '%s' "$output" | tail -n +2)
      if [[ -n "$dir" ]]; then
        cd "$dir"
        [[ -n "$shell_code" ]] && eval "$shell_code"
      fi
      return $exit_code
      ;;
  esac
}
`

func (a *App) runHook(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	cmd.Stdin = a.stdin
	// Stdout is reserved for navigation protocol output such as the target path.
	// Hooks should not be able to corrupt that contract.
	cmd.Stdout = a.stderr
	cmd.Stderr = a.stderr
	return cmd.Run()
}

func (a *App) runVersion() error {
	_, _ = fmt.Fprintln(a.stdout, a.version)
	return nil
}
