package main

import (
	"fmt"
	"os"
)

var verbose bool

// set at the start of the app, not during runtime of cmds.
func setVerbose(enabled bool) {
	verbose = enabled
}

func vprintf(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
