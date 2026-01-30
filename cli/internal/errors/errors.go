package errors

import "fmt"

// SSHError represents errors during SSH operations
type SSHError struct {
	Host    string
	Op      string // "connect", "execute", "upload", "download"
	Wrapped error
}

func (e *SSHError) Error() string {
	return fmt.Sprintf("ssh %s to %s: %v", e.Op, e.Host, e.Wrapped)
}

func (e *SSHError) Unwrap() error {
	return e.Wrapped
}

func NewSSHError(host, op string, err error) *SSHError {
	return &SSHError{Host: host, Op: op, Wrapped: err}
}

// NewSSHKeyError creates an SSHError for key-related failures
func NewSSHKeyError(keyPath string, err error) *SSHError {
	return &SSHError{Host: keyPath, Op: "load_key", Wrapped: err}
}

// NewSSHAuthError creates an SSHError for authentication failures
func NewSSHAuthError(user, host string, err error) *SSHError {
	return &SSHError{Host: fmt.Sprintf("%s@%s", user, host), Op: "authenticate", Wrapped: err}
}

// NewSSHConnectionError creates an SSHError for connection failures
func NewSSHConnectionError(host string, err error) *SSHError {
	return &SSHError{Host: host, Op: "connect", Wrapped: err}
}

// InstallError represents errors during module installation
type InstallError struct {
	Module  string
	Phase   string // "resolve", "download", "compile", "configure", "verify"
	Wrapped error
}

func (e *InstallError) Error() string {
	return fmt.Sprintf("install %s during %s: %v", e.Module, e.Phase, e.Wrapped)
}

func (e *InstallError) Unwrap() error {
	return e.Wrapped
}

func NewInstallError(module, phase string, err error) *InstallError {
	return &InstallError{Module: module, Phase: phase, Wrapped: err}
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error for %s: %s", e.Field, e.Reason)
}

func NewConfigError(field, reason string) *ConfigError {
	return &ConfigError{Field: field, Reason: reason}
}

// SourceError represents source control operation errors
type SourceError struct {
	Op      string // "push", "pull", "status", "link", "unlink"
	Path    string
	Wrapped error
}

func (e *SourceError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("source %s for %s: %v", e.Op, e.Path, e.Wrapped)
	}
	return fmt.Sprintf("source %s: %v", e.Op, e.Wrapped)
}

func (e *SourceError) Unwrap() error {
	return e.Wrapped
}

func NewSourceError(op, path string, err error) *SourceError {
	return &SourceError{Op: op, Path: path, Wrapped: err}
}

// PackageError represents package management errors
type PackageError struct {
	Package string
	Op      string // "install", "uninstall", "publish", "search"
	Wrapped error
}

func (e *PackageError) Error() string {
	return fmt.Sprintf("package %s %s: %v", e.Op, e.Package, e.Wrapped)
}

func (e *PackageError) Unwrap() error {
	return e.Wrapped
}

func NewPackageError(pkg, op string, err error) *PackageError {
	return &PackageError{Package: pkg, Op: op, Wrapped: err}
}

// ServerNotFoundError represents a server lookup failure
type ServerNotFoundError struct {
	Name string
}

func (e *ServerNotFoundError) Error() string {
	return fmt.Sprintf("server %s not found", e.Name)
}

func NewServerNotFoundError(name string) *ServerNotFoundError {
	return &ServerNotFoundError{Name: name}
}

// DetailedError provides rich error information for user-friendly display
type DetailedError struct {
	Operation   string
	Cause       error
	Context     []string
	Suggestions []string
}

func (e *DetailedError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("failed to %s: %v", e.Operation, e.Cause)
	}
	return e.Cause.Error()
}

func (e *DetailedError) Unwrap() error {
	return e.Cause
}

func NewDetailedError(operation string, cause error) *DetailedError {
	return &DetailedError{Operation: operation, Cause: cause}
}

func (e *DetailedError) WithContext(ctx ...string) *DetailedError {
	e.Context = append(e.Context, ctx...)
	return e
}

func (e *DetailedError) WithSuggestions(suggestions ...string) *DetailedError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// ModuleNotFoundError represents a module lookup failure
type ModuleNotFoundError struct {
	Name string
}

func (e *ModuleNotFoundError) Error() string {
	return fmt.Sprintf("module %s not found", e.Name)
}

func NewModuleNotFoundError(name string) *ModuleNotFoundError {
	return &ModuleNotFoundError{Name: name}
}

// InstallationError represents an installation failure with details
type InstallationError struct {
	Module  string
	Server  string
	Phase   string
	Wrapped error
}

func (e *InstallationError) Error() string {
	return fmt.Sprintf("installation of %s on %s failed during %s: %v", e.Module, e.Server, e.Phase, e.Wrapped)
}

func (e *InstallationError) Unwrap() error {
	return e.Wrapped
}

func NewInstallationError(module, server, phase string, err error) *InstallationError {
	return &InstallationError{Module: module, Server: server, Phase: phase, Wrapped: err}
}
