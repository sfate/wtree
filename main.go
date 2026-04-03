package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/sfate/wtree/cli"
	"github.com/sfate/wtree/version"
)

func main() {
	sVersion, err := version.NewSemver()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	cmd := cli.NewRootCmdWithOptions(cli.Options{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Version: sVersion.Current(),
	})
	if err := cmd.Execute(); err != nil {
		var exitErr cli.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
