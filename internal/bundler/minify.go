package bundler

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// minifyDir minifies all .sh files in the given directory in-place.
func minifyDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".sh") {
			return nil
		}
		return minifyFile(path)
	})
}

// minifyFile minifies a single shell script file in-place.
func minifyFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	parser := syntax.NewParser()
	file, err := parser.Parse(bytes.NewReader(data), path)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	var buf bytes.Buffer
	printer := syntax.NewPrinter(syntax.Minify(true))
	if err := printer.Print(&buf, file); err != nil {
		return fmt.Errorf("printing %s: %w", path, err)
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
