package operations

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
	options.UI.Infof("%s\n", m.ProjectDir())
	return nil
}
