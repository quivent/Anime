package cmd

import (
	"log"

	"github.com/joshkornreich/anime/internal/models"
)

// ============================================================================
// MODEL REGISTRY - Wrapper around embedded YAML model database
// ============================================================================
// This file provides backward-compatible access to the model registry.
// All model data is now stored in internal/models/*.yaml and embedded at compile time.
// Last updated: December 2025
// ============================================================================

// Type aliases for backward compatibility
type ModelSpec = models.ModelSpec

// Model Type constants - re-exported for backward compatibility
const (
	TypeLLM        = models.TypeLLM
	TypeCoding     = models.TypeCoding
	TypeMultimodal = models.TypeMultimodal
	TypeImage      = models.TypeImage
	TypeVideo      = models.TypeVideo
	TypeEnhance    = models.TypeEnhance
	TypeControl    = models.TypeControl
)

// Model Provider constants - re-exported for backward compatibility
const (
	ProviderMeta        = models.ProviderMeta
	ProviderDeepSeek    = models.ProviderDeepSeek
	ProviderQwen        = models.ProviderQwen
	ProviderMistral     = models.ProviderMistral
	ProviderMicrosoft   = models.ProviderMicrosoft
	ProviderGoogle      = models.ProviderGoogle
	ProviderYi          = models.ProviderYi
	ProviderCohere      = models.ProviderCohere
	ProviderNVIDIA      = models.ProviderNVIDIA
	ProviderBigCode     = models.ProviderBigCode
	ProviderStability   = models.ProviderStability
	ProviderBlackForest = models.ProviderBlackForest
)

// Architecture constants - re-exported for backward compatibility
const (
	ArchDense = models.ArchDense
	ArchMoE   = models.ArchMoE
)

// getRegistry returns the model registry, logging any errors
func getRegistry() *models.ModelRegistry {
	reg, err := models.GetRegistry()
	if err != nil {
		log.Printf("Warning: failed to load model registry: %v", err)
		return nil
	}
	return reg
}

// ============================================================================
// PUBLIC API - All functions use the embedded YAML registry
// ============================================================================

// GetModelShortcuts returns the mapping of short names to full HuggingFace IDs
func GetModelShortcuts() map[string]string {
	reg := getRegistry()
	if reg == nil {
		return make(map[string]string)
	}
	return reg.Shortcuts
}

// GetAllModels returns all models from the registry
func GetAllModels() []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	// Convert []*ModelSpec to []ModelSpec for backward compatibility
	result := make([]ModelSpec, len(reg.All))
	for i, m := range reg.All {
		result[i] = *m
	}
	return result
}

// GetModelsByType returns models filtered by type
func GetModelsByType(modelType string) []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	models := reg.GetModelsByType(modelType)
	result := make([]ModelSpec, len(models))
	for i, m := range models {
		result[i] = *m
	}
	return result
}

// GetModelsByProvider returns models filtered by provider
func GetModelsByProvider(provider string) []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	models := reg.GetModelsByProvider(provider)
	result := make([]ModelSpec, len(models))
	for i, m := range models {
		result[i] = *m
	}
	return result
}

// GetModelByID returns a model by its ID
func GetModelByID(id string) *ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	model, ok := reg.GetModel(id)
	if !ok {
		return nil
	}
	return model
}

// GetVLLMModels returns all models that have vLLM support
func GetVLLMModels() []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	models := reg.GetVLLMModels()
	result := make([]ModelSpec, len(models))
	for i, m := range models {
		result[i] = *m
	}
	return result
}

// GetOllamaModels returns all models that have Ollama support
func GetOllamaModels() []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	models := reg.GetOllamaModels()
	result := make([]ModelSpec, len(models))
	for i, m := range models {
		result[i] = *m
	}
	return result
}

// ResolveModelID resolves a shortcut or ID to the full HuggingFace/vLLM model path
func ResolveModelID(input string) string {
	reg := getRegistry()
	if reg == nil {
		return input
	}
	if resolved, ok := reg.ResolveModel(input); ok {
		return resolved
	}
	return input
}

// SearchModels searches for models matching a query string
func SearchModels(query string) []ModelSpec {
	reg := getRegistry()
	if reg == nil {
		return nil
	}
	models := reg.SearchModels(query)
	result := make([]ModelSpec, len(models))
	for i, m := range models {
		result[i] = *m
	}
	return result
}
