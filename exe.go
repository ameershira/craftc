package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type executable struct {
	ctx                                                    context.Context
	cc, cfiles, objdir, cflags, ldflags, exePath, libPaths string
	forceBuild                                             bool
}

func (e executable) exeUpToDate(cmdFile, cmd string) (bool, error) {
	// Check if lib file exists
	appStat, err := os.Stat(e.exePath)
	if err != nil {
		if os.IsNotExist(err) {
			vprintf("[link] 🧠 %s: file does not exist.\n", e.exePath)
			return false, nil
		}
		return false, err
	}

	// Check if cmd file exists
	_, err = os.Stat(cmdFile)
	if err != nil {
		if os.IsNotExist(err) {
			vprintf("[link] 🧠 %s: file does not exist.\n", cmdFile)
			return false, nil
		}
		return false, err
	}

	// if we have libs, check if any is newer than the app
	if e.libPaths != "" {
		var libNewer atomic.Bool // tracks if any lib is newer than the app

		ctx, cancel := context.WithCancel(e.ctx)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)
		// limit the io-bound goroutines
		g.SetLimit(runtime.NumCPU() * 4) // can probably be higher

		// kickoff a goroutine per lib file
		for _, lib := range strings.Fields(e.libPaths) {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err() // exit early if already canceled
				default:
				}

				libStat, err := os.Stat(lib)
				if err != nil {
					return err
				}
				if libStat.ModTime().After(appStat.ModTime()) {
					vprintf("[relink] 🧠 %s: lib file %s is newer than exe.\n", e.exePath, lib)
					libNewer.Store(true)
					cancel() // cancel other goroutines
					return nil
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			return false, err
		}
		if libNewer.Load() {
			return false, nil
		}
	}

	// check if previous link cmd is different to current.
	linkcmdData, err := os.ReadFile(cmdFile)
	if err != nil {
		return false, err
	}
	if string(linkcmdData) != cmd {
		vprintf("[link] 🧠 %s: link command changed.\n", e.exePath)
		return false, nil
	}

	return true, nil
}

// run implements the Cmd interface
func (e executable) run() (bool, error) {
	linkCmdFile := filepath.Join(e.objdir, filepath.Base(e.exePath)+".link")

	// compile app objs
	objs := objects{ctx: e.ctx, cc: e.cc, cfiles: e.cfiles, objdir: e.objdir, cflags: e.cflags, forceBuild: e.forceBuild}
	objsWasbuilt, err := objs.run()
	if err != nil {
		os.Remove(e.exePath)
		os.Remove(linkCmdFile)
		return false, err
	}

	// generate the obj list to use when linking.
	// this allows for accurate selection of obj files and
	// wont add stale objs.
	objFiles, err := getObjFiles(e.objdir, e.cfiles)
	if err != nil {
		return false, err
	}

	var args []string
	args = append(args, objFiles...)

	// Tokenize libpaths
	if e.libPaths != "" {
		args = append(args, strings.Fields(e.libPaths)...)
	}

	// Tokenize ldflags
	if e.ldflags != "" {
		args = append(args, strings.Fields(e.ldflags)...)
	}

	// {{.CC}} {{.OBJ_DIR}}/*.o {{.LIBS}} {{.LDFLAGS}} -o {{.APP_PATH}}'
	args = append(args, "-o", e.exePath)
	cmd := exec.CommandContext(e.ctx, e.cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !e.forceBuild && !objsWasbuilt {
		upToDate, err := e.exeUpToDate(linkCmdFile, cmd.String())
		if err != nil {
			os.Remove(e.exePath)
			os.Remove(linkCmdFile)
			return false, err
		}
		if upToDate {
			vprintf("✅ 🚀 %s is up to date.\n", e.exePath)
			return false, nil
		}
	}

	// Clean old files
	os.Remove(linkCmdFile)

	vprintf("[linking] 🔗 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("linking exe failed for %s: %w", e.exePath, err)
	}

	// The `.link` file stores the exact link command used.
	// If the link command changes, the app is relinked.
	if err := os.WriteFile(linkCmdFile, []byte(cmd.String()), 0644); err != nil {
		return false, fmt.Errorf("failed to write link cmd file: %w", err)
	}

	return true, nil
}
