package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/Hayao0819/seira/internal/depgraph"
	"github.com/Hayao0819/seira/internal/doc"
	"github.com/Hayao0819/seira/internal/shellparse"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

func docCmd() *cobra.Command {
	var (
		format string
		output string
		title  string
	)

	cmd := &cobra.Command{
		Use:   "doc [input-file]",
		Short: "Generate documentation from shell scripts",
		Long: `Generate documentation from shell script doc comments.

Supports shdoc-compatible annotations:
  @file, @brief, @description, @param/@arg, @option,
  @return, @exitcode, @example, @see, @internal,
  @set, @stdin, @stdout, @section, @noargs

Output formats: markdown (default), html

If no input file is given, uses the entrypoint from .seirarc.json.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input, baseDir string

			if len(args) > 0 {
				input = args[0]
				baseDir = filepath.Dir(input)
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				cfg, cfgDir, err := config.LoadWithDir(cwd)
				if err != nil || cfg.Entrypoint == "" {
					return fmt.Errorf("no input file specified and no entrypoint in .seirarc.json")
				}
				if cfgDir == "" {
					cfgDir = cwd
				}
				input = filepath.Join(cfgDir, cfg.Entrypoint)
				baseDir = cfgDir
			}

			absBase, err := filepath.Abs(baseDir)
			if err != nil {
				return err
			}

			// Load config for env vars
			cfg, err := config.Load(baseDir)
			if err != nil {
				cfg = config.Default()
			}

			// Resolve dependency graph
			parser := shellparse.NewParser()
			graph, err := depgraph.Resolve(parser, input, cfg.Env)
			if err != nil {
				return errors.Wrap(err, "resolving dependencies")
			}

			// Include extra files from config
			for _, inc := range cfg.Include {
				incPath := inc
				if !filepath.IsAbs(incPath) {
					incPath = filepath.Join(absBase, incPath)
				}
				incPath = filepath.Clean(incPath)
				if err := depgraph.ResolveAdditional(graph, parser, incPath, cfg.Env); err != nil {
					return errors.Wrapf(err, "resolving include %s", inc)
				}
			}

			// Extract documentation
			docTitle := title
			if docTitle == "" {
				docTitle = filepath.Base(absBase)
			}
			pd := doc.ExtractProject(graph, absBase, docTitle)

			if len(pd.Files) == 0 {
				fmt.Fprintln(os.Stderr, "No documented files found.")
				return nil
			}

			// Determine output writer
			var w *os.File
			if output == "" || output == "-" {
				w = os.Stdout
			} else {
				dir := filepath.Dir(output)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return errors.Wrap(err, "creating output directory")
				}
				f, err := os.Create(output)
				if err != nil {
					return errors.Wrap(err, "creating output file")
				}
				defer f.Close()
				w = f
			}

			switch format {
			case "markdown", "md":
				return doc.RenderMarkdown(w, pd)
			case "html":
				return doc.RenderHTML(w, pd)
			default:
				return fmt.Errorf("unknown format %q (supported: markdown, html)", format)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "output format: markdown or html")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().StringVar(&title, "title", "", "documentation title (default: project directory name)")

	return cmd
}
