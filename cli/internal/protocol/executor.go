package protocol

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
)

// Executor handles protocol execution
type Executor struct {
	options ExecutionOptions
	logger  *Logger
}

// Logger handles protocol execution logging
type Logger struct {
	file   *os.File
	stdout io.Writer
	stderr io.Writer
}

// NewExecutor creates a new protocol executor
func NewExecutor(options ExecutionOptions) (*Executor, error) {
	var logger *Logger
	var err error

	if options.LogFile != "" {
		logger, err = NewLogger(options.LogFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	return &Executor{
		options: options,
		logger:  logger,
	}, nil
}

// NewLogger creates a new logger
func NewLogger(logFile string) (*Logger, error) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:   f,
		stdout: io.MultiWriter(os.Stdout, f),
		stderr: io.MultiWriter(os.Stderr, f),
	}, nil
}

// Close closes the logger
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Execute runs a protocol
func (e *Executor) Execute(protocol *Protocol) (*ExecutionResult, error) {
	if protocol == nil {
		return nil, fmt.Errorf("protocol cannot be nil")
	}

	if err := protocol.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol: %w", err)
	}

	// Initialize result
	result := &ExecutionResult{
		Protocol: protocol,
		Success:  false,
		Errors:   []error{},
	}

	protocol.StartTime = time.Now()
	defer func() {
		protocol.EndTime = time.Now()
		result.CompletedAt = protocol.EndTime
		result.TotalDuration = protocol.EndTime.Sub(protocol.StartTime)
	}()

	// Show protocol header
	e.printHeader(protocol)

	// Dry run mode
	if e.options.DryRun {
		return e.executeDryRun(protocol, result)
	}

	// Execute phases
	for i, phase := range protocol.Phases {
		protocol.CurrentPhase = i

		// Check if phase should be skipped
		if e.shouldSkipPhase(phase) {
			phase.MarkSkipped()
			result.PhasesSkipped++
			e.printPhaseSkipped(i+1, len(protocol.Phases), phase)
			continue
		}

		// Check if only specific phases should run
		if len(e.options.OnlyPhases) > 0 && !e.shouldRunPhase(phase) {
			phase.MarkSkipped()
			result.PhasesSkipped++
			e.printPhaseSkipped(i+1, len(protocol.Phases), phase)
			continue
		}

		// Check dependencies
		if err := e.checkDependencies(phase, protocol); err != nil {
			phase.MarkFailed(err)
			result.PhasesFailed++
			result.Errors = append(result.Errors, err)
			e.printPhaseError(i+1, len(protocol.Phases), phase, err)

			if e.options.StopOnError {
				break
			}
			continue
		}

		// Execute phase
		result.PhasesRun++
		if err := e.executePhase(i+1, len(protocol.Phases), phase); err != nil {
			phase.MarkFailed(err)
			result.PhasesFailed++
			result.Errors = append(result.Errors, err)
			e.printPhaseError(i+1, len(protocol.Phases), phase, err)

			if e.options.StopOnError {
				break
			}
			continue
		}

		phase.MarkCompleted()
		result.PhasesPassed++
		e.printPhaseSuccess(i+1, len(protocol.Phases), phase)

		// Run verification if enabled
		if e.options.Verify && phase.Verification != nil {
			if err := e.verifyPhase(phase); err != nil {
				phase.MarkFailed(err)
				result.PhasesFailed++
				result.Errors = append(result.Errors, err)
				e.printVerificationError(phase, err)

				if e.options.StopOnError {
					break
				}
				continue
			}
			e.printVerificationSuccess(phase)
		}
	}

	// Determine overall success
	result.Success = result.PhasesFailed == 0 && len(result.Errors) == 0

	// Print summary
	e.printSummary(result)

	return result, nil
}

