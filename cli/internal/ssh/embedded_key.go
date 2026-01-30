package ssh

import (
	_ "embed"
	"strings"

	"golang.org/x/crypto/ssh"
)

// NOTE: This file embeds the anime internal SSH key at compile time.
// The key files must exist in keys/ when building, but keys/ is gitignored.
// To build: ensure keys/anime_internal exists, then `go build`

//go:embed keys/anime_internal
var embeddedPrivateKey []byte

//go:embed keys/anime_internal.pub
var embeddedPublicKey []byte

// GetEmbeddedSigner returns an ssh.Signer from the embedded private key
func GetEmbeddedSigner() (ssh.Signer, error) {
	return ssh.ParsePrivateKey(embeddedPrivateKey)
}

// GetEmbeddedPublicKey returns the embedded public key bytes
func GetEmbeddedPublicKey() []byte {
	return embeddedPublicKey
}

// GetEmbeddedPublicKeyString returns the embedded public key as a string
func GetEmbeddedPublicKeyString() string {
	return strings.TrimSpace(string(embeddedPublicKey))
}

// GetEmbeddedPrivateKey returns the raw embedded private key bytes
func GetEmbeddedPrivateKey() []byte {
	return embeddedPrivateKey
}

// NewClientWithEmbeddedKey creates an SSH client using the embedded key with strict host key checking
func NewClientWithEmbeddedKey(host, user string) (*Client, error) {
	return NewClientWithEmbeddedKeyOptions(host, user, ClientOptions{
		StrictHostKeyChecking: true,
		Interactive:           true,
	})
}

// NewClientWithEmbeddedKeyInsecure creates an SSH client using the embedded key with host key checking disabled
func NewClientWithEmbeddedKeyInsecure(host, user string) (*Client, error) {
	return NewClientWithEmbeddedKeyOptions(host, user, ClientOptions{
		StrictHostKeyChecking: false,
		Interactive:           false,
	})
}

// NewClientWithEmbeddedKeyOptions creates an SSH client using the embedded key with custom options
func NewClientWithEmbeddedKeyOptions(host, user string, opts ClientOptions) (*Client, error) {
	signer, err := GetEmbeddedSigner()
	if err != nil {
		return nil, err
	}

	// Setup host key verification
	var hostKeyCallback ssh.HostKeyCallback
	if opts.StrictHostKeyChecking {
		// Use proper host key verification
		hostKeyManager, err := NewHostKeyManager(opts.Interactive)
		if err != nil {
			return nil, err
		}
		hostKeyCallback = hostKeyManager.HostKeyCallback()
	} else {
		// Insecure mode - show warning
		hostKeyCallback = InsecureIgnoreHostKeyWithWarning()
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		config: config,
		host:   host,
	}, nil
}
