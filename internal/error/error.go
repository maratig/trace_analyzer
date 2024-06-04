package error

import "errors"

var (
	ErrNilContext          = errors.New("context must not be nil")
	ErrNilApp              = errors.New("application must not be nil")
	ErrEmptyEventsEndpoint = errors.New("events endpoint must not be empty")
)
