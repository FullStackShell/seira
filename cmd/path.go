package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/spf13/cobra"
)

func pathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path <owner/repo/path>",
		Short: "Resolve a package file path in deps",
		Long:  "Resolve owner/repo/path to the actual file path in the deps directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]

			// Parse: owner/repo/path → repo/path
			parts := strings.SplitN(ref, "/", 3)
			if len(parts) < 2 {
				return fmt.Errorf("invalid reference %q: expected owner/repo[/path]", ref)
			}

			// Load config to find deps dir
			cfg, err := config.Load(".")
			if err != nil {
				cfg = config.Default()
			}

			depsDir := cfg.DepsDir

			// Construct: deps/repo[/path]
			var resolved string
			if len(parts) == 3 {
				resolved = filepath.Join(depsDir, parts[1], parts[2])
			} else {
				resolved = filepath.Join(depsDir, parts[1])
			}

			abs, err := filepath.Abs(resolved)
			if err != nil {
				return err
			}

			if _, err := os.Stat(abs); os.IsNotExist(err) {
				return fmt.Errorf("not found: %s", abs)
			}

			fmt.Println(abs)
			return nil
		},
	}
}
