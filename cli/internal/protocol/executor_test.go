package protocol

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	t.Run("executor without logging", func(t *testing.T) {
		executor, err := NewExecutor(ExecutionOptions{})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		if executor == nil {
			t.Fatal("NewExecutor() returned nil executor")
		}

		if executor.logger != nil {
			t.Error("executor should not have logger when no log file specified")
		}
	})

	t.Run("executor with logging", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		executor, err := NewExecutor(ExecutionOptions{
			LogFile: logFile,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		if executor.logger == nil {
			t.Error("executor should have logger when log file specified")
		}

		defer executor.Close()

		// Verify log file was created
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("log file was not created")
		}
	})

	t.Run("executor with invalid log file path", func(t *testing.T) {
		_, err := NewExecutor(ExecutionOptions{
			LogFile: "/nonexistent/dir/test.log",
		})

		if err == nil {
			t.Error("NewExecutor() should return error for invalid log file path")
		}
	})
}

func TestExecuteDryRun(t *testing.T) {
	protocol := &Protocol{
		Name:        "Test Protocol",
		Description: "Test description",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name:        "Phase 1",
				Description: "Test phase 1",
				Commands: []Command{
					{
						Description: "Test command",
						Command:     "echo",
						Args:        []string{"test"},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("dry run should always succeed")
	}

	if result.PhasesRun != 0 {
		t.Errorf("dry run should not execute phases, got %d phases run", result.PhasesRun)
	}

	// All phases should remain pending in dry run
	for _, phase := range protocol.Phases {
		if phase.Status != StatusPending && phase.Status != "" {
			t.Errorf("phase %s status = %s, want pending or empty in dry run", phase.Name, phase.Status)
		}
	}
}

func TestExecuteProtocol(t *testing.T) {
	t.Run("simple protocol execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		protocol := &Protocol{
			Name:        "Test Protocol",
			Description: "Test description",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Create file",
					Description: "Create test file",
					Commands: []Command{
						{
							Description: "Create file",
							Command:     "touch",
							Args:        []string{testFile},
						},
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution should succeed")
		}

		if result.PhasesPassed != 1 {
			t.Errorf("PhasesPassed = %d, want 1", result.PhasesPassed)
		}

		if result.PhasesFailed != 0 {
			t.Errorf("PhasesFailed = %d, want 0", result.PhasesFailed)
		}

		// Verify file was created
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("test file was not created")
		}

		// Verify phase status
		if protocol.Phases[0].Status != StatusCompleted {
			t.Errorf("phase status = %s, want completed", protocol.Phases[0].Status)
		}
	})

	t.Run("protocol with multiple phases", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		protocol := &Protocol{
			Name:        "Multi-Phase Protocol",
			Description: "Test multiple phases",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Phase 1",
					Description: "Create first file",
					Commands: []Command{
						{
							Description: "Create file 1",
							Command:     "touch",
							Args:        []string{file1},
						},
					},
				},
				{
					Name:        "Phase 2",
					Description: "Create second file",
					Commands: []Command{
						{
							Description: "Create file 2",
							Command:     "touch",
							Args:        []string{file2},
						},
					},
					Dependencies: []string{"Phase 1"},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution should succeed")
		}

		if result.PhasesPassed != 2 {
			t.Errorf("PhasesPassed = %d, want 2", result.PhasesPassed)
		}

		// Verify both files were created
		if _, err := os.Stat(file1); os.IsNotExist(err) {
			t.Error("file1 was not created")
		}

		if _, err := os.Stat(file2); os.IsNotExist(err) {
			t.Error("file2 was not created")
		}
	})

	t.Run("protocol with failing command", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Failing Protocol",
			Description: "Test error handling",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Failing phase",
					Description: "This should fail",
					Commands: []Command{
						{
							Description: "Run non-existent command",
							Command:     "nonexistent-command-xyz",
							Args:        []string{},
						},
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.Success {
			t.Error("execution should fail")
		}

		if result.PhasesFailed != 1 {
			t.Errorf("PhasesFailed = %d, want 1", result.PhasesFailed)
		}

		if len(result.Errors) == 0 {
			t.Error("result should contain errors")
		}

		// Verify phase status
		if protocol.Phases[0].Status != StatusFailed {
			t.Errorf("phase status = %s, want failed", protocol.Phases[0].Status)
		}
	})

	t.Run("protocol with ignored error", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		protocol := &Protocol{
			Name:        "Ignore Error Protocol",
			Description: "Test error ignoring",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Phase with ignored error",
					Description: "Should continue despite error",
					Commands: []Command{
						{
							Description: "Failing command",
							Command:     "false",
							IgnoreError: true,
						},
						{
							Description: "Create file",
							Command:     "touch",
							Args:        []string{testFile},
						},
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution should succeed with ignored error")
		}

		// Verify file was created despite earlier failure
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("test file should be created despite ignored error")
		}
	})
}

