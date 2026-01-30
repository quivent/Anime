package models

import (
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ModelRegistry holds all loaded models with various indexes
type ModelRegistry struct {
	// All models indexed by ID
	ByID map[string]*ModelSpec

	// Models grouped by type
	ByType map[string][]*ModelSpec

	// Models grouped by provider
	ByProvider map[string][]*ModelSpec

	// Shortcuts mapping friendly names to HuggingFace IDs
	Shortcuts map[string]string

	// All models as a slice
	All []*ModelSpec
}

// yamlModels is the structure for parsing model YAML files
type yamlModels struct {
	Models []ModelSpec `yaml:"models"`
}

// yamlShortcuts is the structure for parsing shortcuts.yaml
type yamlShortcuts struct {
	Aliases map[string]string `yaml:"aliases"`
}

var (
	registry     *ModelRegistry
	registryOnce sync.Once
	registryErr  error
)

// GetRegistry returns the singleton model registry
func GetRegistry() (*ModelRegistry, error) {
	registryOnce.Do(func() {
		registry, registryErr = loadRegistry()
	})
	return registry, registryErr
}

// loadRegistry loads all model YAML files and builds the registry
func loadRegistry() (*ModelRegistry, error) {
	reg := &ModelRegistry{
		ByID:       make(map[string]*ModelSpec),
		ByType:     make(map[string][]*ModelSpec),
		ByProvider: make(map[string][]*ModelSpec),
		Shortcuts:  make(map[string]string),
		All:        make([]*ModelSpec, 0),
	}

	// Load each model file
	modelFiles := []string{"llm.yaml", "coding.yaml", "vision.yaml", "media.yaml"}
	for _, filename := range modelFiles {
		if err := loadModelFile(reg, filename); err != nil {
			return nil, fmt.Errorf("loading %s: %w", filename, err)
		}
	}

	// Load shortcuts
	if err := loadShortcuts(reg); err != nil {
		return nil, fmt.Errorf("loading shortcuts: %w", err)
	}

	// Generate auto-shortcuts from model IDs
	generateAutoShortcuts(reg)

	return reg, nil
}

// loadModelFile loads a single YAML file and adds models to registry
func loadModelFile(reg *ModelRegistry, filename string) error {
	data, err := ModelFiles.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading embedded file: %w", err)
	}

	var parsed yamlModels
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	for i := range parsed.Models {
		model := &parsed.Models[i]

		// Skip if already exists (shouldn't happen)
		if _, exists := reg.ByID[model.ID]; exists {
			continue
		}

		// Add to main index
		reg.ByID[model.ID] = model
		reg.All = append(reg.All, model)

		// Add to type index
		if model.Type != "" {
			reg.ByType[model.Type] = append(reg.ByType[model.Type], model)
		}

		// Add to provider index
		if model.Provider != "" {
			reg.ByProvider[model.Provider] = append(reg.ByProvider[model.Provider], model)
		}
	}

	return nil
}

// loadShortcuts loads the shortcuts.yaml file
func loadShortcuts(reg *ModelRegistry) error {
	data, err := ModelFiles.ReadFile("shortcuts.yaml")
	if err != nil {
		return fmt.Errorf("reading shortcuts file: %w", err)
	}

	var parsed yamlShortcuts
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("parsing shortcuts YAML: %w", err)
	}

	for alias, target := range parsed.Aliases {
		reg.Shortcuts[alias] = target
	}

	return nil
}

// generateAutoShortcuts creates shortcuts from model IDs
// e.g., "meta-llama/Llama-3.3-70B-Instruct" -> "llama-3.3-70b-instruct"
func generateAutoShortcuts(reg *ModelRegistry) {
	for _, model := range reg.All {
		if model.HuggingFaceID == "" {
			continue
		}

		// Extract model name from HuggingFace ID (after the /)
		parts := strings.SplitN(model.HuggingFaceID, "/", 2)
		if len(parts) != 2 {
			continue
		}

		// Create lowercase shortcut from model name
		shortcut := strings.ToLower(parts[1])

		// Don't overwrite explicit shortcuts
		if _, exists := reg.Shortcuts[shortcut]; !exists {
			reg.Shortcuts[shortcut] = model.HuggingFaceID
		}

		// Also add the model ID as a shortcut if different
		if model.ID != shortcut {
			if _, exists := reg.Shortcuts[model.ID]; !exists {
				reg.Shortcuts[model.ID] = model.HuggingFaceID
			}
		}
	}
}

// ResolveModel resolves a model identifier to a full HuggingFace ID
// Accepts: shortcut name, model ID, or full HuggingFace ID
func (r *ModelRegistry) ResolveModel(identifier string) (string, bool) {
	// Check if it's already a full HuggingFace ID
	if strings.Contains(identifier, "/") {
		return identifier, true
	}

	// Check shortcuts (includes auto-generated ones)
	if hfID, ok := r.Shortcuts[identifier]; ok {
		return hfID, true
	}

	// Check lowercase version
	lower := strings.ToLower(identifier)
	if hfID, ok := r.Shortcuts[lower]; ok {
		return hfID, true
	}

	// Check by model ID
	if model, ok := r.ByID[identifier]; ok && model.HuggingFaceID != "" {
		return model.HuggingFaceID, true
	}

	return "", false
}

// GetModel returns a model by ID
func (r *ModelRegistry) GetModel(id string) (*ModelSpec, bool) {
	model, ok := r.ByID[id]
	return model, ok
}

// GetModelsByType returns all models of a given type
func (r *ModelRegistry) GetModelsByType(modelType string) []*ModelSpec {
	return r.ByType[modelType]
}

// GetModelsByProvider returns all models from a given provider
func (r *ModelRegistry) GetModelsByProvider(provider string) []*ModelSpec {
	return r.ByProvider[provider]
}

// GetVLLMModels returns all models suitable for vLLM deployment
func (r *ModelRegistry) GetVLLMModels() []*ModelSpec {
	var result []*ModelSpec
	for _, model := range r.All {
		if model.VLLMModel != "" {
			result = append(result, model)
		}
	}
	return result
}

// GetOllamaModels returns all models with Ollama IDs
func (r *ModelRegistry) GetOllamaModels() []*ModelSpec {
	var result []*ModelSpec
	for _, model := range r.All {
		if model.OllamaID != "" {
			result = append(result, model)
		}
	}
	return result
}

// SearchModels searches for models matching a query string
func (r *ModelRegistry) SearchModels(query string) []*ModelSpec {
	query = strings.ToLower(query)
	var results []*ModelSpec

	for _, model := range r.All {
		// Search in ID, name, description, provider
		if strings.Contains(strings.ToLower(model.ID), query) ||
			strings.Contains(strings.ToLower(model.Name), query) ||
			strings.Contains(strings.ToLower(model.Description), query) ||
			strings.Contains(strings.ToLower(model.Provider), query) {
			results = append(results, model)
		}
	}

	return results
}
