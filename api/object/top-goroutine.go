package object

import (
	"time"

	"golang.org/x/exp/trace"
)

type TopGoroutine struct {
	ID              trace.GoID    `json:"id"`
	Stack           string        `json:"stack"`
	TransitionStack string        `json:"transition-stack"`
	ExecDuration    time.Duration `json:"execution-duration"`
	IdleDuration    time.Duration `json:"idle-duration"`
	InvokedBy       *TopGoroutine `json:"invoked-by,omitempty"`
}
