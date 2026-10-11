package auth

import (
	"os"

	"golang.org/x/term"
)

// stdinIsTerminal reports whether stdin is attached to a terminal, i.e. whether someone can
// complete the interactive authorization flow. It is a variable so tests can override it.
var stdinIsTerminal = func() bool {
	return isTerminal(os.Stdin)
}

// isTerminal reports whether f is an interactive TTY. Character devices that are not
// terminals, such as /dev/null, are reported as non-interactive.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}
