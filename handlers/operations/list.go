package operations

import (
	"fmt"

	wtreepkg "github.com/sfate/wtree/wtree"
)

type ListService struct {
	OperationServiceImpl
}

func NewListService() OperationService {
	return &ListService{}
}

func (s *ListService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	entries, err := m.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		_, _ = fmt.Fprintf(options.Stdout, "No worktrees found for project: %s\n", m.ProjectName())
		return nil
	}
	_, _ = fmt.Fprintf(options.Stdout, "Worktrees for project: %s\n", m.ProjectName())
	wtreepkg.PrintTableTo(options.Stdout, entries)
	return nil
}
