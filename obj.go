package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

func encodeFilePath(filePath string) string {
	encodedPath := filepath.Clean(filePath)
	encodedPath = strings.TrimPrefix(encodedPath, "./")
	encodedPath = strings.ReplaceAll(encodedPath, string(os.PathSeparator), ".")
	// Optionally, we could hash the path to make it shorter or to handle edge cases
	if len(encodedPath) >= 255 {
		fmt.Fprintln(os.Stderr, "encodeFilePath name size larger than 255")
		os.Exit(1)
	}
	return encodedPath
}

func parseDepFile(depFile string) ([]string, error) {
	data, err := os.ReadFile(depFile)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		trimmed := strings.TrimRight(line, "\\ \t")
		builder.WriteString(trimmed)
		builder.WriteByte(' ')
	}

	content := builder.String()
	colonIndex := strings.Index(content, ":")
	if colonIndex < 0 {
		return nil, fmt.Errorf("invalid depfile format: no colon found")
	}

	return strings.Fields(content[colonIndex+1:]), nil
}

func depsAreUpToDate(ctx context.Context, objFile string, deps []string, objModTime time.Time) (bool, error) {
	if len(deps) == 0 {
		return true, nil
	}

	var triggered atomic.Bool

	g, ctx := errgroup.WithContext(ctx)
	// limit the io-bound goroutines
	g.SetLimit(runtime.NumCPU() * 4) // this can possibly be higher

	for _, dep := range deps {
		// Skip if rebuild already triggered
		if triggered.Load() {
			break
		}

		g.Go(func() error {
			// Short-circuit inside goroutine
			if triggered.Load() {
				return nil
			}

			info, err := os.Stat(dep)
			if err != nil {
				return fmt.Errorf("failed to stat '%s': %w", dep, err)
			}

			if info.ModTime().After(objModTime) {
				vprintf("[rebuild] 🧠 %s: dep %s is newer than object.\n", objFile, dep)
				triggered.Store(true)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	if triggered.Load() {
		return false, nil // Rebuild needed
	}

	return true, nil // all is up to date, no rebuild
}

func objIsUpToDate(ctx context.Context, fileName, objFile, depFile, cmdFile, cmd string) (bool, error) {

	// Check if object file exists
	objStat, err := os.Stat(objFile)
	if err != nil {
		if os.IsNotExist(err) {
			vprintf("[build] 🧠 %s: object file does not exist.\n", objFile)
			return false, nil
		}
		return false, err
	}

	// Check if dep file exists
	if _, err := os.Stat(depFile); err != nil {
		if os.IsNotExist(err) {
			vprintf("[build] 🧠 %s: dep file %s does not exist.\n", objFile, depFile)
			return false, nil
		}
		return false, err
	}

	// Check if cmd file exists
	if _, err := os.Stat(cmdFile); err != nil {
		if os.IsNotExist(err) {
			vprintf("[build] 🧠 %s: cmd file %s does not exist.\n", objFile, cmdFile)
			return false, nil
		}
		return false, err
	}

	// if source file is newer than the obj file
	srcStat, err := os.Stat(fileName)
	if err != nil {
		return false, err
	}
	if srcStat.ModTime().After(objStat.ModTime()) {
		vprintf("[rebuild] 🧠 %s: source file %s is newer than object.\n", objFile, fileName)
		return false, nil
	}

	// if the last cmd is different to the current cmd
	cmdData, err := os.ReadFile(cmdFile)
	if err != nil {
		return false, err
	}
	if string(cmdData) != cmd {
		vprintf("[rebuild] 🧠 %s: compile command changed.\n", objFile)
		return false, nil
	}

	// if any of the dep files are newer than the obj file
	depFiles, err := parseDepFile(depFile)
	if err != nil {
		return false, fmt.Errorf("failed to parse dep file: %w", err)
	}

	// Concurrently check dep files mod time
	upToDate, err := depsAreUpToDate(ctx, objFile, depFiles, objStat.ModTime())
	if err != nil {
		return false, err
	}

	return upToDate, nil
}

func runObj(ctx context.Context, cc, cfile, objdir, cflags string, forceBuild bool) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err() // early exit
	default:
	}

	if cc == "" || cfile == "" || objdir == "" {
		return false, fmt.Errorf("cc, cfile, and objdir are required\n")
	}

	fileName := encodeFilePath(cfile)
	objFile := filepath.Join(objdir, fileName+".o")
	depFile := filepath.Join(objdir, fileName+".d")
	cmdFile := filepath.Join(objdir, fileName+".cmd")

	// Tokenize cflags
	var cflagList []string
	if cflags != "" {
		cflagList = strings.Fields(cflags)
	}

	// Build full command args
	// {{.CC}} {{.CFLAGS}} -MMD -MF {{.DEP_FILE}} -c {{.CFILE}} -o {{.OBJ_FILE}}
	args := append(cflagList, "-MMD", "-MF", depFile, "-c", cfile, "-o", objFile)
	cmd := exec.CommandContext(ctx, cc, args...)

	if !forceBuild {
		// check if we should compile
		upTodate, err := objIsUpToDate(ctx, cfile, objFile, depFile, cmdFile, cmd.String())
		if err != nil {
			return false, err
		}
		if upTodate {
			vprintf("✅ %s is up to date.\n", objFile)
			return false, nil
		}
	}

	// Create the directory if it does not exist
	if err := os.MkdirAll(objdir, os.ModePerm); err != nil {
		return false, err
	}

	// Clean old files
	os.Remove(depFile)
	os.Remove(cmdFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	vprintf("[compile] 🔨 %s\n", cmd.String())
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("compilation failed for %s: %w", cfile, err)
	}

	// The `.cmd` file stores the exact compile command line used.
	// If the compile command changes, the object file is rebuilt.
	if err := os.WriteFile(cmdFile, []byte(cmd.String()), 0644); err != nil {
		return true, fmt.Errorf("failed to write cmd file: %w", err)
	}

	return true, nil
}
