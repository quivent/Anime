package installer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/interfaces"
)

// ProgressUpdate is an alias to the interfaces.ProgressUpdate type.
// This maintains backward compatibility while ensuring interface satisfaction.
type ProgressUpdate = interfaces.ProgressUpdate

type Installer struct {
	client     interfaces.SSHClient
	progress   chan ProgressUpdate
	parallel   bool
	jobs       int
	serverName string
}

func New(client interfaces.SSHClient) *Installer {
	return &Installer{
		client:   client,
		progress: make(chan ProgressUpdate, 100),
	}
}

func (i *Installer) GetProgressChannel() <-chan ProgressUpdate {
	return i.progress
}

func (i *Installer) sendProgress(module, status, output string, err error, done bool) {
	i.progress <- ProgressUpdate{
		Module: module,
		Status: status,
		Output: output,
		Error:  err,
		Done:   done,
	}
}

func (i *Installer) Install(modules []string) error {
	defer close(i.progress)

	// Resolve dependencies
	allModules := i.resolveDependencies(modules)

	i.sendProgress("", "Starting installation", fmt.Sprintf("Installing %d modules", len(allModules)), nil, false)

	// If parallel mode is enabled, use parallel installation
	if i.parallel {
		return i.installParallel(allModules)
	}

	// Sequential installation
	for _, modID := range allModules {
		if err := i.installModule(modID); err != nil {
			i.sendProgress(modID, "Failed", "", err, true)
			return fmt.Errorf("failed to install %s: %w", modID, err)
		}
	}

	i.sendProgress("", "Complete", "All modules installed successfully", nil, true)
	return nil
}

// installParallel installs modules in parallel based on dependency graph
func (i *Installer) installParallel(modules []string) error {
	// Build dependency map
	depMap := make(map[string][]string)
	for _, mod := range config.AvailableModules {
		depMap[mod.ID] = mod.Dependencies
	}

	// Track completed modules
	completed := make(map[string]bool)
	completedMutex := &sync.Mutex{}

	// Track errors
	var firstError error
	errorMutex := &sync.Mutex{}

	// Wait group for all modules
	var wg sync.WaitGroup

	// Dynamically determine parallelism based on system resources
	maxParallel := i.getOptimalParallelism()
	// Channel to limit concurrent installations
	semaphore := make(chan struct{}, maxParallel)

	// Function to check if dependencies are satisfied
	canInstall := func(modID string) bool {
		completedMutex.Lock()
		defer completedMutex.Unlock()
		for _, dep := range depMap[modID] {
			if !completed[dep] {
				return false
			}
		}
		return true
	}

	// Mark module as completed
	markCompleted := func(modID string) {
		completedMutex.Lock()
		defer completedMutex.Unlock()
		completed[modID] = true
	}

	// Set error if first one
	setError := func(err error) {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		if firstError == nil {
			firstError = err
		}
	}

	// Install modules
	for len(completed) < len(modules) {
		// Check if there's an error
		errorMutex.Lock()
		if firstError != nil {
			errorMutex.Unlock()
			break
		}
		errorMutex.Unlock()

		// Find modules ready to install
		for _, modID := range modules {
			completedMutex.Lock()
			alreadyCompleted := completed[modID]
			completedMutex.Unlock()

			if !alreadyCompleted && canInstall(modID) {
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					semaphore <- struct{}{} // Acquire
					defer func() { <-semaphore }() // Release

					if err := i.installModule(id); err != nil {
						i.sendProgress(id, "Failed", "", err, true)
						setError(fmt.Errorf("failed to install %s: %w", id, err))
						return
					}
					markCompleted(id)
				}(modID)
			}
		}

		// Small sleep to avoid busy waiting
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

	if firstError != nil {
		return firstError
	}

	i.sendProgress("", "Complete", "All modules installed successfully", nil, true)
	return nil
}

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
		return fmt.Errorf("module %s not found", modID)
	}

	i.sendProgress(modID, "Starting", fmt.Sprintf("Installing %s", module.Name), nil, false)

	// Get script with job parallelism injected
	script, ok := GetScript(module.Script)
	if !ok {
		return fmt.Errorf("script not found for module %s", modID)
	}

	// Inject CPU core count and jobs for parallel compilation
	nproc := "$(nproc)"
	if i.jobs > 0 {
		nproc = fmt.Sprintf("%d", i.jobs)
	}
	script = "export MAKEFLAGS=\"-j" + nproc + "\"\n" +
		"export CMAKE_BUILD_PARALLEL_LEVEL=" + nproc + "\n" +
		"export PIP_NO_CACHE_DIR=1\n" +
		"export PIP_TIMEOUT=300\n" +
		"export PIP_RETRIES=3\n" +
		"export PIP_PROGRESS_BAR=off\n" +
		"export PIP_NO_INPUT=1\n" +
		"export PIP_NO_COMPILE=1\n" +
		"export PIP_PREFER_BINARY=1\n" +
		"export PIP_DISABLE_PIP_VERSION_CHECK=1\n" +
		script

	// Upload script
	remotePath := fmt.Sprintf("/tmp/anime-install-%s.sh", modID)
	if err := i.client.UploadString(script, remotePath); err != nil {
		return fmt.Errorf("failed to upload script: %w", err)
	}

	// Make executable
	if err := i.client.MakeExecutable(remotePath); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute with progress
	progressChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		errChan <- i.client.RunCommandWithProgress(fmt.Sprintf("bash %s", remotePath), progressChan)
	}()

	// Stream output
	var lastOutput time.Time
	var outputBuffer strings.Builder

	for {
		select {
		case line, ok := <-progressChan:
			if !ok {
				goto DONE
			}
			outputBuffer.WriteString(line)

			// Send progress updates every 500ms or on newline
			if time.Since(lastOutput) > 500*time.Millisecond || strings.Contains(line, "\n") {
				i.sendProgress(modID, "Installing", outputBuffer.String(), nil, false)
				outputBuffer.Reset()
				lastOutput = time.Now()
			}

		case err := <-errChan:
			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}
			goto DONE
		}
	}