// executeDryRun performs a dry run of the protocol
func (e *Executor) executeDryRun(protocol *Protocol, result *ExecutionResult) (*ExecutionResult, error) {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("🔍 DRY RUN MODE - No commands will be executed"))
	fmt.Println()

	for i, phase := range protocol.Phases {
		e.printPhaseDryRun(i+1, len(protocol.Phases), phase)
	}

	result.Success = true
	result.CompletedAt = time.Now()
	result.TotalDuration = time.Since(protocol.StartTime)

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Dry run completed"))
	fmt.Println()

	return result, nil
}

// executePhase executes a single phase
func (e *Executor) executePhase(current, total int, phase *Phase) error {
	phase.MarkRunning()
	e.printPhaseStart(current, total, phase)

	// Execute each command in the phase
	for cmdIdx, cmd := range phase.Commands {
		e.printCommandStart(cmdIdx+1, len(phase.Commands), &cmd)

		output, err := e.executeCommand(&cmd)
		if output != "" {
			phase.AddOutput(strings.Split(output, "\n")...)
		}

		if err != nil && !cmd.IgnoreError {
			return fmt.Errorf("command failed: %w", err)
		}

		e.printCommandSuccess(cmdIdx+1, len(phase.Commands), &cmd)
	}

	return nil
}

// executeCommand executes a single command
func (e *Executor) executeCommand(cmd *Command) (string, error) {
	if cmd.Remote && e.options.Server != "" {
		return e.executeRemoteCommand(cmd)
	}
	return e.executeLocalCommand(cmd)
}

// executeLocalCommand executes a command locally
func (e *Executor) executeLocalCommand(cmd *Command) (string, error) {
	parts := cmd.BuildCommand()
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	execCmd := exec.Command(parts[0], parts[1:]...)

	// Set environment variables
	if len(cmd.Env) > 0 {
		execCmd.Env = append(os.Environ(), cmd.Env...)
	}

	// Set working directory
	if cmd.WorkDir != "" {
		execCmd.Dir = cmd.WorkDir
	}

	// Capture output - stream to console if enabled (default: true)
	var stdout, stderr bytes.Buffer
	streamOutput := e.options.StreamOutput || !e.options.DryRun // Default to streaming unless explicitly disabled

	if streamOutput {
		// Create a prefix writer that indents output
		stdoutWriter := &prefixWriter{w: os.Stdout, prefix: "       ", dim: true}
		stderrWriter := &prefixWriter{w: os.Stderr, prefix: "       ", dim: true}

		execCmd.Stdout = io.MultiWriter(&stdout, stdoutWriter)
		execCmd.Stderr = io.MultiWriter(&stderr, stderrWriter)
	} else {
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
	}

	// Run command
	err := execCmd.Run()

	// Combine output
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	return output, err
}

// executeRemoteCommand executes a command on a remote server
func (e *Executor) executeRemoteCommand(cmd *Command) (string, error) {
	if e.options.SSHHost == "" {
		return "", fmt.Errorf("remote execution requested but no SSH host configured")
	}

	// Build remote command
	remoteCmd := cmd.String()

	// Build SSH command
	sshArgs := []string{}

	if e.options.SSHKey != "" {
		sshArgs = append(sshArgs, "-i", e.options.SSHKey)
	}

	sshArgs = append(sshArgs,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-t", "-t", // Force pseudo-terminal for real-time output
	)

	target := e.options.SSHHost
	if e.options.SSHUser != "" {
		target = e.options.SSHUser + "@" + target
	}

	sshArgs = append(sshArgs, target, remoteCmd)

	execCmd := exec.Command("ssh", sshArgs...)

	// Capture output - stream to console if enabled (default: true)
	var stdout, stderr bytes.Buffer
	streamOutput := e.options.StreamOutput || !e.options.DryRun

	if streamOutput {
		stdoutWriter := &prefixWriter{w: os.Stdout, prefix: "       ", dim: true}
		stderrWriter := &prefixWriter{w: os.Stderr, prefix: "       ", dim: true}

		execCmd.Stdout = io.MultiWriter(&stdout, stdoutWriter)
		execCmd.Stderr = io.MultiWriter(&stderr, stderrWriter)
	} else {
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
	}

	// Run command
	err := execCmd.Run()

	// Combine output
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	return output, err
}

