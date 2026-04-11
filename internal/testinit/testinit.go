package testinit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hayao0819/seira/internal/config"
)

// Init creates ShellSpec project structure in the given directory.
func Init(dir string, cfg *config.Config) error {
	// Create spec directory
	specDir := filepath.Join(dir, "spec")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return fmt.Errorf("creating spec directory: %w", err)
	}

	// Write .shellspec
	if err := writeIfNotExist(filepath.Join(dir, ".shellspec"), dotShellspec()); err != nil {
		return err
	}

	// Write spec/spec_helper.sh
	if err := writeIfNotExist(filepath.Join(specDir, "spec_helper.sh"), specHelper(cfg)); err != nil {
		return err
	}

	// Write example spec
	specName := exampleSpecName(cfg)
	if err := writeIfNotExist(filepath.Join(specDir, specName), exampleSpec(cfg)); err != nil {
		return err
	}

	fmt.Println("Initialized ShellSpec test structure:")
	fmt.Println("  .shellspec")
	fmt.Println("  spec/spec_helper.sh")
	fmt.Printf("  spec/%s\n", specName)
	return nil
}

func writeIfNotExist(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  skip: %s (already exists)\n", path)
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func dotShellspec() string {
	return "--require spec_helper\n"
}

func specHelper(cfg *config.Config) string {
	var b strings.Builder
	b.WriteString("# shellcheck shell=bash\n\n")
	b.WriteString("spec_helper_precheck() {\n")
	b.WriteString("  minimum_version \"0.28.1\"\n")
	b.WriteString("}\n\n")
	b.WriteString("spec_helper_loaded() {\n")
	b.WriteString("  :\n")
	b.WriteString("}\n\n")
	b.WriteString("spec_helper_configure() {\n")
	b.WriteString("  :\n")
	b.WriteString("}\n")
	return b.String()
}

func exampleSpecName(cfg *config.Config) string {
	if cfg.Entrypoint != "" {
		base := filepath.Base(cfg.Entrypoint)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		return name + "_spec.sh"
	}
	return "main_spec.sh"
}

func exampleSpec(cfg *config.Config) string {
	entrypoint := cfg.Entrypoint
	if entrypoint == "" {
		entrypoint = "main.sh"
	}

	var b strings.Builder
	if cfg.Type == "library" {
		b.WriteString(fmt.Sprintf("Describe '%s'\n", filepath.Base(entrypoint)))
		b.WriteString(fmt.Sprintf("  Include %s\n", entrypoint))
		b.WriteString("\n")
		b.WriteString("  It 'loads without error'\n")
		b.WriteString("    When call true\n")
		b.WriteString("    The status should be success\n")
		b.WriteString("  End\n")
		b.WriteString("End\n")
	} else {
		b.WriteString(fmt.Sprintf("Describe '%s'\n", filepath.Base(entrypoint)))
		b.WriteString(fmt.Sprintf("  It 'outputs hello world'\n"))
		b.WriteString(fmt.Sprintf("    When run source %s\n", entrypoint))
		b.WriteString("    The output should include \"Hello\"\n")
		b.WriteString("  End\n")
		b.WriteString("End\n")
	}
	return b.String()
}
