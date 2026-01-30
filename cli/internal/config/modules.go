package config

import (
	_ "embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed modules/system.yaml
var systemYAML string

//go:embed modules/llm-frontier.yaml
var llmFrontierYAML string

//go:embed modules/llm-large.yaml
var llmLargeYAML string

//go:embed modules/llm-medium.yaml
var llmMediumYAML string

//go:embed modules/llm-small.yaml
var llmSmallYAML string

//go:embed modules/image.yaml
var imageYAML string

//go:embed modules/video.yaml
var videoYAML string

//go:embed modules/tools.yaml
var toolsYAML string

// ModuleFile represents a YAML file containing modules
type ModuleFile struct {
	Modules []ModuleYAML `yaml:"modules"`
}

// ModuleYAML represents a module in YAML format
type ModuleYAML struct {
	ID          string        `yaml:"id"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	TimeMinutes int           `yaml:"time_minutes"`
	Script      string        `yaml:"script"`
	Category    string        `yaml:"category"`
	Size        string        `yaml:"size"`
	Dependencies []string     `yaml:"dependencies,omitempty"`
	Cluster     *ClusterYAML  `yaml:"cluster,omitempty"`
}

// ClusterYAML represents cluster configuration in YAML format
type ClusterYAML struct {
	MinGPUs         int    `yaml:"min_gpus"`
	RecommendedGPUs int    `yaml:"recommended_gpus"`
	MinVRAMPerGPU   int    `yaml:"min_vram_per_gpu"`
	TotalVRAM       int    `yaml:"total_vram"`
	Parallelism     string `yaml:"parallelism"`
	InferenceEngine string `yaml:"inference_engine"`
	Notes           string `yaml:"notes"`
}

// LoadModules loads all module definitions from embedded YAML files
func LoadModules() ([]Module, error) {
	var allModules []Module

	// List of all embedded YAML files
	yamlFiles := map[string]string{
		"system":       systemYAML,
		"llm-frontier": llmFrontierYAML,
		"llm-large":    llmLargeYAML,
		"llm-medium":   llmMediumYAML,
		"llm-small":    llmSmallYAML,
		"image":        imageYAML,
		"video":        videoYAML,
		"tools":        toolsYAML,
	}

	// Parse each YAML file
	for name, yamlContent := range yamlFiles {
		var moduleFile ModuleFile
		if err := yaml.Unmarshal([]byte(yamlContent), &moduleFile); err != nil {
			return nil, fmt.Errorf("failed to parse %s.yaml: %w", name, err)
		}

		// Convert YAML modules to Module structs
		for _, m := range moduleFile.Modules {
			module := Module{
				ID:           m.ID,
				Name:         m.Name,
				Description:  m.Description,
				TimeMinutes:  m.TimeMinutes,
				Script:       m.Script,
				Category:     m.Category,
				Size:         m.Size,
				Dependencies: m.Dependencies,
			}

			// Convert cluster config if present
			if m.Cluster != nil {
				module.Cluster = &ClusterConfig{
					MinGPUs:         m.Cluster.MinGPUs,
					RecommendedGPUs: m.Cluster.RecommendedGPUs,
					MinVRAMPerGPU:   m.Cluster.MinVRAMPerGPU,
					TotalVRAM:       m.Cluster.TotalVRAM,
					Parallelism:     m.Cluster.Parallelism,
					InferenceEngine: m.Cluster.InferenceEngine,
					Notes:           m.Cluster.Notes,
				}
			}

			allModules = append(allModules, module)
		}
	}

	return allModules, nil
}

// GetModule returns a module by ID
func GetModule(id string) (*Module, error) {
	for i := range AvailableModules {
		if AvailableModules[i].ID == id {
			return &AvailableModules[i], nil
		}
	}

	return nil, fmt.Errorf("module %s not found", id)
}

// init loads modules from embedded YAML files into AvailableModules
func init() {
	modules, err := LoadModules()
	if err != nil {
		// In production, we want to panic if modules can't be loaded
		// since the CLI is unusable without them
		panic(fmt.Sprintf("failed to load modules: %v", err))
	}
	AvailableModules = modules
}

// GetModulesByCategoryName returns modules filtered by category name
// This is different from the existing GetModulesByCategory which returns a map
func GetModulesByCategoryName(category string) []Module {
	var result []Module
	for _, m := range AvailableModules {
		if strings.EqualFold(m.Category, category) {
			result = append(result, m)
		}
	}

	return result
}
