package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	apiError "github.com/maratig/trace_analyzer/api/error"
)

const (
	profileFetchInterval = 5 * time.Second
	profileRanges        = 3
)

var rangeStepsAndSizes = [profileRanges][2]time.Duration{
	{profileFetchInterval, 30 * time.Minute},
	{1 * time.Minute, 6 * time.Hour},
	{30 * time.Minute, 24 * time.Hour},
}

type (
	HeapProcess struct {
		id   int
		cfg  dataSourceConfig
		mx   sync.Mutex
		err  error
		stat *heapStat
	}

	heapStat struct {
		profiles [][]heapProfile
	}
	heapProfile struct {
		data       []byte
		receivedAt time.Time
	}
)

func NewHeapProcessor(id int, sourcePath string, opts ...ConfigOption) (*HeapProcess, error) {
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	profiles := make([][]heapProfile, profileRanges)
	for i := 0; i < profileRanges; i++ {
		profiles[i] = make([]heapProfile, 0, rangeStepsAndSizes[i][1]/rangeStepsAndSizes[i][0])
	}
	ret := HeapProcess{
		id: id,
		cfg: dataSourceConfig{
			sourcePath:             sourcePath,
			endpointConnectionWait: defaultEndpointConnectionWait,
		},
		stat: &heapStat{profiles: profiles},
	}
	for _, opt := range opts {
		opt(&ret.cfg)
	}

	return &ret, nil
}

func (hp *HeapProcess) IsInProgress(sourcePath string) bool {
	return hp.cfg.sourcePath == sourcePath
}

func (hp *HeapProcess) Run(ctx context.Context) error {
	if ctx == nil {
		return apiError.ErrNilContext
	}

	go func(c context.Context, p *HeapProcess) {
		tmr := time.NewTimer(0)
		now := time.Now()
		defer tmr.Stop()
		for {
			select {
			case <-c.Done():
				return
			case <-tmr.C:
				tmr.Reset(profileFetchInterval)
				now = now.Add(profileFetchInterval)

				resp, err := http.Get(p.cfg.sourcePath)
				if err != nil {
					p.mx.Lock()
					p.err = fmt.Errorf("failed to get heap profile; %w", err)
					p.mx.Unlock()
					return
				}

				data, err := io.ReadAll(resp.Body)
				if err != nil {
					p.mx.Lock()
					p.err = fmt.Errorf("failed to read heap profile response body; %w", err)
					p.mx.Unlock()
					return
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					p.mx.Lock()
					p.err = fmt.Errorf("heap profile response statusCode=%d; status=%s", resp.StatusCode, resp.Status)
					p.mx.Unlock()
					return
				}

				p.mx.Lock()
				p.stat.addProfile(now, data)
				p.mx.Unlock()
			}
		}
	}(ctx, hp)

	return nil
}

func (hs *heapStat) addProfile(tm time.Time, data []byte) {
	var lastFromRange heapProfile
	profileToAdd := heapProfile{data: data, receivedAt: tm}
	for i := 0; i < len(hs.profiles); i++ {
		if len(hs.profiles[i]) == 0 {
			hs.profiles[i] = append(hs.profiles[i], profileToAdd)
			break
		} else {
			interval := rangeStepsAndSizes[i][0]
			if tm.Sub(hs.profiles[i][0].receivedAt) < interval {
				break
			}

			if cap(hs.profiles[i]) == len(hs.profiles[i]) {
				lastFromRange = hs.profiles[i][len(hs.profiles[i])-1]
			}

			copy(hs.profiles[i][1:], hs.profiles[i][0:])
			hs.profiles[i][0] = profileToAdd
			profileToAdd = lastFromRange
		}
	}
}
