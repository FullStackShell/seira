package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/Hayao0819/seira/internal/lint"
	"github.com/spf13/cobra"
)

func lintCmd() *cobra.Command {
	var (
		mode        string
		projectType string
	)

	cmd := &cobra.Command{
		Use:   "lint [input-file]",
		Short: "Check scripts for potential issues",
		Long: `Analyze shell scripts and report potential problems.

Detects patterns that may cause issues in specific bundle modes:
  - BASH_SOURCE / $0 path resolution (breaks in concat mode)
  - Conditional source statements (flattened in concat mode)
  - Function name collisions across files (concat/library modes)

If no input file is given, uses the entrypoint from .seirarc.json.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input, baseDir string

			if len(args) > 0 {
				input = args[0]
				baseDir = filepath.Dir(input)
			} else {
				// Use config entrypoint
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				cfg, err := config.Load(cwd)
				if err != nil {
					return fmt.Errorf("no input file specified and no .seirarc.json found")
				}
				if cfg.Entrypoint == "" {
					return fmt.Errorf("no input file specified and no entrypoint in .seirarc.json")
				}
				input = cfg.Entrypoint
				baseDir = cwd
			}

			// Load config for mode/type defaults
			cfg, err := config.Load(baseDir)
			if err != nil {
				cfg = config.Default()
			}

			m := cfg.Mode
			if cmd.Flags().Changed("mode") {
				m = mode
			}
			if m == "" {
				m = "concat" // lint is most useful for concat mode
			}

			t := cfg.Type
			if cmd.Flags().Changed("type") {
				t = projectType
			}

			diags, err := lint.Lint(lint.Config{
				InputPath: input,
				BaseDir:   baseDir,
				Mode:      m,
				Type:      t,
				Env:       cfg.Env,
			})
			if err != nil {
				return err
			}

			if len(diags) == 0 {
				fmt.Println("No issues found.")
				return nil
			}

			for _, d := range diags {
				icon := "⚠"
				if d.Severity == lint.SeverityError {
					icon = "✗"
				}
				loc := d.File
				if d.Line > 0 {
					loc = fmt.Sprintf("%s:%d", d.File, d.Line)
				}
				fmt.Fprintf(os.Stderr, "  %s %s: %s [%s]\n", icon, loc, d.Message, d.Rule)
			}
			fmt.Fprintf(os.Stderr, "\n%d issue(s) found.\n", len(diags))

			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "", "bundle mode to check against (default: from config or concat)")
	cmd.Flags().StringVar(&projectType, "type", "", "project type: executable or library")

	return cmd
}
