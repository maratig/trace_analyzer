package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/maratig/trace_analyzer/cmd"
	"github.com/maratig/trace_analyzer/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	addr := "127.0.0.1:15000"
	go func() {
		err := cmd.RunExtTestApp(ctx, addr)
		require.NoError(t, err)
	}()

	hp, err := service.NewHeapProcessor(0, "http://"+addr+"/debug/pprof/heap")
	require.NoError(t, err)
	err = hp.Run(ctx)
	require.NoError(t, err)

	time.Sleep(4 * time.Minute)
	profiles := hp.Profiles()
	var current time.Time
	for _, profile := range profiles {
		fmt.Printf("%.2f\n", profile.Sub(current).Seconds())
		current = profile
	}
	assert.NotEmpty(t, profiles)
}
