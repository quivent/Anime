# Structured Logging Implementation Examples

This document provides specific examples of how to integrate structured logging into key files of the anime CLI.

## Example 1: internal/installer/installer.go

Here's how to add structured logging to the installer package:

```go
// Add import
import (
	"github.com/joshkornreich/anime/internal/logger"
)

// In Install() method
func (i *Installer) Install(modules []string) error {
	defer close(i.progress)

	// Resolve dependencies
	allModules := i.resolveDependencies(modules)

	// ADD: Log installation start
	logger.Info("Starting module installation",
		"module_count", len(allModules),
		"modules", allModules,
		"parallel", i.parallel,
		"jobs", i.jobs,
		"server", i.serverName,
	)

	// Keep existing TUI progress
	i.sendProgress("", "Starting installation",
		fmt.Sprintf("Installing %d modules", len(allModules)), nil, false)

	// ... existing code ...

	// ADD: Log completion
	logger.Info("All modules installed successfully",
		"module_count", len(allModules),
		"server", i.serverName,
	)

	return nil
}

// In installModule() method
func (i *Installer) installModule(modID string) error {
	// Find module info
	var module *config.Module
	for _, m := range config.AvailableModules {
		if m.ID == modID {
			module = &m
			break
		}
	}

	if module == nil {
		// ADD: Log error
		logger.Error("Module not found", "module_id", modID)
		return fmt.Errorf("module %s not found", modID)
	}

	// ADD: Create module-specific logger
	modLogger := logger.WithModule(modID)
	modLogger.Info("Installing module",
		"name", module.Name,
		"description", module.Description,
		"estimated_time_minutes", module.TimeMinutes,
		"dependencies", module.Dependencies,
		"server", i.serverName,
	)

	// Keep existing TUI progress
	i.sendProgress(modID, "Starting",
		fmt.Sprintf("Installing %s", module.Name), nil, false)

	// Get script
	script, ok := GetScript(module.Script)
	if !ok {
		// ADD: Log script error
		modLogger.Error("Script not found for module")
		return fmt.Errorf("script not found for module %s", modID)
	}

	// ADD: Log debug info
	modLogger.Debug("Script loaded", "script_id", module.Script)

	// Upload script
	remotePath := fmt.Sprintf("/tmp/anime-install-%s.sh", modID)
	// ADD: Log upload
	modLogger.Debug("Uploading installation script", "remote_path", remotePath)

	if err := i.client.UploadString(script, remotePath); err != nil {
		// ADD: Log upload error
		modLogger.Error("Failed to upload script", "error", err)
		return fmt.Errorf("failed to upload script: %w", err)
	}

	// Make executable
	if err := i.client.MakeExecutable(remotePath); err != nil {
		// ADD: Log chmod error
		modLogger.Error("Failed to make script executable", "error", err)
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// ... execution logic ...

DONE:
	// Cleanup
	i.client.RunCommand(fmt.Sprintf("rm -f %s", remotePath))

	// ADD: Log completion
	modLogger.Info("Module installation complete", "name", module.Name)

	// Keep existing TUI progress
	i.sendProgress(modID, "Complete",
		fmt.Sprintf("%s installed successfully", module.Name), nil, true)

	return nil
}

// In getOptimalParallelism()
func (i *Installer) getOptimalParallelism() int {
	info, err := i.GetSystemInfo()
	if err != nil {
		// ADD: Log warning
		logger.Warn("Failed to get system info for parallelism calculation", "error", err)
		return 3 // Fallback to default
	}

	// ... calculation logic ...

	// ADD: Log calculated value
	logger.Debug("Calculated optimal parallelism",
		"parallelism", optimal,
		"gpu_count", gpuCount,
		"cpu_cores", cpuCores,
	)

	return optimal
}
```

## Example 2: internal/ssh/client.go

Here's how to add structured logging to the SSH client:

