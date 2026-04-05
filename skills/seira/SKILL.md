---
name: seira
description: "Seira shell script framework context. Use when working with seira projects (.seirarc.json), shell script bundling, source dependency resolution, or seira CLI commands."
user-invocable: false
---

# Seira - Shell Script Framework & Bundler

Seira is a Go-based shell script framework and bundler. It resolves `source` dependencies recursively (including `${VAR=default}` variable expansion) and bundles scripts into standalone executables.

Repository: https://github.com/Hayao0819/seira

## CLI Commands

```bash
# Bundle a shell script project into a standalone executable
seira bundle <input.sh> -o output.sh [--mode concat|tarball] [--minify] [--shebang /bin/bash]

# Install a bpkg-compatible package from GitHub
seira install <user/package[@version]> [--save]

# Install all dependencies from .seirarc.json
seira deps

# Resolve a package file path in deps/
seira path <owner/repo/path>

# Parse and display the AST of a shell script
seira ast <file.sh>

# Create a new seira project with scaffolding
seira new <project-name>
```

## Bundle Modes

- **concat** (recommended): Pure single-file shell script. Function definitions are placed at the top, side effects follow in topological order, source lines are stripped, and `main "$@"` is appended.
- **tarball**: Self-extracting archive. Files are tarred, gzipped, base64-encoded into a self-extracting shell script. Requires temp dir extraction at runtime.

## Project Configuration (.seirarc.json)

```json
{
  "entrypoint": "main.sh",
  "shebang": "/bin/bash",
  "mode": "concat",
  "env": {},
  "include": [],
  "exclude": [],
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
| `env` | Variables for source path resolution at build time | `{}` |
| `include` | Extra files to bundle even if not discovered by resolver | `[]` |
| `exclude` | Files to exclude from bundling | `[]` |
| `dependencies` | bpkg-style dependencies (`"user/name": "version"`) | `{}` |
| `deps_dir` | Dependencies directory | `deps` |

## Entrypoint Requirements

The entrypoint script **must** declare a `main()` function. Seira calls `main "$@"` at the end of the bundled output.

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

## Package Management (bpkg)

Seira integrates with the bpkg ecosystem. Install packages from GitHub:

```bash
seira install bpkg/term          # latest (master branch)
seira install user/pkg@v1.0.0    # specific version (git tag)
seira install user/pkg --save    # save to .seirarc.json
seira deps                       # install all from config
```

Installed packages go to `deps/<name>/` with symlinks in `deps/bin/`.

### Accessing Package Files (seira_path)

Use `seira_path` in scripts to access files from installed packages:

```bash
# Resolve owner/repo/path to the actual file path
cat "$(seira_path owner/repo/data.txt)"
source "$(seira_path bpkg/term/term.sh)"
```

The `seira_path` function is automatically embedded when bundling projects that have a `deps/` directory.

## Typical Project Structure

```
my-project/
  .seirarc.json       # Project configuration
  main.sh             # Entrypoint (must have main() function)
  lib/                # Library scripts (sourced by main.sh)
    helper.sh
    utils.sh
  deps/               # Installed packages (via seira install)
    bin/              # Symlinks to package scripts
    term/             # Installed bpkg package
      term.sh
```
