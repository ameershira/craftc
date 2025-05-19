package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func getObjFiles(objdir, cfiles string) ([]string, error) {
	files := strings.Fields(cfiles)
	if len(files) == 0 {
		return nil, fmt.Errorf("no source files specified")
	}
	objs := make([]string, len(files))
	for i, file := range files {
		fileName := encodeFilePath(file)
		objs[i] = filepath.Join(objdir, fileName+".o")
	}
	return objs, nil
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

	// generate the obj list to use when linking.
	// this allows for accurate selection of obj files and
	// wont add stale objs.
	objs, err := getObjFiles(objdir, cfiles)
	if err != nil {
		return err
	}

	// ar rcs {{.LIB_PATH}} {{.OBJ_DIR}}/*.o
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
