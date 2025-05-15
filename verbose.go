package main

import (
	"fmt"
	"os"
	"sync/atomic"
)

var verbose atomic.Bool

func setVerbose(enabled bool) {
	verbose.Store(enabled)
}

func vprintf(format string, args ...any) {
	if verbose.Load() {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
