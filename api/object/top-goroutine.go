package object

import (
	"time"

	"golang.org/x/exp/trace"
)

type TopGoroutine struct {
	ID           trace.GoID    `json:"id"`
	Stack        string        `json:"stack"`
	ExecDuration time.Duration `json:"execution-duration"`
	LiveDuration time.Duration `json:"live-duration"`
}
