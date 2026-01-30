package coverage

import (
	"fmt"
	"time"
)

// ClusterArchitecture defines GPU cluster configurations
type ClusterArchitecture string

const (
	ArchH100_1x ClusterArchitecture = "h100-1x"
	ArchH100_2x ClusterArchitecture = "h100-2x"
	ArchH100_4x ClusterArchitecture = "h100-4x"
	ArchH100_8x ClusterArchitecture = "h100-8x"
	ArchGH200   ClusterArchitecture = "gh200"
	ArchB200_1x ClusterArchitecture = "b200-1x"
	ArchB200_2x ClusterArchitecture = "b200-2x"
	ArchB200_4x ClusterArchitecture = "b200-4x"
	ArchB200_8x ClusterArchitecture = "b200-8x"
)

// ClusterConfig defines the configuration for a GPU cluster
type ClusterConfig struct {
	Architecture     ClusterArchitecture `yaml:"architecture" json:"architecture"`
	GPUCount         int                 `yaml:"gpu_count" json:"gpu_count"`
	GPUModel         string              `yaml:"gpu_model" json:"gpu_model"`
	GPUMemoryGB      int                 `yaml:"gpu_memory_gb" json:"gpu_memory_gb"`
	TotalMemoryGB    int                 `yaml:"total_memory_gb" json:"total_memory_gb"`
	NVLinkEnabled    bool                `yaml:"nvlink_enabled" json:"nvlink_enabled"`
	CostPerHour      float64             `yaml:"cost_per_hour" json:"cost_per_hour"`
	ThroughputPerHr  int                 `yaml:"throughput_per_hr" json:"throughput_per_hr"` // screenplays/hour
	CostPerScreenplay float64            `yaml:"cost_per_screenplay" json:"cost_per_screenplay"`
}

// ModelConfig defines LLM model configuration
type ModelConfig struct {
	Name            string  `yaml:"name" json:"name"`
	Parameters      string  `yaml:"parameters" json:"parameters"`
	Precision       string  `yaml:"precision" json:"precision"`
	MaxContextLen   int     `yaml:"max_context_len" json:"max_context_len"`
	MemoryRequired  int     `yaml:"memory_required_gb" json:"memory_required_gb"`
	TokensPerSecond int     `yaml:"tokens_per_second" json:"tokens_per_second"`
}

// VLLMConfig defines vLLM server configuration
type VLLMConfig struct {
	Model                  string  `yaml:"model" json:"model"`
	DType                  string  `yaml:"dtype" json:"dtype"`
	TensorParallelSize     int     `yaml:"tensor_parallel_size" json:"tensor_parallel_size"`
	PipelineParallelSize   int     `yaml:"pipeline_parallel_size" json:"pipeline_parallel_size"`
	GPUMemoryUtilization   float64 `yaml:"gpu_memory_utilization" json:"gpu_memory_utilization"`
	MaxModelLen            int     `yaml:"max_model_len" json:"max_model_len"`
	MaxNumBatchedTokens    int     `yaml:"max_num_batched_tokens" json:"max_num_batched_tokens"`
	MaxNumSeqs             int     `yaml:"max_num_seqs" json:"max_num_seqs"`
	EnablePrefixCaching    bool    `yaml:"enable_prefix_caching" json:"enable_prefix_caching"`
	EnableChunkedPrefill   bool    `yaml:"enable_chunked_prefill" json:"enable_chunked_prefill"`
	KVCacheDType           string  `yaml:"kv_cache_dtype" json:"kv_cache_dtype"`
	Quantization           string  `yaml:"quantization,omitempty" json:"quantization,omitempty"`
}

// AnalysisConfig defines coverage analysis configuration
type AnalysisConfig struct {
	Dimensions          []string      `yaml:"dimensions" json:"dimensions"`
	QualityThreshold    float64       `yaml:"quality_threshold" json:"quality_threshold"`
	ConfidenceThreshold float64       `yaml:"confidence_threshold" json:"confidence_threshold"`
	Timeout             time.Duration `yaml:"timeout" json:"timeout"`
	IncludeExamples     bool          `yaml:"include_examples" json:"include_examples"`
	OutputFormat        string        `yaml:"output_format" json:"output_format"`
	Template            string        `yaml:"template" json:"template"`
}

// BenchmarkConfig defines benchmark configuration
type BenchmarkConfig struct {
	Suite       string `yaml:"suite" json:"suite"`
	Iterations  int    `yaml:"iterations" json:"iterations"`
	WarmupRuns  int    `yaml:"warmup_runs" json:"warmup_runs"`
	Concurrent  int    `yaml:"concurrent" json:"concurrent"`
	Duration    time.Duration `yaml:"duration" json:"duration"`
}

