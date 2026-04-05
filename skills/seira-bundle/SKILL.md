---
name: seira-bundle
description: "Bundle shell scripts with seira into standalone executables or libraries. Use when the user wants to bundle, build, or package shell scripts, or configure bundle options."
allowed-tools: Bash Read Grep Glob Edit Write
argument-hint: "[input.sh] [--mode concat|tarball] [--type executable|library]"
---

# Seira Bundle

Bundle shell script project into a standalone executable or reusable library.

## Task

Bundle the script specified by the user. If no script is specified, look for `.seirarc.json` in the current directory to determine the entrypoint.

## Steps

1. Check if `.seirarc.json` exists for project configuration
2. Identify the entrypoint script (from argument or config)
3. For executable type: verify the entrypoint has a `main()` function
4. Run the bundle command with appropriate options

## Commands

```bash
# Basic bundle (uses .seirarc.json if available)
seira bundle $ARGUMENTS

# Concat mode (recommended - pure single-file output)
seira bundle <input.sh> -o output.sh --mode concat --shebang /bin/bash

# Tarball mode (self-extracting archive)
seira bundle <input.sh> -o output.sh --mode tarball

# Library mode (sourceable output, no main required)
seira bundle <input.sh> -o output.sh --type library

# With minification
seira bundle <input.sh> -o output.sh --mode concat --minify
```

## Project Types

### executable (default)
Produces a standalone executable script. Entrypoint must declare `main()`.

### library
Produces a sourceable script for distribution. Set via `--type library` or `"type": "library"` in config.
- No `main()` function required
- No shebang line in output
- No `main "$@"` at the end
- Metadata headers: `# seira-library:` and `# seira-exports:`
- Optional `exports` filter to include only specified functions

## Bundle Modes

### concat (recommended)
Produces a pure single-file shell script:
- Function definitions from all dependencies placed at the top
- Side effects (variable assignments, commands) in topological order
- Source lines removed
- `main "$@"` appended at the end (executable type only)
- `seira_path()` function injected if `deps/` directory exists

### tarball
Produces a self-extracting archive:
- All dependency files tarred + gzipped + base64-encoded
- Wrapper script extracts to temp dir and executes
- `deps/` directory included automatically if present
- Cleanup on exit via trap

## Library Auto-Loading

At bundle time, seira automatically scans `deps/` for library projects (`.seirarc.json` with `"type": "library"`). Their functions are inlined into the output without needing explicit `source` statements. Install a library with `seira install`, and its functions become available immediately.

## Entrypoint Requirements

For **executable** type, the entrypoint script **must** declare a `main()` function:

```bash
#!/usr/bin/env bash
source lib/utils.sh

main() {
    echo "Hello"
}
```

For **library** type, no `main()` is needed:

```bash
#!/usr/bin/env bash
source lib/strutils.sh
# Functions from strutils.sh are exported
```

## Configuration (.seirarc.json)

CLI flags override config values:

```json
{
  "entrypoint": "main.sh",
  "shebang": "/bin/bash",
  "mode": "concat",
  "type": "executable",
  "env": {"LIB_DIR": "lib"},
  "include": ["extra/config.sh"],
  "exports": ["func_a", "func_b"]
}
```

## Troubleshooting

- **"does not declare a main() function"**: Add `main() { ... }` to the entrypoint, or use `--type library` for library projects
- **"sourced file not found"**: Check relative paths from the source file's directory
- **Cycle detected**: Break circular `source` dependencies
- **Dynamic path warning**: `${VAR}` without default cannot be resolved statically; use `${VAR:=default}` or add to `env` in config
