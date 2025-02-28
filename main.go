package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/maratig/trace_analyzer/app"
	"github.com/maratig/trace_analyzer/internal/server"
)

var (
	appName = "trace_analyzer"
	version = "0.0.1"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	ver := flag.Bool("version", false, "print version and exit")
	port := flag.Int("port", 0, "port to be used in REST endpoint")
	endpointConnectionWait := flag.Int(
		"endpoint-connection-wait", 0, "time in seconds to wait for connection to endpoint",
	)
	flag.Parse()

	if ver != nil && *ver {
		fmt.Println(appName, "version", version)
		os.Exit(0)
	}

	cfg := app.Config{Port: *port, EndpointConnectionWait: *endpointConnectionWait}
	application := app.NewApp(cfg)

	srv, err := server.StartRestServer(ctx, application)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	srv.Shutdown(ctx)
}
