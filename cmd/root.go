package cmd

import (
	"log/slog"

	"github.com/Hayao0819/nahi/cobrautils"
	"github.com/m-mizutani/clog"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	cmdReg  = cobrautils.Registory{}
)

func init() {
	cmdReg.Add(astCmd(), buildCmd(), newCmd(), installCmd(), depsCmd(), pathCmd(), lintCmd(), docCmd(), testCmd())
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "seira",
		Short:         "Bundle shell scripts into standalone executables",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelWarn
			if verbose {
				level = slog.LevelDebug
			}
			handler := clog.New(clog.WithColor(true), clog.WithLevel(level))
			slog.SetDefault(slog.New(handler))
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	cmdReg.Bind(cmd)

	return cmd
}
