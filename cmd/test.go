package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/Hayao0819/seira/internal/testinit"
	"github.com/spf13/cobra"
)

func testCmd() *cobra.Command {
	var (
		initFlag bool
		shell    string
		jobs     int
		format   string
	)

	cmd := &cobra.Command{
		Use:   "test [-- shellspec-args...]",
		Short: "Run tests with ShellSpec",
		Long: `Run ShellSpec tests for the current project.

Requires ShellSpec to be installed (https://shellspec.info).

Use --init to create the ShellSpec directory structure (.shellspec, spec/).
Extra arguments after -- are passed directly to shellspec.`,
		// Disable flag parsing after -- so extra args pass through
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := config.Load(cwd)
			if err != nil {
				cfg = config.Default()
			}

			// Handle --init
			if initFlag {
				return testinit.Init(cwd, cfg)
			}

			// Check shellspec is installed
			shellspecPath, err := exec.LookPath("shellspec")
			if err != nil {
				return fmt.Errorf("shellspec not found in PATH\nInstall it: curl -fsSL https://git.io/shellspec | sh")
			}

			// Build shellspec arguments
			var ssArgs []string

			// Shell selection: CLI flag > config.Test.Shell > config.Shell
			s := cfg.Test.Shell
			if s == "" {
				s = cfg.Shell
			}
			if cmd.Flags().Changed("shell") {
				s = shell
			}
			if s != "" {
				ssArgs = append(ssArgs, "--shell", s)
			}

			// Jobs: CLI flag > config.Test.Jobs
			j := cfg.Test.Jobs
			if cmd.Flags().Changed("jobs") {
				j = jobs
			}
			if j > 0 {
				ssArgs = append(ssArgs, "--jobs", strconv.Itoa(j))
			}

			// Format: CLI flag > config.Test.Format
			f := cfg.Test.Format
			if cmd.Flags().Changed("format") {
				f = format
			}
			if f != "" {
				ssArgs = append(ssArgs, "--format", f)
			}

			// Append passthrough args (after --)
			ssArgs = append(ssArgs, args...)

			// Execute shellspec
			proc := exec.Command(shellspecPath, ssArgs...)
			proc.Stdin = os.Stdin
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr
			proc.Dir = cwd

			if err := proc.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&initFlag, "init", false, "initialize ShellSpec test structure")
	cmd.Flags().StringVar(&shell, "shell", "", "shell to run tests (e.g. bash, sh, zsh)")
	cmd.Flags().IntVar(&jobs, "jobs", 0, "number of parallel jobs")
	cmd.Flags().StringVar(&format, "format", "", "output format: progress, documentation, tap, junit")

	return cmd
}
