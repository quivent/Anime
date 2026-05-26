package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/ssh"
)

// SSH global flags (registered by root command)
var (
	SSHInsecure              bool
	SSHStrictHostKeyChecking = true
	SSHNonInteractive        bool
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

// sshIdentityArgs returns -i flags for all private keys found in ~/.ssh/
func sshIdentityArgs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil
	}

	var args []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".pub") || name == "known_hosts" ||
			name == "config" || name == "authorized_keys" ||
			strings.HasPrefix(name, "known_hosts") || e.IsDir() {
			continue
		}
		path := filepath.Join(sshDir, name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		buf := make([]byte, 5)
		n, _ := f.Read(buf)
		f.Close()
		if n >= 5 && string(buf) == "-----" {
			args = append(args, "-i", path)
		}
	}
	return args
}

// sshConfigInfo holds parsed SSH config data for a host
type sshConfigInfo struct {
	IsAlias       bool   // hostname differs from input name
	RemoteCommand string // remotecommand value from config
	NeedsWSL      bool   // remotecommand contains "wsl"
}

// getSSHConfigInfo parses ssh -G output for a host name
func getSSHConfigInfo(name string) sshConfigInfo {
	cmd := exec.Command("ssh", "-G", name)
	output, err := cmd.Output()
	if err != nil {
		return sshConfigInfo{}
	}
	var info sshConfigInfo
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "hostname ") {
			hostname := strings.TrimPrefix(line, "hostname ")
			info.IsAlias = hostname != name
		}
		if strings.HasPrefix(line, "remotecommand ") {
			info.RemoteCommand = strings.TrimPrefix(line, "remotecommand ")
			info.NeedsWSL = strings.Contains(info.RemoteCommand, "wsl")
		}
	}
	return info
}

// isSSHConfigAlias returns true if the given name resolves to a different hostname via ssh -G
func isSSHConfigAlias(name string) bool {
	return getSSHConfigInfo(name).IsAlias
}

// sshCommand builds an exec.Command for ssh with all identity keys
func sshCommand(args ...string) *exec.Cmd {
	fullArgs := sshIdentityArgs()
	fullArgs = append(fullArgs, args...)
	return exec.Command("ssh", fullArgs...)
}

// sshCommandNonInteractive builds an SSH command safe for non-interactive use.
// (disables TTY and overrides RemoteCommand from config)
// If the target host has a WSL RemoteCommand, routes the command through "wsl bash -s"
// with the script piped via stdin to avoid Windows cmd.exe quoting issues.
func sshCommandNonInteractive(args ...string) *exec.Cmd {
	fullArgs := sshIdentityArgs()
	fullArgs = append(fullArgs, "-T", "-o", "RequestTTY=no", "-o", "RemoteCommand=none")

	// Find target by skipping -flag and -o value pairs
	var target string
	skipNext := false
	for _, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "-o" || a == "-i" || a == "-p" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		if target == "" {
			target = a
		}
	}
	needsWSL := target != "" && getSSHConfigInfo(target).NeedsWSL

	if needsWSL && len(args) >= 2 {
		// Replace the remote command with "wsl bash -s" and pipe the script via stdin
		lastIdx := len(args) - 1
		remoteCmd := args[lastIdx]
		wrapped := make([]string, lastIdx)
		copy(wrapped, args[:lastIdx])
		wrapped = append(wrapped, "wsl bash -s")
		fullArgs = append(fullArgs, wrapped...)
		cmd := exec.Command("ssh", fullArgs...)
		cmd.Stdin = strings.NewReader(remoteCmd)
		return cmd
	}

	fullArgs = append(fullArgs, args...)
	return exec.Command("ssh", fullArgs...)
}

// shellescape wraps a string in single quotes for shell safety
func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// sshRsyncFlag returns the -e flag value for rsync with all identity keys
func sshRsyncFlag() string {
	args := sshIdentityArgs()
	args = append(args, "-T", "-o", "RequestTTY=no", "-o", "RemoteCommand=none")
	return "ssh " + strings.Join(args, " ")
}
