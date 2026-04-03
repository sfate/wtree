package operations

type CleanService struct {
	OperationServiceImpl
}

func NewCleanService() OperationService {
	return &CleanService{}
}

func (s *CleanService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	return m.Clean()
}
