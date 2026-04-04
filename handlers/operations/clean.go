package operations

import (
	"bufio"
	"strings"
)

type CleanService struct{}

func NewCleanService() OperationService {
	return &CleanService{}
}

func (s *CleanService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}

	entries, err := m.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return m.Clean()
	}

	options.UI.Infof("Remove all %d worktree(s) for project %s? [y/N] ", len(entries), m.ProjectName())
	reader := bufio.NewReader(options.UI.In())
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" {
		options.UI.Infof("Aborted.\n")
		return nil
	}

	return m.Clean()
}
