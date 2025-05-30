package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/trace"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/api/object"
	"github.com/maratig/trace_analyzer/internal/helper"
)

const defaultNumberOfIdlingGoroutines = 20

type (
	TraceProcess struct {
		id            int
		cancel        context.CancelFunc
		sourcePath    string
		err           error
		mx            sync.RWMutex
		lastEventTime trace.Time
		// livingStats contains all active (live) goroutines
		livingStats map[trace.GoID]*goroutineStat
		// terminatedStats contains all destroyed goroutines
		terminatedStats map[trace.GoID]*goroutineStat
		// idlingGors contains a short list of idling goroutines sorted by idling time
		// TODO Сейчас в процессе: умершие горутины переходят в terminatedStats, также idlingGors хранит висящие горутины,
		// умершая горутина удаляется из idlingGors (если она там есть). При этом idlingGors пополняется до конца в
		// момент запроса клиентом TopIdlingGoroutines. Т.е. idlingGors может в этот момент заполняться полностью
		// с нуля, а может частично. В работе: TopIdlingGoroutines
		idlingGors []*goroutineStat
	}

	goroutineStat struct {
		gID         trace.GoID
		firstSeen   trace.Time
		parentStack string
		stack       string
		// goroutine execution time in nanoseconds
		execDuration time.Duration
		// lastRunning is the time when goroutine was switched to Running
		lastRunning trace.Time
		// lastStop is the time when goroutine was switched from Running to another state
		lastStop trace.Time
	}
)

func NewTraceProcessor(id int, cancel context.CancelFunc, sourcePath string) (*TraceProcess, error) {
	if cancel == nil {
		return nil, apiError.ErrNilContext
	}
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	livingStats := make(map[trace.GoID]*goroutineStat)
	terminatedStats := make(map[trace.GoID]*goroutineStat)
	idlingGors := make([]*goroutineStat, 0, defaultNumberOfIdlingGoroutines)

	return &TraceProcess{
		id:              id,
		cancel:          cancel,
		sourcePath:      sourcePath,
		livingStats:     livingStats,
		terminatedStats: terminatedStats,
		idlingGors:      idlingGors,
	}, nil
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
	tip.mx.Lock()
	defer tip.mx.Unlock()

	tip.fillIdling()

	return tip.idlingAsTop()
}

func (tip *TraceProcess) processEvent(ev *trace.Event) {
	tip.mx.Lock()
	defer tip.mx.Unlock()

	tip.lastEventTime = ev.Time()
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

	gStat, ok := tip.livingStats[gID]
	if !ok {
		gStat = &goroutineStat{gID: gID, firstSeen: ev.Time()}
		tip.livingStats[gID] = gStat
	}
}

func (tip *TraceProcess) processTransitionEvent(ev *trace.Event) {
	st := ev.StateTransition()
	// TODO analyze if other kind of events should be considered
	if st.Resource.Kind != trace.ResourceGoroutine {
		return
	}

	gID := st.Resource.Goroutine()
	from, to := st.Goroutine()
	if to == trace.GoNotExist {
		tip.handleTerminated(gID)
		return
	}

	gStat, ok := tip.livingStats[gID]
	if !ok {
		var sb strings.Builder
		for frame := range st.Stack.Frames() {
			sb.WriteString(fmt.Sprintf("\t%s @ 0x%x\n\t\t%s:%d\n", frame.Func, frame.PC, frame.File, frame.Line))
		}

		var psb strings.Builder
		for frame := range ev.Stack().Frames() {
			psb.WriteString(fmt.Sprintf("\t%s @ 0x%x\n\t\t%s:%d\n", frame.Func, frame.PC, frame.File, frame.Line))
		}

		gStat = &goroutineStat{gID: gID, firstSeen: ev.Time(), parentStack: psb.String(), stack: sb.String()}
		tip.livingStats[gID] = gStat
	}

	if to == trace.GoRunning {
		gStat.lastRunning = ev.Time()
		tip.removeFromIdling(gStat)
	}
	if from == trace.GoRunning {
		gStat.execDuration += ev.Time().Sub(gStat.lastRunning)
		gStat.lastStop = ev.Time()
	}
}

// handleTerminated moves the corresponding goroutineStat from livingStats to terminatedStats and removes the goroutine
// from idlingGors (if exists)
func (tip *TraceProcess) handleTerminated(gID trace.GoID) {
	stat, ok := tip.livingStats[gID]
	if ok {
		delete(tip.livingStats, gID)
		tip.terminatedStats[gID] = stat
		tip.removeFromIdling(stat)
	}
}

func (tip *TraceProcess) fillIdling() {
	if len(tip.idlingGors) == cap(tip.idlingGors) {
		return
	}

	keys := make([]trace.Time, 0, cap(tip.idlingGors))
	for _, ig := range tip.idlingGors {
		keys = append(keys, ig.lastStop)
	}

	maxIndex := cap(keys) - 1
	for _, stat := range tip.livingStats {
		if stat.lastStop == 0 || len(keys) == cap(keys) && stat.lastStop > keys[maxIndex] {
			continue
		}

		idx, found := slices.BinarySearch(keys, stat.lastStop)
		if found || idx == maxIndex && len(keys) == cap(keys) {
			keys[idx], tip.idlingGors[idx] = stat.lastStop, stat
			continue
		}

		if len(keys) < cap(keys) {
			keys = append(keys, stat.lastStop)
			tip.idlingGors = append(tip.idlingGors, stat)
		}
		if idx == len(keys)-1 {
			continue
		}

		copy(keys[idx+1:], keys[idx:])
		copy(tip.idlingGors[idx+1:], tip.idlingGors[idx:])
		keys[idx], tip.idlingGors[idx] = stat.lastStop, stat
	}
}

func (tip *TraceProcess) removeFromIdling(stat *goroutineStat) {
	lastIdling := len(tip.idlingGors) - 1
	if lastIdling == -1 || stat.lastStop > tip.idlingGors[lastIdling].lastStop {
		return
	}

	index := -1
	for i, idling := range tip.idlingGors {
		if idling.gID == stat.gID {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	if index == lastIdling {
		tip.idlingGors = tip.idlingGors[:lastIdling]
		return
	}

	copy(tip.idlingGors[index:], tip.idlingGors[index+1:])
	tip.idlingGors = tip.idlingGors[:len(tip.idlingGors)-1]
}

func (tip *TraceProcess) idlingAsTop() []object.TopGoroutine {
	if len(tip.idlingGors) == 0 {
		return nil
	}

	ret := make([]object.TopGoroutine, 0, len(tip.idlingGors))
	for _, stat := range tip.idlingGors {
		var idleDur time.Duration
		if stat.lastRunning < stat.lastStop {
			idleDur = tip.lastEventTime.Sub(stat.lastStop)
		}
		ret = append(ret, object.TopGoroutine{
			ID:           stat.gID,
			ParentStack:  stat.parentStack,
			Stack:        stat.stack,
			ExecDuration: stat.execDuration,
			IdleDuration: idleDur,
		})
	}

	return ret
}
