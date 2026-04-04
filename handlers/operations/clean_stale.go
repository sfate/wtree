package operations

import (
	"bufio"
	"strings"
	"time"

	wtreepkg "github.com/sfate/wtree/wtree"
)

type CleanStaleService struct{}

func NewCleanStaleService() OperationService {
	return &CleanStaleService{}
}

func (s *CleanStaleService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	stale, err := m.CleanStale(14 * 24 * time.Hour)
	if err != nil {
		return err
	}
	if len(stale) == 0 {
		options.UI.Infof("No stale worktrees found for project: %s\n", m.ProjectName())
		return nil
	}

	options.UI.Infof("Stale worktrees for project: %s (no activity in 2+ weeks)\n", m.ProjectName())
	wtreepkg.PrintTableTo(options.UI.Out(), stale)

	options.UI.Infof("\nRemove %d worktree(s)? [y/N] ", len(stale))
	reader := bufio.NewReader(options.UI.In())
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" {
		options.UI.Infof("Aborted.\n")
		return nil
	}

	if err := m.DeleteEntries(stale); err != nil {
		return err
	}
	options.UI.Infof("Done.\n")
	return nil
}
