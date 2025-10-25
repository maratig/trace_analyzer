package heap_process

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeapProfilesSummary(t *testing.T) {
	data, err := os.ReadFile("test_data/profile.pprof")
	require.NoError(t, err)
	profiles := []heapProfile{
		{
			data:       data,
			receivedAt: time.Now(),
		},
	}
	heapProc := HeapProcess{stat: &heapStat{profiles: [][]heapProfile{profiles}}}
	summaries, err := heapProc.HeapProfilesSummary()
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	require.Len(t, summaries[0], 1)
	sm := summaries[0][0]
	inuseSpace := float64(sm.InuseSpace) / 1024
	allocSpace := float64(sm.AllocSpace) / 1024
	assert.InDelta(t, 2050.61, inuseSpace, 0.001)
	assert.InDelta(t, 2050.61, allocSpace, 0.001)
	assert.Equal(t, int64(6428), sm.InuseObjects)
	assert.Equal(t, int64(6428), sm.AllocObjects)
}
