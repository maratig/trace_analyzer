package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    = "0.0.1"
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version number of Trace analyzer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Trace analyzer version", version)
		},
	}
)
