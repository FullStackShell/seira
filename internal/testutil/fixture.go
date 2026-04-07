// Package testutil provides test fixture helpers for seira tests.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteFixture writes a map of relative-path → content into a temp directory
// and returns the base directory path.
func WriteFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
