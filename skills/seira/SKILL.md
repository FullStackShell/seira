---
name: seira
description: "Seira shell script framework context. Use when working with seira projects (.seirarc.json), shell script bundling, source dependency resolution, or seira CLI commands."
user-invocable: false
---

# Seira - Shell Script Framework & Bundler

Seira is a Go-based shell script framework and bundler. It resolves `source` dependencies recursively (including `${VAR=default}` variable expansion) and bundles scripts into standalone executables or reusable libraries.

Repository: https://github.com/Hayao0819/seira

## CLI Commands

```bash
# Bundle a shell script project into a standalone executable
seira bundle <input.sh> -o output.sh [--mode concat|tarball] [--type executable|library] [--minify] [--shebang /bin/bash]

# Install a bpkg-compatible package from GitHub
seira install <user/package[@version]> [--save]

# Install all dependencies from .seirarc.json
seira deps

# Resolve a package file path in deps/
seira path <owner/repo/path>

# Parse and display the AST of a shell script
seira ast <file.sh>

# Create a new seira project with scaffolding
seira new <project-name> [--library]
```

## Project Types

### executable (default)
Standard shell script project. Entrypoint must declare a `main()` function. Output is a standalone executable script.

### library
Reusable shell script library. No `main()` required. Output is a sourceable script containing only function definitions and side effects, without shebang or `main "$@"`.

Set via `"type": "library"` in `.seirarc.json` or `--type library` CLI flag.

## Bundle Modes

- **concat** (recommended): Pure single-file shell script. Function definitions are placed at the top, side effects follow in topological order, source lines are stripped, and `main "$@"` is appended (executable type only).
- **tarball**: Self-extracting archive. Files are tarred, gzipped, base64-encoded into a self-extracting shell script. Requires temp dir extraction at runtime.
- **library**: Automatically selected when `type` is `"library"`. Outputs a sourceable script with function definitions and side effects, plus metadata headers (`# seira-library:`, `# seira-exports:`).

## Project Configuration (.seirarc.json)

```json
{
  "entrypoint": "main.sh",
  "shebang": "/bin/bash",
  "mode": "concat",
  "type": "executable",
  "env": {},
  "include": [],
  "exclude": [],
  "exports": [],
  "dependencies": {
    "user/package": "version"
  },
  "deps_dir": "deps"
}
```

| Field | Description | Default |
|-------|-------------|---------|
| `entrypoint` | Main script file | - |
| `shebang` | Shebang line for output | `/bin/sh` |
| `mode` | Bundle mode (`concat` or `tarball`) | `tarball` |
| `type` | Project type (`executable` or `library`) | `executable` |
| `env` | Variables for source path resolution at build time | `{}` |
| `include` | Extra files to bundle even if not discovered by resolver | `[]` |
| `exclude` | Files to exclude from bundling | `[]` |
| `exports` | Library mode: function names to export (empty = all) | `[]` |
| `dependencies` | bpkg-style dependencies (`"user/name": "version"`) | `{}` |
| `deps_dir` | Dependencies directory | `deps` |

## Entrypoint Requirements

For **executable** projects, the entrypoint **must** declare a `main()` function. For **library** projects, no `main()` is needed.

```bash
#!/usr/bin/env bash
source lib/helper.sh

main() {
    greet "World"
}
```

## Source Resolution

Seira resolves `source` and `.` commands recursively, building a dependency graph (DAG). It supports:

- Relative paths: `source ./lib/helper.sh`
- Variable expansion with defaults: `source "${LIB_DIR:=lib}/helper.sh"`
- Cycle detection (errors if circular dependencies found)
- Dynamic paths that cannot be statically resolved are warned and skipped

## Library Auto-Loading from deps/

At bundle time, seira scans the `deps/` directory for seira library projects (directories with `.seirarc.json` containing `"type": "library"`). Their entrypoints are automatically added to the dependency graph — **no `source` statement needed in consumer scripts**. Library functions are inlined directly into the bundle output.

```
my-project/
  .seirarc.json
  main.sh              # Can use str_upper() without sourcing anything
  deps/
    strutils/          # Installed via: seira install user/strutils
      .seirarc.json    # {"type": "library", "entrypoint": "index.sh"}
      index.sh
      lib/strutils.sh  # str_upper(), str_lower()
```

## Package Management (bpkg)

Seira integrates with the bpkg ecosystem. Install packages from GitHub:

```bash
seira install bpkg/term          # latest (master branch)
seira install user/pkg@v1.0.0    # specific version (git tag)
seira install user/pkg --save    # save to .seirarc.json
seira deps                       # install all from config
```

Installed packages go to `deps/<name>/` with symlinks in `deps/bin/`.

### Seira Project Detection

When installing a package that is a seira project (has `.seirarc.json` with `dependencies`), seira recursively installs those dependencies as well.

### Lockfile (.seira-lock.json)

Every `seira install` and `seira deps` command records all installed dependencies' git commit hashes in `.seira-lock.json`:

```json
{
  "dependencies": {
    "user/repo": {
      "version": "v1.0",
      "commit": "abc123def456..."
    }
  }
}
```

When `seira deps` runs with an existing lockfile, it uses the locked commit hashes to ensure reproducible installs.

### Accessing Package Files (seira_path)

Use `seira_path` in scripts to access files from installed packages:

```bash
cat "$(seira_path owner/repo/data.txt)"
source "$(seira_path bpkg/term/term.sh)"
```

The `seira_path` function is automatically embedded when bundling projects that have a `deps/` directory.

## Typical Project Structures

### Executable Project

```
my-project/
  .seirarc.json       # {"entrypoint": "main.sh", "mode": "concat"}
  main.sh             # Entrypoint (must have main() function)
  lib/                # Library scripts (sourced by main.sh)
    helper.sh
  deps/               # Installed packages (via seira install)
    bin/
    term/
```

### Library Project

```
my-library/
  .seirarc.json       # {"type": "library", "entrypoint": "index.sh"}
  index.sh            # Sources internal modules
  lib/
    utils.sh          # Exported functions
```
