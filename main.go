package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

var (
	// obj cmd flags
	cmdObj           = flag.NewFlagSet("obj", flag.ExitOnError)
	ccCmdObj         = cmdObj.String("cc", "", "C compiler")
	cfileCmdObj      = cmdObj.String("cfile", "", "C source file")
	objDirCmdObj     = cmdObj.String("objdir", "", "Output object directory")
	cflagsCmdObj     = cmdObj.String("cflags", "", "Additional compiler flags")
	forceBuildCmdObj = cmdObj.Bool("f", false, "Force a complete build")
	verboseCmdObj    = cmdObj.Bool("v", false, "Verbose output")

	// objs cmd flags
	cmdObjs           = flag.NewFlagSet("objs", flag.ExitOnError)
	ccCmdObjs         = cmdObjs.String("cc", "", "C compiler")
	cfilesCmdObjs     = cmdObjs.String("cfiles", "", "Space-separated list of C source files")
	objDirCmdObjs     = cmdObjs.String("objdir", "", "Output object directory")
	cflagsCmdObjs     = cmdObjs.String("cflags", "", "Additional compiler flags")
	forceBuildCmdObjs = cmdObjs.Bool("f", false, "Force a complete build")
	verboseCmdObjs    = cmdObjs.Bool("v", false, "Verbose output")

	// static-lib cmd flags
	cmdStaticLib           = flag.NewFlagSet("static-lib", flag.ExitOnError)
	libPathCmdStaticLib    = cmdStaticLib.String("lib-path", "", "Lib path")
	ccCmdStaticLib         = cmdStaticLib.String("cc", "", "C compiler")
	cfilesCmdStaticLib     = cmdStaticLib.String("cfiles", "", "Space-separated list of C source files")
	objDirCmdStaticLib     = cmdStaticLib.String("objdir", "", "Output object directory")
	cflagsCmdStaticLib     = cmdStaticLib.String("cflags", "", "Additional compiler flags")
	forceBuildCmdStaticLib = cmdStaticLib.Bool("f", false, "Force a complete build")
	verboseCmdStaticLib    = cmdStaticLib.Bool("v", false, "Verbose output")

	// exe cmd flags
	cmdExe           = flag.NewFlagSet("exe", flag.ExitOnError)
	exePathCmdExe    = cmdExe.String("exe-path", "", "Executable path")
	libPathsCmdExe   = cmdExe.String("lib-paths", "", "Space-separated list of Lib paths")
	ccCmdExe         = cmdExe.String("cc", "", "C compiler")
	cfilesCmdExe     = cmdExe.String("cfiles", "", "Space-separated list of C source files")
	objDirCmdExe     = cmdExe.String("objdir", "", "Output object directory")
	cflagsCmdExe     = cmdExe.String("cflags", "", "Additional compiler flags")
	ldflagsCmdExe    = cmdExe.String("ldflags", "", "Additional linker flags")
	forceBuildCmdExe = cmdExe.Bool("f", false, "Force a complete build")
	verboseCmdExe    = cmdExe.Bool("v", false, "Verbose output")
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

	switch os.Args[1] {
	case "obj":
		cmdObj.Parse(os.Args[2:])
		setVerbose(*verboseCmdObj)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *cfileCmdObj)
		obj := object{ctx: ctx, cc: *ccCmdObj, cfile: *cfileCmdObj, objdir: *objDirCmdObj, cflags: *cflagsCmdObj, forceBuild: *forceBuildCmdObj}
		runCmd(obj)
	case "objs":
		cmdObjs.Parse(os.Args[2:])
		setVerbose(*verboseCmdObjs)
		vprintf("⚙️  Running cmd %s\n", os.Args[1])
		objs := objects{ctx: ctx, cc: *ccCmdObjs, cfiles: *cfilesCmdObjs, objdir: *objDirCmdObjs, cflags: *cflagsCmdObjs, forceBuild: *forceBuildCmdObjs}
		runCmd(objs)
	case "static-lib":
		cmdStaticLib.Parse(os.Args[2:])
		if *ccCmdStaticLib == "" || *cfilesCmdStaticLib == "" || *objDirCmdStaticLib == "" || *libPathCmdStaticLib == "" {
			fmt.Fprintf(os.Stderr, "❌ cc, cfiles, objdir, and libpath are required\n")
			os.Exit(1)
		}
		setVerbose(*verboseCmdStaticLib)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *libPathCmdStaticLib)
		sl := staticLib{ctx: ctx, cc: *ccCmdStaticLib, cfiles: *cfilesCmdStaticLib, objdir: *objDirCmdStaticLib,
			cflags: *cflagsCmdStaticLib, libPath: *libPathCmdStaticLib, forceBuild: *forceBuildCmdStaticLib}
		runCmd(sl)
	case "exe":
		cmdExe.Parse(os.Args[2:])
		if *ccCmdExe == "" || *cfilesCmdExe == "" || *objDirCmdExe == "" || *exePathCmdExe == "" {
			fmt.Fprintf(os.Stderr, "❌ cc, cfiles, objdir, and app-path are required\n")
			os.Exit(1)
		}
		setVerbose(*verboseCmdExe)
		vprintf("⚙️  Running cmd %s: %s\n", os.Args[1], *exePathCmdExe)
		exe := executable{ctx: ctx, cc: *ccCmdExe, cfiles: *cfilesCmdExe, objdir: *objDirCmdExe, cflags: *cflagsCmdExe,
			ldflags: *ldflagsCmdExe, exePath: *exePathCmdExe, libPaths: *libPathsCmdExe, forceBuild: *forceBuildCmdExe}
		runCmd(exe)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "❌ unknown subcommand: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
