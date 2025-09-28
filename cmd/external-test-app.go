package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	extApp "github.com/maratig/trace_analyzer/pkg/ext_app"
)

const defaultExtTestAppAddr = "127.0.0.1:11000"

var extTestAppCmd = &cobra.Command{
	Use:   "ext-test-app",
	Short: "Run an external application to be used for testing the analyzer",
	PreRun: func(cmd *cobra.Command, args []string) {
		ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
		cmd.SetContext(ctx)
	},
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			panic(fmt.Sprintf("failed to parse external test app addr; %v", err))
		}

		if err := RunExtTestApp(cmd.Context(), addr); err != nil {
			panic(fmt.Sprintf("RunExtTestApp command failed; %v", err))
		}
	},
}

func initExtTestAppCmdFlags() {
	extTestAppCmd.Flags().StringP("addr", "a", "", "Address to be exposed to clients")
}

// RunExtTestApp is public for running integration tests
func RunExtTestApp(ctx context.Context, addr string) error {
	if addr == "" {
		addr = defaultExtTestAppAddr
	}
	if err := extApp.RunExternalApp(ctx, addr); err != nil {
		return fmt.Errorf("failed to run external app on addr=%s; %w", addr, err)
	}

	fmt.Printf("External test app is running on %s\n", addr)

	<-ctx.Done()
	return nil
}
