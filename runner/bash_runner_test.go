package runner

import (
	"os"
	"testing"
	"time"
)

func TestBashRunner_Run_Success(t *testing.T) {
	// Create a temporary bash script
	scriptContent := "#!/bin/bash\necho Hello, World!"
	scriptPath := "/tmp/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	bashRunner := NewBashJob(scriptPath)
	exitCode := bashRunner.Run(5 * time.Second)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestBashRunner_Run_Timeout(t *testing.T) {
	// Create a temporary bash script that sleeps for 10 seconds
	scriptContent := "#!/bin/bash\nsleep 10"
	scriptPath := "/tmp/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	bashRunner := NewBashJob(scriptPath)
	exitCode := bashRunner.Run(2 * time.Second)

	if exitCode != 2 {
		t.Errorf("expected code 3 for timeout, got %d", exitCode)
	}
}

func TestBashRunner_Run_Failure(t *testing.T) {
	// Create a temporary bash script that exits with an error
	scriptContent := "#!/bin/bash\nexit 1"
	scriptPath := "/tmp/test_script.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	bashRunner := NewBashJob(scriptPath)
	exitCode := bashRunner.Run(5 * time.Second)

	if exitCode != 1 {
		t.Errorf("expected code 1 for script failure, got %d", exitCode)
	}
}
