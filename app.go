package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runApp(ctx context.Context, cc, cfiles, objdir, cflags, ldflags, appPath, libPaths string, forceBuild bool) error {

	// compile app objs
	// objsWasbuilt, err := runObjs(ctx, cc, cfiles, objdir, cflags, forceBuild)
	_, err := runObjs(ctx, cc, cfiles, objdir, cflags, forceBuild)
	if err != nil {
		os.Remove(appPath)
		return err
	}

	objs, err := filepath.Glob(objdir + "/*.o")
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		return fmt.Errorf("no object files found in %s.", objdir)
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

	args = append(args, "-o", appPath)
	cmd := exec.CommandContext(ctx, cc, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	vprintf("[linking] 🔗 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("linking app failed for %s: %w", appPath, err)
	}

	vprintf("🚀 Ready to launch: %s", appPath)

	return nil
}
