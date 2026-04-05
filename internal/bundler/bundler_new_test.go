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
		InputPath:  "../../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/simple",
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
		InputPath:  "../../testdata/deps/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/deps",
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
		InputPath:  "../../testdata/deps/lib/helper.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/deps/lib",
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
		InputPath:  "../../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/simple",
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
		InputPath:  "../../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/simple",
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
		InputPath:  "../../testdata/deps/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/deps",
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
		InputPath:  "../../testdata/varpath/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/varpath",
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

func TestBundle_E2E_Concat_Simple(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/simple/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/simple",
		Shebang:    "/bin/bash",
		Mode:       "concat",
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

func TestBundle_E2E_Concat_WithDeps(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/deps/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/deps",
		Shebang:    "/bin/bash",
		Mode:       "concat",
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

func TestBundle_E2E_Concat_SideEffect(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/sideeffect/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/sideeffect",
		Shebang:    "/bin/bash",
		Mode:       "concat",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		data, _ := os.ReadFile(outPath)
		t.Fatalf("execution failed: %v\nstdout: %s\nscript:\n%s", err, out, data)
	}
	if got := strings.TrimSpace(string(out)); got != "Hello, World!" {
		data, _ := os.ReadFile(outPath)
		t.Errorf("output: got %q, want %q\nscript:\n%s", got, "Hello, World!", data)
	}
}

func TestBundle_Library(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/consumer/deps/strutils/index.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/consumer/deps/strutils",
		Shebang:    "/bin/bash",
		Type:       "library",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Library output should NOT have shebang or main "$@"
	if strings.HasPrefix(content, "#!/") {
		t.Error("library output should not have a shebang line")
	}
	if strings.Contains(content, `main "$@"`) {
		t.Error("library output should not have main entry point")
	}
	// Should contain the library header
	if !strings.Contains(content, "library mode") {
		t.Errorf("library output should contain library mode header, got:\n%s", content)
	}
	// Should contain function definitions
	if !strings.Contains(content, "str_upper") {
		t.Error("library output should contain str_upper function")
	}
	if !strings.Contains(content, "str_lower") {
		t.Error("library output should contain str_lower function")
	}
}

func TestBundle_Library_NoMainRequired(t *testing.T) {
	// Verify that library mode does NOT require a main() function
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/consumer/deps/strutils/index.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/consumer/deps/strutils",
		Shebang:    "/bin/bash",
		Type:       "library",
	}
	// Should succeed even though there's no main()
	if err := New(cfg).Bundle(); err != nil {
		t.Fatalf("library mode should not require main(): %v", err)
	}
}

func TestBundle_Library_WithExports(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/consumer/deps/strutils/index.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/consumer/deps/strutils",
		Shebang:    "/bin/bash",
		Type:       "library",
		Exports:    []string{"str_upper"},
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "str_upper") {
		t.Error("library output should contain exported str_upper function")
	}
	if strings.Contains(content, "str_lower()") {
		t.Error("library output should NOT contain non-exported str_lower function")
	}
}

func TestBundle_E2E_Consumer_WithLibrary(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/consumer/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/consumer",
		Shebang:    "/bin/bash",
		Mode:       "concat",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should contain library functions inlined
	if !strings.Contains(content, "str_upper") {
		t.Error("consumer output should contain str_upper from library")
	}
	// Should contain main entry point
	if !strings.Contains(content, `main "$@"`) {
		t.Error("consumer output should have main entry point")
	}

	// E2E: execute and verify output
	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("execution failed: %v\noutput: %s\nscript:\n%s", err, out, content)
	}
	if got := strings.TrimSpace(string(out)); got != "HELLO" {
		t.Errorf("output: got %q, want %q\nscript:\n%s", got, "HELLO", content)
	}
}

func TestBundle_E2E_Concat_Directive_Ignore(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/directive/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/directive",
		Shebang:    "/bin/bash",
		Mode:       "concat",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// The ignored source should remain in output
	if !strings.Contains(content, "source /etc/myapp/config.sh") {
		t.Errorf("ignored source should be preserved in output, got:\n%s", content)
	}

	// The bundled source (lib/helper.sh) should NOT appear as a source line
	if strings.Contains(content, "source lib/helper.sh") {
		t.Errorf("bundled source should be removed, got:\n%s", content)
	}

	// greet function should be inlined from lib/helper.sh
	if !strings.Contains(content, "greet()") {
		t.Errorf("greet function should be inlined, got:\n%s", content)
	}
}

func TestBundle_E2E_Concat_UsingNamespace(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/using/main.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/using",
		Shebang:    "/bin/bash",
		Mode:       "concat",
	}
	if err := New(cfg).Bundle(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should contain the original namespaced functions
	if !strings.Contains(content, "str::upper()") {
		t.Errorf("should contain str::upper function, got:\n%s", content)
	}

	// Should contain generated aliases
	if !strings.Contains(content, `upper() { str::upper "$@"; }`) {
		t.Errorf("should contain upper alias, got:\n%s", content)
	}
	if !strings.Contains(content, `lower() { str::lower "$@"; }`) {
		t.Errorf("should contain lower alias, got:\n%s", content)
	}

	// E2E: execute and verify
	cmd := exec.Command(outPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("execution failed: %v\nscript:\n%s", err, content)
	}
	if got := strings.TrimSpace(string(out)); got != "HELLO" {
		t.Errorf("output: got %q, want %q\nscript:\n%s", got, "HELLO", content)
	}
}

func TestBundle_Circular(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "out.sh")

	cfg := Config{
		InputPath:  "../../testdata/circular/a.sh",
		OutputPath: outPath,
		BaseDir:    "../../testdata/circular",
		Shebang:    "/bin/bash",
	}
	b := New(cfg)
	err := b.Bundle()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}
