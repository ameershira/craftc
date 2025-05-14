package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	cmdObj = flag.NewFlagSet("obj", flag.ExitOnError)
	cc     = cmdObj.String("cc", "", "C compiler")
	cfile  = cmdObj.String("cfile", "", "C source file")
	objDir = cmdObj.String("objdir", "", "Output object directory")
	cflags = cmdObj.String("cflags", "", "Additional compiler flags")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "expected subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "obj":
		cmdObj.Parse(os.Args[2:])
		runObj()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func encodeFilePath(filePath string) string {
	// Replace path separators with underscores or another safe character
	encodedPath := strings.TrimPrefix(filePath, "./")
	encodedPath = strings.ReplaceAll(encodedPath, string(os.PathSeparator), "_")
	// Optionally, we could hash the path to make it shorter or to handle edge cases
	return encodedPath
}

func runObj() {
	if *cc == "" || *cfile == "" || *objDir == "" {
		fmt.Fprintln(os.Stderr, "cc, cfile, and objdir are required")
		os.Exit(1)
	}

	// fileName := filepath.Base(*cfile)
	fileName := encodeFilePath(*cfile)
	objFile := filepath.Join(*objDir, fileName+".o")
	depFile := filepath.Join(*objDir, fileName+".d")

	// check if we should compile
	if objIsUpToDate(*cfile, objFile, depFile) {
		fmt.Fprintf(os.Stderr, "%s is up to date.\n", objFile)
		os.Exit(0)
	}

	// Create the directory if it does not exist
	if err := os.MkdirAll(*objDir, os.ModePerm); err != nil {
		os.Exit(1)
	}

	// Clean old files
	os.Remove(depFile)

	// Tokenize cflags
	var cflagList []string
	if *cflags != "" {
		cflagList = strings.Fields(*cflags)
	}

	// Build full command args
	args := append(cflagList, "-MMD", "-MF", depFile, "-c", *cfile, "-o", objFile)

	// Compile
	cmd := exec.Command(*cc, args...)
	fmt.Println(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "compilation failed: %v\n", err)
		os.Exit(1)
	}

}

func parseDepFile(depFile string) ([]string, error) {
	data, err := os.ReadFile(depFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	joined := ""
	for _, line := range lines {
		joined += strings.TrimRight(line, "\\ \t") + " "
	}

	parts := strings.SplitN(joined, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid depfile format")
	}

	return strings.Fields(parts[1]), nil
}

func objIsUpToDate(fileName, objFile, depFile string) bool {
	// if the obj file does not exist
	objStat, err := os.Stat(objFile)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// if the dep file does not exist
	_, err = os.Stat(depFile)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// if source file is newer than the obj file
	srcStat, err := os.Stat(fileName)
	if err != nil || srcStat.ModTime().After(objStat.ModTime()) {
		return false
	}

	// if any of the dep files are newer than the obj file
	depFiles, err := parseDepFile(depFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse dep file: %v\n", err)
		os.Exit(1)
	}
	// fmt.Println(depFiles)
	for _, dep := range depFiles {
		depStat, err := os.Stat(dep)
		if err != nil || depStat.ModTime().After(objStat.ModTime()) {
			return false
		}
	}

	return true
}
