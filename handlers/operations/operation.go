package operations

import wtreepkg "github.com/sfate/wtree/wtree"

type ServiceOptions struct {
	Version string
	UI      *wtreepkg.UI

	ManagerFactory func() (*wtreepkg.Manager, error)
	HookRunner     func(script string, args ...string) error
}

type ServiceArgs struct {
	Ref        string
	Branch     string
	BaseBranch string
	Shell      string
	DeleteRef  string
}

type OperationServiceFactory func() OperationService

type OperationService interface {
	Process(options ServiceOptions, args ServiceArgs) error
}
