package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lazy",
	Short: "lazy command line tools",
}

func main() {
	rootCmd.Execute()
}
