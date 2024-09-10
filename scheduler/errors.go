package scheduler

import (
	"errors"
	"fmt"
)

var ErrJobQueueEmpty = errors.New("job queue is empty")
var ErrJobAlreadyAdded = errors.New("job with given id is already added")
var ErrJobNotFound = errors.New("job with given id not found")
var ErrJobIdPreviouslyDeleted = errors.New("job with given id was previously deleted. please use new Id")

func ErrSchedulerNotRunning(currentState schedulerState) error {
	return errors.New(fmt.Sprintf("scheduler not running. current state is %v", currentState))
}
