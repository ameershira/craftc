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

func appUpToDate(ctx context.Context, appPath, libPaths, cmdFile, cmd string) (bool, error) {
	// Check if lib file exists
	appStat, err := os.Stat(appPath)
	if err != nil {
		if os.IsNotExist(err) {
			vprintf("[link] 🧠 %s: file does not exist.\n", appPath)
			return false, nil
		}
		return false, err
	}

	// if we have libs, check if any is newer than the app
	if libPaths != "" {
		var libNewer atomic.Bool // tracks if any lib is newer than the app

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)
		// limit the io-bound goroutines
		g.SetLimit(runtime.NumCPU() * 4) // can probably be higher

		// kickoff a goroutine per lib file
		for _, lib := range strings.Fields(libPaths) {
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
					vprintf("[relink] 🧠 %s: lib file %s is newer than app.\n", appPath, lib)
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
		vprintf("[link] 🧠 %s: link command changed.\n", appPath)
		return false, nil
	}

	return true, nil
}

func runApp(ctx context.Context, cc, cfiles, objdir, cflags, ldflags, appPath, libPaths string, forceBuild bool) error {
	linkCmdFile := filepath.Join(objdir, filepath.Base(appPath)+".link")

	// compile app objs
	objsWasbuilt, err := runObjs(ctx, cc, cfiles, objdir, cflags, forceBuild)
	if err != nil {
		os.Remove(appPath)
		os.Remove(linkCmdFile)
		return err
	}

	// generate the obj list to use when linking.
	// this allows for accurate selection of obj files and
	// wont add stale objs.
	objs, err := getObjFiles(objdir, cfiles)
	if err != nil {
		return err
	}

	var args []string
	args = append(args, objs...)

	// Tokenize libpaths
	if libPaths != "" {
		args = append(args, strings.Fields(libPaths)...)
	}

	// Tokenize ldflags
	if ldflags != "" {
		args = append(args, strings.Fields(ldflags)...)
	}

	// {{.CC}} {{.OBJ_DIR}}/*.o {{.LIBS}} {{.LDFLAGS}} -o {{.APP_PATH}}'
	args = append(args, "-o", appPath)
	cmd := exec.CommandContext(ctx, cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !forceBuild && !objsWasbuilt {
		upToDate, err := appUpToDate(ctx, appPath, libPaths, linkCmdFile, cmd.String())
		if err != nil {
			os.Remove(appPath)
			os.Remove(linkCmdFile)
			return err
		}
		if upToDate {
			vprintf("✅ 🚀 %s is up to date.\n", appPath)
			return nil
		}
	}

	// Clean old files
	os.Remove(linkCmdFile)

	vprintf("[linking] 🔗 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("linking app failed for %s: %w", appPath, err)
	}

	// The `.link` file stores the exact link command used.
	// If the link command changes, the app is relinked.
	if err := os.WriteFile(linkCmdFile, []byte(cmd.String()), 0644); err != nil {
		return fmt.Errorf("failed to write link cmd file: %w", err)
	}

	return nil
}
