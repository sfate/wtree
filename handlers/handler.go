package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sfate/wtree/config"
	gitpkg "github.com/sfate/wtree/git"
	"github.com/sfate/wtree/handlers/operations"
	wtreepkg "github.com/sfate/wtree/wtree"
)

type Options struct {
	UI      *wtreepkg.UI
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
	ui      *wtreepkg.UI
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
	ui := opts.UI
	if ui == nil {
		ui = wtreepkg.NewUI(os.Stdin, os.Stdout, os.Stderr)
	}

	h := &Handler{ui: ui, version: opts.Version}

	options := operations.ServiceOptions{
		Version:        opts.Version,
		UI:             h.ui,
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
		UI: h.ui,
	}), nil
}

func (h *Handler) runHook(script string, args ...string) error {
	cmd := exec.Command(script, args...)
	cmd.Stdin = h.ui.In()
	cmd.Stdout = h.ui.Err()
	cmd.Stderr = h.ui.Err()
	return cmd.Run()
}
