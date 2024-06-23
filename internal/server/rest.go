package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/app"
)

func StartRestServer(ctx context.Context, application *app.App) (*http.Server, error) {
	if ctx == nil {
		return nil, apiError.ErrNilContext
	}
	if application == nil {
		return nil, apiError.ErrNilApp
	}

	h, err := NewHandler(ctx, application)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler; %w", err)
	}
	router := http.NewServeMux()
	router.HandleFunc("/trace-events/listen", h.RunTraceEventsListening)
	router.HandleFunc("/trace-events/top-goroutines", h.TopGoroutines)

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
