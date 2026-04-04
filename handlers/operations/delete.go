package operations

type DeleteService struct{}

func NewDeleteService() OperationService {
	return &DeleteService{}
}

func (s *DeleteService) Process(options ServiceOptions, args ServiceArgs) error {
	m, err := options.ManagerFactory()
	if err != nil {
		return err
	}
	return m.Delete(args.DeleteRef)
}
