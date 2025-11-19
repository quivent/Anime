package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers []Server `yaml:"servers"`
	APIKeys APIKeys  `yaml:"api_keys"`
}

type Server struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	SSHKey      string `yaml:"ssh_key"`
	CostPerHour float64 `yaml:"cost_per_hour"`
	Modules     []string `yaml:"modules,omitempty"`
}

type APIKeys struct {
	Anthropic   string `yaml:"anthropic,omitempty"`
	OpenAI      string `yaml:"openai,omitempty"`
	HuggingFace string `yaml:"huggingface,omitempty"`
	LambdaLabs  string `yaml:"lambda_labs,omitempty"`
}

type Module struct {
	ID          string
	Name        string
	Description string
	TimeMinutes int
	Dependencies []string
	Script      string
}

var AvailableModules = []Module{
	{
		ID:          "core",
		Name:        "Core System",
		Description: "CUDA, Python, Node.js, Docker",
		TimeMinutes: 5,
		Script:      "core",
	},
	{
		ID:          "pytorch",
		Name:        "PyTorch + AI Libraries",
		Description: "PyTorch, Transformers, Diffusers, xformers",
		TimeMinutes: 2,
		Dependencies: []string{"core"},
		Script:      "pytorch",
	},
	{
		ID:          "ollama",
		Name:        "Ollama Server",
		Description: "Ollama LLM server (no models)",
		TimeMinutes: 1,
		Dependencies: []string{"core"},
		Script:      "ollama",
	},
	{
		ID:          "models-small",
		Name:        "Small Models (7B)",
		Description: "Mistral, Llama 3.3 8B, Qwen 2.5 7B",
		TimeMinutes: 8,
		Dependencies: []string{"ollama"},
		Script:      "models-small",
	},
	{
		ID:          "models-medium",
		Name:        "Medium Models (14-34B)",
		Description: "Qwen 2.5 14B, Mixtral 8x7B, DeepSeek Coder",
		TimeMinutes: 25,
		Dependencies: []string{"ollama"},
		Script:      "models-medium",
	},
	{
		ID:          "models-large",
		Name:        "Large Models (70B+)",
		Description: "Llama 3.3 70B, Qwen 2.5 72B",
		TimeMinutes: 40,
		Dependencies: []string{"ollama"},
		Script:      "models-large",
	},
	{
		ID:          "comfyui",
		Name:        "ComfyUI",
		Description: "Stable Diffusion UI with Manager",
		TimeMinutes: 2,
		Dependencies: []string{"pytorch"},
		Script:      "comfyui",
	},
	{
		ID:          "claude",
		Name:        "Claude Code CLI",
		Description: "Anthropic Claude Code CLI",
		TimeMinutes: 1,
		Dependencies: []string{"core"},
		Script:      "claude",
	},
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anime", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Servers: []Server{},
			APIKeys: APIKeys{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) AddServer(server Server) {
	c.Servers = append(c.Servers, server)
}

func (c *Config) GetServer(name string) (*Server, error) {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i], nil
		}
	}
	return nil, fmt.Errorf("server %s not found", name)
}

func (c *Config) UpdateServer(name string, server Server) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers[i] = server
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

func (c *Config) DeleteServer(name string) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("server %s not found", name)
}

func EstimateCost(modules []string, costPerHour float64) float64 {
	totalMinutes := 0
	moduleMap := make(map[string]bool)

	// Add dependencies
	var addDeps func(string)
	addDeps = func(id string) {
		if moduleMap[id] {
			return
		}
		moduleMap[id] = true

		for _, mod := range AvailableModules {
			if mod.ID == id {
				for _, dep := range mod.Dependencies {
					addDeps(dep)
				}
				break
			}
		}
	}

	for _, id := range modules {
		addDeps(id)
	}

	// Calculate total time
	for _, mod := range AvailableModules {
		if moduleMap[mod.ID] {
			totalMinutes += mod.TimeMinutes
		}
	}

	return (float64(totalMinutes) / 60.0) * costPerHour
}

func GetModulesByID(ids []string) []Module {
	var result []Module
	for _, id := range ids {
		for _, mod := range AvailableModules {
			if mod.ID == id {
				result = append(result, mod)
				break
			}
		}
	}
	return result
}
