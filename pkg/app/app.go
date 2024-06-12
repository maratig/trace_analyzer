package app

import (
	"context"
	"errors"
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
		mx sync.Mutex
		// nextID starts from 1
		nextID         int
		traceProcesses []*service.TraceProcess
	}
)

func New() *App {
	once.Do(func() {
		appInstance = &App{nextID: -1}
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

	for _, tp := range a.traceProcesses {
		if tp.IsInProgress(sourcePath) {
			return 0, apiError.ErrTraceAlreadyRunning
		}
	}

	newCtx, cancel := context.WithCancel(ctx)
	a.nextID++
	tp, err := service.NewTraceProcessor(a.nextID, cancel, sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create a trace processor: %w", err)
	}
	a.traceProcesses = append(a.traceProcesses, tp)

	if err = tp.RunListening(newCtx); err != nil {
		return 0, fmt.Errorf("failed to run trace listening; %w", err)
	}

	return a.nextID, nil
}

func (a *App) Stats(ctx context.Context, id int) (int, error) {
	if ctx == nil {
		return 0, apiError.ErrNilContext
	}
	if id < 0 {
		return 0, errors.New("id must not be negative")
	}
	if len(a.traceProcesses) == 0 || id >= len(a.traceProcesses) {
		return 0, errors.New("no item with given id")
	}

	return a.traceProcesses[id].NumberOfGoroutines(), nil
}
