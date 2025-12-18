package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

const version = "0.5.4"

// common flags
var (
	ccFlag     = &cli.StringFlag{Name: "cc", Usage: "C compiler", Required: true}
	cflagsFlag = &cli.StringFlag{Name: "cflags", Usage: "Additional compiler flags"}
	objdirFlag = &cli.StringFlag{Name: "objdir", Usage: "Output object directory", Required: true}
	cfilesFlag = &cli.StringFlag{Name: "cfiles", Usage: "Space-separated list of C source files", Required: true}
)

func runCmd(cmd Cmd) {
	_, err := cmd.run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s: cmd `%s` failed: %v\n", os.Args[0], os.Args[1], err)
		os.Exit(1)
	}
}

func filterArgs(args []string, cmd *cli.Command, verbose bool) []string {
	var filterArgs = make(map[string]bool)
	var allFlags []cli.Flag
	finalArgs := []string{args[0]}

	// strip the app name
	argsTmp := args[1:]

	for _, f := range cmd.Flags {
		allFlags = append(allFlags, f)
	}

	for _, c := range cmd.Commands {
		filterArgs[c.Name] = false
		for _, f := range c.Flags {
			allFlags = append(allFlags, f)
		}
	}

	for _, f := range allFlags {
		switch tf := f.(type) {
		case *cli.BoolFlag:
			for _, k := range tf.Names() {
				filterArgs[k] = false // does not expect value
			}
		default:
			for _, k := range f.Names() {
				filterArgs[k] = true // conservative fallback: assume expects value
			}
		}
	}

	skip := false
	for i, arg := range argsTmp {
		if skip {
			skip = false
			continue
		}

		if strings.HasPrefix(arg, "-") {
			// is a flag
			trimmedArg := strings.TrimLeft(arg, "-")
			// check if exists in filter
			expectValue, ok := filterArgs[trimmedArg]
			if ok {
				finalArgs = append(finalArgs, arg)
				if expectValue {
					if i+1 < len(argsTmp) {
						finalArgs = append(finalArgs, argsTmp[i+1])
					}
					skip = true
				}
			} else if verbose {
				fmt.Fprintf(os.Stderr, "⚠️  Ignoring unknown arg: %q\n", arg)
			}
		} else {
			// is a cmd
			// check if exists in filter
			_, ok := filterArgs[arg]
			if ok {
				finalArgs = append(finalArgs, arg)
			} else if verbose {
				fmt.Fprintf(os.Stderr, "⚠️  Ignoring unknown arg: %q\n", arg)
			}
		}
	}

	return finalArgs
}

func main() {
	cmd := &cli.Command{
		Name:    "craftc",
		Usage:   "A fast, minimal C build tool",
		Version: version,
		// Global options.
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "v",
				Aliases: []string{"verbose"},
				Usage:   "Enable verbose output",
			},
			&cli.BoolFlag{
				Name:    "i",
				Aliases: []string{"ignore"},
				Usage:   "Ignore unknown commands and flags",
			},
			&cli.BoolFlag{
				Name:    "f",
				Aliases: []string{"force"},
				Usage:   "Force a complete build",
			},
		},
		EnableShellCompletion: true,

		// Sub-commands.
		Commands: []*cli.Command{
			{
				Name:  "obj",
				Usage: "Compile a single source file to object file",
				Flags: []cli.Flag{
					ccFlag,
					objdirFlag,
					cflagsFlag,
					&cli.StringFlag{Name: "cfile", Usage: "C source file", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					setVerbose(cmd.Bool("v"))
					obj := object{
						ctx:        ctx,
						cc:         cmd.String("cc"),
						cfile:      cmd.String("cfile"),
						objdir:     cmd.String("objdir"),
						cflags:     cmd.String("cflags"),
						forceBuild: cmd.Bool("f"),
					}
					runCmd(obj)
					return nil
				},
			},
			{
				Name:  "objs",
				Usage: "Compile multiple source files to object files",
				Flags: []cli.Flag{
					ccFlag,
					cfilesFlag,
					objdirFlag,
					cflagsFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					setVerbose(cmd.Bool("v"))
					objs := objects{
						ctx:        ctx,
						cc:         cmd.String("cc"),
						cfiles:     cmd.String("cfiles"),
						objdir:     cmd.String("objdir"),
						cflags:     cmd.String("cflags"),
						forceBuild: cmd.Bool("f"),
					}
					runCmd(objs)
					return nil
				},
			},
			{
				Name:  "static-lib",
				Usage: "Build a static library from multiple source files",
				Flags: []cli.Flag{
					ccFlag,
					cfilesFlag,
					objdirFlag,
					cflagsFlag,
					&cli.StringFlag{Name: "lib-path", Usage: "Library path", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					setVerbose(cmd.Bool("v"))
					sl := staticLib{
						ctx:        ctx,
						cc:         cmd.String("cc"),
						cfiles:     cmd.String("cfiles"),
						objdir:     cmd.String("objdir"),
						cflags:     cmd.String("cflags"),
						libPath:    cmd.String("lib-path"),
						forceBuild: cmd.Bool("f"),
					}
					runCmd(sl)
					return nil
				},
			},
			{
				Name:  "exe",
				Usage: "Build an application binary from source files and libraries",
				Flags: []cli.Flag{
					ccFlag,
					cfilesFlag,
					objdirFlag,
					cflagsFlag,
					&cli.StringFlag{Name: "exe-path", Usage: "Executable path", Required: true},
					&cli.StringFlag{Name: "lib-paths", Usage: "Space-separated list of library paths"},
					&cli.StringFlag{Name: "ldflags", Usage: "Additional linker flags"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					setVerbose(cmd.Bool("v"))
					exe := executable{
						ctx:        ctx,
						cc:         cmd.String("cc"),
						cfiles:     cmd.String("cfiles"),
						objdir:     cmd.String("objdir"),
						cflags:     cmd.String("cflags"),
						ldflags:    cmd.String("ldflags"),
						exePath:    cmd.String("exe-path"),
						libPaths:   cmd.String("lib-paths"),
						forceBuild: cmd.Bool("f"),
					}
					runCmd(exe)
					return nil
				},
			},
		},
	}

	// Determine if the ignore flag was set
	ignore := false
	verbose := false
	for _, arg := range os.Args {
		if arg == "-i" || arg == "--ignore" {
			ignore = true
		}
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		}
	}

	args := os.Args
	if ignore {
		args = filterArgs(os.Args, cmd, verbose)
	}

	if err := cmd.Run(context.Background(), args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s\n", err)
		os.Exit(1)
	}
}
