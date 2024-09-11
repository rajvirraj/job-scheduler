package jobqueue

import (
	"errors"
)

const (
	ErrJobQueueEmptyMsg = "job queue is empty"
)

var (
	ErrJobQueueEmpty error
)

func init() {
	ErrJobQueueEmpty = errors.New(ErrJobQueueEmptyMsg)
}
