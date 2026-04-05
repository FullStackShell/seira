package bundler

import (
	"io"

	"github.com/Hayao0819/seira/internal/depgraph"
)

// Mode defines the interface for different bundling strategies.
type Mode interface {
	// Generate produces the bundled output from resolved dependencies.
	Generate(ctx *BundleContext) error
}

// BundleContext holds the shared state for all bundle modes.
type BundleContext struct {
	Graph   *depgraph.Graph
	Order   []string // topologically sorted paths
	BaseDir string
	Shebang string
	Minify  bool
	Output  io.Writer
	DepsDir string // deps directory path (for seira_path support)
	HasDeps bool   // whether a deps directory exists
}
