package handlers

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

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

	h := &Handler{
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		version: opts.Version,
	}

	options := operations.ServiceOptions{
		Stdin:          stdin,
		Stdout:         stdout,
		Stderr:         stderr,
		Version:        opts.Version,
		ManagerFactory: h.newManager,
		HookRunner:     h.runHook,
	}

	operationHandlers := make(map[HandlerType]operations.OperationService)
	for kind, service := range operationFactories {
		operationHandlers[kind] = service()
	}

	h.options = options
	h.handlers = operationHandlers
	return h
}

func (h *Handler) Call(kind HandlerType, args operations.ServiceArgs) error {
	if service, ok := h.handlers[kind]; ok {
		return service.Process(h.options, args)
	}

	return fmt.Errorf("unsupported handler type: %q", kind)
}

func (h *Handler) newManager() (*wtreepkg.Manager, error) {
	projectDir, err := gitpkg.FindRoot()
	if err != nil {
		return nil, err
	}

	projectName := filepath.Base(projectDir)
	cfg, err := config.LoadOrCreateProjectConfig(projectName, projectDir)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	cfg.BaseDir = cfg.EffectiveBaseDir()
	if script := cfg.Hooks.ExpandedPostNavigation(); script != "" {
		cfg.PostNavigation = func(ref, projectName, worktreeDir string) error {
			return h.runHook(script, ref, projectName, worktreeDir)
		}
	}
	if script := cfg.Hooks.ExpandedPostDelete(); script != "" {
		cfg.PostDelete = func(ref string) error {
			return h.runHook(script, ref)
		}
	}

	return wtreepkg.NewManager(projectDir, cfg, wtreepkg.ManagerDeps{
		Reporter: wtreepkg.NewCLIReporter(h.stdout, h.stderr),
	}), nil
}

func (h *Handler) runHook(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	cmd.Stdin = h.stdin
	cmd.Stdout = h.stderr
	cmd.Stderr = h.stderr
	return cmd.Run()
}
