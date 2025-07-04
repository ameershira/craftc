package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

type commonFlags struct {
	cc         *string
	cflags     *string
	objDir     *string
	forceBuild *bool
	verbose    *bool
}

func addCommonFlags(fs *flag.FlagSet) *commonFlags {
	return &commonFlags{
		cc:         fs.String("cc", "", "C compiler"),
		cflags:     fs.String("cflags", "", "Additional compiler flags"),
		objDir:     fs.String("objdir", "", "Output object directory"),
		forceBuild: fs.Bool("f", false, "Force a complete build"),
		verbose:    fs.Bool("v", false, "Verbose output"),
	}
}

func addCFilesFlag(fs *flag.FlagSet) *string {
	return fs.String("cfiles", "", "Space-separated list of C source files")
}

var (
	// obj cmd
	cmdObj      = flag.NewFlagSet("obj", flag.ExitOnError)
	flagsCmdObj = addCommonFlags(cmdObj)
	cfileCmdObj = cmdObj.String("cfile", "", "C source file")

	// objs cmd
	cmdObjs       = flag.NewFlagSet("objs", flag.ExitOnError)
	flagsCmdObjs  = addCommonFlags(cmdObjs)
	cfilesCmdObjs = addCFilesFlag(cmdObjs)

	// static-lib cmd
	cmdStaticLib        = flag.NewFlagSet("static-lib", flag.ExitOnError)
	flagsCmdStaticLib   = addCommonFlags(cmdStaticLib)
	cfilesCmdStaticLib  = addCFilesFlag(cmdStaticLib)
	libPathCmdStaticLib = cmdStaticLib.String("lib-path", "", "Library path")

	// exe cmd
	cmdExe         = flag.NewFlagSet("exe", flag.ExitOnError)
	flagsCmdExe    = addCommonFlags(cmdExe)
	cfilesCmdExe   = addCFilesFlag(cmdExe)
	exePathCmdExe  = cmdExe.String("exe-path", "", "Executable path")
	libPathsCmdExe = cmdExe.String("lib-paths", "", "Space-separated list of library paths")
	ldflagsCmdExe  = cmdExe.String("ldflags", "", "Additional linker flags")
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s <command> [options]

Available commands:
  obj         Compile a single source file to object file
  objs        Compile multiple source files to object files
  static-lib  Build a static library from multiple C source files
  exe         Build an application binary from source files and libraries

Use %s "<command> -h" for command-specific options.

`, os.Args[0], os.Args[0])
}

func runCmd(cmd Cmd) {
	_, err := cmd.run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s: cmd `%s` failed: %v\n", os.Args[0], os.Args[1], err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx := context.Background()

	subCmd := os.Args[1]
	subArgs := os.Args[2:]

	switch subCmd {
	case "obj":
		cmdObj.Parse(subArgs)
		setVerbose(*flagsCmdObj.verbose)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *cfileCmdObj)
		obj := object{ctx: ctx, cc: *flagsCmdObj.cc, cfile: *cfileCmdObj, objdir: *flagsCmdObj.objDir, cflags: *flagsCmdObj.cflags, forceBuild: *flagsCmdObj.forceBuild}
		runCmd(obj)
	case "objs":
		cmdObjs.Parse(subArgs)
		setVerbose(*flagsCmdObjs.verbose)
		vprintf("⚙️  Running cmd %s\n", os.Args[1])
		objs := objects{ctx: ctx, cc: *flagsCmdObjs.cc, cfiles: *cfilesCmdObjs, objdir: *flagsCmdObjs.objDir, cflags: *flagsCmdObjs.cflags, forceBuild: *flagsCmdObjs.forceBuild}
		runCmd(objs)
	case "static-lib":
		cmdStaticLib.Parse(subArgs)
		if *flagsCmdStaticLib.cc == "" || *cfilesCmdStaticLib == "" || *flagsCmdStaticLib.objDir == "" || *libPathCmdStaticLib == "" {
			fmt.Fprintf(os.Stderr, "❌ cc, cfiles, objdir, and libpath are required\n")
			os.Exit(1)
		}
		setVerbose(*flagsCmdStaticLib.verbose)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *libPathCmdStaticLib)
		sl := staticLib{ctx: ctx, cc: *flagsCmdStaticLib.cc, cfiles: *cfilesCmdStaticLib, objdir: *flagsCmdStaticLib.objDir,
			cflags: *flagsCmdStaticLib.cflags, libPath: *libPathCmdStaticLib, forceBuild: *flagsCmdStaticLib.forceBuild}
		runCmd(sl)
	case "exe":
		cmdExe.Parse(subArgs)
		if *flagsCmdExe.cc == "" || *cfilesCmdExe == "" || *flagsCmdExe.objDir == "" || *exePathCmdExe == "" {
			fmt.Fprintf(os.Stderr, "❌ cc, cfiles, objdir, and app-path are required\n")
			os.Exit(1)
		}
		setVerbose(*flagsCmdExe.verbose)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *exePathCmdExe)
		exe := executable{ctx: ctx, cc: *flagsCmdExe.cc, cfiles: *cfilesCmdExe, objdir: *flagsCmdExe.objDir, cflags: *flagsCmdExe.cflags,
			ldflags: *ldflagsCmdExe, exePath: *exePathCmdExe, libPaths: *libPathsCmdExe, forceBuild: *flagsCmdExe.forceBuild}
		runCmd(exe)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "❌ unknown subcommand: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