// verifyPhase runs verification for a phase
func (e *Executor) verifyPhase(phase *Phase) error {
	if phase.Verification == nil {
		return nil
	}

	v := phase.Verification

	// Check if any command in this phase was remote - if so, run verification remotely too
	isRemote := false
	for _, cmd := range phase.Commands {
		if cmd.Remote {
			isRemote = true
			break
		}
	}

	var stdout, stderr bytes.Buffer
	var err error

	if isRemote && e.options.SSHHost != "" {
		// Run verification on remote server
		verifyCmd := v.Command
		for _, arg := range v.Args {
			verifyCmd += " " + arg
		}

		sshArgs := []string{}
		if e.options.SSHKey != "" {
			sshArgs = append(sshArgs, "-i", e.options.SSHKey)
		}
		sshArgs = append(sshArgs,
			"-o", "BatchMode=yes",
			"-o", "StrictHostKeyChecking=no",
		)

		target := e.options.SSHHost
		if e.options.SSHUser != "" {
			target = e.options.SSHUser + "@" + target
		}
		sshArgs = append(sshArgs, target, verifyCmd)

		execCmd := exec.Command("ssh", sshArgs...)
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr

		// Set timeout
		if v.Timeout > 0 {
			timer := time.AfterFunc(v.Timeout, func() {
				if execCmd.Process != nil {
					execCmd.Process.Kill()
				}
			})
			defer timer.Stop()
		}

		err = execCmd.Run()
	} else {
		// Run verification locally
		execCmd := exec.Command(v.Command, v.Args...)
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr

		// Set timeout
		if v.Timeout > 0 {
			timer := time.AfterFunc(v.Timeout, func() {
				if execCmd.Process != nil {
					execCmd.Process.Kill()
				}
			})
			defer timer.Stop()
		}

		err = execCmd.Run()
	}

	// Check expected output
	if v.ExpectedOut != "" {
		if !strings.Contains(stdout.String(), v.ExpectedOut) {
			return fmt.Errorf("verification failed: expected output not found: %s", v.ExpectedOut)
		}
	}

	// Check expected error
	if v.ExpectedErr != "" {
		if !strings.Contains(stderr.String(), v.ExpectedErr) {
			return fmt.Errorf("verification failed: expected error not found: %s", v.ExpectedErr)
		}
	} else if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	return nil
}

// checkDependencies verifies all phase dependencies are met
func (e *Executor) checkDependencies(phase *Phase, protocol *Protocol) error {
	for _, depName := range phase.Dependencies {
		dep := protocol.GetPhaseByName(depName)
		if dep == nil {
			return fmt.Errorf("dependency not found: %s", depName)
		}
		if dep.Status != StatusCompleted {
			return fmt.Errorf("dependency not met: %s (status: %s)", depName, dep.Status)
		}
	}
	return nil
}

// shouldSkipPhase checks if a phase should be skipped
func (e *Executor) shouldSkipPhase(phase *Phase) bool {
	for _, skipName := range e.options.SkipPhases {
		if phase.Name == skipName {
			return true
		}
	}
	return false
}

// shouldRunPhase checks if a phase should run (when OnlyPhases is set)
func (e *Executor) shouldRunPhase(phase *Phase) bool {
	for _, runName := range e.options.OnlyPhases {
		if phase.Name == runName {
			return true
		}
	}
	return false
}

// Print functions for various execution stages

func (e *Executor) printHeader(protocol *Protocol) {
	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("🚀 %s 🚀", strings.ToUpper(protocol.Name))))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(protocol.Description))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Version: %s", protocol.Version)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Phases: %d", len(protocol.Phases))))
	if protocol.Requirements.GPUs > 0 {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  GPUs Required: %d x %dGB", protocol.Requirements.GPUs, protocol.Requirements.GPUMemoryGB)))
	}
	fmt.Println()
}

