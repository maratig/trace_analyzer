package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maratig/trace_analyzer/cmd"
	traceProcess "github.com/maratig/trace_analyzer/internal/service/trace_process"
)

func TestGoroutineStat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	addr := "127.0.0.1:11000"
	go func() {
		err := cmd.RunExtTestApp(ctx, addr)
		require.NoError(t, err)
	}()

	tp, err := traceProcess.NewTraceProcessor("http://" + addr + "/debug/pprof/trace")
	require.NoError(t, err)
	require.NoError(t, tp.Run(ctx))
	time.Sleep(5 * time.Second)
	gors := tp.TopIdlingGoroutines()
	assert.NotEmpty(t, gors)
	fmt.Printf("%v", gors[0])
}
