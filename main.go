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
	forceBuildCmdObj = cmdObj.Bool("f", false, "force a complete build")
	verboseCmdObj    = cmdObj.Bool("v", false, "verbose output")

	// objs cmd flags
	cmdObjs           = flag.NewFlagSet("objs", flag.ExitOnError)
	ccCmdObjs         = cmdObjs.String("cc", "", "C compiler")
	cfilesCmdObjs     = cmdObjs.String("cfiles", "", "Space-separated list of C source files")
	objDirCmdObjs     = cmdObjs.String("objdir", "", "Output object directory")
	cflagsCmdObjs     = cmdObjs.String("cflags", "", "Additional compiler flags")
	forceBuildCmdObjs = cmdObjs.Bool("f", false, "force a complete build")
	verboseCmdObjs    = cmdObjs.Bool("v", false, "verbose output")

	// static-lib cmd flags
	cmdStaticLib           = flag.NewFlagSet("static-lib", flag.ExitOnError)
	libPathCmdStaticLib    = cmdStaticLib.String("libpath", "", "Lib path")
	ccCmdStaticLib         = cmdStaticLib.String("cc", "", "C compiler")
	cfilesCmdStaticLib     = cmdStaticLib.String("cfiles", "", "Space-separated list of C source files")
	objDirCmdStaticLib     = cmdStaticLib.String("objdir", "", "Output object directory")
	cflagsCmdStaticLib     = cmdStaticLib.String("cflags", "", "Additional compiler flags")
	forceBuildCmdStaticLib = cmdStaticLib.Bool("f", false, "force a complete build")
	verboseCmdStaticLib    = cmdStaticLib.Bool("v", false, "verbose output")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "❌ expected subcommand")
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "obj":
		cmdObj.Parse(os.Args[2:])
		setVerbose(*verboseCmdObj)
		vprintf("⚙️  Running cmd: %s...\n", os.Args[1])
		_, err := runObj(ctx, *ccCmdObj, *cfileCmdObj, *objDirCmdObj, *cflagsCmdObj, *forceBuildCmdObj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: cmd `obj` failed: %v\n", os.Args[0], err)
			os.Exit(1)
		}
	case "objs":
		cmdObjs.Parse(os.Args[2:])
		setVerbose(*verboseCmdObjs)
		vprintf("⚙️  Running cmd: %s...\n", os.Args[1])
		_, err := runObjs(ctx, *ccCmdObjs, *cfilesCmdObjs, *objDirCmdObjs, *cflagsCmdObjs, *forceBuildCmdObjs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: cmd `objs` failed: %v\n", os.Args[0], err)
			os.Exit(1)
		}
	case "static-lib":
		cmdStaticLib.Parse(os.Args[2:])
		if *ccCmdStaticLib == "" || *cfilesCmdStaticLib == "" || *objDirCmdStaticLib == "" || *libPathCmdStaticLib == "" {
			fmt.Fprintf(os.Stderr, "❌ cc, cfiles, objdir, and libpath are required")
			os.Exit(1)
		}
		setVerbose(*verboseCmdStaticLib)
		vprintf("⚙️  Running cmd: %s...\n", os.Args[1])
		err := runStaticLib(ctx, *ccCmdStaticLib, *cfilesCmdStaticLib, *objDirCmdStaticLib, *cflagsCmdStaticLib, *libPathCmdStaticLib, *forceBuildCmdStaticLib)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: cmd `static-lib` failed: %v\n", os.Args[0], err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "❌ unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}
