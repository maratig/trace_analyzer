package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maratig/trace_analyzer/cmd"
	heapProcess "github.com/maratig/trace_analyzer/internal/service/heap_process"
)

// TestProfiles runs a test application and a heap profiles collector. Then checks that collector returns a list of
// heap profiles (a slice of bytes for now)
func TestProfiles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	addr := "127.0.0.1:15000"
	go func() {
		err := cmd.RunExtTestApp(ctx, addr)
		require.NoError(t, err)
	}()

	opts := []heapProcess.Option{
		heapProcess.WithProfileRangeConfig(1*time.Second, 30*time.Second),
		heapProcess.WithProfileRangeConfig(10*time.Second, 30*time.Second),
	}
	hp, err := heapProcess.NewHeapProcessor("http://"+addr+"/debug/pprof/heap", opts...)
	require.NoError(t, err)
	err = hp.Run(ctx)
	require.NoError(t, err)

	time.Sleep(31 * time.Second)
	profiles, err := hp.HeapProfilesSummary()
	require.NoError(t, err)
	assert.Len(t, profiles, 2)
	assert.True(t, len(profiles[0]) > 29 && len(profiles[0]) < 32)
	assert.True(t, len(profiles[0]) > 3 && len(profiles[1]) < 5)
}
