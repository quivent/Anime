package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadStackConfig loads and parses an anime.yaml file from the given path
func LoadStackConfig(path string) (*StackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack config: %w", err)
	}

	var cfg StackConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse stack config: %w", err)
	}

	return &cfg, nil
}

// FindStackConfig searches for anime.yaml in the given directory and parent directories
// Returns the path to the found config file, or an error if not found
func FindStackConfig(dir string) (string, error) {
	// Start from the given directory and search upward
	current := dir

	// Get absolute path
	absPath, err := filepath.Abs(current)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	current = absPath

	for {
		// Check for anime.yaml in current directory
		configPath := filepath.Join(current, "anime.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Also check for anime.yml
		configPathYml := filepath.Join(current, "anime.yml")
		if _, err := os.Stat(configPathYml); err == nil {
			return configPathYml, nil
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root, config not found
			return "", fmt.Errorf("anime.yaml not found in %s or any parent directory", dir)
		}
		current = parent
	}
}

// ValidateStackConfig validates a stack configuration
func ValidateStackConfig(cfg *StackConfig) error {
	if cfg == nil {
		return fmt.Errorf("stack config is nil")
	}

	if cfg.Name == "" {
		return fmt.Errorf("stack name is required")
	}

	if len(cfg.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}

	// Validate each service
	for name, svc := range cfg.Services {
		if svc == nil {
			return fmt.Errorf("service %q is nil", name)
		}

		// Check if it's a database service
		if svc.Type == "postgres" {
			// Database services only need name
			if svc.Name == "" && svc.URL == "" {
				return fmt.Errorf("database service %q must have either 'name' or 'url' set", name)
			}
			continue
		}

		// Non-database services need a start command or build
		if svc.Start == "" && svc.Build == "" {
			return fmt.Errorf("service %q must have either 'start' or 'build' set", name)
		}
	}

	// Validate depends_on references
	for name, svc := range cfg.Services {
		for _, dep := range svc.DependsOn {
			if _, exists := cfg.Services[dep]; !exists {
				return fmt.Errorf("service %q depends on unknown service %q", name, dep)
			}
		}
	}

	// Check for circular dependencies
	if err := validateNoCycles(cfg.Services); err != nil {
		return err
	}

	// Validate routing config if present
	if cfg.Routing != nil {
		if cfg.Routing.Domain == "" && len(cfg.Routing.Paths) == 0 {
			return fmt.Errorf("routing config must have domain or paths set")
		}

		for path, serviceName := range cfg.Routing.Paths {
			if _, exists := cfg.Services[serviceName]; !exists {
				return fmt.Errorf("routing path %q references unknown service %q", path, serviceName)
			}
		}
	}

	// Validate auth config if present
	if cfg.Auth != nil {
		validAuthTypes := map[string]bool{
			"oauth2-google": true,
			"oauth2-github": true,
			"basic":         true,
			"none":          true,
		}
		if cfg.Auth.Type != "" && !validAuthTypes[cfg.Auth.Type] {
			return fmt.Errorf("invalid auth type %q", cfg.Auth.Type)
		}
	}

	return nil
}

// validateNoCycles checks for circular dependencies in the service graph
func validateNoCycles(services map[string]*ServiceConfig) error {
	// Track visit state: 0 = unvisited, 1 = visiting, 2 = visited
	state := make(map[string]int)

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		if state[name] == 1 {
			// Currently visiting - found a cycle
			cycle := append(path, name)
			return fmt.Errorf("circular dependency detected: %v", cycle)
		}
		if state[name] == 2 {
			// Already visited
			return nil
		}

		state[name] = 1 // Mark as visiting
		path = append(path, name)

		svc := services[name]
		if svc != nil {
			for _, dep := range svc.DependsOn {
				if err := visit(dep, path); err != nil {
					return err
				}
			}
		}

		state[name] = 2 // Mark as visited
		return nil
	}

	for name := range services {
		if state[name] == 0 {
			if err := visit(name, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
