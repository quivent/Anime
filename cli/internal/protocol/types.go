package protocol

import (
	"fmt"
	"strings"
	"time"
)

// PhaseStatus represents the current status of a phase
type PhaseStatus string

const (
	StatusPending   PhaseStatus = "pending"
	StatusRunning   PhaseStatus = "running"
	StatusCompleted PhaseStatus = "completed"
	StatusFailed    PhaseStatus = "failed"
	StatusSkipped   PhaseStatus = "skipped"
)

// Phase represents a single step in a protocol
type Phase struct {
	Name         string
	Description  string
	Commands     []Command
	Verification *Verification
	Dependencies []string // Names of phases that must complete first
	Status       PhaseStatus
	StartTime    time.Time
	EndTime      time.Time
	Error        error
	Output       []string
}

// Command represents a single command to execute
type Command struct {
	Description string   // Human-readable description
	Command     string   // The command to run (e.g., "apt-get", "python3")
	Args        []string // Command arguments
	Env         []string // Environment variables (KEY=VALUE format)
	WorkDir     string   // Working directory for the command
	Sudo        bool     // Whether to run with sudo
	Remote      bool     // Whether to run on remote server
	IgnoreError bool     // Continue even if this command fails
}

// Verification defines how to verify a phase completed successfully
type Verification struct {
	Description string
	Command     string
	Args        []string
	ExpectedOut string // Expected output substring
	ExpectedErr string // Expected error substring (empty = no error expected)
	Timeout     time.Duration
}

// Protocol represents a complete setup protocol
type Protocol struct {
	Name         string
	Description  string
	Category     string // e.g., "LLM", "Video", "Development"
	Version      string
	Requirements Requirements
	Phases       []*Phase

	// Runtime tracking
	CurrentPhase int
	StartTime    time.Time
	EndTime      time.Time
}

// Requirements defines what's needed to run the protocol
type Requirements struct {
	GPUs         int      // Number of GPUs required
	GPUMemoryGB  int      // GPU memory per GPU in GB
	SystemMemGB  int      // System RAM in GB
	DiskSpaceGB  int      // Free disk space in GB
	CUDA         string   // Required CUDA version (e.g., "12.4+")
	Python       string   // Required Python version (e.g., "3.11+")
	OS           []string // Supported operating systems
	Arch         []string // Supported architectures (e.g., "arm64", "amd64")
}

// ExecutionOptions configures how the protocol is executed
type ExecutionOptions struct {
	DryRun       bool     // Preview only, don't execute
	Verify       bool     // Run verification after each phase
	AutoContinue bool     // Continue automatically on success
	Server       string   // Server name/alias for remote execution
	SSHKey       string   // SSH key for remote execution
	SSHUser      string   // SSH user for remote execution
	SSHHost      string   // SSH host for remote execution
	LogFile      string   // File to log output
	StopOnError  bool     // Stop execution on first error (default: true)
	SkipPhases   []string // Phase names to skip
	OnlyPhases   []string // Only run these phases
	StreamOutput bool     // Stream command output in real-time (default: true)
}

// ExecutionResult contains the results of protocol execution
type ExecutionResult struct {
	Protocol      *Protocol
	Success       bool
	CompletedAt   time.Time
	TotalDuration time.Duration
	PhasesRun     int
	PhasesPassed  int
	PhasesFailed  int
	PhasesSkipped int
	Errors        []error
}

// GetDuration returns the duration of a phase
func (p *Phase) GetDuration() time.Duration {
	if p.EndTime.IsZero() || p.StartTime.IsZero() {
		return 0
	}
	return p.EndTime.Sub(p.StartTime)
}

// IsComplete returns true if the phase has completed (success or failure)
func (p *Phase) IsComplete() bool {
	return p.Status == StatusCompleted || p.Status == StatusFailed || p.Status == StatusSkipped
}

// MarkRunning marks the phase as running
func (p *Phase) MarkRunning() {
	p.Status = StatusRunning
	p.StartTime = time.Now()
}

// MarkCompleted marks the phase as completed
func (p *Phase) MarkCompleted() {
	p.Status = StatusCompleted
	p.EndTime = time.Now()
}

// MarkFailed marks the phase as failed with an error
func (p *Phase) MarkFailed(err error) {
	p.Status = StatusFailed
	p.Error = err
	p.EndTime = time.Now()
}

// MarkSkipped marks the phase as skipped
func (p *Phase) MarkSkipped() {
	p.Status = StatusSkipped
	p.EndTime = time.Now()
}

// AddOutput adds output lines to the phase
func (p *Phase) AddOutput(lines ...string) {
	p.Output = append(p.Output, lines...)
}

// GetProgress returns current progress as (current, total)
func (pr *Protocol) GetProgress() (int, int) {
	completed := 0
	for _, phase := range pr.Phases {
		if phase.IsComplete() {
			completed++
		}
	}
	return completed, len(pr.Phases)
}

// GetCurrentPhase returns the currently executing or next phase
func (pr *Protocol) GetCurrentPhase() *Phase {
	if pr.CurrentPhase >= 0 && pr.CurrentPhase < len(pr.Phases) {
		return pr.Phases[pr.CurrentPhase]
	}
	return nil
}

// GetPhaseByName finds a phase by name
func (pr *Protocol) GetPhaseByName(name string) *Phase {
	for _, phase := range pr.Phases {
		if phase.Name == name {
			return phase
		}
	}
	return nil
}

// Validate checks if the protocol is valid
func (pr *Protocol) Validate() error {
	if pr.Name == "" {
		return fmt.Errorf("protocol name cannot be empty")
	}
	if len(pr.Phases) == 0 {
		return fmt.Errorf("protocol must have at least one phase")
	}

	// Check for duplicate phase names
	phaseNames := make(map[string]bool)
	for _, phase := range pr.Phases {
		if phase.Name == "" {
			return fmt.Errorf("phase name cannot be empty")
		}
		if phaseNames[phase.Name] {
			return fmt.Errorf("duplicate phase name: %s", phase.Name)
		}
		phaseNames[phase.Name] = true

		// Validate dependencies
		for _, dep := range phase.Dependencies {
			if !phaseNames[dep] {
				return fmt.Errorf("phase %s has invalid dependency: %s", phase.Name, dep)
			}
		}
	}

	return nil
}

// Summary returns a summary of the protocol
func (pr *Protocol) Summary() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Protocol: %s v%s\n", pr.Name, pr.Version))
	b.WriteString(fmt.Sprintf("Description: %s\n", pr.Description))
	b.WriteString(fmt.Sprintf("Phases: %d\n", len(pr.Phases)))

	if pr.Requirements.GPUs > 0 {
		b.WriteString(fmt.Sprintf("Requirements: %d x GPU (%dGB each)\n",
			pr.Requirements.GPUs, pr.Requirements.GPUMemoryGB))
	}

	return b.String()
}

// BuildCommand constructs the full command string for execution
func (c *Command) BuildCommand() []string {
	cmd := []string{}

	if c.Sudo {
		cmd = append(cmd, "sudo")
	}

	cmd = append(cmd, c.Command)
	cmd = append(cmd, c.Args...)

	return cmd
}

// String returns a string representation of the command
func (c *Command) String() string {
	parts := c.BuildCommand()
	return strings.Join(parts, " ")
}
