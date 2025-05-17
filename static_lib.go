package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func staticLibUpToDate(libPath string) (bool, error) {
	// Check if lib file exists
	_, err := os.Stat(libPath)
	if err != nil {
		if os.IsNotExist(err) {
			vprintf("[build-lib] 🧠 %s: file does not exist.\n", libPath)
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func runStaticLib(ctx context.Context, cc, cfiles, objdir, cflags, libPath string, forceBuild bool) error {

	objsWasbuilt, err := runObjs(ctx, cc, cfiles, objdir, cflags, forceBuild)
	if err != nil {
		os.Remove(libPath)
		return err
	}

	// Create the libpath directory if it does not exist
	if err := os.MkdirAll(filepath.Dir(libPath), os.ModePerm); err != nil {
		return err
	}

	if !forceBuild && !objsWasbuilt {
		upToDate, err := staticLibUpToDate(libPath)
		if err != nil {
			os.Remove(libPath)
			return err
		}
		if upToDate {
			vprintf("✅ 📦 %s is up to date.\n", libPath)
			return nil
		}
	}

	os.Remove(libPath)

	objs, err := filepath.Glob(objdir + "/*.o")
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		return fmt.Errorf("no object files found in %s.", objdir)
	}

	args := append([]string{"rcs", libPath}, objs...)
	cmd := exec.CommandContext(ctx, "ar", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	vprintf("[archive] 📦 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("archive failed for %s: %w", libPath, err)
	}

	return nil
}