// CoverageConfig is the main configuration structure
type CoverageConfig struct {
	Cluster   ClusterConfig   `yaml:"cluster" json:"cluster"`
	Model     ModelConfig     `yaml:"model" json:"model"`
	VLLM      VLLMConfig      `yaml:"vllm" json:"vllm"`
	Analysis  AnalysisConfig  `yaml:"analysis" json:"analysis"`
	Benchmark BenchmarkConfig `yaml:"benchmark" json:"benchmark"`
}

// ClusterSpecs contains specifications for all supported cluster architectures
var ClusterSpecs = map[ClusterArchitecture]ClusterConfig{
	ArchH100_1x: {
		Architecture:     ArchH100_1x,
		GPUCount:         1,
		GPUModel:         "H100 SXM5",
		GPUMemoryGB:      80,
		TotalMemoryGB:    80,
		NVLinkEnabled:    false,
		CostPerHour:      4.50,
		ThroughputPerHr:  7,
		CostPerScreenplay: 0.64,
	},
	ArchH100_2x: {
		Architecture:     ArchH100_2x,
		GPUCount:         2,
		GPUModel:         "H100 SXM5",
		GPUMemoryGB:      80,
		TotalMemoryGB:    160,
		NVLinkEnabled:    true,
		CostPerHour:      9.00,
		ThroughputPerHr:  18,
		CostPerScreenplay: 0.50,
	},
	ArchH100_4x: {
		Architecture:     ArchH100_4x,
		GPUCount:         4,
		GPUModel:         "H100 SXM5",
		GPUMemoryGB:      80,
		TotalMemoryGB:    320,
		NVLinkEnabled:    true,
		CostPerHour:      18.00,
		ThroughputPerHr:  15,
		CostPerScreenplay: 1.20,
	},
	ArchH100_8x: {
		Architecture:     ArchH100_8x,
		GPUCount:         8,
		GPUModel:         "H100 SXM5",
		GPUMemoryGB:      80,
		TotalMemoryGB:    640,
		NVLinkEnabled:    true,
		CostPerHour:      36.00,
		ThroughputPerHr:  30,
		CostPerScreenplay: 1.20,
	},
	ArchGH200: {
		Architecture:     ArchGH200,
		GPUCount:         1,
		GPUModel:         "GH200 Grace Hopper",
		GPUMemoryGB:      96,
		TotalMemoryGB:    576, // 96GB GPU + 480GB unified
		NVLinkEnabled:    false,
		CostPerHour:      6.50,
		ThroughputPerHr:  12,
		CostPerScreenplay: 0.54,
	},
	ArchB200_1x: {
		Architecture:     ArchB200_1x,
		GPUCount:         1,
		GPUModel:         "B200 Blackwell",
		GPUMemoryGB:      192,
		TotalMemoryGB:    192,
		NVLinkEnabled:    false,
		CostPerHour:      8.00,
		ThroughputPerHr:  30,
		CostPerScreenplay: 0.27,
	},
	ArchB200_2x: {
		Architecture:     ArchB200_2x,
		GPUCount:         2,
		GPUModel:         "B200 Blackwell",
		GPUMemoryGB:      192,
		TotalMemoryGB:    384,
		NVLinkEnabled:    true,
		CostPerHour:      16.00,
		ThroughputPerHr:  55,
		CostPerScreenplay: 0.29,
	},
	ArchB200_4x: {
		Architecture:     ArchB200_4x,
		GPUCount:         4,
		GPUModel:         "B200 Blackwell",
		GPUMemoryGB:      192,
		TotalMemoryGB:    768,
		NVLinkEnabled:    true,
		CostPerHour:      32.00,
		ThroughputPerHr:  100,
		CostPerScreenplay: 0.32,
	},
	ArchB200_8x: {
		Architecture:     ArchB200_8x,
		GPUCount:         8,
		GPUModel:         "B200 Blackwell",
		GPUMemoryGB:      192,
		TotalMemoryGB:    1536,
		NVLinkEnabled:    true,
		CostPerHour:      64.00,
		ThroughputPerHr:  175,
		CostPerScreenplay: 0.37,
	},
}

