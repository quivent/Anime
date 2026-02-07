package stack

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joshkornreich/anime/internal/launch"
)

// DeployStack orchestrates the deployment of a multi-service stack
func DeployStack(cfg *StackConfig, serverName, sudoPassword string, runner launch.CommandRunner) error {
	if err := ValidateStackConfig(cfg); err != nil {
		return fmt.Errorf("invalid stack config: %w", err)
	}

	// Get deployment order using topological sort
	order, err := topologicalSort(cfg.Services)
	if err != nil {
		return fmt.Errorf("failed to determine deployment order: %w", err)
	}

	// Track provisioned databases for variable interpolation
	databaseURLs := make(map[string]string)

	// Deploy services in order
	for _, serviceName := range order {
		svc := cfg.Services[serviceName]

		// Handle database services specially
		if svc.Type == "postgres" {
			url, err := deployDatabase(cfg.Name, serviceName, svc, sudoPassword, runner)
			if err != nil {
				return fmt.Errorf("failed to deploy database %q: %w", serviceName, err)
			}
			databaseURLs[serviceName] = url
			continue
		}

		// Deploy regular service
		if err := deployService(cfg.Name, serviceName, svc, databaseURLs, sudoPassword, runner); err != nil {
			return fmt.Errorf("failed to deploy service %q: %w", serviceName, err)
		}
	}

	return nil
}

// topologicalSort returns services in dependency order (dependencies first)
func topologicalSort(services map[string]*ServiceConfig) ([]string, error) {
	// Build adjacency list and in-degree count
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	// Initialize
	for name := range services {
		inDegree[name] = 0
	}

	// Build graph
	for name, svc := range services {
		for _, dep := range svc.DependsOn {
			dependents[dep] = append(dependents[dep], name)
			inDegree[name]++
		}
	}

	// Find all nodes with no dependencies
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Process nodes in topological order
	var result []string
	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Reduce in-degree for dependents
		for _, dependent := range dependents[node] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check if all nodes were processed
	if len(result) != len(services) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// deployDatabase provisions a database and returns its connection URL
func deployDatabase(stackName, serviceName string, svc *ServiceConfig, sudoPassword string, runner launch.CommandRunner) (string, error) {
	// If external URL is provided, use it
	if svc.URL != "" {
		return svc.URL, nil
	}

	// Provision local postgres database
	dbName := svc.Name
	if dbName == "" {
		dbName = stackName + "_" + serviceName
	}

	dbUser := stackName + "_user"
	dbPassword := launch.GenerateRandomPassword(16)

	if err := launch.ProvisionPostgres(dbName, dbUser, dbPassword, sudoPassword, runner); err != nil {
		return "", err
	}

	// Return connection URL
	return fmt.Sprintf("postgresql://%s:%s@localhost:5432/%s", dbUser, dbPassword, dbName), nil
}

// deployService deploys a regular (non-database) service
func deployService(stackName, serviceName string, svc *ServiceConfig, databaseURLs map[string]string, sudoPassword string, runner launch.CommandRunner) error {
	// Prepare environment with interpolated values
	env := make(map[string]string)
	for k, v := range svc.Env {
		env[k] = interpolateEnv(v, databaseURLs)
	}

	// Determine working directory
	workDir := svc.Path
	if workDir == "" {
		workDir = "/home/" + runner.User() + "/apps/" + stackName + "/" + serviceName
	}

	// Run build command if specified
	if svc.Build != "" {
		buildCmd := fmt.Sprintf("cd %s && %s", workDir, svc.Build)
		if _, err := runner.Run(buildCmd); err != nil {
			return fmt.Errorf("build command failed: %w", err)
		}
	}

	// Determine port
	port := svc.Port
	if port == 0 {
		port = 3000 // Default port
	}

	// Set PORT in environment
	env["PORT"] = fmt.Sprintf("%d", port)

	// Create systemd service
	fullServiceName := launch.ServiceName(stackName + "-" + serviceName)

	systemdCfg := launch.SystemdConfig{
		Name:        fullServiceName,
		Description: fmt.Sprintf("%s - %s service", stackName, serviceName),
		ExecStart:   svc.Start,
		WorkingDir:  workDir,
		User:        runner.User(),
		Port:        port,
		Environment: env,
	}

	unitContent, err := launch.GenerateSystemdUnit(systemdCfg)
	if err != nil {
		return fmt.Errorf("failed to generate systemd unit: %w", err)
	}

	if err := launch.InstallSystemdUnit(fullServiceName, unitContent, sudoPassword, runner); err != nil {
		return fmt.Errorf("failed to install systemd unit: %w", err)
	}

	// Set up nginx if domain is specified
	if svc.Domain != "" {
		nginxCfg := launch.NginxConfig{
			Domain:   svc.Domain,
			Port:     port,
			AppName:  stackName + "-" + serviceName,
			AuthType: "none",
		}

		nginxContent, err := launch.GenerateNginxConfig(nginxCfg)
		if err != nil {
			return fmt.Errorf("failed to generate nginx config: %w", err)
		}

		if err := launch.InstallNginxConfig(stackName+"-"+serviceName, nginxContent, sudoPassword, runner); err != nil {
			return fmt.Errorf("failed to install nginx config: %w", err)
		}
	}

	return nil
}

// interpolateEnv replaces ${database.url} patterns in environment values
func interpolateEnv(value string, databaseURLs map[string]string) string {
	// Pattern: ${serviceName.url}
	pattern := regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\.url\}`)

	return pattern.ReplaceAllStringFunc(value, func(match string) string {
		// Extract service name from ${serviceName.url}
		serviceName := strings.TrimPrefix(match, "${")
		serviceName = strings.TrimSuffix(serviceName, ".url}")

		if url, exists := databaseURLs[serviceName]; exists {
			return url
		}
		// Return original if not found
		return match
	})
}
