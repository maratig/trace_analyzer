package heap_process

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/pprof/profile"

	apiError "github.com/maratig/trace_analyzer/api/error"
)

// defaultProfileFetchInterval is a time interval used for fetching heap profile data from source path (endpoint)
const defaultProfileFetchInterval = 5 * time.Second

// defaultRangeConfigs is a default configuration of profile ranges. It has 3 range configs, every range config describes
// what time interval will be used for collecting profiles and what the range size is
var defaultRangeConfigs = []rangeConfig{
	{defaultProfileFetchInterval, 30 * time.Minute},
	{1 * time.Minute, 3 * time.Hour},
	{30 * time.Minute, 24 * time.Hour},
}

type (
	HeapProcess struct {
		cfg  config
		mx   sync.RWMutex
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

	config struct {
		sourcePath string
		ranges     []rangeConfig
	}

	rangeConfig struct {
		interval time.Duration
		size     time.Duration
	}

	Option func(hp *HeapProcess)
)

func WithProfileRangeConfig(interval, size time.Duration) Option {
	return func(hp *HeapProcess) {
		if interval > 0 && size > 0 && size >= interval {
			hp.cfg.ranges = append(hp.cfg.ranges, rangeConfig{interval: interval, size: size})
		}
	}
}

func NewHeapProcessor(sourcePath string, opts ...Option) (*HeapProcess, error) {
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	ret := HeapProcess{
		cfg: config{
			sourcePath: sourcePath,
		},
	}
	// Applying options
	for _, opt := range opts {
		opt(&ret)
	}
	// Calculating profile ranges according to configuration
	rangeConfigs := defaultRangeConfigs
	if len(ret.cfg.ranges) > 0 {
		rangeConfigs = ret.cfg.ranges
	}
	profiles := make([][]heapProfile, len(rangeConfigs))
	for i, r := range rangeConfigs {
		profiles[i] = make([]heapProfile, 0, r.size/r.interval+1)
	}
	ret.stat = &heapStat{profiles}

	return &ret, nil
}

func (hp *HeapProcess) IsInProgress(sourcePath string) bool {
	return hp.cfg.sourcePath == sourcePath
}

func (hp *HeapProcess) Run(ctx context.Context) error {
	if ctx == nil {
		return apiError.ErrNilContext
	}

	for i := range hp.cfg.ranges {
		go func(c context.Context, rangeIndex int, p *HeapProcess) {
			tmr := time.NewTimer(0)
			defer tmr.Stop()
			rangeCfg := hp.cfg.ranges[rangeIndex]

			for {
				select {
				case <-c.Done():
					return
				case <-tmr.C:
					tmr.Reset(rangeCfg.interval)

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
					p.stat.addProfile(time.Now(), rangeIndex, data)
					p.mx.Unlock()
				}
			}
		}(ctx, i, hp)
	}

	return nil
}

func (hp *HeapProcess) Profiles() ([][]*profile.Profile, error) {
	hp.mx.RLock()
	defer hp.mx.RUnlock()

	ret := make([][]*profile.Profile, 0, len(hp.stat.profiles))
	for _, profiles := range hp.stat.profiles {
		profilesInRange := make([]*profile.Profile, 0, len(profiles))
		for _, pf := range profiles {
			pfToAdd, err := profile.ParseData(pf.data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse profile; %w", err)
			}
			profilesInRange = append(profilesInRange, pfToAdd)
		}
		ret = append(ret, profilesInRange)
	}

	return ret, nil
}

func (hs *heapStat) addProfile(receivedAt time.Time, rangeIndex int, data []byte) {
	profileToAdd := heapProfile{data, receivedAt}
	if len(hs.profiles[rangeIndex]) == 0 {
		hs.profiles[rangeIndex] = append(hs.profiles[rangeIndex], profileToAdd)
		return
	}

	if len(hs.profiles[rangeIndex]) < cap(hs.profiles[rangeIndex]) {
		hs.profiles[rangeIndex] = append(hs.profiles[rangeIndex], heapProfile{})
	}
	copy(hs.profiles[rangeIndex][1:], hs.profiles[rangeIndex])
	hs.profiles[rangeIndex][0] = profileToAdd
}
