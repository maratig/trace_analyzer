package app

import (
	"context"
	"fmt"
	"sync"

	errPkg "github.com/maratig/trace_analyzer/internal/error"
)

type App struct {
	mx               sync.Mutex
	tracesInProgress []string
}

func New() *App {
	return &App{}
}

func (a *App) ListenTraceEvents(ctx context.Context, sourcePath string) error {
	if ctx == nil {
		return errPkg.ErrNilContext
	}
	if sourcePath == "" {
		return errPkg.ErrEmptyEventsEndpoint
	}

	a.mx.Lock()
	defer a.mx.Unlock()

	for _, trace := range a.tracesInProgress {
		if sourcePath == trace {
			return nil
		}
	}

	a.tracesInProgress = append(a.tracesInProgress, sourcePath)
	if err := a.runListening(ctx, sourcePath); err != nil {
		return fmt.Errorf("failed to run trace listening; %w", err)
	}

	return nil
}

func (a *App) runListening(ctx context.Context, sourcePath string) error {

}
