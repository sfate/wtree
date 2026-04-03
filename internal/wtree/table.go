package wtree

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintTable renders an ASCII table of worktree entries to stdout.
// Headers: Ref | Branch | Last Activity. Borders use +, -, |.
func PrintTable(entries []WorktreeEntry) {
	PrintTableTo(os.Stdout, entries)
}

// PrintTableTo renders an ASCII table of worktree entries to the provided writer.
func PrintTableTo(w io.Writer, entries []WorktreeEntry) {
	if w == nil {
		w = os.Stdout
	}
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

	_, _ = fmt.Fprintln(w, hr)
	_, _ = fmt.Fprintf(w, "| %-*s | %-*s | %-*s |\n", wRef, "Ref", wBranch, "Branch", wActivity, "Last Activity")
	_, _ = fmt.Fprintln(w, hr)
	for _, e := range entries {
		_, _ = fmt.Fprintf(w, "| %-*s | %-*s | %-*s |\n", wRef, e.Ref, wBranch, e.Branch, wActivity, e.RelativeAge())
	}
	_, _ = fmt.Fprintln(w, hr)
}
