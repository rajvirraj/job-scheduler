package runner

import "time"

type Runnable interface {
	Run(timeout time.Duration) int32
}
