package matrixcfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Homeserver HomeserverConfig `yaml:"homeserver"`
	Synapse    SynapseConfig    `yaml:"synapse"`
	Agents     []AgentConfig    `yaml:"agents,omitempty"`
	Daemons    []DaemonConfig   `yaml:"daemons,omitempty"`
}

type HomeserverConfig struct {
	URL        string `yaml:"url"`
	Domain     string `yaml:"domain"`
	AdminToken string `yaml:"admin_token,omitempty"`
	AdminUser  string `yaml:"admin_user,omitempty"`
}

type SynapseConfig struct {
	DataDir      string `yaml:"data_dir"`
	SharedSecret string `yaml:"shared_secret,omitempty"`
	Running      bool   `yaml:"running"`
}

type AgentConfig struct {
	Name        string   `yaml:"name"`
	UserID      string   `yaml:"user_id"`
	AccessToken string   `yaml:"access_token"`
	Rooms       []string `yaml:"rooms,omitempty"`
	Model       string   `yaml:"model,omitempty"`
	Status      string   `yaml:"status"`
	PID         int      `yaml:"pid,omitempty"`
	LogFile     string   `yaml:"log_file,omitempty"`
}

type DaemonConfig struct {
	Name      string `yaml:"name"`
	PID       int    `yaml:"pid"`
	Status    string `yaml:"status"`
	StartedAt string `yaml:"started_at,omitempty"`
	Type      string `yaml:"type"`
	LogFile   string `yaml:"log_file,omitempty"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".matrix")
}

func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

func Load() (*Config, error) {
	cfg := &Config{
		Homeserver: HomeserverConfig{
			URL:    "http://localhost:8008",
			Domain: "localhost",
		},
		Synapse: SynapseConfig{
			DataDir: filepath.Join(Dir(), "data"),
		},
	}

	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(Path(), data, 0600)
}

func (c *Config) AddAgent(agent AgentConfig) {
	for i, a := range c.Agents {
		if a.Name == agent.Name {
			c.Agents[i] = agent
			return
		}
	}
	c.Agents = append(c.Agents, agent)
}

func (c *Config) RemoveAgent(name string) error {
	for i, a := range c.Agents {
		if a.Name == name {
			c.Agents = append(c.Agents[:i], c.Agents[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("agent %q not found", name)
}

func (c *Config) GetAgent(name string) *AgentConfig {
	for i, a := range c.Agents {
		if a.Name == name {
			return &c.Agents[i]
		}
	}
	return nil
}

func (c *Config) AddDaemon(d DaemonConfig) {
	for i, existing := range c.Daemons {
		if existing.Name == d.Name {
			c.Daemons[i] = d
			return
		}
	}
	c.Daemons = append(c.Daemons, d)
}

func (c *Config) RemoveDaemon(name string) {
	for i, d := range c.Daemons {
		if d.Name == name {
			c.Daemons = append(c.Daemons[:i], c.Daemons[i+1:]...)
			return
		}
	}
}
