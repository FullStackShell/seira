# Seira Architecture

## Package Structure

All implementation packages live under `internal/`:

```
internal/
  bundler/      Pipeline orchestrator with Mode interface
  config/       .seirarc.json loading with upward directory search
  depgraph/     DAG for dependency resolution (BFS + topological sort + cycle detection)
  shellparse/   Shell script analysis using mvdan.cc/sh/v3
  shell/        Shell execution utilities (EvalSh)
  bpkg/         bpkg package manager integration (manifest parsing, GitHub install)
```

## Key Dependencies

- `mvdan.cc/sh/v3` - Shell parser and printer (AST, minification)
- `github.com/spf13/cobra` - CLI framework
- `github.com/Hayao0819/nahi` - Cobra utilities (maintained by project author, DO NOT remove)
- `github.com/cockroachdb/errors` - Error wrapping with stack traces (DO NOT remove)
- `github.com/samber/lo` - Generic utility functions
- `github.com/m-mizutani/clog` - Structured logging

## Data Flow

1. **Entry**: `cmd/bundle.go` loads config and CLI flags
2. **Orchestration**: `bundler.New(cfg).Bundle()`
3. **Graph Building**: `depgraph.Resolve()` recursively parses scripts via `shellparse.Parser`
4. **Sorting**: `graph.TopologicalSort()` produces dependency-first file order
5. **Mode Dispatch**: `resolveMode()` returns `TarballMode` or `ConcatMode`
6. **Generation**: Mode-specific `Generate(ctx)` produces the output

## Bundle Mode Interface

```go
type Mode interface {
    Generate(ctx *BundleContext) error
}
```

`BundleContext` holds: Graph, Order (topo-sorted paths), BaseDir, Shebang, Minify, Output, DepsDir, HasDeps.

## Statement Classification (concat mode)

`shellparse.ClassifyStmts()` categorizes top-level statements:
- `StmtFunc` - Function definitions (placed at top of output)
- `StmtSource` - Source/. commands (removed, deps already inlined)
- `StmtEffect` - Side effects: variable assignments, commands (preserved in topo order)

## Shell Parser

`shellparse.Parser` wraps `mvdan.cc/sh/v3/syntax.Parser`. `Analyze()` extracts:
- `SourceRef` - source/. targets with Raw, Resolved, IsDynamic fields
- Top-level function names

`EvaluateSourcePath()` handles `${VAR=default}`, `${VAR:-default}` expansion for build-time resolution.

## Config Loading

`config.Load(dir)` walks upward from `dir` to find `.seirarc.json`. Returns `Default()` if not found (no error).

## Long-term Vision

Seira aims to be a comprehensive shell script development framework, not just a bundler. Planned features include:
- Tree-sitter integration
- Package manager
- Test framework
- Object-oriented programming support
