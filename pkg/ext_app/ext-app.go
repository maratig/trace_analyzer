package ext_app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

// RunExternalApp runs a test application, "addr" can be used by client for collecting pprof and trace metrics
func RunExternalApp(ctx context.Context, addr string) error {
	if ctx == nil {
		return errors.New("ctx must be not nil")
	}
	if addr == "" {
		return errors.New("addr must be not empty")
	}

	srv := startServerWithPprofAndTrace(addr)
	cancelAppFn := startApp(ctx)

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			panic(fmt.Sprintf("failed to shutdown http server; %v", err))
		}
		cancelAppFn()
	}()

	return nil
}

func startServerWithPprofAndTrace(addr string) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux, // handles pprof as well
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("failed to listen and serve, addr=%s; %v", addr, err))
		}
	}()

	return server
}

func startApp(ctx context.Context) context.CancelFunc {
	newCtx, cancel := context.WithCancel(ctx)

	go func(c context.Context) {
		a := make([]int, 0, 10000)
		for i := 0; i < cap(a); i++ {
			if c.Err() != nil {
				return
			}
			a = append(a, i)
			time.Sleep(10 * time.Millisecond)

			if len(a) == cap(a) {
				a = make([]int, 0, 10000)
				time.Sleep(1 * time.Second)
			}
		}
	}(newCtx)

	return cancel
}
