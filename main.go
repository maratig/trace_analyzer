package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/maratig/trace_analyzer/internal/server"
	"github.com/maratig/trace_analyzer/pkg/app"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	application := app.New()
	srv, err := server.StartRestServer(ctx, application)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	srv.Shutdown(ctx)
}
