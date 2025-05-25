
# ⚒️ craftC

> A fast, minimal C build tool inspired by Taskfile & Make. Designed for speed, clarity, and cross-platform simplicity.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ameergituser/craftc)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ameergituser/craftc/go.yml)
![GitHub](https://img.shields.io/github/license/ameergituser/craftc)

---

## 🚀 Features

- ⚡ Lightning-fast incremental builds
- 🧠 Smart recompilation detection with reasoning
- 🔨 C compiler integration with support for custom flags
- 📦 Static library archiving
- 🔗 Application linking
- ✅ Clear verbose output with emoji feedback
- 🧩 Simple CLI
- 🧰 Cross-platform: Linux, (macOS, Windows not tested yet)

---

## 📦 Installation

Install with Go:

```sh
go install github.com/ameergituser/craftc@latest
```
Or clone and build manually:
```sh
git clone https://github.com/ameergituser/craftc
cd craftc
go build .
```
## ✅ Sample output
Using the verbose option:
```sh
⚙️ Running cmd static-lib: ./build/test4/test4.a
✅ build/test4/obj/libsrc.libsrc1.c.o is up to date.
✅ build/test4/obj/libsrc.libsrc2.c.o is up to date.
✅ 📦 ./build/test4/test4.a is up to date.
⚙️ Running cmd app: ./build/test6/test6-app
✅ build/test6/obj/app-src.main.c.o is up to date.
✅ 🚀 ./build/test6/test6-app is up to date.
```