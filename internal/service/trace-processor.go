package service

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/exp/trace"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/internal/helper"
)

const defaultNumberOfGoroutines = 1000

type (
	TraceProcessor struct {
		id         int
		cancel     context.CancelFunc
		sourcePath string
		mx         sync.RWMutex
		err        error
		// gID is a key, index in stats is a value
		statIndex map[int64]int
		stats     []traceStat
	}

	traceStat struct {
		gID int64
	}
)

func NewTraceProcessor(id int, cancel context.CancelFunc, sourcePath string) (*TraceProcessor, error) {
	if cancel == nil {
		return nil, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	statIndex := make(map[int64]int, defaultNumberOfGoroutines)
	stats := make([]traceStat, 0, defaultNumberOfGoroutines)

	return &TraceProcessor{id: id, cancel: cancel, sourcePath: sourcePath, statIndex: statIndex, stats: stats}, nil
}

func (tip *TraceProcessor) IsInProgress(sourcePath string) bool {
	return tip.sourcePath == sourcePath
}

// TODO temporary method
func (tip *TraceProcessor) NumberOfGoroutines() int {
	tip.mx.RLock()
	defer tip.mx.RUnlock()

	return len(tip.stats)
}

func (tip *TraceProcessor) RunListening(ctx context.Context) error {
	if ctx == nil {
		return apiError.ErrNilContext
	}

	r, closer, err := helper.CreateTraceReader(tip.sourcePath)
	if err != nil {
		return fmt.Errorf("failed to create trace reader; %w", err)
	}

	// TODO temporary limit for development and debugging purposes
	var counter int
	go func() {
		defer closer.Close()

		for counter < defaultNumberOfGoroutines {
			if ctx.Err() != nil {
				return
			}

			event, err := r.ReadEvent()
			// TODO consider not breaking the process
			if err != nil {
				tip.mx.Lock()
				tip.err = fmt.Errorf("failed to read event; %w", err)
				tip.mx.Unlock()
				return
			}

			tip.processEvent(event)
			counter++
		}
	}()

	return nil
}

func (tip *TraceProcessor) processEvent(ev trace.Event) {
	tip.mx.Lock()
	defer tip.mx.Unlock()

	// TODO allocate new space and move tip.stats there
	if len(tip.stats) == cap(tip.stats) {
	}

	gID := int64(ev.Goroutine())
	_, ok := tip.statIndex[gID]
	if !ok {
		tip.stats = append(tip.stats, traceStat{gID: gID})
		tip.statIndex[gID] = len(tip.stats) - 1
	} else {
		// TODO
	}
}
