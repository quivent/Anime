package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Client struct {
	client *ssh.Client
	config *ssh.ClientConfig
	host   string
}

// ClientOptions contains optional parameters for creating an SSH client
type ClientOptions struct {
	StrictHostKeyChecking bool
	Interactive           bool
}

// NewClient creates a new SSH client with the default options (strict host key checking enabled)
func NewClient(host, user, keyPath string) (*Client, error) {
	return NewClientWithOptions(host, user, keyPath, ClientOptions{
		StrictHostKeyChecking: true,
		Interactive:           true,
	})
}

// NewClientInsecure creates a new SSH client with host key checking disabled
// This function prints a warning about the security implications
func NewClientInsecure(host, user, keyPath string) (*Client, error) {
	return NewClientWithOptions(host, user, keyPath, ClientOptions{
		StrictHostKeyChecking: false,
		Interactive:           false,
	})
}

// NewClientWithOptions creates a new SSH client with custom options
func NewClientWithOptions(host, user, keyPath string, opts ClientOptions) (*Client, error) {
	var authMethods []ssh.AuthMethod

	// Try SSH agent first
	if agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		agentClient := agent.NewClient(agentConn)
		authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
	}

	// If key path specified, use it
	if keyPath != "" {
		// Expand home directory
		if strings.HasPrefix(keyPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			keyPath = filepath.Join(home, keyPath[2:])
		}

		// Read private key
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, errors.NewSSHKeyError(keyPath, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, errors.NewSSHKeyError(keyPath, err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if len(authMethods) == 0 {
		// No agent and no key path - try common key locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		// Try common SSH key locations
		possibleKeys := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		}

		for _, path := range possibleKeys {
			if _, err := os.Stat(path); err == nil {
				key, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					continue
				}
				authMethods = append(authMethods, ssh.PublicKeys(signer))
				break
			}
		}

		if len(authMethods) == 0 {
			return nil, errors.NewSSHAuthError(user, host, fmt.Errorf("no SSH authentication method available"))
		}
	}

	// Setup host key verification
	var hostKeyCallback ssh.HostKeyCallback
	if opts.StrictHostKeyChecking {
		// Use proper host key verification
		hostKeyManager, err := NewHostKeyManager(opts.Interactive)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize host key verification: %w", err)
		}
		hostKeyCallback = hostKeyManager.HostKeyCallback()
	} else {
		// Insecure mode - show warning
		hostKeyCallback = InsecureIgnoreHostKeyWithWarning()
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, errors.NewSSHConnectionError(host, err)
	}

	return &Client{
		client: client,
		config: config,
		host:   host,
	}, nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *Client) RunCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

func (c *Client) RunCommandWithProgress(cmd string, progress chan<- string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	// Stream output
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				progress <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				progress <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	return session.Wait()
}

func (c *Client) UploadFile(localPath, remotePath string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Read local file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}

	// Create remote file
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "C0644 %d %s\n", len(data), filepath.Base(remotePath))
		stdin.Write(data)
		fmt.Fprint(stdin, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return err
	}

	return nil
}

func (c *Client) UploadString(content, remotePath string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "C0644 %d %s\n", len(content), filepath.Base(remotePath))
		io.WriteString(stdin, content)
		fmt.Fprint(stdin, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return err
	}

	return nil
}

func (c *Client) FileExists(path string) (bool, error) {
	_, err := c.RunCommand(fmt.Sprintf("test -f %s && echo exists", path))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) MakeExecutable(path string) error {
	_, err := c.RunCommand(fmt.Sprintf("chmod +x %s", path))
	return err
}

// Host returns the hostname without port
func (c *Client) Host() string {
	// Strip port if present
	if strings.Contains(c.host, ":") {
		parts := strings.Split(c.host, ":")
		return parts[0]
	}
	return c.host
}
