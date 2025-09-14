package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "trace_analyzer",
	Short: "Trace analyzer helps to analyze Go profiles and traces",
}

func init() {
	rootCmd.AddCommand(versionCmd)
	initAnalyzerCmdFlags()
	rootCmd.AddCommand(analyzerCmd)
	initExtTestAppCmdFlags()
	rootCmd.AddCommand(extTestAppCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
