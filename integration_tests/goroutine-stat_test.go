package integration_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"github.com/maratig/trace_analyzer/cmd"
	traceProcess "github.com/maratig/trace_analyzer/internal/service/trace_process"
)

func TestGoroutineStat(t *testing.T) {
	addr := "127.0.0.1:11000"

	t.Run("endpoint does not respond, should return a retryable error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		tp, err := traceProcess.NewTraceProcessor("http://some-fake-url/debug/pprof/trace")
		require.NoError(t, err)
		require.NoError(t, tp.Run(ctx))
		time.Sleep(8 * time.Second)
		gors, err := tp.TopIdlingGoroutines()
		assert.Error(t, err)
		assert.Nil(t, gors)
		assert.True(t, errors.Is(err, apiError.ErrRetryable))
	})

	extTestAppCtx, extTextAppCancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		extTextAppCancel()
	})
	go func() {
		err := cmd.RunExtTestApp(extTestAppCtx, addr)
		require.NoError(t, err)
	}()

	t.Run("endpoint responds, should return goroutines", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		tp, err := traceProcess.NewTraceProcessor("http://" + addr + "/debug/pprof/trace")
		require.NoError(t, err)
		require.NoError(t, tp.Run(ctx))
		time.Sleep(3 * time.Second)
		gors, err := tp.TopIdlingGoroutines()
		require.NoError(t, err)
		assert.NotEmpty(t, gors)
	})

	// Wait for the previous subtest's trace connection to fully close before connecting again.
	time.Sleep(500 * time.Millisecond)

	t.Run("endpoint responds, then fails, should return a retryable error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		tp, err := traceProcess.NewTraceProcessor("http://" + addr + "/debug/pprof/trace")
		require.NoError(t, err)
		require.NoError(t, tp.Run(ctx))
		time.Sleep(3 * time.Second)
		gors, err := tp.TopIdlingGoroutines()
		require.NoError(t, err)
		assert.NotEmpty(t, gors)

		extTextAppCancel()
		time.Sleep(3 * time.Second)
		gors, err = tp.TopIdlingGoroutines()
		assert.Error(t, err)
		assert.Nil(t, gors)
		assert.True(t, errors.Is(err, apiError.ErrRetryable))
	})
}
