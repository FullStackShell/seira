# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Seira is a Go-based shell script framework and bundler. It resolves `source` dependencies recursively (including `${VAR=default}` variable expansion) and bundles scripts into standalone executables via two modes: tarball (self-extracting archive) and concat (single pure shell script).

Repository: https://github.com/Hayao0819/seira

## Build & Test Commands

```bash
# Build
go build -o seira ./main.go

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/shellparse/ -run TestEvaluateSourcePath

# CLI usage
go run ./main.go bundle <input.sh> -o output.sh --shebang /bin/bash --mode concat
go run ./main.go bundle <input.sh> -o output.sh --mode tarball -m
go run ./main.go ast <file.sh>
go run ./main.go new <project-name>
```

## Architecture

### Package Structure (internal/ based)

- **`internal/config/`** — `.seirarc.json` loading with upward directory search. `Load(dir)` walks up; `Default()` returns sensible defaults.
- **`internal/shellparse/`** — Shell script analysis using `mvdan.cc/sh/v3`. `Parser` struct (no global state). `Analyze()` extracts `SourceRef` and top-level functions. `ClassifyStmts()` classifies statements into `StmtFunc`, `StmtSource`, `StmtEffect`.
- **`internal/depgraph/`** — DAG for dependency resolution. `Resolve()` builds graph via BFS. `TopologicalSort()` orders files. `DetectCycles()` returns `CycleError`.
- **`internal/bundler/`** — Pipeline orchestrator with `Mode` interface. Two modes:
  - **tarball** (`TarballMode`): copy → minify → tar.gz → base64 → self-extracting template
  - **concat** (`ConcatMode`): classify stmts → functions to top → side effects in topo order → remove source lines → single script
- **`cmd/`** — CLI layer (Cobra + nahi). Commands: `ast`, `bundle` (`--mode`, `--minify`, `--shebang`), `new`.
- **`internal/shell/`** — `EvalSh()` runs shell code via `sh -c`.
- **`assets/`** — `execute.sh` template embedded via `//go:embed`. Used by tarball mode.

### Bundle Modes

**tarball** (default): Files are tarred, gzipped, base64-encoded into a self-extracting shell script. Requires temp dir extraction at runtime.

**concat**: Produces a pure single-file shell script. Function definitions from all dependencies are placed at the top, side effects (variable assignments, commands) follow in topological order, source lines are stripped, and `main "$@"` is appended.

### Key Dependencies

- `mvdan.cc/sh/v3` — Shell parser and printer (minification)
- `github.com/spf13/cobra` — CLI framework
- `github.com/Hayao0819/nahi` — Cobra utilities (maintained by project author)
- `github.com/cockroachdb/errors` — Error wrapping with stack traces
- `github.com/samber/lo` — Generic utility functions
- `github.com/m-mizutani/clog` — Structured logging

### Configuration (.seirarc.json)

```json
{"entrypoint": "main.sh", "shebang": "/bin/bash", "mode": "concat", "env": {}, "include": [], "exclude": []}
```

- `mode`: "tarball" or "concat" (CLI `--mode` flag overrides)
- `env`: Variables for source path resolution at build time
- `include`: Extra files to bundle even if not discovered by resolver

## Language

Go 1.24. Target scripts are Bash/POSIX shell.
