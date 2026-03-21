package wtree

import (
	"fmt"
	"strings"
)

// PrintTable renders an ASCII table of worktree entries to stdout.
// Headers: Ref | Branch | Last Activity. Borders use +, -, |.
func PrintTable(entries []WorktreeEntry) {
	wRef := len("Ref")
	wBranch := len("Branch")
	wActivity := len("Last Activity")

	for _, e := range entries {
		if len(e.Ref) > wRef {
			wRef = len(e.Ref)
		}
		if len(e.Branch) > wBranch {
			wBranch = len(e.Branch)
		}
		age := e.RelativeAge()
		if len(age) > wActivity {
			wActivity = len(age)
		}
	}

	hr := fmt.Sprintf("+-%s-+-%s-+-%s-+",
		strings.Repeat("-", wRef),
		strings.Repeat("-", wBranch),
		strings.Repeat("-", wActivity),
	)

	fmt.Println(hr)
	fmt.Printf("| %-*s | %-*s | %-*s |\n", wRef, "Ref", wBranch, "Branch", wActivity, "Last Activity")
	fmt.Println(hr)
	for _, e := range entries {
		fmt.Printf("| %-*s | %-*s | %-*s |\n", wRef, e.Ref, wBranch, e.Branch, wActivity, e.RelativeAge())
	}
	fmt.Println(hr)
}
