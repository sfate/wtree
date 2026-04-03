package operations

import wtreepkg "github.com/sfate/wtree/wtree"

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
		options.UI.Infof("No worktrees found for project: %s\n", m.ProjectName())
		return nil
	}
	options.UI.Infof("Worktrees for project: %s\n", m.ProjectName())
	wtreepkg.PrintTableTo(options.UI.Out(), entries)
	return nil
}
