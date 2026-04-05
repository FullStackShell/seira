package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/config"
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	return &cobra.Command{
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
			cfg.Entrypoint = "main.sh"
			cfg.Shebang = "/bin/bash"
			cfgData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			cfgPath := filepath.Join(name, ".seirarc.json")
			if err := os.WriteFile(cfgPath, append(cfgData, '\n'), 0644); err != nil {
				return err
			}

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
			return nil
		},
	}
}
