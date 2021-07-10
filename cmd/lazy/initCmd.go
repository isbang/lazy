package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/isbang/lazy/internal/generate"
)

var jobNameRegex = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

var initCmd = &cobra.Command{
	Use: "init [JobName]...",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if !jobNameRegex.MatchString(arg) {
				// if !jobNameRegex.MatchString(arg) ||
				// 	arg == "After" { // DoAfter 라는 함수와 이름이 충돌하여 어쩔 수 없이 적용
				return fmt.Errorf("invalid job name: %s", arg)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll("lazy/job", 0755); err != nil {
			return err
		}

		if err := generate.GoGen("lazy/generate.go"); err != nil {
			return err
		}

		for _, job := range args {
			path := "lazy/job/" + strings.ToLower(job) + ".go"
			if err := generate.JobStub(path, job); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
