![craftc — Build when it matters](banner.png)

# craftc

`craftc` is a small command-line build tool for C projects. It compiles object
files, creates static libraries, and links executables. It is designed to be
called directly or from tools such as Make and [Task](https://taskfile.dev/).

It tracks source files, headers, and command-line changes so unchanged targets
do not need to be rebuilt.

[![Go Reference](https://pkg.go.dev/badge/github.com/ameershira/craftc.svg)](https://pkg.go.dev/github.com/ameershira/craftc)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ameershira/craftc)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ameershira/craftc/go.yml)
![GitHub](https://img.shields.io/github/license/ameershira/craftc)

## What it does

- Rebuilds objects when a source file, header, or compile command changes
- Compiles multiple source files concurrently
- Creates static libraries with `ar`
- Links executables and relinks when an input library changes
- Accepts custom compiler and linker flags
- Provides optional verbose output

Linux is supported. macOS and Windows have not been tested.

## Module path change

This module was previously published as:

```text
github.com/ameergituser/craftc
```

New versions are published under:

```text
github.com/ameershira/craftc
```

## Installation

Install with Go:

```sh
go install github.com/ameershira/craftc@latest
```

Or clone and build it locally:

```sh
git clone https://github.com/ameershira/craftc
cd craftc
go build .
```

## Usage examples

These examples use `-v` to print the commands being run.

### 1. Build an object file

Use the `obj` command:

```sh
craftc obj -cc cc -cfile ./libsrc1.c -cflags "-Wall -O2" -objdir ./build/obj -v
```
Output:
```sh
[build] 🧠 build/obj/libsrc1.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -O2 -MMD -MF build/obj/libsrc.libsrc1.c.d -c ./libsrc1.c -o build/obj/libsrc1.c.o
```
Re-run:
```sh
✅ build/obj/libsrc1.c.o is up to date.
```

### 2. Build multiple object files concurrently

Use the `objs` command:

```sh
craftc objs -cc cc -cfiles "./libsrc1.c ./libsrc2.c" -cflags "-Wall" -objdir ./build/obj -v
```
Output:
```sh
[build] 🧠 build/obj/libsrc2.c.o: object file does not exist.
[build] 🧠 build/obj/libsrc1.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -MMD -MF build/obj/libsrc2.c.d -c ./libsrc2.c -o build/obj/libsrc2.c.o
[compile] 🔨 /usr/bin/cc -Wall -MMD -MF build/obj/libsrc1.c.d -c ./libsrc1.c -o build/obj/libsrc1.c.o
```

### 3. Build a static library

Use the `static-lib` command:

```sh
craftc static-lib -cc cc -cfiles "./libsrc1.c ./libsrc2.c" -cflags "-Wall -O2" -lib-path "./build/lib.a" -objdir ./build/obj -v
```
Output:
```sh
[build] 🧠 build/obj/libsrc2.c.o: object file does not exist.
[build] 🧠 build/obj/libsrc1.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -O2 -MMD -MF build/obj/libsrc2.c.d -c ./libsrc2.c -o build/obj/libsrc2.c.o
[compile] 🔨 /usr/bin/cc -Wall -O2 -MMD -MF build/obj/libsrc1.c.d -c ./libsrc1.c -o build/obj/libsrc1.c.o
[archive] 📦 /usr/bin/ar rcs ./build/lib.a build/obj/libsrc1.c.o build/obj/libsrc2.c.o
```
Re-run:
```sh
✅ build/obj/libsrc2.c.o is up to date.
✅ build/obj/libsrc1.c.o is up to date.
✅ 📦 ./build/lib.a is up to date.
```

### 4. Build an executable
This example also links a static library:

```sh
craftc exe -cc cc -cfiles "main.c" -objdir ./build/obj -cflags "-Wall -O2 -I ./libsrc" -lib-paths "./build/lib.a" -exe-path ./app
```
Output:
```sh
[build] 🧠 build/obj/main.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -O2 -I ./libsrc -MMD -MF build/obj/main.c.d -c ./main.c -o build/obj/main.c.o
[linking] 🔗 /usr/bin/cc build/obj/main.c.o ./build/lib.a -o ./build/app
```
Re-run:
```sh
✅ build/obj/main.c.o is up to date.
✅ 🚀 ./build/app is up to date.
```

### 5. Task integration

Each command can be used on its own, which makes `craftc` straightforward to
call from [Task](https://taskfile.dev/) or another task runner. This Taskfile
builds a static library before linking the executable:

```yaml
version: '3'

tasks:

  lib:
    desc: Build a static library with a few source files
    vars:
      BUILD_DIR: ./build/{{.TASK}}
      OBJ_DIR: '{{.BUILD_DIR}}/obj'
      SRC: ./libsrc/libsrc1.c ./libsrc/libsrc2.c
      STATIC_LIB: '{{.BUILD_DIR}}/{{.TASK}}.a'
      CFLAGS: -Wall -O2
    cmds:
      - ./craftc static-lib -cc cc -cfiles "{{.SRC}}" -objdir {{.OBJ_DIR}} -cflags "{{.CFLAGS}}" -lib-path "{{.STATIC_LIB}}" -i {{.CLI_ARGS}}


  exe:
    desc: Build an exe cmd with a few source files
    vars:
      BUILD_DIR: ./build/{{.TASK}}
      OBJ_DIR: '{{.BUILD_DIR}}/obj'
      SRC: ./appsrc/main.c
      STATIC_LIB: ./build/lib/lib.a
      CFLAGS: -Wall -O2 -I ./libsrc
      # LDFLAGS: -Wl,--trace
      APP_PATH: '{{.BUILD_DIR}}/{{.TASK}}-app'
    deps:
      - task: lib
    cmds:
      - >
        ./craftc exe
        -cc cc
        -cfiles "{{.SRC}}"
        -objdir {{.OBJ_DIR}}
        -cflags "{{.CFLAGS}}"
        -ldflags "{{.LDFLAGS}}"
        -exe-path {{.APP_PATH}}
        -lib-paths "{{.STATIC_LIB}}"
        -i {{.CLI_ARGS}}
```
Output:

```sh
[build] 🧠 build/lib/obj/libsrc.libsrc2.c.o: object file does not exist.
[build] 🧠 build/lib/obj/libsrc.libsrc1.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -O2 -MMD -MF build/lib/obj/libsrc.libsrc1.c.d -c ./libsrc/libsrc1.c -o build/lib/obj/libsrc.libsrc1.c.o
[compile] 🔨 /usr/bin/cc -Wall -O2 -MMD -MF build/test4/obj/libsrc.libsrc2.c.d -c ./libsrc/libsrc2.c -o build/lib/obj/libsrc.libsrc2.c.o
[archive] 📦 /usr/bin/ar rcs ./build/lib/lib.a build/lib/obj/libsrc.libsrc1.c.o build/lib/obj/libsrc.libsrc2.c.o
[build] 🧠 build/exe/obj/appsrc.main.c.o: object file does not exist.
[compile] 🔨 /usr/bin/cc -Wall -O2 -I ./libsrc -MMD -MF build/exe/obj/appsrc.main.c.d -c ./appsrc/main.c -o build/exe/obj/appsrc.main.c.o
[linking] 🔗 /usr/bin/cc build/exe/obj/appsrc.main.c.o ./build/lib/lib.a -o ./build/exe/app
```
Re-run:
```sh
✅ build/lib/obj/libsrc.libsrc2.c.o is up to date.
✅ build/lib/obj/libsrc.libsrc1.c.o is up to date.
✅ 📦 ./build/lib/lib.a is up to date.
✅ build/exe/obj/appsrc.main.c.o is up to date.
✅ 🚀 ./build/exe/app is up to date.
```
