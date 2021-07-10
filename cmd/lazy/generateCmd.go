package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/isbang/lazy/internal/generate"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}

		stat, err := os.Stat(args[0])
		if err != nil {
			return err
		}

		if !stat.IsDir() {
			return fmt.Errorf("invalid argument")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// if err := os.MkdirAll("lazy", 0755); err != nil {
		// 	return err
		// }

		if err := generate.ClientStub(args[0], ""); err != nil {
			return err
		}

		if err := generate.ServerStub(args[0], ""); err != nil {
			return err
		}

		return nil
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
