package operations

import "fmt"

type VersionService struct {
	OperationServiceImpl
}

func NewVersionService() OperationService {
	return &VersionService{}
}

func (s *VersionService) Process(options ServiceOptions, args ServiceArgs) error {
	_, err := fmt.Fprintln(options.Stdout, options.Version)
	return err
}
