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

	// Copy deps directory if it exists
	if ctx.HasDeps {
		depsRel, err := filepath.Rel(ctx.BaseDir, ctx.DepsDir)
		if err != nil {
			depsRel = "deps"
		}
		destDeps := filepath.Join(workDir, depsRel)
		if err := os.MkdirAll(destDeps, 0755); err != nil {
			return errors.Wrap(err, "creating deps dir in work directory")
		}
		if err := copyDirRecursive(ctx.DepsDir, destDeps); err != nil {
			return errors.Wrap(err, "copying deps directory")
		}
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
	return renderOutput(ctx.Output, tarball, ctx.Shebang, entryRel, ctx.HasDeps)
}

// copyDirRecursive copies all contents from src to dst, excluding .git.
func copyDirRecursive(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			os.Remove(dstPath)
			if err := os.Symlink(target, dstPath); err != nil {
				return err
			}
			continue
		}

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
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
