package bundler

import (
	"io"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/depgraph"
)

// Mode defines the interface for different bundling strategies.
type Mode interface {
	// Generate produces the bundled output from resolved dependencies.
	Generate(ctx *BundleContext) error
}

// BundleContext holds the shared state for all bundle modes.
type BundleContext struct {
	Graph       *depgraph.Graph
	Order       []string // topologically sorted paths
	BaseDir     string
	Shebang     string
	Minify      bool
	Output      io.Writer
	DepsDir     string // deps directory path (for seira_path support)
	HasDeps     bool   // whether a deps directory exists
	Type        string // "executable" or "library"
	LibraryName string // library name (used in library mode header)
	Exports     []string // library mode: exported function names
}

// relativePath returns a relative path from base to target, falling back to target on error.
func relativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