func TestExecuteWithVerification(t *testing.T) {
	t.Run("successful verification", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		protocol := &Protocol{
			Name:        "Verified Protocol",
			Description: "Test verification",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Create and verify",
					Description: "Create file and verify",
					Commands: []Command{
						{
							Description: "Create file",
							Command:     "touch",
							Args:        []string{testFile},
						},
					},
					Verification: &Verification{
						Description: "Check file exists",
						Command:     "test",
						Args:        []string{"-f", testFile},
						Timeout:     5 * time.Second,
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			Verify:      true,
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution with verification should succeed")
		}
	})

	t.Run("failed verification", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Failed Verification",
			Description: "Test failed verification",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Phase with bad verification",
					Description: "Verification should fail",
					Commands: []Command{
						{
							Description: "Echo test",
							Command:     "echo",
							Args:        []string{"test"},
						},
					},
					Verification: &Verification{
						Description: "Check non-existent file",
						Command:     "test",
						Args:        []string{"-f", "/nonexistent/file.txt"},
						Timeout:     5 * time.Second,
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			Verify:      true,
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.Success {
			t.Error("execution should fail on verification failure")
		}

		if result.PhasesFailed == 0 {
			t.Error("should have failed phases")
		}
	})

	t.Run("verification with expected output", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Output Verification",
			Description: "Test output verification",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name:        "Echo and verify",
					Description: "Echo and verify output",
					Commands: []Command{
						{
							Description: "Echo message",
							Command:     "echo",
							Args:        []string{"hello world"},
						},
					},
					Verification: &Verification{
						Description: "Verify echo",
						Command:     "echo",
						Args:        []string{"verification success"},
						ExpectedOut: "verification success",
						Timeout:     5 * time.Second,
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			Verify:      true,
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution should succeed with matching output")
		}
	})
}

