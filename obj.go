package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// func parseDepFile(depFile string) ([]string, error) {
// 	data, err := os.ReadFile(depFile)
// 	if err != nil {
// 		return nil, err
// 	}

// 	lines := strings.Split(string(data), "\n")
// 	joined := ""
// 	for _, line := range lines {
// 		joined += strings.TrimRight(line, "\\ \t") + " "
// 	}

// 	parts := strings.SplitN(joined, ":", 2)
// 	if len(parts) != 2 {
// 		return nil, fmt.Errorf("invalid depfile format")
// 	}

// 	return strings.Fields(parts[1]), nil
// }

func parseDepFile(depFile string) ([]string, error) {
	data, err := os.ReadFile(depFile)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	// lines := strings.Split(string(data), "\n")
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

func objIsUpToDate(fileName, objFile, depFile, cmdFile, cmd string) (bool, error) {
	// if the obj file does not exist
	objStat, err := os.Stat(objFile)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// if the dep file does not exist
	_, err = os.Stat(depFile)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// if the cmd file does not exist
	_, err = os.Stat(cmdFile)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// if source file is newer than the obj file
	srcStat, err := os.Stat(fileName)
	if err != nil || srcStat.ModTime().After(objStat.ModTime()) {
		return false, nil
	}

	// if the last cmd is different to the current cmd
	cmdData, err := os.ReadFile(cmdFile)
	if err != nil {
		return false, err
	}
	if string(cmdData) != cmd {
		return false, nil
	}

	// if any of the dep files are newer than the obj file
	depFiles, err := parseDepFile(depFile)
	if err != nil {
		return false, fmt.Errorf("failed to parse dep file: %w", err)
	}

	for _, dep := range depFiles {
		depStat, err := os.Stat(dep)
		if err != nil || depStat.ModTime().After(objStat.ModTime()) {
			return false, nil
		}
	}

	return true, nil
}

func runObj(ctx context.Context, cc, cfile, objdir, cflags string) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // early exit
	default:
	}

	if cc == "" || cfile == "" || objdir == "" {
		return fmt.Errorf("cc, cfile, and objdir are required")
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
	args := append(cflagList, "-MMD", "-MF", depFile, "-c", cfile, "-o", objFile)
	// cmd := exec.Command(cc, args...)
	cmd := exec.CommandContext(ctx, cc, args...)

	// check if we should compile
	upTodate, err := objIsUpToDate(cfile, objFile, depFile, cmdFile, cmd.String())
	if err != nil {
		return err
	}
	if upTodate {
		fmt.Fprintf(os.Stderr, "%s is up to date.\n", objFile)
		return nil
	}

	// Create the directory if it does not exist
	if err := os.MkdirAll(objdir, os.ModePerm); err != nil {
		return err
	}

	// Clean old files
	os.Remove(depFile)
	os.Remove(cmdFile)

	fmt.Println(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed for %s: %w", cfile, err)
	}

	// The `.cmd` file stores the exact compile command line used.
	// If the compile command changes, the object file is rebuilt.
	if err := os.WriteFile(cmdFile, []byte(cmd.String()), 0644); err != nil {
		return fmt.Errorf("failed to write cmd file: %w", err)
	}

	return nil
}
