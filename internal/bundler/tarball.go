package bundler

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// TarballMode bundles scripts into a self-extracting tarball.
type TarballMode struct{}

func (m *TarballMode) Generate(ctx *BundleContext) error {
	// Create work directory
	workDir, err := os.MkdirTemp("", "seira-tarball-*")
	if err != nil {
		return errors.Wrap(err, "creating work directory")
	}
	defer os.RemoveAll(workDir)

	// Copy files to work directory
	if err := copyFiles(ctx.Order, ctx.BaseDir, workDir); err != nil {
		return errors.Wrap(err, "copying files")
	}

	// Minify if enabled
	if ctx.Minify {
		if err := minifyDir(workDir); err != nil {
			return errors.Wrap(err, "minifying")
		}
	}

	// Create tarball
	tarball, err := createTarball(workDir)
	if err != nil {
		return errors.Wrap(err, "creating tarball")
	}

	// Determine entrypoint relative path
	entryRel, err := filepath.Rel(ctx.BaseDir, ctx.Order[len(ctx.Order)-1])
	if err != nil {
		return errors.Wrap(err, "computing entrypoint relative path")
	}

	// Render output
	return renderOutput(ctx.Output, tarball, ctx.Shebang, entryRel)
}

// createTarball creates a tar.gz archive of the directory contents.
func createTarball(dir string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