func (e *Executor) printPhaseStart(current, total int, phase *Phase) {
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("[%d/%d] %s %s", current, total, theme.SymbolLoading, phase.Name)))
	if phase.Description != "" {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  %s", phase.Description)))
	}
}

func (e *Executor) printPhaseSuccess(current, total int, phase *Phase) {
	duration := phase.GetDuration()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  %s Phase completed (%.1fs)", theme.SymbolSuccess, duration.Seconds())))
	fmt.Println()
}

func (e *Executor) printPhaseError(current, total int, phase *Phase, err error) {
	fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  %s Phase failed: %v", theme.SymbolError, err)))
	fmt.Println()
}

func (e *Executor) printPhaseSkipped(current, total int, phase *Phase) {
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("[%d/%d] ⊘ %s (skipped)", current, total, phase.Name)))
}

func (e *Executor) printPhaseDryRun(current, total int, phase *Phase) {
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("[%d/%d] %s", current, total, phase.Name)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  %s", phase.Description)))

	for i, cmd := range phase.Commands {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("    %d. %s", i+1, cmd.Description)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("       $ %s", cmd.String())))
	}

	if phase.Verification != nil {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("    ✓ Verification: %s", phase.Verification.Description)))
	}
	fmt.Println()
}

func (e *Executor) printCommandStart(current, total int, cmd *Command) {
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("    [%d/%d] %s", current, total, cmd.Description)))
}

func (e *Executor) printCommandSuccess(current, total int, cmd *Command) {
	// Silent success for commands
}

func (e *Executor) printVerificationSuccess(phase *Phase) {
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  %s Verification passed", theme.SymbolSuccess)))
}

func (e *Executor) printVerificationError(phase *Phase, err error) {
	fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  %s Verification failed: %v", theme.SymbolError, err)))
}

func (e *Executor) printSummary(result *ExecutionResult) {
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))

	if result.Success {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  %s Protocol completed successfully!", theme.SymbolSuccess)))
	} else {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  %s Protocol completed with errors", theme.SymbolError)))
	}

	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()

	// Stats
	fmt.Println(theme.InfoStyle.Render("📊 Summary:"))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Total duration: %.1fs", result.TotalDuration.Seconds())))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Phases run: %d", result.PhasesRun)))
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  Phases passed: %d", result.PhasesPassed)))

	if result.PhasesFailed > 0 {
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("  Phases failed: %d", result.PhasesFailed)))
	}

	if result.PhasesSkipped > 0 {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Phases skipped: %d", result.PhasesSkipped)))
	}

	fmt.Println()

	// Errors
	if len(result.Errors) > 0 {
		fmt.Println(theme.ErrorStyle.Render("❌ Errors:"))
		for i, err := range result.Errors {
			fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  %d. %v", i+1, err)))
		}
		fmt.Println()
	}
}

// Close closes the executor and cleans up resources
func (e *Executor) Close() error {
	if e.logger != nil {
		return e.logger.Close()
	}
	return nil
}

// prefixWriter wraps a writer and adds a prefix to each line
type prefixWriter struct {
	w         io.Writer
	prefix    string
	dim       bool
	atNewLine bool
}

// Write implements io.Writer
func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	// Process byte by byte to handle line prefixes
	for i, b := range p {
		if pw.atNewLine || i == 0 {
			// Write prefix at start of each line
			if pw.dim {
				fmt.Fprint(pw.w, theme.DimTextStyle.Render(pw.prefix))
			} else {
				pw.w.Write([]byte(pw.prefix))
			}
			pw.atNewLine = false
		}

		if pw.dim {
			// For dim output, we write character by character
			fmt.Fprint(pw.w, theme.DimTextStyle.Render(string(b)))
		} else {
			pw.w.Write([]byte{b})
		}

		if b == '\n' {
			pw.atNewLine = true
		}
	}
	return len(p), nil
}
