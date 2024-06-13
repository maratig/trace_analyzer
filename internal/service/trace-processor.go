package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"

	"golang.org/x/exp/trace"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/internal/helper"
)

const defaultNumberOfGoroutines = 1000

type (
	TraceProcess struct {
		id         int
		cancel     context.CancelFunc
		sourcePath string
		mx         sync.RWMutex
		err        error
		// gID is a key, index in stats is a value
		statIndex map[trace.GoID]int
		stats     []traceStat
	}

	traceStat struct {
		gID         trace.GoID
		occurrences int
		states      map[trace.EventKind]int
	}
)

func NewTraceProcessor(id int, cancel context.CancelFunc, sourcePath string) (*TraceProcess, error) {
	if cancel == nil {
		return nil, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	statIndex := make(map[trace.GoID]int, defaultNumberOfGoroutines)
	stats := make([]traceStat, 0, defaultNumberOfGoroutines)

	return &TraceProcess{id: id, cancel: cancel, sourcePath: sourcePath, statIndex: statIndex, stats: stats}, nil
}

func (tip *TraceProcess) IsInProgress(sourcePath string) bool {
	return tip.sourcePath == sourcePath
}

// TODO temporary method
func (tip *TraceProcess) GoroutineStat() (trace.GoID, []string) {
	tip.mx.RLock()
	defer tip.mx.RUnlock()

	randomIdx := rand.IntN(len(tip.stats))
	states := make([]string, 0, len(tip.stats[randomIdx].states))
	for stateNum := range tip.stats[randomIdx].states {
		states = append(states, stateNum.String())
	}

	return tip.stats[randomIdx].gID, states
}

func (tip *TraceProcess) RunListening(ctx context.Context) error {
	if ctx == nil {
		return apiError.ErrNilContext
	}

	r, closer, err := helper.CreateTraceReader(tip.sourcePath)
	if err != nil {
		return fmt.Errorf("failed to create trace reader; %w", err)
	}

	go func() {
		defer closer.Close()

		for {
			if ctx.Err() != nil {
				return
			}

			event, err := r.ReadEvent()
			// TODO consider not breaking the process
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				tip.mx.Lock()
				tip.err = fmt.Errorf("failed to read event; %w", err)
				tip.mx.Unlock()
				return
			}

			tip.processEvent(event)
		}
	}()

	return nil
}

func (tip *TraceProcess) processEvent(ev trace.Event) {
	tip.mx.Lock()
	defer tip.mx.Unlock()

	if len(tip.stats) == cap(tip.stats) {
		newStats := make([]traceStat, len(tip.stats), len(tip.stats)*2)
		copy(newStats, tip.stats)
		tip.stats = newStats
	}

	gID := ev.Goroutine()
	statIdx, ok := tip.statIndex[gID]
	if !ok {
		states := map[trace.EventKind]int{ev.Kind(): 1}
		tip.stats = append(tip.stats, traceStat{gID: gID, occurrences: 1, states: states})
		tip.statIndex[gID] = len(tip.stats) - 1
	} else {
		tip.stats[statIdx].occurrences++
		tip.stats[statIdx].states[ev.Kind()]++
	}
}
