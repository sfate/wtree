package operations

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	wtreepkg "github.com/sfate/wtree/wtree"
)

type CleanStaleService struct {
	OperationServiceImpl
}

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
		_, _ = fmt.Fprintf(options.Stdout, "No stale worktrees found for project: %s\n", m.ProjectName())
		return nil
	}

	_, _ = fmt.Fprintf(options.Stdout, "Stale worktrees for project: %s (no activity in 2+ weeks)\n", m.ProjectName())
	wtreepkg.PrintTableTo(options.Stdout, stale)

	_, _ = fmt.Fprintf(options.Stdout, "\nRemove %d worktree(s)? [y/N] ", len(stale))
	reader := bufio.NewReader(options.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" {
		_, _ = fmt.Fprintln(options.Stdout, "Aborted.")
		return nil
	}

	if err := m.DeleteEntries(stale); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(options.Stdout, "Done.")
	return nil
}
