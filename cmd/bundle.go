package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/bundler"
	"github.com/Hayao0819/seira/internal/config"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	var (
		output      string
		minify      bool
		shebang     string
		mode        string
		projectType string
		treeshake   bool
	)

	cmd := &cobra.Command{
		Use:     "build [input-file]",
		Aliases: []string{"bundle"},
		Short:   "Bundle shell scripts into a standalone executable",
		Long: `Bundle shell scripts into a standalone executable.

If input-file is specified, it is used as the entry point.
If omitted, the entrypoint is read from .seirarc.json.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input, baseDir string

			if len(args) == 1 {
				// Explicit input file specified
				input = args[0]
				baseDir = filepath.Dir(input)
			} else {
				// No input file: resolve from config
				cfg, cfgDir, err := config.LoadWithDir(".")
				if err != nil {
					return err
				}
				if cfg.Entrypoint == "" {
					return fmt.Errorf("no input file specified and no entrypoint defined in .seirarc.json")
				}
				if cfgDir == "" {
					return fmt.Errorf("no input file specified and no .seirarc.json found")
				}
				input = filepath.Join(cfgDir, cfg.Entrypoint)
				baseDir = cfgDir
			}

			// Load config from base directory
			cfg, err := config.Load(baseDir)
			if err != nil {
				return err
			}

			// CLI flags override config values
			sh := cfg.Shebang
			if cmd.Flags().Changed("shebang") {
				sh = shebang
			}

			m := cfg.Mode
			if cmd.Flags().Changed("mode") {
				m = mode
			}

			t := cfg.Type
			if cmd.Flags().Changed("type") {
				t = projectType
			}

			ts := cfg.TreeShake
			if cmd.Flags().Changed("treeshake") {
				ts = treeshake
			}

			return bundler.New(bundler.Config{
				InputPath:  input,
				OutputPath: output,
				BaseDir:    baseDir,
				Minify:     minify,
				Shebang:    sh,
				Mode:       m,
				Type:       t,
				Env:        cfg.Env,
				Include:    cfg.Include,
				Exports:    cfg.Exports,
				DepsDir:    cfg.DepsDir,
				TreeShake:  ts,
			}).Bundle()
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "output.sh", "output file path")
	cmd.Flags().BoolVarP(&minify, "minify", "m", false, "minify shell scripts")
	cmd.Flags().StringVar(&shebang, "shebang", "", "shebang line (default: from config or /bin/sh)")
	cmd.Flags().StringVar(&mode, "mode", "", "bundle mode: tarball or concat (default: from config or tarball)")
	cmd.Flags().StringVar(&projectType, "type", "", "project type: executable or library (default: from config or executable)")
	cmd.Flags().BoolVar(&treeshake, "treeshake", false, "enable tree shaking to remove unused functions")

	return cmd
}
