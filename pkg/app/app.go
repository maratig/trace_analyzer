package app

import (
	"context"
	"fmt"
	"sync"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/internal/service"
)

var (
	appInstance *App
	once        sync.Once
)

type (
	App struct {
		mx               sync.Mutex
		nextID           int
		tracesInProgress []*service.TraceProcessor
	}
)

func New() *App {
	once.Do(func() {
		appInstance = &App{}
	})

	return appInstance
}

func (a *App) ListenTraceEvents(ctx context.Context, sourcePath string) (int, error) {
	if ctx == nil {
		return 0, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return 0, apiError.ErrEmptySourcePath
	}

	a.mx.Lock()
	defer a.mx.Unlock()

	for _, tip := range a.tracesInProgress {
		if tip.IsInProgress(sourcePath) {
			return 0, apiError.ErrTraceAlreadyRunning
		}
	}

	newCtx, cancel := context.WithCancel(ctx)
	a.nextID++
	tip, err := service.NewTraceProcessor(a.nextID, cancel, sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create a trace processor: %w", err)
	}
	a.tracesInProgress = append(a.tracesInProgress, tip)

	if err := tip.RunListening(newCtx); err != nil {
		return 0, fmt.Errorf("failed to run trace listening; %w", err)
	}

	return a.nextID, nil
}
