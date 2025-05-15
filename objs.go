package main

import (
	"context"
	"craftc/semaphore"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
)

// func runObjs(ctx context.Context, cc, cfiles, objdir, cflags string) error {
// 	if cc == "" || cfiles == "" || objdir == "" {
// 		return fmt.Errorf("cc, cfiles, and objdir are required")
// 	}

// 	files := strings.Fields(cfiles)
// 	if len(files) == 0 {
// 		return fmt.Errorf("no source files specified")
// 	}

// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	maxGoroutines := runtime.NumCPU() // Limit to CPU cores
// 	sem := make(chan struct{}, maxGoroutines)
// 	errChan := make(chan error, 1)
// 	var wg sync.WaitGroup

// 	for _, f := range files {
// 		select {
// 		case <-ctx.Done():
// 			break
// 		default:
// 		}

// 		wg.Add(1)
// 		sem <- struct{}{} // acquire slot

// 		go func(file string) {
// 			defer wg.Done()
// 			defer func() { <-sem }() // release slot

// 			if err := runObj(ctx, cc, file, objdir, cflags); err != nil {
// 				select {
// 				case errChan <- err:
// 					cancel() // first error triggers cancel
// 				default:
// 					// error already sent
// 				}
// 			}
// 		}(f)
// 	}

// 	wg.Wait()

// 	select {
// 	case err := <-errChan:
// 		return err
// 	default:
// 		return nil
// 	}
// }

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

	sem := semaphore.New(runtime.NumCPU()) // lightweight semaphore

	for _, file := range files {
		g.Go(func() error {
			if err := sem.Acquire(ctx); err != nil {
				return err // context canceled
			}
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
