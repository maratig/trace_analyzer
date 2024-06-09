package error

import "errors"

var (
	ErrNilContext          = errors.New("context must not be nil")
	ErrNilApp              = errors.New("application must not be nil")
	ErrEmptySourcePath     = errors.New("source path must not be empty")
	ErrTraceAlreadyRunning = errors.New("trace with given sourcePath is running already")
)
