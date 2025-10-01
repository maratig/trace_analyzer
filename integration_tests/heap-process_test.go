//go:build integration

package integration_tests

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maratig/trace_analyzer/cmd"
	"github.com/maratig/trace_analyzer/internal/service"
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

	hp, err := service.NewHeapProcessor(0, "http://"+addr+"/debug/pprof/heap")
	require.NoError(t, err)
	err = hp.Run(ctx)
	require.NoError(t, err)

	time.Sleep(1 * time.Minute)
	profiles := hp.Profiles()
	assert.NotEmpty(t, profiles)
	r := bytes.NewReader(profiles[0])
	p, err := profile.Parse(r)
	require.NoError(t, err)
	assert.NotNil(t, p)
}
