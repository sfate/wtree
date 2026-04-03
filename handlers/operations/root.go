package operations

import "fmt"

type RootService struct {
	OperationServiceImpl
}

func NewRootService() OperationService {
	return &RootService{}
}

func (s *RootService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(options.Stdout, m.ProjectDir())
	return nil
}
