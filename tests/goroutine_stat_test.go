package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maratig/trace_analyzer/internal/service"
	extApp "github.com/maratig/trace_analyzer/pkg/ext_app"
)

func TestGoroutineStat(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
	})

	addr := "127.0.0.1:11000"
	require.NoError(t, extApp.RunExternalApp(ctx, addr))
	tp, err := service.NewTraceProcessor(0, fmt.Sprintf("http://%s/debug/pprof/trace", addr))
	require.NoError(t, err)
	require.NoError(t, tp.Run(ctx))
	time.Sleep(5 * time.Second)
	gors := tp.TopIdlingGoroutines()
	assert.NotEmpty(t, gors)
}