DONE:
	// Cleanup
	i.client.RunCommand(fmt.Sprintf("rm -f %s", remotePath))

	i.sendProgress(modID, "Complete", fmt.Sprintf("%s installed successfully", module.Name), nil, true)
	return nil
}

func (i *Installer) resolveDependencies(modules []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	var addDeps func(string)
	addDeps = func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true

		// Find module
		for _, mod := range config.AvailableModules {
			if mod.ID == id {
				// Add dependencies first
				for _, dep := range mod.Dependencies {
					addDeps(dep)
				}
				result = append(result, id)
				break
			}
		}
	}

	for _, id := range modules {
		addDeps(id)
	}

	return result
}

func (i *Installer) TestConnection() error {
	output, err := i.client.RunCommand("echo 'Connection successful'")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "Connection successful") {
		return fmt.Errorf("unexpected output: %s", output)
	}
	return nil
}

func (i *Installer) CheckNVIDIA() (bool, error) {
	output, err := i.client.RunCommand("nvidia-smi --query-gpu=name --format=csv,noheader")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) != "", nil
}

func (i *Installer) GetSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// OS
	output, _ := i.client.RunCommand("cat /etc/os-release | grep PRETTY_NAME | cut -d'=' -f2 | tr -d '\"'")
	info["os"] = strings.TrimSpace(output)

	// Kernel
	output, _ = i.client.RunCommand("uname -r")
	info["kernel"] = strings.TrimSpace(output)

	// Architecture
	output, _ = i.client.RunCommand("uname -m")
	info["arch"] = strings.TrimSpace(output)

	// CPU cores
	output, _ = i.client.RunCommand("nproc")
	info["cpu_cores"] = strings.TrimSpace(output)

	// GPU
	output, _ = i.client.RunCommand("nvidia-smi --query-gpu=name --format=csv,noheader")
	if strings.TrimSpace(output) != "" {
		info["gpu"] = strings.TrimSpace(output)
	} else {
		info["gpu"] = "Not detected"
	}

	// GPU count
	output, _ = i.client.RunCommand("nvidia-smi --list-gpus | wc -l")
	info["gpu_count"] = strings.TrimSpace(output)

	// Disk space
	output, _ = i.client.RunCommand("df -h / | tail -1 | awk '{print $4}'")
	info["disk_free"] = strings.TrimSpace(output)

	// Memory
	output, _ = i.client.RunCommand("free -h | grep Mem | awk '{print $7}'")
	info["mem_free"] = strings.TrimSpace(output)

	return info, nil
}

// SetParallel enables or disables parallel installation
func (i *Installer) SetParallel(parallel bool) {
	i.parallel = parallel
}

// SetJobs sets the number of parallel compilation jobs
func (i *Installer) SetJobs(jobs int) {
	i.jobs = jobs
}

// SetServerName sets the server name for error reporting
func (i *Installer) SetServerName(name string) {
	i.serverName = name
}

// GetServerName returns the server name
func (i *Installer) GetServerName() string {
	return i.serverName
}

// getOptimalParallelism calculates the optimal number of parallel installations
// based on GPU count and CPU cores
func (i *Installer) getOptimalParallelism() int {
	info, err := i.GetSystemInfo()
	if err != nil {
		return 3 // Fallback to default
	}

	// Parse GPU count
	gpuCount := 1
	if gpuCountStr, ok := info["gpu_count"]; ok {
		if count, err := fmt.Sscanf(gpuCountStr, "%d", &gpuCount); err == nil && count == 1 {
			if gpuCount == 0 {
				gpuCount = 1
			}
		}
	}

	// Parse CPU cores
	cpuCores := 4
	if cpuCoresStr, ok := info["cpu_cores"]; ok {
		if count, err := fmt.Sscanf(cpuCoresStr, "%d", &cpuCores); err == nil && count == 1 {
			if cpuCores < 4 {
				cpuCores = 4
			}
		}
	}

	// Calculate optimal parallelism:
	// - For systems with multiple GPUs, we can be more aggressive
	// - Base: 2 parallel installations
	// - Add 1 per GPU (up to 4 additional)
	// - Add 1 per 16 CPU cores (up to 2 additional)
	optimal := 2
	optimal += min(gpuCount, 4)
	optimal += min(cpuCores/16, 2)

	// Cap at reasonable maximum
	if optimal > 8 {
		optimal = 8
	}

	return optimal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compile-time check to ensure Installer implements interfaces.Installer
var _ interfaces.Installer = (*Installer)(nil)
