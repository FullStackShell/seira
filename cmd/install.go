package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/bpkg"
	"github.com/Hayao0819/seira/internal/config"
	"github.com/Hayao0819/seira/internal/lockfile"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	var depsDir string
	var save bool

	cmd := &cobra.Command{
		Use:   "install [user/package[@version]]",
		Short: "Install a bpkg-compatible package",
		Long: `Install a bpkg-compatible package from GitHub into the deps directory.

If no package is specified, installs all dependencies from .seira-lock.json
(or .seirarc.json if no lockfile exists).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return installFromLockOrConfig(depsDir)
			}

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
			result, err := inst.Install(ref)
			if err != nil {
				return err
			}

			fmt.Printf("Installed %s to %s/%s\n", ref, inst.DepsDir, ref.Name)

			// Update lockfile
			lf, err := lockfile.Load(".")
			if err != nil {
				return fmt.Errorf("loading lockfile: %w", err)
			}
			recordResults(lf, result)
			if err := lf.Save("."); err != nil {
				return fmt.Errorf("saving lockfile: %w", err)
			}

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

// installFromLockOrConfig installs all dependencies from lockfile, falling back to config.
func installFromLockOrConfig(depsDir string) error {
	cfg, cfgErr := config.Load(".")
	if cfgErr == nil && depsDir == "" {
		depsDir = cfg.DepsDir
	}

	lf, err := lockfile.Load(".")
	if err != nil {
		return fmt.Errorf("loading lockfile: %w", err)
	}

	// If lockfile has entries, install from it
	if len(lf.Dependencies) > 0 {
		return installFromLock(lf, depsDir)
	}

	// Fall back to .seirarc.json dependencies
	if cfgErr != nil {
		return fmt.Errorf("no .seira-lock.json or .seirarc.json found")
	}
	if len(cfg.Dependencies) == 0 {
		fmt.Println("No dependencies to install.")
		return nil
	}

	return installFromConfig(cfg, lf, depsDir)
}

// installFromLock installs all packages recorded in the lockfile.
func installFromLock(lf *lockfile.LockFile, depsDir string) error {
	inst := bpkg.NewInstaller(depsDir)
	newLf := lockfile.New()

	for pkg, entry := range lf.Dependencies {
		ref, err := bpkg.ParsePackageRef(pkg)
		if err != nil {
			return fmt.Errorf("parsing lockfile entry %s: %w", pkg, err)
		}

		// Pin to the locked commit hash
		if entry.Commit != "" {
			ref.Version = entry.Commit
		} else if entry.Version != "" {
			ref.Version = entry.Version
		}

		result, err := inst.Install(ref)
		if err != nil {
			return fmt.Errorf("installing %s: %w", pkg, err)
		}
		recordResults(newLf, result)
		fmt.Printf("Installed %s\n", ref)
	}

	// Preserve lockfile
	if err := newLf.Save("."); err != nil {
		return fmt.Errorf("saving lockfile: %w", err)
	}

	return nil
}

// installFromConfig installs dependencies listed in .seirarc.json.
func installFromConfig(cfg *config.Config, lf *lockfile.LockFile, depsDir string) error {
	inst := bpkg.NewInstaller(depsDir)

	for pkg, ver := range cfg.Dependencies {
		ref, err := bpkg.ParsePackageRef(pkg)
		if err != nil {
			return fmt.Errorf("parsing dependency %s: %w", pkg, err)
		}

		if entry, ok := lf.Get(pkg); ok && entry.Commit != "" {
			ref.Version = entry.Commit
		} else if ver != "" && ver != "*" {
			ref.Version = ver
		}

		result, err := inst.Install(ref)
		if err != nil {
			return fmt.Errorf("installing %s: %w", pkg, err)
		}
		recordResults(lf, result)
		fmt.Printf("Installed %s\n", ref)
	}

	if err := lf.Save("."); err != nil {
		return fmt.Errorf("saving lockfile: %w", err)
	}

	return nil
}

func depsCmd() *cobra.Command {
	var depsDir string

	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Install all dependencies from .seirarc.json (alias for install with no args)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installFromLockOrConfig(depsDir)
		},
	}

	cmd.Flags().StringVar(&depsDir, "deps-dir", "", "deps directory (default: from config or ./deps)")
	return cmd
}

// recordResults recursively records install results into the lockfile.
func recordResults(lf *lockfile.LockFile, result *bpkg.InstallResult) {
	if result == nil {
		return
	}
	version := result.Ref.Version
	if version == "master" {
		version = ""
	}
	lf.Set(result.PkgKey(), version, result.Commit)

	for _, sub := range result.Sub {
		recordResults(lf, sub)
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
