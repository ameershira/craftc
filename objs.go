package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type objects struct {
	ctx                        context.Context
	cc, cfiles, objdir, cflags string
	forceBuild                 bool
}

// run implements the Cmd interface
func (o objects) run() (bool, error) {
	if o.cc == "" || o.cfiles == "" || o.objdir == "" {
		return false, fmt.Errorf("cc, cfiles, and objdir are required\n")
	}

	files := strings.Fields(o.cfiles)
	if len(files) == 0 {
		return false, fmt.Errorf("no source files specified")
	}

	var anyObjsBuilt atomic.Bool // tracks if any object was built

	g, ctx := errgroup.WithContext(o.ctx)
	// limit the cpu bound goroutines to the number of logical cpus
	g.SetLimit(runtime.NumCPU())

	// kickoff a goroutine per source file
	for _, file := range files {
		g.Go(func() error {
			obj := object{ctx: ctx, cc: o.cc, cfile: file, objdir: o.objdir, cflags: o.cflags, forceBuild: o.forceBuild}
			builtObj, err := obj.run()
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
