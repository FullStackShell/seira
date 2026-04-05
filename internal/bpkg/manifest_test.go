package bpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()

	data := `{
  "name": "term",
  "version": "0.1.1",
  "description": "Terminal utility functions",
  "scripts": ["term.sh"],
  "files": ["README.md"],
  "dependencies": {
    "bpkg/colors": "0.0.1"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "bpkg.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != "term" {
		t.Errorf("name: got %q, want %q", m.Name, "term")
	}
	if m.Version != "0.1.1" {
		t.Errorf("version: got %q, want %q", m.Version, "0.1.1")
	}
	if len(m.Scripts) != 1 || m.Scripts[0] != "term.sh" {
		t.Errorf("scripts: got %v, want [term.sh]", m.Scripts)
	}
	if len(m.Dependencies) != 1 {
		t.Errorf("dependencies: got %d, want 1", len(m.Dependencies))
	}
}

func TestLoadManifest_PackageJsonFallback(t *testing.T) {
	dir := t.TempDir()

	data := `{"name": "legacy", "scripts": ["main.sh"]}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "legacy" {
		t.Errorf("name: got %q, want %q", m.Name, "legacy")
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadManifest(dir)
	if !os.IsNotExist(err) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}
