package cmd

import (
	"fmt"
	"os"
)

func Execute() error {
	root := rootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		return err
	}
	return nil
}
