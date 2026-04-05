---
name: seira-bundle
description: "Bundle shell scripts with seira into standalone executables. Use when the user wants to bundle, build, or package shell scripts, or configure bundle options."
allowed-tools: Bash Read Grep Glob Edit Write
argument-hint: "[input.sh] [--mode concat|tarball]"
---

# Seira Bundle

Bundle shell script project into a standalone executable.

## Task

Bundle the script specified by the user. If no script is specified, look for `.seirarc.json` in the current directory to determine the entrypoint.

## Steps

1. Check if `.seirarc.json` exists for project configuration
2. Identify the entrypoint script (from argument or config)
3. Verify the entrypoint has a `main()` function
4. Run the bundle command with appropriate options

## Commands

```bash
# Basic bundle (uses .seirarc.json if available)
seira bundle $ARGUMENTS

# Concat mode (recommended - pure single-file output)
seira bundle <input.sh> -o output.sh --mode concat --shebang /bin/bash

# Tarball mode (self-extracting archive)
seira bundle <input.sh> -o output.sh --mode tarball

# With minification
seira bundle <input.sh> -o output.sh --mode concat --minify
```

## Bundle Modes

### concat (recommended)
Produces a pure single-file shell script:
- Function definitions from all dependencies placed at the top
- Side effects (variable assignments, commands) in topological order
- Source lines removed
- `main "$@"` appended at the end
- `seira_path()` function injected if `deps/` directory exists

### tarball
Produces a self-extracting archive:
- All dependency files tarred + gzipped + base64-encoded
- Wrapper script extracts to temp dir and executes
- `deps/` directory included automatically if present
- Cleanup on exit via trap

## Entrypoint Requirements

The entrypoint script **must** declare a `main()` function:

```bash
#!/usr/bin/env bash
source lib/utils.sh

main() {
    echo "Hello"
}
```

## Configuration (.seirarc.json)

CLI flags override config values:

```json
{
  "entrypoint": "main.sh",
  "shebang": "/bin/bash",
  "mode": "concat",
  "env": {"LIB_DIR": "lib"},
  "include": ["extra/config.sh"]
}
```

## Troubleshooting

- **"does not declare a main() function"**: Add `main() { ... }` to the entrypoint
- **"sourced file not found"**: Check relative paths from the source file's directory
- **Cycle detected**: Break circular `source` dependencies
- **Dynamic path warning**: `${VAR}` without default cannot be resolved statically; use `${VAR:=default}` or add to `env` in config
