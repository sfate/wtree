package handlers

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sfate/wtree/config"
	gitpkg "github.com/sfate/wtree/git"
	"github.com/sfate/wtree/handlers/operations"
	wtreepkg "github.com/sfate/wtree/wtree"
)

type Options struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
}

type Set struct {
	Create     func(ref, branch, baseBranch string) error
	List       func() error
	Delete     func(ref string) error
	Clean      func() error
	CleanStale func() error
	Root       func() error
	ShellInit  func(shell string) error
	Version    func() error
}

type HandlerType string

const (
	HandlerTypeCreate     HandlerType = "create"
	HandlerTypeList       HandlerType = "list"
	HandlerTypeDelete     HandlerType = "delete"
	HandlerTypeClean      HandlerType = "clean"
	HandlerTypeCleanStale HandlerType = "clean_stale"
	HandlerTypeRoot       HandlerType = "root"
	HandlerTypeShellInit  HandlerType = "shell_init"
	HandlerTypeVersion    HandlerType = "version"
)

var operationFactories = map[HandlerType]operations.OperationServiceFactory{
	HandlerTypeCreate:     operations.NewCreateService,
	HandlerTypeList:       operations.NewListService,
	HandlerTypeDelete:     operations.NewDeleteService,
	HandlerTypeClean:      operations.NewCleanService,
	HandlerTypeCleanStale: operations.NewCleanStaleService,
	HandlerTypeRoot:       operations.NewRootService,
	HandlerTypeShellInit:  operations.NewShellInitService,
	HandlerTypeVersion:    operations.NewVersionService,
}

type Handler struct {
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	version string

	options  operations.ServiceOptions
	handlers map[HandlerType]operations.OperationService
}

type HandlerArgs struct {
	Ref        string
	Branch     string
	BaseBranch string
	Shell      string
}

func NewHandler(opts Options) *Handler {
	options := operations.ServiceOptions{
		Stdin:   opts.Stdin,
		Stdout:  opts.Stdout,
		Stderr:  opts.Stderr,
		Version: opts.Version,
		ManagerFactory: func() (*wtreepkg.Manager, error) {
			return nil, fmt.Errorf("manager factory not set")
		},
		HookRunner: func(script string, args ...string) error {
			return fmt.Errorf("hook runner not set")
		},
	}

	operationHandlers := make(map[HandlerType]operations.OperationService)
	for kind, service := range operationFactories {
		operationHandlers[kind] = service()
	}

	return &Handler{
		options:  options,
		handlers: operationHandlers,
	}
}

func (h *Handler) Call(kind HandlerType, args operations.ServiceArgs) error {
	if service, ok := h.handlers[kind]; ok {
		return service.Process(h.options, args)
	}

	return fmt.Errorf("unsupported handler type: %q", kind)
}

func (h *Handler) contextWithDefaults(ctx operations.Context) operations.Context {
	ctx.ManagerFactory = h.newManager
	ctx.HookRunner = h.runHook
	return ctx
}

func (h *Handler) newManager() (*wtreepkg.Manager, error) {
	projectDir, err := gitpkg.FindRoot()
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
			_, _ = fmt.Fprintf(h.stderr, "Warning: could not save config: %v\n", err)
		}
		return nil, fmt.Errorf("added project %q to config — please fill in the required fields and re-run", projCfg.Name)
	}
	cfg := wtreepkg.DefaultConfig()
	if fileCfg.BaseDir != "" {
		cfg.BaseDir = fileCfg.ExpandedBaseDir()
	}
	if projCfg.TicketPrefix != "" {
		cfg.TicketPrefix = projCfg.TicketPrefix
		cfg.BranchPrefix = projCfg.BranchPrefix + strings.ToLower(projCfg.TicketPrefix)
	}

	if script := projCfg.Hooks.ExpandedPostNavigation(); script != "" {
		cfg.PostNavigation = func(ref, projectName, worktreeDir string) error {
			return h.runHook(script, ref, projectName, worktreeDir)
		}
	}
	if script := projCfg.Hooks.ExpandedPostDelete(); script != "" {
		cfg.PostDelete = func(ref string) error {
			return h.runHook(script, ref)
		}
	}

	return wtreepkg.NewAt(cfg, projectDir), nil
}

func (h *Handler) runHook(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	cmd.Stdin = h.stdin
	cmd.Stdout = h.stderr
	cmd.Stderr = h.stderr
	return cmd.Run()
}
