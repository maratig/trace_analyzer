package service

import (
	"context"

	apiError "github.com/maratig/trace_analyzer/api/error"
)

type (
	HeapProcess struct {
		id  int
		cfg dataSourceConfig
		err error
	}
)

func NewHeapProcessor(id int, sourcePath string, opts ...ConfigOption) (*HeapProcess, error) {
	if sourcePath == "" {
		return nil, apiError.ErrEmptySourcePath
	}

	ret := HeapProcess{
		id: id,
		cfg: dataSourceConfig{
			sourcePath:             sourcePath,
			endpointConnectionWait: defaultEndpointConnectionWait,
		},
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

	go func(c context.Context, hp *HeapProcess) {
		// TODO
		//r, closer, err := helper.CreateTraceReader(c, tp.cfg.sourcePath, tp.cfg.endpointConnectionWait)
		//if err != nil {
		//	tp.err = fmt.Errorf("failed to create trace reader; %w", err)
		//	return
		//}
		//defer closer.Close()
		//
		//for {
		//	if c.Err() != nil {
		//		return
		//	}
		//
		//	event, err := r.ReadEvent()
		//	// TODO consider not break the process
		//	if err != nil {
		//		if errors.Is(err, io.EOF) {
		//			return
		//		}
		//
		//		tp.mx.Lock()
		//		tp.err = fmt.Errorf("failed to read event; %w", err)
		//		tp.mx.Unlock()
		//		return
		//	}
		//
		//	tp.processEvent(&event)
		//}
	}(ctx, hp)

	return nil
}
