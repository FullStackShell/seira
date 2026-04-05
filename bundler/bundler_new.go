package bundler

import (
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/depgraph"
	"github.com/Hayao0819/seira/shellparse"
	"github.com/cockroachdb/errors"
)

// Config holds all settings for the bundling process.
type Config struct {
	InputPath  string
	OutputPath string
	BaseDir    string
	WorkDir    string // empty → use os.MkdirTemp
	Minify     bool
	Shebang    string
	Env        map[string]string
	Include    []string
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
// parse → resolve deps → validate main() → topo sort → copy → minify → tarball → render output
func (b *Bundler) Bundle() error {
	// 1. Resolve entrypoint absolute path
	absInput, err := filepath.Abs(b.cfg.InputPath)
	if err != nil {
		return errors.Wrap(err, "resolving input path")
	}

	// 2. Build dependency graph
	graph, err := depgraph.Resolve(b.parser, absInput, b.cfg.Env)
	if err != nil {
		return errors.Wrap(err, "resolving dependencies")
	}

	// 3. Validate entrypoint has main()
	entryNode := graph.Node(absInput)
	if entryNode == nil || entryNode.Script == nil {
		return errors.Newf("entrypoint not found in graph: %s", absInput)
	}
	if !entryNode.Script.HasFunc("main") {
		return errors.Newf("entrypoint %s does not declare a main() function", absInput)
	}

	// 4. Topological sort
	order, err := graph.TopologicalSort()
	if err != nil {
		return errors.Wrap(err, "topological sort")
	}

	// 5. Add include files
	for _, inc := range b.cfg.Include {
		absInc := inc
		if !filepath.IsAbs(inc) {
			absInc = filepath.Join(b.cfg.BaseDir, inc)
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

	// 6. Create work directory
	workDir := b.cfg.WorkDir
	if workDir == "" {
		workDir, err = os.MkdirTemp("", "seira-work-*")
		if err != nil {
			return errors.Wrap(err, "creating work directory")
		}
		defer os.RemoveAll(workDir)
	} else {
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return errors.Wrap(err, "creating work directory")
		}
	}

	// 7. Copy files to work directory
	baseDir := b.cfg.BaseDir
	if baseDir == "" {
		baseDir = filepath.Dir(absInput)
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return errors.Wrap(err, "resolving base dir")
	}
	if err := copyFiles(order, baseDir, workDir); err != nil {
		return errors.Wrap(err, "copying files")
	}

	// 8. Minify if enabled
	if b.cfg.Minify {
		if err := minifyDir(workDir); err != nil {
			return errors.Wrap(err, "minifying")
		}
	}

	// 9. Create tarball
	tarball, err := createTarball(workDir)
	if err != nil {
		return errors.Wrap(err, "creating tarball")
	}

	// 10. Determine entrypoint relative path for the template
	entryRel, err := filepath.Rel(baseDir, absInput)
	if err != nil {
		return errors.Wrap(err, "computing entrypoint relative path")
	}

	// 11. Render output
	shebang := b.cfg.Shebang
	if shebang == "" {
		shebang = "/bin/sh"
	}

	outPath := b.cfg.OutputPath
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return errors.Wrap(err, "creating output directory")
	}
	outFile, err := os.Create(outPath)
	if err != nil {
		return errors.Wrap(err, "creating output file")
	}
	defer outFile.Close()

	if err := renderOutput(outFile, tarball, shebang, entryRel); err != nil {
		return errors.Wrap(err, "rendering output")
	}

	// Make output executable
	if err := os.Chmod(outPath, 0755); err != nil {
		return errors.Wrap(err, "setting output permissions")
	}

	return nil
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
