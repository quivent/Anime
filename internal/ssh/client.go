package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	client *ssh.Client
	config *ssh.ClientConfig
	host   string
}

func NewClient(host, user, keyPath string) (*Client, error) {
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
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Improve this
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
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
