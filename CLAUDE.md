# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Seira is a Go tool that bundles shell scripts into distributable standalone executables. It resolves `source` dependencies recursively (including `${VAR=default}` variable expansion), packages all files into a gzipped tarball, and generates a self-extracting shell script.

Repository: https://github.com/Hayao0819/seira

## Build & Test Commands

```bash
# Build
go build -o seira ./main.go

# Run all tests
go test ./...

# Run a single package's tests
go test ./shellparse/ -run TestEvaluateSourcePath

# CLI usage
go run ./main.go bundle <input.sh> -o output.sh --shebang /bin/bash -m
go run ./main.go ast <file.sh>
go run ./main.go new <project-name>
```

## Architecture

### Package Structure

- **`config/`** — `.seirarc.json` loading with upward directory search. `Load(dir)` walks up from dir to find config; `Default()` returns sensible defaults.
- **`shellparse/`** — Shell script analysis using `mvdan.cc/sh/v3` parser. `Parser` struct (no global state). `Analyze()` extracts `SourceRef` (with variable default-value expansion) and top-level functions. Key: `evaluateSourcePathParts()` handles `${VAR=default}` / `${VAR:-default}` resolution statically.
- **`depgraph/`** — Directed acyclic graph for dependency resolution. `Resolve()` builds graph from entrypoint via BFS. `TopologicalSort()` orders files. `DetectCycles()` returns cycle paths as `CycleError`.
- **`bundler/`** — Pipeline orchestrator with `Config` struct (not functional options). `Bundle()` pipeline: resolve deps → validate main() → topo sort → copy files → minify (optional) → create tar.gz → render output template.
- **`cmd/`** — CLI layer (Cobra). Flat structure (no sub-packages). Commands: `ast`, `bundle`, `new`. `--verbose/-v` flag controls log level (default: Warn).
- **`shell/`** — `EvalSh()` runs shell code via `sh -c`.
- **`assets/`** — `execute.sh` template (embedded via `//go:embed`). Runtime bootstrap: extract tarball to temp dir, cd into it, source entrypoint, call main().

### Bundling Pipeline

```
Input script → Parse AST → Resolve source deps (BFS, variable expansion)
→ Detect cycles (error) → Topological sort → Copy to work dir
→ Minify (optional, syntax.Printer) → Create tar.gz → Base64 encode
→ Render execute.sh template → Write output
```

### Key Dependencies

- `mvdan.cc/sh/v3` — Shell parser (core engine, also used for minification)
- `github.com/spf13/cobra` — CLI framework
- `github.com/samber/lo` — Generic utility functions
- `github.com/m-mizutani/clog` — Structured logging

### Configuration (.seirarc.json)

```json
{"entrypoint": "main.sh", "shebang": "/bin/bash", "env": {}, "include": [], "exclude": []}
```

- `env`: Variables for source path resolution at build time
- `include`: Extra files to bundle even if not discovered by dependency resolver
- Config is auto-discovered by walking parent directories

## Language

Go 1.24. Target scripts are Bash/POSIX shell.
