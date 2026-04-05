package bundler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundle_Simple(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/simple",
		Shebang:    "/bin/bash",
	}
	b := New(cfg)
	if err := b.Bundle(); err != nil {
		t.Fatal(err)
	}

	// Verify output file exists and is executable
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0100 == 0 {
		t.Error("output file is not executable")
	}

	// Verify output contains base64 data
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Error("output file seems too small")
	}
}

func TestBundle_WithDeps(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/deps/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/deps",
		Shebang:    "/bin/bash",
	}
	b := New(cfg)
	if err := b.Bundle(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

func TestBundle_NoMainFunc(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/deps/lib/helper.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/deps/lib",
		Shebang:    "/bin/bash",
	}
	b := New(cfg)
	err := b.Bundle()
	if err == nil {
		t.Fatal("expected error for missing main(), got nil")
	}
}

func TestBundle_WithMinify(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/simple",
		Shebang:    "/bin/bash",
		Minify:     true,
	}
	b := New(cfg)
	if err := b.Bundle(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

func TestBundle_E2E_Simple(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/simple",
		Shebang:    "/bin/bash",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("execution failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "Hello from simple!" {
		t.Errorf("output: got %q, want %q", got, "Hello from simple!")
	}
}

func TestBundle_E2E_WithDeps(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/deps/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/deps",
		Shebang:    "/bin/bash",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("execution failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "Hello, World!" {
		t.Errorf("output: got %q, want %q", got, "Hello, World!")
	}
}

func TestBundle_E2E_VarPath(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/varpath/main.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/varpath",
		Shebang:    "/bin/bash",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("execution failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "Hello from dep!" {
		t.Errorf("output: got %q, want %q", got, "Hello from dep!")
	}
}

func TestBundle_Circular(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../testdata/circular/a.sh",
		OutputPath: outPath,
		BaseDir:    "../testdata/circular",
		Shebang:    "/bin/bash",
	}
	b := New(cfg)
	err := b.Bundle()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}
