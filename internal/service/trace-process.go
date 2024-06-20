package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/exp/trace"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/internal/helper"
)

const defaultNumberOfGoroutines = 10000

type (
	TraceProcess struct {
		id         int
		cancel     context.CancelFunc
		sourcePath string
		mx         sync.RWMutex
		err        error
		statIndex  map[trace.GoID]*goroutineStat
		stats      []*goroutineStat
	}

	goroutineStat struct {
		gID         trace.GoID
		firstStart  trace.Time
		startStack  string
		execTime    time.Duration
		lastRunning trace.Time
		lastSeen    trace.Time
	}
)

func NewTraceProcessor(id int, cancel context.CancelFunc, sourcePath string) (*TraceProcess, error) {
	if cancel == nil {
		return nil, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	statIndex := make(map[trace.GoID]*goroutineStat, defaultNumberOfGoroutines)
	stats := make([]*goroutineStat, 0, defaultNumberOfGoroutines)

	return &TraceProcess{id: id, cancel: cancel, sourcePath: sourcePath, statIndex: statIndex, stats: stats}, nil
}

func (tip *TraceProcess) IsInProgress(sourcePath string) bool {
	return tip.sourcePath == sourcePath
}

func (tip *TraceProcess) Run(ctx context.Context) error {
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

			tip.processEvent(&event)
		}
	}()

	return nil
}

func (tip *TraceProcess) TopIdles() (trace.GoID, time.Duration, time.Duration) {
	tip.mx.RLock()
	defer tip.mx.RUnlock()

	gStat, ok := tip.statIndex[4765]
	if !ok {
		return 0, 0, 0
	}

	return gStat.gID, gStat.execTime, gStat.lastSeen.Sub(gStat.firstStart)
}

func (tip *TraceProcess) processEvent(ev *trace.Event) {
	tip.mx.Lock()
	defer tip.mx.Unlock()

	if len(tip.stats) == cap(tip.stats) {
		// TODO improve allocation logic. Something like +100%, +80% etc. Maybe add some rule depending on the number
		// of goroutines
		newStats := make([]*goroutineStat, len(tip.stats), len(tip.stats)*2)
		copy(newStats, tip.stats)
		tip.stats = newStats
	}

	switch ev.Kind() {
	case trace.EventStateTransition:
		tip.processTransitionEvent(ev)
	default:
		tip.processGenericEvent(ev)
	}
}

func (tip *TraceProcess) processGenericEvent(ev *trace.Event) {
	gID := ev.Goroutine()
	gStat, ok := tip.statIndex[gID]
	if !ok {
		gStat = &goroutineStat{gID: gID, firstStart: ev.Time()}
		tip.statIndex[gID] = gStat
		tip.stats = append(tip.stats, gStat)
	}
	gStat.lastSeen = ev.Time()
}

func (tip *TraceProcess) processTransitionEvent(ev *trace.Event) {
	st := ev.StateTransition()
	// TODO analyze if other kind of events should be considered
	if st.Resource.Kind != trace.ResourceGoroutine {
		return
	}

	gID := st.Resource.Goroutine()
	gStat, ok := tip.statIndex[gID]
	if !ok {
		gStat = &goroutineStat{gID: gID, firstStart: ev.Time(), startStack: fmt.Sprintf("%v", st.Stack)}
		tip.statIndex[gID] = gStat
		tip.stats = append(tip.stats, gStat)
	}

	from, to := st.Goroutine()
	if to == trace.GoRunning {
		gStat.lastRunning = ev.Time()
	}
	if from == trace.GoRunning {
		gStat.execTime += ev.Time().Sub(gStat.lastRunning)
	}
	gStat.lastSeen = ev.Time()
}
