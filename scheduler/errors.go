package scheduler

import (
	"errors"
	"fmt"
)

const (
	ErrJobAlreadyAddedMsg        = "job with given id is already added"
	ErrJobNotFoundMsg            = "job with given id not found"
	ErrJobIdPreviouslyDeletedMsg = "job with given id was previously deleted. please use new Id"
)

var (
	ErrJobAlreadyAdded        error
	ErrJobNotFound            error
	ErrJobIdPreviouslyDeleted error
)

func init() {
	ErrJobAlreadyAdded = errors.New(ErrJobAlreadyAddedMsg)
	ErrJobNotFound = errors.New(ErrJobNotFoundMsg)
	ErrJobIdPreviouslyDeleted = errors.New(ErrJobIdPreviouslyDeletedMsg)
}

func ErrSchedulerNotRunning(currentState schedulerState) error {
	return errors.New(fmt.Sprintf("scheduler not running. current state is %v", currentState))
}
