package ssh

import (
	"sync"
)

// Pool manages reusable SSH connections
type Pool struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

var (
	defaultPool *Pool
	poolOnce    sync.Once
)

// GetPool returns the global connection pool
func GetPool() *Pool {
	poolOnce.Do(func() {
		defaultPool = &Pool{
			clients: make(map[string]*Client),
		}
	})
	return defaultPool
}

// poolKey generates a unique key for a connection
func poolKey(host, user string) string {
	return user + "@" + host
}

// Get retrieves or creates a connection for the given target
func (p *Pool) Get(host, user, keyPath string) (*Client, error) {
	return p.GetWithOptions(host, user, keyPath, ClientOptions{
		StrictHostKeyChecking: true,
		Interactive:           true,
	})
}

// GetWithOptions retrieves or creates a connection with custom options
func (p *Pool) GetWithOptions(host, user, keyPath string, opts ClientOptions) (*Client, error) {
	key := poolKey(host, user)

	// Check for existing connection
	p.mu.RLock()
	if client, ok := p.clients[key]; ok {
		p.mu.RUnlock()
		// Verify connection is still alive
		if _, err := client.RunCommand("echo 1"); err == nil {
			return client, nil
		}
		// Connection dead, remove it
		p.mu.Lock()
		delete(p.clients, key)
		p.mu.Unlock()
	} else {
		p.mu.RUnlock()
	}

	// Create new connection
	client, err := NewClientWithOptions(host, user, keyPath, opts)
	if err != nil {
		return nil, err
	}

	// Store in pool
	p.mu.Lock()
	p.clients[key] = client
	p.mu.Unlock()

	return client, nil
}

// Close closes all connections in the pool
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, client := range p.clients {
		client.Close()
		delete(p.clients, key)
	}
}

// CloseHost closes the connection for a specific host
func (p *Pool) CloseHost(host, user string) {
	key := poolKey(host, user)

	p.mu.Lock()
	defer p.mu.Unlock()

	if client, ok := p.clients[key]; ok {
		client.Close()
		delete(p.clients, key)
	}
}

// Size returns the number of active connections
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}
