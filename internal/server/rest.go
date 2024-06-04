package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/maratig/trace_analyzer/internal/app"
	errPkg "github.com/maratig/trace_analyzer/internal/error"
)

func StartRestServer(ctx context.Context, application *app.App) (*http.Server, error) {
	if ctx == nil {
		return nil, errPkg.ErrNilContext
	}
	if application == nil {
		return nil, errPkg.ErrNilApp
	}

	router := http.NewServeMux()
	router.HandleFunc("/trace-events/listen", application.RunTraceEventsListening)

	srv := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second, // nolint:gomnd
		WriteTimeout:      15 * time.Second, // nolint:gomnd
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to listen and serve; %v", err)
		}
	}()

	return srv, nil
}
