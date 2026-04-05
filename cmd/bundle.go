package cmd

import (
	"path/filepath"

	"github.com/Hayao0819/seira/internal/bundler"
	"github.com/Hayao0819/seira/internal/config"
	"github.com/spf13/cobra"
)

func bundleCmd() *cobra.Command {
	var (
		output  string
		minify  bool
		shebang string
		mode    string
	)

	cmd := &cobra.Command{
		Use:   "bundle <input-file>",
		Short: "Bundle shell scripts into a standalone executable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			baseDir := filepath.Dir(input)

			// Load config from input file's directory
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

			return bundler.New(bundler.Config{
				InputPath:  input,
				OutputPath: output,
				BaseDir:    baseDir,
				Minify:     minify,
				Shebang:    sh,
				Mode:       m,
				Env:        cfg.Env,
				Include:    cfg.Include,
				DepsDir:    cfg.DepsDir,
			}).Bundle()
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "output.sh", "output file path")
	cmd.Flags().BoolVarP(&minify, "minify", "m", false, "minify shell scripts")
	cmd.Flags().StringVar(&shebang, "shebang", "", "shebang line (default: from config or /bin/sh)")
	cmd.Flags().StringVar(&mode, "mode", "", "bundle mode: tarball or concat (default: from config or tarball)")

	return cmd
}
