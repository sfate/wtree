package cli

import (
	"errors"
	"io"
	"os"

	arg "github.com/alexflint/go-arg"
	"github.com/sfate/wtree/handlers"
	"github.com/sfate/wtree/handlers/operations"
	wtreepkg "github.com/sfate/wtree/wtree"
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
	ui      *wtreepkg.UI
	version string

	handler *handlers.Handler
}

type Command struct {
	app  *App
	args []string
}

func NewRootCmd(opts Options) *Command {
	app := NewApp(opts)
	return app.NewRootCmd()
}

func NewRootCmdWithOptions(opts Options) *Command {
	return NewRootCmd(opts)
}

func NewApp(opts Options) *App {
	ui := wtreepkg.NewUI(opts.Stdin, opts.Stdout, opts.Stderr)
	handler := handlers.NewHandler(handlers.Options{
		UI:      ui,
		Version: opts.Version,
	})

	return &App{
		ui:      ui,
		version: opts.Version,
		handler: handler,
	}
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
		Out:     c.app.ui.Out(),
	}, args)
	if err != nil {
		return err
	}

	if len(argv) == 0 {
		parser.WriteHelp(c.app.ui.Out())
		return ExitError{Code: 1}
	}

	err = parser.Parse(argv)
	switch {
	case err == nil:
		return c.app.runParsedArgs(args)
	case errors.Is(err, arg.ErrHelp):
		parser.WriteHelp(c.app.ui.Out())
		return nil
	case errors.Is(err, arg.ErrVersion):
		// Keep version output off stdout so stale shell wrappers do not treat it
		// as a navigation target and attempt to cd into it.
		c.app.ui.Errorf("%s\n", c.app.version)
		return nil
	default:
		return err
	}
}

func (a *App) runParsedArgs(args *rootArgs) error {
	var handlerType handlers.HandlerType
	switch {
	case args.ShellInit != "":
		handlerType = handlers.HandlerTypeShellInit
	case args.List:
		handlerType = handlers.HandlerTypeList
	case args.DeleteRef != "":
		handlerType = handlers.HandlerTypeDelete
	case args.Clean:
		handlerType = handlers.HandlerTypeClean
	case args.CleanStale:
		handlerType = handlers.HandlerTypeCleanStale
	case args.Root:
		handlerType = handlers.HandlerTypeRoot
	case args.Version:
		handlerType = handlers.HandlerTypeVersion
	default:
		handlerType = handlers.HandlerTypeCreate
	}

	hasPositionals := args.Ref != "" || args.Branch != "" || args.BaseBranch != ""
	switch handlerType {
	case handlers.HandlerTypeShellInit:
		if hasPositionals {
			return errors.New("--shell-init does not accept positional arguments")
		}
	case handlers.HandlerTypeList:
		if hasPositionals {
			return errors.New("--list does not accept positional arguments")
		}
	case handlers.HandlerTypeDelete:
		if hasPositionals {
			return errors.New("--delete does not accept positional arguments")
		}
	case handlers.HandlerTypeClean:
		if hasPositionals {
			return errors.New("--clean/--clear does not accept positional arguments")
		}
	case handlers.HandlerTypeCleanStale:
		if hasPositionals {
			return errors.New("--clean-stale does not accept positional arguments")
		}
	case handlers.HandlerTypeRoot:
		if hasPositionals {
			return errors.New("--root does not accept positional arguments")
		}
	case handlers.HandlerTypeVersion:
		if hasPositionals {
			return errors.New("--version does not accept positional arguments")
		}
	}

	serviceArgs := operations.ServiceArgs{
		Ref:        args.Ref,
		Branch:     args.Branch,
		BaseBranch: args.BaseBranch,
		Shell:      args.ShellInit,
	}
	return a.handler.Call(handlerType, serviceArgs)
}
