package wtree

import gitpkg "github.com/sfate/wtree/git"

type Config struct {
	BaseDir      string
	TicketPrefix string
	BranchPrefix string

	PostNavigation func(ref, projectName, worktreeDir string) error
	PostDelete     func(ref string) error
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

type discardLogger struct{}

func (discardLogger) Infof(format string, args ...any)  {}
func (discardLogger) Errorf(format string, args ...any) {}

// ManagerDeps holds optional runtime dependencies for Manager.
type ManagerDeps struct {
	Git    gitpkg.Client
	Logger Logger
}
