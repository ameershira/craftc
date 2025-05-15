package main

import (
	"context"
	"craftc/semaphore"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
)

func runObjs(ctx context.Context, cc, cfiles, objdir, cflags string) error {
	if cc == "" || cfiles == "" || objdir == "" {
		return fmt.Errorf("cc, cfiles, and objdir are required")
	}

	files := strings.Fields(cfiles)
	if len(files) == 0 {
		return fmt.Errorf("no source files specified")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// limit the cpu bound tasks to number of logical cpus
	sem := semaphore.New(runtime.NumCPU()) // lightweight semaphore

	for _, file := range files {
		// Acquire before spawning goroutine
		if err := sem.Acquire(ctx); err != nil {
			cancel()
			return err
		}

		g.Go(func() error {
			defer sem.Release()

			if err := runObj(ctx, cc, file, objdir, cflags); err != nil {
				cancel() // trigger early cancelation
				return err
			}
			return nil
		})
	}

	return g.Wait()
}
