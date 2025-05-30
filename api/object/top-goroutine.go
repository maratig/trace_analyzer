package object

import (
	"time"

	"golang.org/x/exp/trace"
)

type TopGoroutine struct {
	ID           trace.GoID    `json:"id"`
	ParentStack  string        `json:"parent-stack"`
	Stack        string        `json:"stack"`
	ExecDuration time.Duration `json:"execution-duration"`
	IdleDuration time.Duration `json:"idle-duration"`
}
