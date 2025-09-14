package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/maratig/trace_analyzer/app"
	"github.com/maratig/trace_analyzer/internal/server"
	"github.com/spf13/cobra"
)

const defaultAnalyzerPort = 10000

var analyzerCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the analyzer application",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			panic(fmt.Sprintf("failed to parse analyzer port; %v", err))
		}

		if port <= 0 {
			port = defaultAnalyzerPort
		}

		runAnalyzer(port)
	},
}

func initAnalyzerCmdFlags() {
	analyzerCmd.Flags().IntP("port", "p", 0, "Port to be used in REST endpoint")
}

func runAnalyzer(port int) {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	cfg := app.Config{ApiPort: port}
	application := app.NewApp(cfg)

	srv, err := server.StartRestServer(ctx, application)
	if err != nil {
		panic(fmt.Sprintf("failed to start REST server; %v", err))
	}

	<-ctx.Done()
	srv.Shutdown(ctx)
}
