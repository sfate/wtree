package operations

type VersionService struct {
	OperationServiceImpl
}

func NewVersionService() OperationService {
	return &VersionService{}
}

func (s *VersionService) Process(options ServiceOptions, args ServiceArgs) error {
	options.UI.Errorf("%s\n", options.Version)
	return nil
}
