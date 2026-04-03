package operations

import (
	_ "embed"
	"fmt"
)

//go:embed shell_init_bash.sh
var shellInitBash string

//go:embed shell_init_zsh.sh
var shellInitZsh string

type ShellInitService struct {
	OperationServiceImpl
}

func NewShellInitService() OperationService {
	return &ShellInitService{}
}

func (s *ShellInitService) Process(options ServiceOptions, args ServiceArgs) error {
	switch args.Shell {
	case "zsh":
		options.UI.Infof("%s", shellInitZsh)
		return nil
	case "bash":
		options.UI.Infof("%s", shellInitBash)
		return nil
	default:
		return fmt.Errorf("unsupported shell %q — supported: zsh, bash", args.Shell)
	}
}
