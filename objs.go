package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

func runObjs(ctx context.Context, cc, cfiles, objdir, cflags string, forceBuild bool) (bool, error) {
	if cc == "" || cfiles == "" || objdir == "" {
		return false, fmt.Errorf("cc, cfiles, and objdir are required")
	}

	files := strings.Fields(cfiles)
	if len(files) == 0 {
		return false, fmt.Errorf("no source files specified")
	}

	var anyObjsBuilt atomic.Bool // tracks if any object was built

	g, ctx := errgroup.WithContext(ctx)
	// limit the cpu bound goroutines to the number of logical cpus
	g.SetLimit(runtime.NumCPU())

	// kickoff a goroutine per source file
	for _, file := range files {
		g.Go(func() error {
			builtObj, err := runObj(ctx, cc, file, objdir, cflags, forceBuild)
			if err != nil {
				return err
			}
			if builtObj {
				anyObjsBuilt.Store(true)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	// returns true if at least one obj was built
	return anyObjsBuilt.Load(), nil
}
