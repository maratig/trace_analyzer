package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/api/object"
	"github.com/maratig/trace_analyzer/internal/service"
)

// defaultApiPort is a default port for program's REST API
const defaultApiPort = 10000

var (
	appInstance *App
	once        sync.Once
)

type (
	App struct {
		cfg            Config
		mx             sync.Mutex
		nextID         int
		traceProcesses []*service.TraceProcess
		heapProcesses  []*service.HeapProcess
	}

	Config struct {
		ApiPort int
	}
)

func initConfig(cfg Config) Config {
	if cfg.ApiPort <= 0 {
		cfg.ApiPort = defaultApiPort
	}

	return cfg
}

func NewApp(cfg Config) *App {
	once.Do(func() {
		cfg = initConfig(cfg)
		appInstance = &App{cfg: cfg, nextID: -1}
	})

	return appInstance
}

func (a *App) GetConfig() Config {
	return a.cfg
}

// ProcessTraceSource creates a worker for reading and processing events from source. The returned int value is an id
// of the whole process for the given source. Later using that id one can get analytical info
func (a *App) ProcessTraceSource(ctx context.Context, sourcePath string) (int, error) {
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

	tp, err := service.NewTraceProcessor(len(a.traceProcesses), sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create a trace processor: %w", err)
	}
	a.traceProcesses = append(a.traceProcesses, tp)

	if err = tp.Run(ctx); err != nil {
		return 0, fmt.Errorf("failed to run trace listening; %w", err)
	}

	return len(a.traceProcesses) - 1, nil
}

// ProcessHeapSource creates a worker for reading and processing heap profile data from source. The returned int value
// is an id of the whole process for the given source. Later using that id one can get analytical info
func (a *App) ProcessHeapSource(ctx context.Context, sourcePath string) (int, error) {
	if ctx == nil {
		return 0, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return 0, apiError.ErrEmptySourcePath
	}

	a.mx.Lock()
	defer a.mx.Unlock()

	for _, hp := range a.heapProcesses {
		if hp.IsInProgress(sourcePath) {
			return 0, apiError.ErrHeapProcAlreadyRunning
		}
	}

	newCtx, cancel := context.WithCancel(ctx)
	hp, err := service.NewHeapProcessor(len(a.heapProcesses), cancel, sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create a heap profile processor: %w", err)
	}
	a.heapProcesses = append(a.heapProcesses, hp)

	if err = hp.Run(newCtx); err != nil {
		return 0, fmt.Errorf("failed to run heap profile processing; %w", err)
	}

	return len(a.heapProcesses) - 1, nil
}

// TopIdlingGoroutines returns the top n inactive goroutines
func (a *App) TopIdlingGoroutines(ctx context.Context, id int) ([]object.TopGoroutine, error) {
	if ctx == nil {
		return nil, apiError.ErrNilContext
	}
	if id < 0 {
		return nil, errors.New("id must not be negative")
	}
	if len(a.traceProcesses) == 0 || id >= len(a.traceProcesses) {
		return nil, errors.New("no item with given id")
	}

	return a.traceProcesses[id].TopIdlingGoroutines(), nil
}
