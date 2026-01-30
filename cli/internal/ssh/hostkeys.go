package ssh

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyManager manages SSH host key verification and known_hosts file
type HostKeyManager struct {
	knownHostsPath string
	callback       ssh.HostKeyCallback
	interactive    bool
}

// NewHostKeyManager creates a new host key manager
func NewHostKeyManager(interactive bool) (*HostKeyManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(knownHostsPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Create known_hosts if it doesn't exist
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		f, err := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to create known_hosts: %w", err)
		}
		f.Close()
	}

	// Load known_hosts
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts: %w", err)
	}

	return &HostKeyManager{
		knownHostsPath: knownHostsPath,
		callback:       callback,
		interactive:    interactive,
	}, nil
}

// HostKeyCallback returns an ssh.HostKeyCallback that verifies host keys
func (m *HostKeyManager) HostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Try the standard callback first
		err := m.callback(hostname, remote, key)
		if err == nil {
			return nil
		}

		// Check if this is a "key not found" error
		var keyErr *knownhosts.KeyError
		if !strings.Contains(err.Error(), "knownhosts:") {
			// Some other error - reject
			return err
		}

		// Parse the error to see if it's unknown host or key mismatch
		keyErr, ok := err.(*knownhosts.KeyError)
		if !ok {
			return err
		}

		// If we have mismatched keys, this is serious - reject
		if len(keyErr.Want) > 0 {
			return fmt.Errorf("WARNING: HOST KEY VERIFICATION FAILED!\n"+
				"The host key for %s has CHANGED!\n"+
				"This could indicate a man-in-the-middle attack.\n"+
				"Expected fingerprint(s):\n%s\n"+
				"Got fingerprint: %s\n"+
				"If you are certain this is correct, remove the old key from %s",
				hostname,
				formatFingerprints(keyErr.Want),
				formatFingerprint(key),
				m.knownHostsPath)
		}

		// Unknown host - prompt user if interactive
		if m.interactive {
			return m.promptAddHost(hostname, remote, key)
		}

		// Non-interactive and unknown host - reject
		return fmt.Errorf("host key verification failed: %s not found in known_hosts (use --insecure to skip verification)", hostname)
	}
}

// promptAddHost prompts the user to accept an unknown host key
func (m *HostKeyManager) promptAddHost(hostname string, remote net.Addr, key ssh.PublicKey) error {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("WARNING: Unknown SSH Host Key")
	fmt.Println("========================================")
	fmt.Printf("Host: %s (%s)\n", hostname, remote.String())
	fmt.Printf("Key Type: %s\n", key.Type())
	fmt.Println()
	fmt.Println("Fingerprints:")
	fmt.Printf("  SHA256: %s\n", formatSHA256Fingerprint(key))
	fmt.Printf("  MD5:    %s\n", formatMD5Fingerprint(key))
	fmt.Println()
	fmt.Println("The authenticity of this host cannot be verified.")
	fmt.Print("Are you sure you want to continue connecting? (yes/no): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read user input")
	}

	response := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if response != "yes" {
		return fmt.Errorf("host key verification failed: user rejected host key")
	}

	// Add the key to known_hosts
	if err := m.addHostKey(hostname, key); err != nil {
		return fmt.Errorf("failed to add host key: %w", err)
	}

	fmt.Println()
	fmt.Printf("Host key added to %s\n", m.knownHostsPath)
	fmt.Println("========================================")
	fmt.Println()

	return nil
}

// addHostKey adds a host key to the known_hosts file
func (m *HostKeyManager) addHostKey(hostname string, key ssh.PublicKey) error {
	f, err := os.OpenFile(m.knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Format: hostname algorithm base64-encoded-key
	line := knownhosts.Line([]string{hostname}, key)
	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}

	return nil
}

// formatFingerprint formats a single key fingerprint
func formatFingerprint(key ssh.PublicKey) string {
	return fmt.Sprintf("%s %s", key.Type(), formatSHA256Fingerprint(key))
}

// formatFingerprints formats multiple key fingerprints
func formatFingerprints(keys []knownhosts.KnownKey) string {
	var lines []string
	for _, kk := range keys {
		lines = append(lines, "  "+formatFingerprint(kk.Key))
	}
	return strings.Join(lines, "\n")
}

// formatSHA256Fingerprint returns the SHA256 fingerprint of a public key
func formatSHA256Fingerprint(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
}

// formatMD5Fingerprint returns the MD5 fingerprint of a public key (legacy format)
func formatMD5Fingerprint(key ssh.PublicKey) string {
	hash := md5.Sum(key.Marshal())
	var parts []string
	for _, b := range hash {
		parts = append(parts, fmt.Sprintf("%02x", b))
	}
	return strings.Join(parts, ":")
}

// InsecureIgnoreHostKeyWithWarning returns a callback that ignores host keys but prints a warning
func InsecureIgnoreHostKeyWithWarning() ssh.HostKeyCallback {
	warned := false
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if !warned {
			fmt.Println()
			fmt.Println("========================================")
			fmt.Println("WARNING: INSECURE SSH CONNECTION")
			fmt.Println("========================================")
			fmt.Println("Host key verification is DISABLED.")
			fmt.Println("This connection is vulnerable to man-in-the-middle attacks.")
			fmt.Println()
			fmt.Printf("Connecting to: %s (%s)\n", hostname, remote.String())
			fmt.Printf("Key Type: %s\n", key.Type())
			fmt.Printf("Fingerprint: %s\n", formatSHA256Fingerprint(key))
			fmt.Println()
			fmt.Println("For secure connections, remove the --insecure flag.")
			fmt.Println("========================================")
			fmt.Println()
			warned = true
		}
		return nil
	}
}
