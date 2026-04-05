---
name: seira-install
description: "Install bpkg-compatible shell script packages with seira. Use when the user wants to add dependencies, install packages, or manage deps."
allowed-tools: Bash Read Grep Glob Edit Write
argument-hint: "<user/package[@version]> [--save]"
---

# Seira Package Install

Install bpkg-compatible packages from GitHub into the project's `deps/` directory.

## Task

Install the package specified by the user. If the user asks to install all dependencies, use `seira deps`.

## Commands

```bash
# Install a specific package
seira install $ARGUMENTS

# Install and save to .seirarc.json
seira install <user/package[@version]> --save

# Install all dependencies from .seirarc.json (uses lockfile for pinned versions)
seira deps

# Resolve a package file path
seira path <owner/repo/path>
```

## Package Reference Format

```
user/package            # latest (master branch)
user/package@v1.0.0     # specific git tag
user/package@main       # specific branch
```

## Installation Behavior

### bpkg-compatible packages (have bpkg.json)
- Downloads only files listed in `scripts` and `files` arrays
- Creates symlinks in `deps/bin/`
- Installs transitive bpkg dependencies recursively
- Resolves commit hash via GitHub API

### Non-bpkg repositories
- Falls back to `git clone --depth 1`
- Copies all files (excluding `.git/`)
- Auto-detects `.sh` files for `deps/bin/` symlinks
- Records HEAD commit hash from clone

### Seira project detection
After installing any package, seira checks if it is a seira project (has `.seirarc.json`). If the project has `dependencies`, they are recursively installed as well. This enables transitive dependency resolution across seira library projects.

## Lockfile (.seira-lock.json)

Every install operation updates `.seira-lock.json` with the exact git commit hash for each dependency:

```json
{
  "dependencies": {
    "user/repo": {
      "version": "v1.0",
      "commit": "abc123def456789..."
    },
    "user/other": {
      "version": "",
      "commit": "def789abc123456..."
    }
  }
}
```

- `version`: The tag/branch specified at install time (empty if default/master)
- `commit`: The exact git commit SHA resolved at install time

When `seira deps` runs and a lockfile exists, it uses the locked commit hashes to ensure reproducible installs. This guarantees that all developers get the same dependency versions.

## Library Dependencies

If an installed package has `.seirarc.json` with `"type": "library"`, its functions are **automatically available** at bundle time — no `source` statement needed in your scripts. The bundler scans `deps/` for library projects and inlines their entrypoints into the dependency graph.

```bash
# Install a seira library
seira install user/strutils --save

# Use its functions directly in main.sh (no source needed)
main() {
    str_upper "hello"  # from strutils library
}
```

## Installed Structure

```
deps/
  bin/              # Symlinks (script.sh → ../pkg/script.sh, extension stripped)
  <package-name>/   # Package contents
    .seirarc.json   # If seira project (enables auto-loading for libraries)
    bpkg.json       # Manifest (if bpkg-compatible)
    *.sh            # Script files
```

## Using Packages in Scripts

```bash
# For library deps: just use the functions (auto-loaded at bundle time)
str_upper "hello"

# For non-library deps: source directly
source ./deps/term/term.sh

# Use seira_path for owner/repo/path resolution
cat "$(seira_path bpkg/term/term.sh)"

# Add deps/bin to PATH for direct execution
export PATH="./deps/bin:$PATH"
term
```

## Configuration (.seirarc.json)

```json
{
  "dependencies": {
    "bpkg/term": "0.1.1",
    "user/strutils": "v1.0"
  },
  "deps_dir": "deps"
}
```

## Notes

- The `seira_path` shell function is automatically embedded in bundled output when `deps/` exists
- `SEIRA_DEPS_DIR` env var overrides the deps directory at runtime
- `--save` flag writes the dependency to `.seirarc.json` in the current directory
- `.seira-lock.json` should be committed to version control for reproducible builds
