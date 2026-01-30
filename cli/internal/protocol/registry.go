package protocol

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages available protocols
type Registry struct {
	protocols map[string]*Protocol
	mu        sync.RWMutex
}

// NewRegistry creates a new protocol registry
func NewRegistry() *Registry {
	return &Registry{
		protocols: make(map[string]*Protocol),
	}
}

// Register adds a protocol to the registry
func (r *Registry) Register(protocol *Protocol) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if protocol == nil {
		return fmt.Errorf("cannot register nil protocol")
	}

	if err := protocol.Validate(); err != nil {
		return fmt.Errorf("invalid protocol: %w", err)
	}

	if _, exists := r.protocols[protocol.Name]; exists {
		return fmt.Errorf("protocol %s already registered", protocol.Name)
	}

	r.protocols[protocol.Name] = protocol
	return nil
}

// Get retrieves a protocol by name
func (r *Registry) Get(name string) (*Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	protocol, exists := r.protocols[name]
	if !exists {
		return nil, fmt.Errorf("protocol %s not found", name)
	}

	// Return a deep copy to prevent modifications
	return r.copyProtocol(protocol), nil
}

// List returns all registered protocols
func (r *Registry) List() []*Protocol {
	r.mu.RLock()
	defer r.mu.RUnlock()

	protocols := make([]*Protocol, 0, len(r.protocols))
	for _, p := range r.protocols {
		protocols = append(protocols, r.copyProtocol(p))
	}

	// Sort by name for consistent output
	sort.Slice(protocols, func(i, j int) bool {
		return protocols[i].Name < protocols[j].Name
	})

	return protocols
}

// ListByCategory returns protocols in a specific category
func (r *Registry) ListByCategory(category string) []*Protocol {
	r.mu.RLock()
	defer r.mu.RUnlock()

	protocols := make([]*Protocol, 0)
	for _, p := range r.protocols {
		if p.Category == category {
			protocols = append(protocols, r.copyProtocol(p))
		}
	}

	// Sort by name for consistent output
	sort.Slice(protocols, func(i, j int) bool {
		return protocols[i].Name < protocols[j].Name
	})

	return protocols
}

// Exists checks if a protocol is registered
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.protocols[name]
	return exists
}

// Count returns the number of registered protocols
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.protocols)
}

// Categories returns all unique categories
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoryMap := make(map[string]bool)
	for _, p := range r.protocols {
		if p.Category != "" {
			categoryMap[p.Category] = true
		}
	}

	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}

	sort.Strings(categories)
	return categories
}

// copyProtocol creates a deep copy of a protocol
func (r *Registry) copyProtocol(p *Protocol) *Protocol {
	if p == nil {
		return nil
	}

	copy := &Protocol{
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Version:     p.Version,
		Requirements: Requirements{
			GPUs:        p.Requirements.GPUs,
			GPUMemoryGB: p.Requirements.GPUMemoryGB,
			SystemMemGB: p.Requirements.SystemMemGB,
			DiskSpaceGB: p.Requirements.DiskSpaceGB,
			CUDA:        p.Requirements.CUDA,
			Python:      p.Requirements.Python,
			OS:          append([]string{}, p.Requirements.OS...),
			Arch:        append([]string{}, p.Requirements.Arch...),
		},
		Phases: make([]*Phase, len(p.Phases)),
	}

	// Deep copy phases
	for i, phase := range p.Phases {
		copy.Phases[i] = r.copyPhase(phase)
	}

	return copy
}

// copyPhase creates a deep copy of a phase
func (r *Registry) copyPhase(p *Phase) *Phase {
	if p == nil {
		return nil
	}

	copy := &Phase{
		Name:         p.Name,
		Description:  p.Description,
		Dependencies: append([]string{}, p.Dependencies...),
		Status:       StatusPending, // Always start fresh
		Commands:     make([]Command, len(p.Commands)),
		Output:       []string{},
	}

	// Copy commands
	for i, cmd := range p.Commands {
		copy.Commands[i] = Command{
			Description: cmd.Description,
			Command:     cmd.Command,
			Args:        append([]string{}, cmd.Args...),
			Env:         append([]string{}, cmd.Env...),
			WorkDir:     cmd.WorkDir,
			Sudo:        cmd.Sudo,
			Remote:      cmd.Remote,
			IgnoreError: cmd.IgnoreError,
		}
	}

	// Copy verification if present
	if p.Verification != nil {
		copy.Verification = &Verification{
			Description: p.Verification.Description,
			Command:     p.Verification.Command,
			Args:        append([]string{}, p.Verification.Args...),
			ExpectedOut: p.Verification.ExpectedOut,
			ExpectedErr: p.Verification.ExpectedErr,
			Timeout:     p.Verification.Timeout,
		}
	}

	return copy
}

// Global registry instance
var globalRegistry *Registry
var once sync.Once

// GetGlobalRegistry returns the global protocol registry
func GetGlobalRegistry() *Registry {
	once.Do(func() {
		globalRegistry = NewRegistry()
		// Register built-in protocols
		registerBuiltinProtocols()
	})
	return globalRegistry
}

// registerBuiltinProtocols registers all built-in protocols
func registerBuiltinProtocols() {
	// Coverage protocol will be registered from coverage.go
	// This allows for modular protocol definitions
	if err := globalRegistry.Register(NewCoverageProtocol()); err != nil {
		// Log error but don't panic - allows graceful degradation
		fmt.Printf("Warning: failed to register coverage protocol: %v\n", err)
	}
}
