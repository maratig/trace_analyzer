package error

import "errors"

var (
	ErrNilContext             = errors.New("context must not be nil")
	ErrNilApp                 = errors.New("application must not be nil")
	ErrEmptySourcePath        = errors.New("source path must not be empty")
	ErrInvalidSourcePath      = errors.New("invalid source path")
	ErrTraceAlreadyRunning    = errors.New("trace with given sourcePath is running already")
	ErrHeapProcAlreadyRunning = errors.New("heap profile processing with given sourcePath is running already")
	ErrConnectionFailed       = errors.New("failed to connect to the endpoint")
	ErrRetryable              = errors.New("retryable error")
)