```go
// Add import
import (
	"github.com/joshkornreich/anime/internal/logger"
)

// In NewClient()
func NewClient(host, user, keyPath string) (*Client, error) {
	// ADD: Log connection attempt
	logger.Debug("Creating SSH client",
		"host", host,
		"user", user,
		"key_path", keyPath,
	)

	var authMethods []ssh.AuthMethod

	// Try SSH agent first
	if agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		// ADD: Log agent usage
		logger.Debug("Using SSH agent authentication")
		agentClient := agent.NewClient(agentConn)
		authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
	} else {
		// ADD: Log agent failure
		logger.Debug("SSH agent not available", "error", err)
	}

	// If key path specified, use it
	if keyPath != "" {
		// ... key loading logic ...

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// ADD: Log key parse error
			logger.Error("Failed to parse private key",
				"key_path", keyPath,
				"error", err,
			)
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		// ADD: Log key loaded
		logger.Debug("SSH private key loaded", "key_path", keyPath)
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// ... connection logic ...

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		// ADD: Log connection failure
		logger.Error("SSH connection failed",
			"host", host,
			"user", user,
			"error", err,
		)
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	// ADD: Log successful connection
	logger.Info("SSH connection established",
		"host", host,
		"user", user,
	)

	return &Client{
		client: client,
		config: config,
		host:   host,
	}, nil
}

// In RunCommand()
func (c *Client) RunCommand(cmd string) (string, error) {
	// ADD: Log command execution
	logger.Debug("Executing SSH command",
		"host", c.Host(),
		"command", cmd,
	)

	session, err := c.client.NewSession()
	if err != nil {
		// ADD: Log session error
		logger.Error("Failed to create SSH session",
			"host", c.Host(),
			"error", err,
		)
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		// ADD: Log command failure
		logger.Warn("SSH command failed",
			"host", c.Host(),
			"command", cmd,
			"error", err,
			"output", string(output),
		)
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	// ADD: Log success (debug only to avoid spam)
	logger.Debug("SSH command completed",
		"host", c.Host(),
		"output_length", len(output),
	)

	return string(output), nil
}
```

## Example 3: internal/config/config.go

Here's how to add structured logging to the config package:

```go
// Add import
import (
	"github.com/joshkornreich/anime/internal/logger"
)

// In Load()
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		// ADD: Log path error
		logger.Error("Failed to get config path", "error", err)
		return nil, err
	}

	// ADD: Log load attempt
	logger.Debug("Loading configuration", "path", path)

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// ADD: Log default creation
		logger.Info("Config file not found, creating default",
			"path", path,
		)
		return &Config{
			Servers:     []Server{},
			APIKeys:     APIKeys{},
			Aliases:     make(map[string]string),
			Collections: []Collection{},
			Users:       []User{},
			ActiveUser:  "",
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// ADD: Log read error
		logger.Error("Failed to read config file",
			"path", path,
			"error", err,
		)
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// ADD: Log parse error
		logger.Error("Failed to parse config YAML",
			"path", path,
			"error", err,
		)
		return nil, err
	}

	// ... initialization logic ...

	// ADD: Log successful load
	logger.Info("Configuration loaded successfully",
		"path", path,
		"server_count", len(cfg.Servers),
		"alias_count", len(cfg.Aliases),
	)

	return &cfg, nil
}

// In Save()
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		// ADD: Log path error
		logger.Error("Failed to get config path", "error", err)
		return err
	}

	// ADD: Log save attempt
	logger.Debug("Saving configuration",
		"path", path,
		"server_count", len(c.Servers),
		"alias_count", len(c.Aliases),
	)

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// ADD: Log directory error
		logger.Error("Failed to create config directory",
			"dir", dir,
			"error", err,
		)
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		// ADD: Log marshal error
		logger.Error("Failed to marshal config to YAML",
			"error", err,
		)
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		// ADD: Log write error
		logger.Error("Failed to write config file",
			"path", path,
			"error", err,
		)
		return err
	}

	// ADD: Log successful save
	logger.Info("Configuration saved successfully", "path", path)

	return nil
}

// In AddServer()
func (c *Config) AddServer(server Server) {
	// ADD: Log server addition
	logger.Info("Adding server to configuration",
		"name", server.Name,
		"host", server.Host,
		"user", server.User,
	)

	c.Servers = append(c.Servers, server)
}

// In DeleteServer()
func (c *Config) DeleteServer(name string) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			// ADD: Log server deletion
			logger.Info("Deleting server from configuration",
				"name", name,
				"host", c.Servers[i].Host,
			)

			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			return nil
		}
	}

	// ADD: Log not found
	logger.Warn("Server not found for deletion", "name", name)

	return fmt.Errorf("server %s not found", name)
}
```

## Example 4: cmd/deploy.go

Here's how to add structured logging to deploy command:

