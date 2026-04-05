package depgraph

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/shellparse"
)

// Resolve builds a dependency graph starting from the given entrypoint.
// env provides variable values for source path resolution.
func Resolve(parser *shellparse.Parser, entrypoint string, env map[string]string) (*Graph, error) {
	absEntry, err := filepath.Abs(entrypoint)
	if err != nil {
		return nil, err
	}

	g := New()
	if err := resolveFile(g, parser, absEntry, env); err != nil {
		return nil, err
	}

	cycles := g.DetectCycles()
	if len(cycles) > 0 {
		return nil, &CycleError{Cycles: cycles}
	}

	return g, nil
}

// ResolveAdditional resolves a file and its dependencies into an existing graph.
// Use this to inject additional entrypoints (e.g. library deps) after the initial Resolve.
func ResolveAdditional(g *Graph, parser *shellparse.Parser, absPath string, env map[string]string) error {
	return resolveFile(g, parser, absPath, env)
}

func resolveFile(g *Graph, parser *shellparse.Parser, absPath string, env map[string]string) error {
	if g.HasNode(absPath) {
		return nil
	}

	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	script, err := parser.Analyze(f, absPath)
	f.Close() // close immediately, not defer in a loop
	if err != nil {
		return err
	}

	g.AddNode(absPath, script)

	baseDir := filepath.Dir(absPath)
	for _, src := range script.Sources {
		// Skip sources marked with @seira:ignore directive
		if src.Directives.Has("ignore") {
			slog.Info("source ignored by directive", "file", absPath, "source", src.Raw)
			continue
		}

		// Re-evaluate with the provided env
		ref := shellparse.EvaluateSourcePath(src, env)

		if ref.IsDynamic {
			slog.Warn("cannot resolve dynamic source path statically, skipping",
				"file", absPath, "source", ref.Raw)
			continue
		}

		depPath := ref.Resolved
		if !filepath.IsAbs(depPath) {
			depPath = filepath.Join(baseDir, depPath)
		}
		depPath = filepath.Clean(depPath)

		// Check if the dependency file exists
		if _, err := os.Stat(depPath); os.IsNotExist(err) {
			slog.Warn("sourced file not found, skipping",
				"file", absPath, "source", depPath)
			continue
		}

		g.AddEdge(absPath, depPath)

		if err := resolveFile(g, parser, depPath, env); err != nil {
			return err
		}
	}

	return nil
}