func TestSkipPhases(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	protocol := &Protocol{
		Name:        "Skip Test Protocol",
		Description: "Test phase skipping",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "Phase 1",
				Commands: []Command{
					{
						Description: "Create file 1",
						Command:     "touch",
						Args:        []string{file1},
					},
				},
			},
			{
				Name: "Phase 2",
				Commands: []Command{
					{
						Description: "Create file 2",
						Command:     "touch",
						Args:        []string{file2},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{
		SkipPhases:  []string{"Phase 2"},
		StopOnError: true,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("execution should succeed")
	}

	if result.PhasesSkipped != 1 {
		t.Errorf("PhasesSkipped = %d, want 1", result.PhasesSkipped)
	}

	// File 1 should exist, file 2 should not
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("file1 should be created")
	}

	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should not be created (phase skipped)")
	}

	// Verify phase statuses
	if protocol.Phases[0].Status != StatusCompleted {
		t.Errorf("Phase 1 status = %s, want completed", protocol.Phases[0].Status)
	}

	if protocol.Phases[1].Status != StatusSkipped {
		t.Errorf("Phase 2 status = %s, want skipped", protocol.Phases[1].Status)
	}
}

func TestOnlyPhases(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	protocol := &Protocol{
		Name:        "Only Phases Test",
		Description: "Test only-phases execution",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "Phase 1",
				Commands: []Command{
					{
						Description: "Create file 1",
						Command:     "touch",
						Args:        []string{file1},
					},
				},
			},
			{
				Name: "Phase 2",
				Commands: []Command{
					{
						Description: "Create file 2",
						Command:     "touch",
						Args:        []string{file2},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{
		OnlyPhases:  []string{"Phase 2"},
		StopOnError: true,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("execution should succeed")
	}

	if result.PhasesRun != 1 {
		t.Errorf("PhasesRun = %d, want 1", result.PhasesRun)
	}

	// File 1 should not exist, file 2 should exist
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should not be created (phase not in only list)")
	}

	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("file2 should be created")
	}
}

func TestDependencyHandling(t *testing.T) {
	t.Run("successful dependency resolution", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		protocol := &Protocol{
			Name:        "Dependency Test",
			Description: "Test dependencies",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name: "Base",
					Commands: []Command{
						{
							Description: "Create base file",
							Command:     "touch",
							Args:        []string{file1},
						},
					},
				},
				{
					Name: "Dependent",
					Commands: []Command{
						{
							Description: "Create dependent file",
							Command:     "touch",
							Args:        []string{file2},
						},
					},
					Dependencies: []string{"Base"},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !result.Success {
			t.Error("execution should succeed")
		}

		if result.PhasesPassed != 2 {
			t.Errorf("PhasesPassed = %d, want 2", result.PhasesPassed)
		}
	})

	t.Run("failed dependency blocks phase", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Failed Dependency",
			Description: "Test failed dependency",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name: "Base",
					Commands: []Command{
						{
							Description: "Failing command",
							Command:     "false",
						},
					},
				},
				{
					Name:         "Dependent",
					Commands:     []Command{},
					Dependencies: []string{"Base"},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{
			StopOnError: true,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		result, err := executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if result.Success {
			t.Error("execution should fail")
		}

		// Base phase should fail, dependent phase should not run
		if protocol.Phases[0].Status != StatusFailed {
			t.Error("base phase should be failed")
		}
	})
}

func TestProtocolValidation(t *testing.T) {
	t.Run("valid protocol", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Valid Protocol",
			Description: "Test",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name: "Phase 1",
					Commands: []Command{
						{Command: "echo", Args: []string{"test"}},
					},
				},
			},
		}

		executor, err := NewExecutor(ExecutionOptions{})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		_, err = executor.Execute(protocol)
		if err != nil {
			t.Fatalf("Execute() error = %v for valid protocol", err)
		}
	})

	t.Run("nil protocol", func(t *testing.T) {
		executor, err := NewExecutor(ExecutionOptions{})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		_, err = executor.Execute(nil)
		if err == nil {
			t.Error("Execute() should return error for nil protocol")
		}
	})

	t.Run("protocol with no phases", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "No Phases",
			Description: "Test",
			Version:     "1.0.0",
			Phases:      []*Phase{},
		}

		executor, err := NewExecutor(ExecutionOptions{})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		_, err = executor.Execute(protocol)
		if err == nil {
			t.Error("Execute() should return error for protocol with no phases")
		}
	})

	t.Run("protocol with duplicate phase names", func(t *testing.T) {
		protocol := &Protocol{
			Name:        "Duplicate Names",
			Description: "Test",
			Version:     "1.0.0",
			Phases: []*Phase{
				{
					Name: "Phase 1",
					Commands: []Command{
						{Command: "echo", Args: []string{"test"}},
					},
				},
				{
					Name: "Phase 1",
					Commands: []Command{
						{Command: "echo", Args: []string{"test2"}},
					},
				},
			},
		}

		err := protocol.Validate()
		if err == nil {
			t.Error("Validate() should return error for duplicate phase names")
		}

		if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("error should mention duplicate: %v", err)
		}
	})
}