```go
// Add import
import (
	"github.com/joshkornreich/anime/internal/logger"
)

// In runDeploy()
func runDeploy(cmd *cobra.Command, args []string) error {
	// ... argument validation ...

	serverName := args[0]

	// ADD: Create operation-specific logger
	deployLogger := logger.WithOperation("deploy")
	deployLogger.Info("Starting deployment",
		"server", serverName,
		"modules", server.Modules,
		"module_count", len(server.Modules),
	)

	// Keep existing TUI output
	fmt.Println(theme.InfoStyle.Render("🚀 Deploying to " + serverName))
	fmt.Println()

	// ... server config loading ...

	// ADD: Log installer creation
	deployLogger.Debug("Creating installer instance",
		"host", server.Host,
		"user", server.User,
	)

	m := tui.NewInstallModel(server)
	p := tea.NewProgram(m)

	startTime := time.Now()

	if _, err := p.Run(); err != nil {
		// ADD: Log deployment failure
		deployLogger.Error("Deployment failed",
			"server", serverName,
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return fmt.Errorf("error running installation: %w", err)
	}

	// ADD: Log successful deployment
	deployLogger.Info("Deployment completed successfully",
		"server", serverName,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	return nil
}
```

## Example 5: cmd/push.go

Here's how to add structured logging to push command (partial):

```go
// Add import
import (
	"github.com/joshkornreich/anime/internal/logger"
)

// In runPush()
func runPush(cmd *cobra.Command, args []string) error {
	// ... argument parsing ...

	// ADD: Create operation-specific logger
	pushLogger := logger.WithOperation("push")
	pushLogger.Info("Starting push operation",
		"target", target,
		"arch", pushArch,
		"include_source", pushIncludeSource,
		"fast_mode", pushFast,
		"binary_only", pushBinaryOnly,
	)

	// Keep existing TUI output
	if pushFast {
		fmt.Println(theme.InfoStyle.Render("⚡ Fast push to remote server"))
	} else {
		fmt.Println(theme.InfoStyle.Render("🚀 Pushing anime to remote server"))
	}
	fmt.Println()

	// Keep TUI progress indicator
	fmt.Print(theme.DimTextStyle.Render("▶ Preparing (build + connect)... "))

	// ADD: Log build start
	pushLogger.Debug("Building Linux binary",
		"arch", pushArch,
		"source_dir", sourceDir,
	)

	binaryPath, version, buildTime, sourceDir, err := buildLinuxBinary()
	if err != nil {
		// ADD: Log build failure
		pushLogger.Error("Binary build failed",
			"error", err,
			"arch", pushArch,
		)

		// Keep TUI error indicator
		fmt.Println(theme.ErrorStyle.Render("✗"))
		showBuildSuggestions(err)
		return fmt.Errorf("build failed")
	}

	// ADD: Log build success
	pushLogger.Info("Binary built successfully",
		"version", version,
		"arch", pushArch,
		"binary_path", binaryPath,
	)

	// Keep TUI success indicator
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// ... rest of push logic with similar pattern ...

	return nil
}

// In buildLinuxBinary()
func buildLinuxBinary() (binaryPath, version, buildTime, sourceDir string, err error) {
	// ADD: Log source directory detection
	sourceDir, err = findSourceDir()
	if err != nil {
		logger.Error("Failed to find source directory", "error", err)
		return "", "", "", "", fmt.Errorf("could not find source directory: %w", err)
	}

	logger.Debug("Found source directory", "path", sourceDir)

	// ... build logic ...

	// ADD: Log build command
	logger.Debug("Executing go build",
		"output", binaryPath,
		"GOOS", "linux",
		"GOARCH", pushArch,
		"ldflags", ldflags,
	)

	output, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		// ADD: Log build error with output
		logger.Error("Go build failed",
			"error", buildErr,
			"output", string(output),
		)
		return "", "", "", "", fmt.Errorf("%w: %s", buildErr, string(output))
	}

	return binaryPath, version, buildTime, sourceDir, nil
}
```

## Summary

Key patterns to follow:

1. **Import the logger package**: `import "github.com/joshkornreich/anime/internal/logger"`

2. **Log operational events**: Use `logger.Info()` for important operations, `logger.Debug()` for details

3. **Add structured fields**: Always include relevant context (server, module, operation, error, etc.)

4. **Keep TUI output separate**: Don't replace `fmt.Print` calls that are part of the user interface

5. **Use contextual loggers**: Create operation/server/module-specific loggers with `WithOperation()`, `WithServer()`, `WithModule()`

6. **Log errors with context**: Always log errors before returning them

7. **Log timing information**: Include duration_ms for long-running operations

8. **Be selective with debug logs**: Use debug level for verbose information that helps troubleshooting

These examples demonstrate the incremental approach to adding structured logging while preserving the existing TUI functionality.