// ModelSpecs contains specifications for supported models
var ModelSpecs = map[string]ModelConfig{
	"llama-70b-int4": {
		Name:           "Llama 3.1 70B",
		Parameters:     "70B",
		Precision:      "INT4",
		MaxContextLen:  16384,
		MemoryRequired: 40,
		TokensPerSecond: 1000,
	},
	"llama-70b-bf16": {
		Name:           "Llama 3.1 70B",
		Parameters:     "70B",
		Precision:      "BF16",
		MaxContextLen:  65536,
		MemoryRequired: 140,
		TokensPerSecond: 4000,
	},
	"llama-405b-bf16": {
		Name:           "Llama 3.1 405B",
		Parameters:     "405B",
		Precision:      "BF16",
		MaxContextLen:  32768,
		MemoryRequired: 800,
		TokensPerSecond: 2500,
	},
	"llama-405b-fp8": {
		Name:           "Llama 3.1 405B",
		Parameters:     "405B",
		Precision:      "FP8",
		MaxContextLen:  65536,
		MemoryRequired: 400,
		TokensPerSecond: 6000,
	},
	"llama-405b-fp4": {
		Name:           "Llama 3.1 405B",
		Parameters:     "405B",
		Precision:      "FP4",
		MaxContextLen:  16384,
		MemoryRequired: 200,
		TokensPerSecond: 12000,
	},
}

// DefaultAnalysisDimensions are the standard coverage analysis dimensions
var DefaultAnalysisDimensions = []string{
	"structure",
	"character",
	"dialogue",
	"theme",
	"tone",
	"concept",
	"marketability",
	"pacing",
	"craft",
}

// GetVLLMConfig generates vLLM configuration for a given cluster and model
func GetVLLMConfig(arch ClusterArchitecture, modelKey string) VLLMConfig {
	cluster := ClusterSpecs[arch]
	model := ModelSpecs[modelKey]

	config := VLLMConfig{
		Model:                "meta-llama/Llama-3.1-405B-Instruct",
		DType:                "bfloat16",
		TensorParallelSize:   cluster.GPUCount,
		PipelineParallelSize: 1,
		GPUMemoryUtilization: 0.92,
		MaxModelLen:          model.MaxContextLen,
		MaxNumBatchedTokens:  model.MaxContextLen * 2,
		MaxNumSeqs:           256,
		EnablePrefixCaching:  true,
		EnableChunkedPrefill: true,
		KVCacheDType:         "auto",
	}

	// Adjust for model precision
	switch model.Precision {
	case "FP8":
		config.Quantization = "fp8"
		config.KVCacheDType = "fp8"
	case "INT4":
		config.Quantization = "awq"
	case "FP4":
		config.Quantization = "fp4"
	}

	// Special handling for GH200 unified memory
	if arch == ArchGH200 {
		config.TensorParallelSize = 1
		config.GPUMemoryUtilization = 0.95
	}

	return config
}

// DefaultCoverageConfig returns a default configuration
func DefaultCoverageConfig() *CoverageConfig {
	return &CoverageConfig{
		Cluster: ClusterSpecs[ArchGH200],
		Model:   ModelSpecs["llama-405b-bf16"],
		VLLM:    GetVLLMConfig(ArchGH200, "llama-405b-bf16"),
		Analysis: AnalysisConfig{
			Dimensions:          DefaultAnalysisDimensions,
			QualityThreshold:    0.85,
			ConfidenceThreshold: 0.80,
			Timeout:             5 * time.Minute,
			IncludeExamples:     true,
			OutputFormat:        "json",
			Template:            "hollywood",
		},
		Benchmark: BenchmarkConfig{
			Suite:      "standard",
			Iterations: 3,
			WarmupRuns: 2,
			Concurrent: 4,
			Duration:   30 * time.Minute,
		},
	}
}

// RecommendCluster suggests optimal cluster based on requirements
func RecommendCluster(targetCost float64, targetThroughput int) ClusterArchitecture {
	bestArch := ArchGH200 // default
	bestScore := 0.0

	for arch, spec := range ClusterSpecs {
		// Calculate score based on cost efficiency and meeting throughput target
		if spec.ThroughputPerHr >= targetThroughput && spec.CostPerScreenplay <= targetCost {
			score := float64(spec.ThroughputPerHr) / spec.CostPerHour
			if score > bestScore {
				bestScore = score
				bestArch = arch
			}
		}
	}

	return bestArch
}

// FormatClusterComparison generates a comparison table
func FormatClusterComparison() string {
	output := "| Cluster   | GPUs | Model | Memory | Throughput | Cost/Hr | Cost/SP |\n"
	output += "|-----------|------|-------|--------|------------|---------|--------|\n"

	order := []ClusterArchitecture{
		ArchH100_1x, ArchH100_2x, ArchH100_4x, ArchH100_8x,
		ArchGH200,
		ArchB200_1x, ArchB200_2x, ArchB200_4x, ArchB200_8x,
	}

	for _, arch := range order {
		spec := ClusterSpecs[arch]
		output += fmt.Sprintf("| %-9s | %dx   | %-13s | %4dGB  | %3d/hr     | $%.2f  | $%.2f  |\n",
			arch, spec.GPUCount, spec.GPUModel, spec.TotalMemoryGB,
			spec.ThroughputPerHr, spec.CostPerHour, spec.CostPerScreenplay)
	}

	return output
}
