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

const (
	defaultPort                   = "10000"
	defaultEndpointConnectionWait = 10
)

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
	}

	Config struct {
		Port                   string
		EndpointConnectionWait int
	}
)

func initConfig(cfg Config) Config {
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.EndpointConnectionWait <= 0 {
		cfg.EndpointConnectionWait = defaultEndpointConnectionWait
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

	newCtx, cancel := context.WithCancel(ctx)
	a.nextID++
	tp, err := service.NewTraceProcessor(a.nextID, cancel, sourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create a trace processor: %w", err)
	}
	a.traceProcesses = append(a.traceProcesses, tp)

	if err = tp.Run(newCtx); err != nil {
		return 0, fmt.Errorf("failed to run trace listening; %w", err)
	}

	return a.nextID, nil
}

// TopIdlingGoroutines returns the top n goroutines having small execution_time/live_time ratio
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
