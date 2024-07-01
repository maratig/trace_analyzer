package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/trace"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/api/object"
	"github.com/maratig/trace_analyzer/internal/helper"
)

const (
	defaultNumberOfGoroutines    = 10000
	defaultNumberOfTopGoroutines = 10
)

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
		parentStack string
		stack       string
		// goroutine execution time in nanoseconds
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
			// TODO consider not break the process
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

// TopIdlingGoroutines returns defaultNumberOfTopGoroutines most idling goroutines
func (tip *TraceProcess) TopIdlingGoroutines() []object.TopGoroutine {
	tip.mx.RLock()
	defer tip.mx.RUnlock()

	numberOfTopGoroutines := defaultNumberOfTopGoroutines
	if numberOfTopGoroutines > len(tip.stats) {
		numberOfTopGoroutines = len(tip.stats)
	}

	top := helper.NewKeyValueSorter[float64, *goroutineStat](numberOfTopGoroutines)
	// First "numberOfTopGoroutines" goroutines are considered as top idling
	for i := 0; i < numberOfTopGoroutines; i++ {
		gStat := tip.stats[i]
		ratio := float64(gStat.execTime) / float64(gStat.lastSeen.Sub(gStat.firstStart))
		top.Add(ratio, gStat)
	}
	sort.Sort(top)

	// The rest part of goroutines are being compared with threshold i.e. goroutines with threshold
	threshold := top.LastKey()
	for i := numberOfTopGoroutines; i < len(tip.stats); i++ {
		gStat := tip.stats[i]
		ratio := float64(gStat.execTime) / float64(gStat.lastSeen.Sub(gStat.firstStart))
		if ratio > threshold {
			continue
		}

		// Insert found idling goroutine and push away the last one from the "top"
		top.InsertAndShift(ratio, gStat)
		threshold = top.LastKey()
	}

	ret := make([]object.TopGoroutine, 0, numberOfTopGoroutines)
	for _, t := range top.Values() {
		ret = append(ret, object.TopGoroutine{
			ID:           t.gID,
			ParentStack:  t.parentStack,
			Stack:        t.stack,
			ExecDuration: t.execTime,
			LiveDuration: t.lastSeen.Sub(t.firstStart),
		})
	}

	return ret
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
	if gID == trace.NoGoroutine {
		return
	}

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
		var sb strings.Builder
		st.Stack.Frames(func(f trace.StackFrame) bool {
			fmt.Fprintf(&sb, "\t%s @ 0x%x\n", f.Func, f.PC)
			fmt.Fprintf(&sb, "\t\t%s:%d\n", f.File, f.Line)
			return true
		})

		var psb strings.Builder
		ev.Stack().Frames(func(f trace.StackFrame) bool {
			fmt.Fprintf(&psb, "\t%s @ 0x%x\n", f.Func, f.PC)
			fmt.Fprintf(&psb, "\t\t%s:%d\n", f.File, f.Line)
			return true
		})

		gStat = &goroutineStat{gID: gID, firstStart: ev.Time(), parentStack: psb.String(), stack: sb.String()}
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
