package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/bpkg"
	"github.com/Hayao0819/seira/internal/config"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	var depsDir string
	var save bool

	cmd := &cobra.Command{
		Use:   "install <user/package[@version]>",
		Short: "Install a bpkg-compatible package",
		Long:  "Download and install a bpkg-compatible package from GitHub into the deps directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bpkg.ParsePackageRef(args[0])
			if err != nil {
				return err
			}

			// Load config to determine deps directory
			cfg, cfgErr := config.Load(".")
			if cfgErr == nil && depsDir == "" {
				depsDir = cfg.DepsDir
			}

			inst := bpkg.NewInstaller(depsDir)
			if err := inst.Install(ref); err != nil {
				return err
			}

			fmt.Printf("Installed %s to %s/%s\n", ref, inst.DepsDir, ref.Name)

			// Save to .seirarc.json if --save flag is set
			if save && cfgErr == nil {
				return saveDependency(cfg, ref)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&depsDir, "deps-dir", "", "deps directory (default: from config or ./deps)")
	cmd.Flags().BoolVarP(&save, "save", "s", false, "save dependency to .seirarc.json")

	return cmd
}

func depsCmd() *cobra.Command {
	var depsDir string

	return &cobra.Command{
		Use:   "deps",
		Short: "Install all dependencies from .seirarc.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(".")
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if len(cfg.Dependencies) == 0 {
				fmt.Println("No dependencies configured in .seirarc.json")
				return nil
			}

			if depsDir == "" {
				depsDir = cfg.DepsDir
			}

			inst := bpkg.NewInstaller(depsDir)

			for pkg, ver := range cfg.Dependencies {
				ref, err := bpkg.ParsePackageRef(pkg)
				if err != nil {
					return fmt.Errorf("parsing dependency %s: %w", pkg, err)
				}
				if ver != "" && ver != "*" {
					ref.Version = ver
				}
				if err := inst.Install(ref); err != nil {
					return fmt.Errorf("installing %s: %w", pkg, err)
				}
				fmt.Printf("Installed %s\n", ref)
			}

			return nil
		},
	}
}

// saveDependency adds or updates a dependency in .seirarc.json.
func saveDependency(cfg *config.Config, ref *bpkg.PackageRef) error {
	pkgKey := fmt.Sprintf("%s/%s", ref.User, ref.Name)
	cfg.Dependencies[pkgKey] = ref.Version

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Find the config file path by searching upward
	cfgPath := filepath.Join(".", ".seirarc.json")
	return os.WriteFile(cfgPath, append(data, '\n'), 0644)
}
