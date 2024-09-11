package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type BashRunner struct {
	path string
}

func (bashJob *BashRunner) Run(maxDuration time.Duration) int32 {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/bash", bashJob.path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdStartErr := cmd.Start()
	if cmdStartErr != nil {
		return -1
	}
	// Channel to signal that the command is done
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	// Context timeout reached, attempt to kill the process
	case <-ctx.Done():
		fmt.Println("Command timed out. Killing process...")
		if killErr := cmd.Process.Kill(); killErr != nil {
			fmt.Printf("Failed to kill process: %v\n", killErr)
			return 2
		} else {
			fmt.Println("Process killed successfully")
			return 3
		}
	// Command completed within the timeout
	case cmdErr := <-done:
		if cmdErr != nil {
			fmt.Printf("Error running script: %v\n", cmdErr)
			return 1
		}
		return 0
	}
}

func NewBashJob(path string) *BashRunner {
	return &BashRunner{path: path}
}
