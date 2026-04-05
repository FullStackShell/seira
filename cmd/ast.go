package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/Hayao0819/seira/shellparse"
	"github.com/spf13/cobra"
)

func astCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ast <file>",
		Short: "Parse a shell script and output its AST as JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parser := shellparse.NewParser()
			parsed, err := parser.ParseFile(args[0])
			if err != nil {
				return err
			}
			data, err := json.Marshal(parsed)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}
