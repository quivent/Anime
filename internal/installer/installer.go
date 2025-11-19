package installer

import (
	"fmt"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
)

type ProgressUpdate struct {
	Module  string
	Status  string
	Output  string
	Error   error
	Done    bool
}

type Installer struct {
	client   *ssh.Client
	progress chan ProgressUpdate
}

func New(client *ssh.Client) *Installer {
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

	for _, modID := range allModules {
		if err := i.installModule(modID); err != nil {
			i.sendProgress(modID, "Failed", "", err, true)
			return fmt.Errorf("failed to install %s: %w", modID, err)
		}
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

	// Get script
	script, ok := GetScript(module.Script)
	if !ok {
		return fmt.Errorf("script not found for module %s", modID)
	}

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

	// GPU
	output, _ = i.client.RunCommand("nvidia-smi --query-gpu=name --format=csv,noheader")
	if strings.TrimSpace(output) != "" {
		info["gpu"] = strings.TrimSpace(output)
	} else {
		info["gpu"] = "Not detected"
	}

	// Disk space
	output, _ = i.client.RunCommand("df -h / | tail -1 | awk '{print $4}'")
	info["disk_free"] = strings.TrimSpace(output)

	// Memory
	output, _ = i.client.RunCommand("free -h | grep Mem | awk '{print $7}'")
	info["mem_free"] = strings.TrimSpace(output)

	return info, nil
}
