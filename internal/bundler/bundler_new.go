package bundler

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/Hayao0819/seira/internal/depgraph"
	"github.com/Hayao0819/seira/internal/shellparse"
	"github.com/cockroachdb/errors"
)

// Config holds all settings for the bundling process.
type Config struct {
	InputPath  string
	OutputPath string
	BaseDir    string
	WorkDir    string // only used by tarball mode for explicit work dir
	Minify     bool
	Shebang    string
	Mode       string // "tarball" or "concat", default "tarball"
	Type       string // "executable" (default) or "library"
	Env        map[string]string
	Include    []string
	Exports    []string // library mode: exported function names
	DepsDir    string   // deps directory (default: deps relative to BaseDir)
	TreeShake  bool     // enable tree shaking (remove unused functions)
}

// Bundler orchestrates the full bundle pipeline.
type Bundler struct {
	cfg    Config
	parser *shellparse.Parser
}

func New(cfg Config) *Bundler {
	return &Bundler{
		cfg:    cfg,
		parser: shellparse.NewParser(),
	}
}

// Bundle executes the full pipeline:
// parse → resolve deps → validate → topo sort → dispatch to mode
func (b *Bundler) Bundle() error {
	// 1. Resolve entrypoint absolute path
	absInput, err := filepath.Abs(b.cfg.InputPath)
	if err != nil {
		return errors.Wrap(err, "resolving input path")
	}

	// 2. Resolve base directory
	baseDir := b.cfg.BaseDir
	if baseDir == "" {
		baseDir = filepath.Dir(absInput)
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return errors.Wrap(err, "resolving base dir")
	}

	// 3. Resolve deps directory
	depsDir := filepath.Join(baseDir, "deps")
	if b.cfg.DepsDir != "" {
		depsDir = b.cfg.DepsDir
		if !filepath.IsAbs(depsDir) {
			depsDir = filepath.Join(baseDir, depsDir)
		}
	}

	// 4. Build dependency graph (follows source statements recursively)
	graph, err := depgraph.Resolve(b.parser, absInput, b.cfg.Env)
	if err != nil {
		return errors.Wrap(err, "resolving dependencies")
	}

	// 5. Auto-discover seira libraries in deps/ and inject into graph
	if err := b.resolveLibraryDeps(graph, absInput, depsDir); err != nil {
		return errors.Wrap(err, "resolving library dependencies")
	}

	// 6. Validate entrypoint
	entryNode := graph.Node(absInput)
	if entryNode == nil || entryNode.Script == nil {
		return errors.Newf("entrypoint not found in graph: %s", absInput)
	}
	if b.cfg.Type != "library" && !entryNode.Script.HasFunc("main") {
		return errors.Newf("entrypoint %s does not declare a main() function", absInput)
	}

	// 7. Topological sort
	order, err := graph.TopologicalSort()
	if err != nil {
		return errors.Wrap(err, "topological sort")
	}

	// 8. Add include files
	for _, inc := range b.cfg.Include {
		absInc := inc
		if !filepath.IsAbs(inc) {
			absInc = filepath.Join(baseDir, inc)
		}
		found := false
		for _, p := range order {
			if p == absInc {
				found = true
				break
			}
		}
		if !found {
			order = append([]string{absInc}, order...)
		}
	}

	// 9. Prepare output file
	outPath := b.cfg.OutputPath
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return errors.Wrap(err, "creating output directory")
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		return errors.Wrap(err, "creating output file")
	}
	defer outFile.Close()

	// 10. Build context and dispatch to mode
	shebang := b.cfg.Shebang
	if shebang == "" {
		shebang = "/bin/sh"
	}

	hasDeps := false
	if info, err := os.Stat(depsDir); err == nil && info.IsDir() {
		hasDeps = true
	}

	ctx := &BundleContext{
		Graph:       graph,
		Order:       order,
		BaseDir:     baseDir,
		Shebang:     shebang,
		Minify:      b.cfg.Minify,
		Output:      outFile,
		DepsDir:     depsDir,
		HasDeps:     hasDeps,
		Type:        b.cfg.Type,
		LibraryName: "",
		Exports:     b.cfg.Exports,
		TreeShake:   b.cfg.TreeShake,
	}

	mode := resolveMode(b.cfg.Mode, b.cfg.Type)
	if err := mode.Generate(ctx); err != nil {
		return errors.Wrap(err, "generating output")
	}

	// Make output executable
	if err := os.Chmod(outPath, 0755); err != nil {
		return errors.Wrap(err, "setting output permissions")
	}

	return nil
}

// resolveLibraryDeps scans the deps directory for seira library projects
// and injects their entrypoints into the dependency graph.
func (b *Bundler) resolveLibraryDeps(graph *depgraph.Graph, entrypoint string, depsDir string) error {
	entries, err := os.ReadDir(depsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		depDir := filepath.Join(depsDir, entry.Name())
		cfg, err := config.Load(depDir)
		if err != nil || cfg.Type != "library" {
			continue
		}

		if cfg.Entrypoint == "" {
			slog.Warn("library has no entrypoint, skipping", "library", entry.Name())
			continue
		}

		libEntry := filepath.Join(depDir, cfg.Entrypoint)
		if _, err := os.Stat(libEntry); err != nil {
			slog.Warn("library entrypoint not found, skipping",
				"library", entry.Name(), "entrypoint", libEntry)
			continue
		}

		slog.Info("auto-loading library from deps", "library", entry.Name())

		if err := depgraph.ResolveAdditional(graph, b.parser, libEntry, b.cfg.Env); err != nil {
			return errors.Wrapf(err, "resolving library %s", entry.Name())
		}

		graph.AddEdge(entrypoint, libEntry)
	}

	return nil
}

// resolveMode returns the Mode implementation for the given mode name and project type.
func resolveMode(name string, projectType string) Mode {
	if projectType == "library" {
		return &LibraryMode{}
	}
	switch name {
	case "concat":
		return &ConcatMode{}
	case "tarball", "":
		return &TarballMode{}
	default:
		return &TarballMode{}
	}
}

// copyFiles copies files to work directory preserving relative paths from baseDir.
func copyFiles(files []string, baseDir string, workDir string) error {
	for _, file := range files {
		rel, err := filepath.Rel(baseDir, file)
		if err != nil {
			return errors.Wrapf(err, "computing relative path for %s", file)
		}
		dest := filepath.Join(workDir, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return errors.Wrapf(err, "creating directory for %s", dest)
		}
		if err := copyFile(file, dest); err != nil {
			return errors.Wrapf(err, "copying %s to %s", file, dest)
		}
	}
	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
