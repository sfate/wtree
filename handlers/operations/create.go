package operations

import "fmt"

type CreateService struct {
	OperationServiceImpl
}

func NewCreateService() OperationService {
	return &CreateService{}
}

func (s *CreateService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	dir, _, err := m.Create(args.Ref, args.Branch, args.BaseBranch)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(options.Stdout, dir)
	return nil
}
