package main

import (
	"fmt"
	"os"

	"github.com/sfate/wtree/version"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: go run ./version/cmd <get|set|bump> [args...]")
	}

	sVersion, err := version.NewSemver()
	if err != nil {
		fail(fmt.Sprintf("failed to load version: %v", err))
	}

	switch os.Args[1] {
	case "get":
		fmt.Println(sVersion.Current())
	case "set":
		if len(os.Args) != 3 {
			fail("usage: go run ./version/cmd set <version>")
		}
		if err := sVersion.Set(os.Args[2]); err != nil {
			fail(err.Error())
		}
	case "bump":
		if len(os.Args) != 3 {
			fail("usage: go run ./version/cmd bump <patch|minor|major>")
		}
		next, err := sVersion.Bump(os.Args[2])
		if err != nil {
			fail(err.Error())
		}
		fmt.Println(next)
	default:
		fail(fmt.Sprintf("unknown subcommand: %q", os.Args[1]))
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
