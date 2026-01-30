package models

// ModelSpec represents a complete model specification
type ModelSpec struct {
	// Identity
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Classification
	Type     string `yaml:"type"`
	Category string `yaml:"category"`
	Provider string `yaml:"provider"`

	// Parameters
	TotalParams  string `yaml:"total_params,omitempty"`
	ActiveParams string `yaml:"active_params,omitempty"`
	Architecture string `yaml:"architecture,omitempty"`

	// Size & Requirements
	Size string `yaml:"size"`
	VRAM string `yaml:"vram"`

	// Model IDs for different providers
	HuggingFaceID string `yaml:"huggingface_id"`
	OllamaID      string `yaml:"ollama_id,omitempty"`
	VLLMModel     string `yaml:"vllm_model,omitempty"`

	// Metadata
	ReleaseDate string   `yaml:"release_date,omitempty"`
	UseCases    []string `yaml:"use_cases,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	HFLink      string   `yaml:"hf_link,omitempty"`
	License     string   `yaml:"license,omitempty"`
}

// ModelFile represents the structure of a YAML model file
type ModelFile struct {
	Models []ModelSpec `yaml:"models"`
}

// ShortcutsFile represents the structure of the shortcuts YAML
type ShortcutsFile struct {
	Aliases map[string]string `yaml:"aliases"`
}

// Model Type constants
const (
	TypeLLM        = "LLM"
	TypeCoding     = "Coding"
	TypeMultimodal = "Multimodal"
	TypeImage      = "Image"
	TypeVideo      = "Video"
	TypeEnhance    = "Enhance"
	TypeControl    = "Control"
)

// Model Provider constants
const (
	ProviderMeta        = "Meta"
	ProviderDeepSeek    = "DeepSeek"
	ProviderQwen        = "Qwen"
	ProviderAlibaba     = "Alibaba"
	ProviderMistral     = "Mistral"
	ProviderMicrosoft   = "Microsoft"
	ProviderGoogle      = "Google"
	ProviderYi          = "01.AI"
	ProviderCohere      = "Cohere"
	ProviderNVIDIA      = "NVIDIA"
	ProviderBigCode     = "BigCode"
	ProviderStability   = "Stability"
	ProviderBlackForest = "Black Forest Labs"
)

// Architecture constants
const (
	ArchDense = "Dense"
	ArchMoE   = "MoE"
)
