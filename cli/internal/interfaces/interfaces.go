// Package interfaces defines core abstractions for the anime CLI components.
// These interfaces enable loose coupling, dependency injection, and testability.
package interfaces

// SSHClient defines the interface for SSH operations.
// Implementations must provide secure remote command execution,
// file transfer, and connection management.
type SSHClient interface {
	// RunCommand executes a command on the remote host and returns the output.
	RunCommand(cmd string) (string, error)

	// RunCommandWithProgress executes a command and streams output through the progress channel.
	// The channel is written to by the implementation and should be read by the caller.
	// The progress parameter is send-only from the implementation's perspective.
	RunCommandWithProgress(cmd string, progress chan<- string) error

	// UploadString uploads a string content to a remote path.
	UploadString(content, path string) error

	// MakeExecutable sets executable permissions on a remote file.
	MakeExecutable(path string) error

	// Close closes the SSH connection and releases resources.
	Close() error
}

// ProgressUpdate represents a progress event during module installation.
// It provides structured information about the installation state.
type ProgressUpdate struct {
	Module string // Module identifier
	Status string // Current status (e.g., "Starting", "Installing", "Complete", "Failed")
	Output string // Output text from the installation process
	Error  error  // Error if the operation failed
	Done   bool   // True if this module's installation is complete
}

// Installer defines the interface for module installation operations.
// Implementations handle dependency resolution, parallel installation,
// and progress reporting.
type Installer interface {
	// Install installs the specified modules and their dependencies.
	// Returns an error if any module installation fails.
	Install(modules []string) error

	// GetProgressChannel returns a read-only channel for receiving progress updates.
	// The channel is closed when installation completes or fails.
	GetProgressChannel() <-chan ProgressUpdate

	// SetParallel enables or disables parallel module installation.
	// When enabled, modules are installed concurrently based on dependencies.
	SetParallel(parallel bool)

	// SetJobs sets the number of parallel compilation jobs for build systems.
	// Affects MAKEFLAGS and CMAKE_BUILD_PARALLEL_LEVEL environment variables.
	SetJobs(jobs int)

	// TestConnection verifies that the SSH connection is working.
	TestConnection() error
}

// SourceController defines the interface for source control operations.
// This provides rsync-based version control functionality for managing
// code synchronization between local and remote hosts.
type SourceController interface {
	// Push syncs local changes to the remote repository.
	Push() error

	// Pull syncs remote changes to the local repository.
	Pull() error

	// Status checks the synchronization status between local and remote.
	Status() error

	// Link associates the current directory with a remote path.
	Link(remotePath string) error

	// Unlink removes the association with a remote path.
	Unlink() error
}

// PackageManager defines the interface for package management operations.
// This provides functionality similar to npm/pip for managing code packages
// in a distributed development environment.
type PackageManager interface {
	// Install installs a package from the registry.
	// The name can include version specification (e.g., "package@1.0.0").
	Install(name string, global bool, force bool) error

	// Uninstall removes an installed package.
	Uninstall(name string, global bool) error

	// Publish publishes the current package to the registry.
	Publish() error

	// Search searches for packages matching the query.
	Search(query string) error

	// List lists installed packages.
	List(global bool) error
}
