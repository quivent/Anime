package errors

import (
	"errors"
	"testing"
)

func TestSSHError(t *testing.T) {
	baseErr := errors.New("connection refused")
	sshErr := NewSSHError("192.168.1.100", "connect", baseErr)

	t.Run("Error message format", func(t *testing.T) {
		expected := "ssh connect to 192.168.1.100: connection refused"
		if got := sshErr.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		if got := sshErr.Unwrap(); got != baseErr {
			t.Errorf("Unwrap() = %v, want %v", got, baseErr)
		}
	})

	t.Run("Constructor sets fields correctly", func(t *testing.T) {
		if sshErr.Host != "192.168.1.100" {
			t.Errorf("Host = %q, want %q", sshErr.Host, "192.168.1.100")
		}
		if sshErr.Op != "connect" {
			t.Errorf("Op = %q, want %q", sshErr.Op, "connect")
		}
		if sshErr.Wrapped != baseErr {
			t.Errorf("Wrapped = %v, want %v", sshErr.Wrapped, baseErr)
		}
	})

	t.Run("Different operations", func(t *testing.T) {
		tests := []struct {
			op   string
			want string
		}{
			{"execute", "ssh execute to 192.168.1.100: connection refused"},
			{"upload", "ssh upload to 192.168.1.100: connection refused"},
			{"download", "ssh download to 192.168.1.100: connection refused"},
		}

		for _, tt := range tests {
			t.Run(tt.op, func(t *testing.T) {
				err := NewSSHError("192.168.1.100", tt.op, baseErr)
				if got := err.Error(); got != tt.want {
					t.Errorf("Error() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestInstallError(t *testing.T) {
	baseErr := errors.New("checksum mismatch")
	installErr := NewInstallError("comfyui", "download", baseErr)

	t.Run("Error message format", func(t *testing.T) {
		expected := "install comfyui during download: checksum mismatch"
		if got := installErr.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		if got := installErr.Unwrap(); got != baseErr {
			t.Errorf("Unwrap() = %v, want %v", got, baseErr)
		}
	})

	t.Run("Constructor sets fields correctly", func(t *testing.T) {
		if installErr.Module != "comfyui" {
			t.Errorf("Module = %q, want %q", installErr.Module, "comfyui")
		}
		if installErr.Phase != "download" {
			t.Errorf("Phase = %q, want %q", installErr.Phase, "download")
		}
		if installErr.Wrapped != baseErr {
			t.Errorf("Wrapped = %v, want %v", installErr.Wrapped, baseErr)
		}
	})

	t.Run("Different phases", func(t *testing.T) {
		tests := []struct {
			phase string
			want  string
		}{
			{"resolve", "install comfyui during resolve: checksum mismatch"},
			{"compile", "install comfyui during compile: checksum mismatch"},
			{"configure", "install comfyui during configure: checksum mismatch"},
			{"verify", "install comfyui during verify: checksum mismatch"},
		}

		for _, tt := range tests {
			t.Run(tt.phase, func(t *testing.T) {
				err := NewInstallError("comfyui", tt.phase, baseErr)
				if got := err.Error(); got != tt.want {
					t.Errorf("Error() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestConfigError(t *testing.T) {
	configErr := NewConfigError("host", "must not be empty")

	t.Run("Error message format", func(t *testing.T) {
		expected := "config error for host: must not be empty"
		if got := configErr.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Constructor sets fields correctly", func(t *testing.T) {
		if configErr.Field != "host" {
			t.Errorf("Field = %q, want %q", configErr.Field, "host")
		}
		if configErr.Reason != "must not be empty" {
			t.Errorf("Reason = %q, want %q", configErr.Reason, "must not be empty")
		}
	})

	t.Run("Different field errors", func(t *testing.T) {
		tests := []struct {
			field  string
			reason string
			want   string
		}{
			{"port", "must be between 1 and 65535", "config error for port: must be between 1 and 65535"},
			{"username", "invalid characters", "config error for username: invalid characters"},
			{"path", "does not exist", "config error for path: does not exist"},
		}

		for _, tt := range tests {
			t.Run(tt.field, func(t *testing.T) {
				err := NewConfigError(tt.field, tt.reason)
				if got := err.Error(); got != tt.want {
					t.Errorf("Error() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestSourceError(t *testing.T) {
	baseErr := errors.New("remote rejected")
	sourceErr := NewSourceError("push", "/path/to/repo", baseErr)

	t.Run("Error message format with path", func(t *testing.T) {
		expected := "source push for /path/to/repo: remote rejected"
		if got := sourceErr.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Error message format without path", func(t *testing.T) {
		err := NewSourceError("pull", "", baseErr)
		expected := "source pull: remote rejected"
		if got := err.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		if got := sourceErr.Unwrap(); got != baseErr {
			t.Errorf("Unwrap() = %v, want %v", got, baseErr)
		}
	})

	t.Run("Constructor sets fields correctly", func(t *testing.T) {
		if sourceErr.Op != "push" {
			t.Errorf("Op = %q, want %q", sourceErr.Op, "push")
		}
		if sourceErr.Path != "/path/to/repo" {
			t.Errorf("Path = %q, want %q", sourceErr.Path, "/path/to/repo")
		}
		if sourceErr.Wrapped != baseErr {
			t.Errorf("Wrapped = %v, want %v", sourceErr.Wrapped, baseErr)
		}
	})

	t.Run("Different operations", func(t *testing.T) {
		tests := []struct {
			op   string
			path string
			want string
		}{
			{"pull", "/repo", "source pull for /repo: remote rejected"},
			{"status", "/repo", "source status for /repo: remote rejected"},
			{"link", "/repo", "source link for /repo: remote rejected"},
			{"unlink", "/repo", "source unlink for /repo: remote rejected"},
		}

		for _, tt := range tests {
			t.Run(tt.op, func(t *testing.T) {
				err := NewSourceError(tt.op, tt.path, baseErr)
				if got := err.Error(); got != tt.want {
					t.Errorf("Error() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestPackageError(t *testing.T) {
	baseErr := errors.New("not found")
	pkgErr := NewPackageError("custom-nodes", "install", baseErr)

	t.Run("Error message format", func(t *testing.T) {
		expected := "package install custom-nodes: not found"
		if got := pkgErr.Error(); got != expected {
			t.Errorf("Error() = %q, want %q", got, expected)
		}
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		if got := pkgErr.Unwrap(); got != baseErr {
			t.Errorf("Unwrap() = %v, want %v", got, baseErr)
		}
	})

	t.Run("Constructor sets fields correctly", func(t *testing.T) {
		if pkgErr.Package != "custom-nodes" {
			t.Errorf("Package = %q, want %q", pkgErr.Package, "custom-nodes")
		}
		if pkgErr.Op != "install" {
			t.Errorf("Op = %q, want %q", pkgErr.Op, "install")
		}
		if pkgErr.Wrapped != baseErr {
			t.Errorf("Wrapped = %v, want %v", pkgErr.Wrapped, baseErr)
		}
	})

	t.Run("Different operations", func(t *testing.T) {
		tests := []struct {
			op   string
			want string
		}{
			{"uninstall", "package uninstall custom-nodes: not found"},
			{"publish", "package publish custom-nodes: not found"},
			{"search", "package search custom-nodes: not found"},
		}

		for _, tt := range tests {
			t.Run(tt.op, func(t *testing.T) {
				err := NewPackageError("custom-nodes", tt.op, baseErr)
				if got := err.Error(); got != tt.want {
					t.Errorf("Error() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestErrorsImplementErrorInterface(t *testing.T) {
	baseErr := errors.New("base error")

	// Verify all custom errors implement the error interface
	var _ error = &SSHError{}
	var _ error = &InstallError{}
	var _ error = &ConfigError{}
	var _ error = &SourceError{}
	var _ error = &PackageError{}

	// Verify errors.Is and errors.As work correctly with wrapped errors
	t.Run("errors.Is works with wrapped errors", func(t *testing.T) {
		sshErr := NewSSHError("host", "connect", baseErr)
		if !errors.Is(sshErr, baseErr) {
			t.Error("errors.Is should find the wrapped error")
		}
	})

	t.Run("errors.As works with custom error types", func(t *testing.T) {
		sshErr := NewSSHError("host", "connect", baseErr)
		var target *SSHError
		if !errors.As(sshErr, &target) {
			t.Error("errors.As should match the custom error type")
		}
		if target.Host != "host" {
			t.Errorf("errors.As target.Host = %q, want %q", target.Host, "host")
		}
	})
}
