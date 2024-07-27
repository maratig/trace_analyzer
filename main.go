package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/maratig/trace_analyzer/app"
	"github.com/maratig/trace_analyzer/internal/server"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	port := flag.String("port", "", "port usage")
	endpointConnectionWait := flag.Int("endpoint-connection-wait", 0, "wait connection endpoint")
	flag.Parse()
	cfg := app.Config{Port: *port, EndpointConnectionWait: *endpointConnectionWait}
	application := app.NewApp(cfg)

	srv, err := server.StartRestServer(ctx, application)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	srv.Shutdown(ctx)
}
