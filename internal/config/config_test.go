package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	data := `{"entrypoint": "main.sh", "shebang": "/bin/bash", "env": {"FOO": "bar"}, "include": ["extra.sh"], "exclude": []}`
	if err := os.WriteFile(filepath.Join(dir, ".seirarc.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Entrypoint != "main.sh" {
		t.Errorf("Entrypoint: got %q, want %q", cfg.Entrypoint, "main.sh")
	}
	if cfg.Shebang != "/bin/bash" {
		t.Errorf("Shebang: got %q, want %q", cfg.Shebang, "/bin/bash")
	}
	if cfg.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO]: got %q, want %q", cfg.Env["FOO"], "bar")
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "extra.sh" {
		t.Errorf("Include: got %v, want [extra.sh]", cfg.Include)
	}
}

func TestLoadFromDirNoDot(t *testing.T) {
	dir := t.TempDir()
	data := `{"entrypoint": "entry.sh", "shebang": "/bin/bash"}`
	if err := os.WriteFile(filepath.Join(dir, "seirarc.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Entrypoint != "entry.sh" {
		t.Errorf("Entrypoint: got %q, want %q", cfg.Entrypoint, "entry.sh")
	}
}

func TestLoadNoDotTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "seirarc.json"), []byte(`{"entrypoint": "nodot.sh"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".seirarc.json"), []byte(`{"entrypoint": "dot.sh"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Entrypoint != "nodot.sh" {
		t.Errorf("seirarc.json should take precedence: got %q, want %q", cfg.Entrypoint, "nodot.sh")
	}
}

func TestLoadFromParent(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "sub", "dir")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	data := `{"entrypoint": "test.sh"}`
	if err := os.WriteFile(filepath.Join(parent, ".seirarc.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(child)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Entrypoint != "test.sh" {
		t.Errorf("Entrypoint: got %q, want %q", cfg.Entrypoint, "test.sh")
	}
}

func TestLoadDefault(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Shebang != "/bin/sh" {
		t.Errorf("Shebang: got %q, want %q", cfg.Shebang, "/bin/sh")
	}
	if cfg.Entrypoint != "" {
		t.Errorf("Entrypoint: got %q, want empty", cfg.Entrypoint)
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".seirarc.json"), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}
