package bundler

import (
	"io"
	"log/slog"
	"path/filepath"
	"strings"

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

// namespaceAlias represents a short-name alias for a namespaced function.
type namespaceAlias struct {
	fullName  string // e.g., "oop::new"
	shortName string // e.g., "new"
}

// resolveNamespaceAliases builds alias entries for @seira:using directives.
// For each using namespace, it finds all functions matching "namespace::*" and
// creates short-name aliases, skipping any that collide with existing function names.
func resolveNamespaceAliases(funcs []funcBlock, usingNamespaces map[string]bool) []namespaceAlias {
	if len(usingNamespaces) == 0 {
		return nil
	}

	// Build set of existing non-namespaced function names for collision detection
	existingNames := map[string]bool{}
	for _, f := range funcs {
		if !strings.Contains(f.name, "::") {
			existingNames[f.name] = true
		}
	}

	var aliases []namespaceAlias
	aliasNames := map[string]bool{} // track generated aliases to avoid duplicates
	for _, f := range funcs {
		ns, shortName, ok := strings.Cut(f.name, "::")
		if !ok || !usingNamespaces[ns] {
			continue
		}
		if existingNames[shortName] {
			slog.Warn("namespace alias skipped: name collision with existing function",
				"alias", shortName, "function", f.name)
			continue
		}
		if aliasNames[shortName] {
			slog.Warn("namespace alias skipped: duplicate short name from another namespace",
				"alias", shortName, "function", f.name)
			continue
		}
		aliasNames[shortName] = true
		aliases = append(aliases, namespaceAlias{fullName: f.name, shortName: shortName})
	}
	return aliases
}

// relativePath returns a relative path from base to target, falling back to target on error.
func relativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