func TestCommandBuilding(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		want    string
		wantLen int
	}{
		{
			name: "simple command",
			cmd: Command{
				Command: "echo",
				Args:    []string{"hello"},
			},
			want:    "echo hello",
			wantLen: 2,
		},
		{
			name: "command with sudo",
			cmd: Command{
				Command: "apt-get",
				Args:    []string{"install", "vim"},
				Sudo:    true,
			},
			want:    "sudo apt-get install vim",
			wantLen: 4,
		},
		{
			name: "command with multiple args",
			cmd: Command{
				Command: "git",
				Args:    []string{"commit", "-m", "test message"},
			},
			want:    "git commit -m test message",
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cmd.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}

			parts := tt.cmd.BuildCommand()
			if len(parts) != tt.wantLen {
				t.Errorf("BuildCommand() length = %d, want %d", len(parts), tt.wantLen)
			}
		})
	}
}

func TestExecutorClose(t *testing.T) {
	t.Run("close without logger", func(t *testing.T) {
		executor, err := NewExecutor(ExecutionOptions{})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		err = executor.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("close with logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		executor, err := NewExecutor(ExecutionOptions{
			LogFile: logFile,
		})
		if err != nil {
			t.Fatalf("NewExecutor() error = %v", err)
		}

		err = executor.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		// Log file should still exist after close
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("log file should exist after close")
		}
	})
}

func TestExecutionResult(t *testing.T) {
	protocol := &Protocol{
		Name:        "Result Test",
		Description: "Test execution result",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "Success Phase",
				Commands: []Command{
					{Command: "echo", Args: []string{"success"}},
				},
			},
			{
				Name: "Fail Phase",
				Commands: []Command{
					{Command: "false"},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{
		StopOnError: false, // Continue on error
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify result fields
	if result.Protocol != protocol {
		t.Error("result should reference the protocol")
	}

	if result.CompletedAt.IsZero() {
		t.Error("result should have completion time")
	}

	if result.TotalDuration == 0 {
		t.Error("result should have non-zero duration")
	}

	if result.PhasesRun != 2 {
		t.Errorf("PhasesRun = %d, want 2", result.PhasesRun)
	}

	if result.Success {
		t.Error("result should not be successful with failed phase")
	}

	if len(result.Errors) == 0 {
		t.Error("result should contain errors")
	}
}

func TestPhaseOutput(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")
	testContent := "test output content"

	// Write test content to file
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	protocol := &Protocol{
		Name:        "Output Test",
		Description: "Test phase output capture",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "Output Phase",
				Commands: []Command{
					{
						Description: "Cat file",
						Command:     "cat",
						Args:        []string{testFile},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{
		StreamOutput: false, // Don't stream to allow output capture
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("execution should succeed")
	}

	// Verify phase has output
	phase := protocol.Phases[0]
	if len(phase.Output) == 0 {
		t.Error("phase should have output")
	}

	// Output should contain the test content
	outputStr := strings.Join(phase.Output, "")
	if !strings.Contains(outputStr, testContent) {
		t.Errorf("phase output should contain %q, got %q", testContent, outputStr)
	}
}

func TestCommandEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "env_test.txt")

	protocol := &Protocol{
		Name:        "Environment Test",
		Description: "Test environment variables",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "Env Phase",
				Commands: []Command{
					{
						Description: "Echo env var to file",
						Command:     "sh",
						Args:        []string{"-c", fmt.Sprintf("echo $TEST_VAR > %s", testFile)},
						Env:         []string{"TEST_VAR=test_value"},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("execution should succeed")
	}

	// Verify environment variable was set
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if !strings.Contains(string(content), "test_value") {
		t.Errorf("file should contain env var value, got: %s", string(content))
	}
}

func TestWorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	protocol := &Protocol{
		Name:        "WorkDir Test",
		Description: "Test working directory",
		Version:     "1.0.0",
		Phases: []*Phase{
			{
				Name: "WorkDir Phase",
				Commands: []Command{
					{
						Description: "Create file in workdir",
						Command:     "touch",
						Args:        []string{"test.txt"},
						WorkDir:     tmpDir,
					},
				},
			},
		},
	}

	executor, err := NewExecutor(ExecutionOptions{})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	result, err := executor.Execute(protocol)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("execution should succeed")
	}

	// Verify file was created in working directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("file should be created in working directory")
	}
}
