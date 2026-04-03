package operations

import (
	"fmt"

	wtreepkg "github.com/sfate/wtree/wtree"
)

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
}

type OperationServiceFactory func() OperationService

type OperationServiceImpl struct {
}

type OperationService interface {
	Process(options ServiceOptions, args ServiceArgs) error
}

func (s *OperationServiceImpl) Process(options ServiceOptions, args ServiceArgs) error {
	return fmt.Errorf("not implemented")
}
