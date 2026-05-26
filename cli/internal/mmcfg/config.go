package mmcfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Install InstallConfig  `yaml:"install,omitempty"`
	Agents  []AgentConfig  `yaml:"agents,omitempty"`
	Daemons []DaemonConfig `yaml:"daemons,omitempty"`
}

type ServerConfig struct {
	URL      string `yaml:"url"`
	Token    string `yaml:"token,omitempty"`
	Username string `yaml:"username,omitempty"`
	TeamID   string `yaml:"team_id,omitempty"`
	TeamName string `yaml:"team_name,omitempty"`
}

type InstallConfig struct {
	DataDir string `yaml:"data_dir,omitempty"`
	BinPath string `yaml:"bin_path,omitempty"`
	Running bool   `yaml:"running,omitempty"`
}

type AgentConfig struct {
	Name     string   `yaml:"name"`
	UserID   string   `yaml:"user_id"`
	Token    string   `yaml:"token"`
	Channels []string `yaml:"channels,omitempty"`
	Model    string   `yaml:"model,omitempty"`
	Status   string   `yaml:"status"`
	PID      int      `yaml:"pid,omitempty"`
	LogFile  string   `yaml:"log_file,omitempty"`
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
		Server: ServerConfig{URL: "http://localhost:8065"},
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

func (c *Config) AddAgent(a AgentConfig) {
	for i, existing := range c.Agents {
		if existing.Name == a.Name {
			c.Agents[i] = a
			return
		}
	}
	c.Agents = append(c.Agents, a)
}

func (c *Config) RemoveAgent(name string) {
	for i, a := range c.Agents {
		if a.Name == name {
			c.Agents = append(c.Agents[:i], c.Agents[i+1:]...)
			return
		}
	}
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
