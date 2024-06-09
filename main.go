package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/maratig/trace_analyzer/internal/server"
	"github.com/maratig/trace_analyzer/pkg/app"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	application := app.New()
	srv, err := server.StartRestServer(ctx, application)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	srv.Shutdown(ctx)
}

func main2() {
	f, err := os.Open("sync-03062024.out")
	if err != nil {
		panic(err)
	}
	r, err := trace.NewReader(f)
	if err != nil {
		panic(err)
	}

	//writes := 0
	invokers := make(map[trace.GoID][]trace.Time)
	for {
		//for writes < 100 {
		ev, err := r.ReadEvent()
		if err != nil {
			break
		}
		if ev.Kind() != trace.EventStateTransition {
			continue
		}
		st := ev.StateTransition()
		if st.Resource.Kind == trace.ResourceGoroutine {
			id := st.Resource.Goroutine()
			if id == 4777 && ev.Goroutine() != 4777 && ev.Goroutine() != trace.NoGoroutine {
				println(ev.String())
				invokers[ev.Goroutine()] = append(invokers[ev.Goroutine()], ev.Time())
				//writes++
			}
		}
	}
	for inID, ts := range invokers {
		fmt.Printf("invoker: %d, time: %d\n", inID, ts)
	}
}
