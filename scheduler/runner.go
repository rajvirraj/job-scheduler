package scheduler

import "time"

type Runnable interface {
	Run(timeout time.Duration) int32
}
