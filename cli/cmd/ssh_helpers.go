package cmd

import (
	"github.com/joshkornreich/anime/internal/ssh"
)

// GetSSHClientOptions returns SSH client options based on global flags
func GetSSHClientOptions() ssh.ClientOptions {
	// If --insecure is set, disable host key checking
	if SSHInsecure {
		return ssh.ClientOptions{
			StrictHostKeyChecking: false,
			Interactive:           false,
		}
	}

	// Use strict host key checking by default
	// Non-interactive mode disables prompts for unknown hosts
	return ssh.ClientOptions{
		StrictHostKeyChecking: SSHStrictHostKeyChecking,
		Interactive:           !SSHNonInteractive,
	}
}

// NewSSHClient creates an SSH client using global flags for security options
func NewSSHClient(host, user, keyPath string) (*ssh.Client, error) {
	opts := GetSSHClientOptions()
	return ssh.NewClientWithOptions(host, user, keyPath, opts)
}

// NewSSHClientWithEmbeddedKey creates an SSH client using the embedded key and global flags
func NewSSHClientWithEmbeddedKey(host, user string) (*ssh.Client, error) {
	opts := GetSSHClientOptions()
	return ssh.NewClientWithEmbeddedKeyOptions(host, user, opts)
}
