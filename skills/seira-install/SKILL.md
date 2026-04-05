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

# Install all dependencies from .seirarc.json
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
- Installs transitive dependencies recursively

### Non-bpkg repositories
- Falls back to `git clone --depth 1`
- Copies all files (excluding `.git/`)
- Auto-detects `.sh` files for `deps/bin/` symlinks

## Installed Structure

```
deps/
  bin/              # Symlinks (script.sh → ../pkg/script.sh, extension stripped)
  <package-name>/   # Package contents
    bpkg.json       # Manifest (if bpkg-compatible)
    *.sh            # Script files
```

## Using Packages in Scripts

```bash
# Source directly
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
    "user/package": "master"
  },
  "deps_dir": "deps"
}
```

## Notes

- The `seira_path` shell function is automatically embedded in bundled output when `deps/` exists
- `SEIRA_DEPS_DIR` env var overrides the deps directory at runtime
- `--save` flag writes the dependency to `.seirarc.json` in the current directory
