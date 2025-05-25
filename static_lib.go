package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type staticLib struct {
	ctx                                 context.Context
	cc, cfiles, objdir, cflags, libPath string
	forceBuild                          bool
}

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

// run implements the Cmd interface
func (s staticLib) run() (bool, error) {
	objs := objects{ctx: s.ctx, cc: s.cc, cfiles: s.cfiles, objdir: s.objdir, cflags: s.cflags, forceBuild: s.forceBuild}
	objsWasbuilt, err := objs.run()
	if err != nil {
		os.Remove(s.libPath)
		return false, err
	}

	// Create the libpath directory if it does not exist
	if err := os.MkdirAll(filepath.Dir(s.libPath), os.ModePerm); err != nil {
		return false, err
	}

	if !s.forceBuild && !objsWasbuilt {
		upToDate, err := staticLibUpToDate(s.libPath)
		if err != nil {
			os.Remove(s.libPath)
			return false, err
		}
		if upToDate {
			vprintf("✅ 📦 %s is up to date.\n", s.libPath)
			return false, nil
		}
	}

	os.Remove(s.libPath)

	// generate the obj list to use when linking.
	// this allows for accurate selection of obj files and
	// wont add stale objs.
	objFiles, err := getObjFiles(s.objdir, s.cfiles)
	if err != nil {
		return false, err
	}

	// ar rcs {{.LIB_PATH}} {{.OBJ_DIR}}/*.o
	args := append([]string{"rcs", s.libPath}, objFiles...)
	cmd := exec.CommandContext(s.ctx, "ar", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	vprintf("[archive] 📦 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("archive failed for %s: %w", s.libPath, err)
	}

	// return true if we built the static lib
	return true, nil
}
