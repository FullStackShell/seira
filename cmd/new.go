package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	var isLibrary bool

	cmd := &cobra.Command{
		Use:   "new <project-name>",
		Short: "Create a new seira project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Create project directory structure
			dirs := []string{
				name,
				filepath.Join(name, "lib"),
			}
			for _, d := range dirs {
				if err := os.MkdirAll(d, 0755); err != nil {
					return fmt.Errorf("creating directory %s: %w", d, err)
				}
			}

			// Write .seirarc.json
			cfg := config.Default()
			cfg.Shebang = "/bin/bash"

			if isLibrary {
				cfg.Type = "library"
				cfg.Entrypoint = "index.sh"
			} else {
				cfg.Entrypoint = "main.sh"
			}

			cfgData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			cfgPath := filepath.Join(name, ".seirarc.json")
			if err := os.WriteFile(cfgPath, append(cfgData, '\n'), 0644); err != nil {
				return err
			}

			if isLibrary {
				// Write index.sh (library entrypoint)
				indexSh := `#!/usr/bin/env bash

source lib/utils.sh
`
				indexPath := filepath.Join(name, "index.sh")
				if err := os.WriteFile(indexPath, []byte(indexSh), 0644); err != nil {
					return err
				}

				// Write lib/utils.sh (example library functions)
				utilsSh := `#!/usr/bin/env bash

# Add your library functions here
example_func() {
    echo "Hello from library: %s"
}
`
				utilsSh = fmt.Sprintf(utilsSh, name)
				utilsPath := filepath.Join(name, "lib", "utils.sh")
				if err := os.WriteFile(utilsPath, []byte(utilsSh), 0644); err != nil {
					return err
				}

				fmt.Printf("Created new seira library: %s\n", name)
			} else {
				// Write main.sh
				mainSh := `#!/usr/bin/env bash

main() {
    echo "Hello, World!"
}
`
				mainPath := filepath.Join(name, "main.sh")
				if err := os.WriteFile(mainPath, []byte(mainSh), 0644); err != nil {
					return err
				}

				fmt.Printf("Created new seira project: %s\n", name)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&isLibrary, "library", false, "create a library project instead of an executable")

	return cmd
}
