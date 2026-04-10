---
name: seira-doc
description: "Generate documentation from shell script doc comments with seira. Use when the user wants to generate docs, API references, or documentation from shell scripts."
allowed-tools: Bash Read Grep Glob Edit Write
argument-hint: "[input.sh] [--format markdown|html] [--output file]"
---

# Seira Doc

Generate documentation from shell script projects using shdoc-compatible annotations.

## Task

Generate documentation for the script or project specified by the user. If no script is specified, look for `.seirarc.json` in the current directory to determine the entrypoint.

## Commands

```bash
# Generate Markdown documentation (default)
seira doc $ARGUMENTS

# Generate HTML documentation
seira doc <input.sh> --format html -o docs/api.html

# Generate Markdown to a file
seira doc <input.sh> --format markdown -o docs/api.md

# Use project config (reads entrypoint from .seirarc.json)
seira doc -f html -o docs/index.html

# Custom title
seira doc --title "My API Reference" -f html -o docs/index.html
```

## Supported Annotations

Document shell scripts using comment blocks immediately before function declarations:

### File-Level Tags

```bash
# @file mylib.sh
# @brief A utility library for string operations
# @description This library provides functions for
# common string manipulation tasks.
```

| Tag | Description |
|-----|-------------|
| `@file` | File name/title |
| `@brief` | One-line summary |
| `@description` | Detailed description (multi-line) |

### Function-Level Tags

```bash
# @description Convert a string to uppercase
# @param $1 Input string
# @option -n --no-newline Suppress trailing newline
# @stdout The uppercased string
# @exitcode 0 Success
# @exitcode 1 Empty input
# @example
#   str_upper "hello"
#   # Output: HELLO
# @see str_lower
str_upper() {
    echo "${1^^}"
}
```

| Tag | Description |
|-----|-------------|
| `@description` | Function description (multi-line) |
| `@param` / `@arg` | Parameter: `@param $1 Description` |
| `@option` | Flag/option: `@option -f --force Description` |
| `@stdin` | Expected stdin input |
| `@stdout` | What is written to stdout |
| `@return` | Return value description |
| `@exitcode` | Exit code: `@exitcode 0 Success` |
| `@example` | Code example block (multi-line) |
| `@set` | Global variable set: `@set VAR Description` |
| `@see` | Cross-reference to another function |
| `@internal` | Mark function as internal (hidden from output) |
| `@noargs` | Marker indicating function takes no arguments |
| `@section` | Group subsequent functions under a named section |

### Sections

Use `@section` in a comment before a function to group functions:

```bash
# @section String Functions

# @description Convert to uppercase
str_upper() { ... }

# @description Convert to lowercase
str_lower() { ... }

# @section Math Functions

# @description Add two numbers
math_add() { ... }
```

## Output Formats

### Markdown (`--format markdown` or `--format md`)
- GitHub-flavored Markdown
- Table of contents for multi-file projects
- Code blocks for examples
- Default format when no `--format` is specified

### HTML (`--format html`)
- Self-contained HTML with embedded CSS
- Dark mode support (via `prefers-color-scheme`)
- Navigation sidebar for multi-file projects
- Function index with anchor links

## Multi-File Projects

When the input file sources other scripts, seira resolves the full dependency graph and generates documentation for all files that contain doc comments. Files are ordered topologically (dependencies first).

## Internal Functions

Functions marked with `@internal` are excluded from the generated documentation. Use this for private helper functions:

```bash
# @description Internal helper
# @internal
_validate_input() {
    [[ -n "$1" ]]
}
```

## Configuration

The `doc` command respects `.seirarc.json`:
- `entrypoint`: Used when no input file is specified
- `env`: Variables for source path resolution
- `include`: Extra files to include in documentation

## Troubleshooting

- **"No documented files found"**: Add `@description`, `@file`, or other doc tags to your scripts
- **Missing functions**: Ensure doc comments are placed immediately before the function declaration
- **"no input file specified"**: Either pass a script path or create `.seirarc.json` with an `entrypoint`
