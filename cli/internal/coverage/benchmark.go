package coverage

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// BenchmarkSuite defines a benchmark suite configuration
type BenchmarkSuite struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Iterations  int           `json:"iterations"`
	WarmupRuns  int           `json:"warmup_runs"`
	Concurrent  int           `json:"concurrent"`
	Screenplays int           `json:"screenplays"`
	Duration    time.Duration `json:"duration"`
}

// BenchmarkMetrics contains performance metrics from a benchmark run
type BenchmarkMetrics struct {
	// Timing metrics
	TTFT         time.Duration `json:"ttft"`          // Time to first token
	TotalTime    time.Duration `json:"total_time"`    // Total processing time
	LatencyP50   time.Duration `json:"latency_p50"`   // 50th percentile latency
	LatencyP90   time.Duration `json:"latency_p90"`   // 90th percentile latency
	LatencyP99   time.Duration `json:"latency_p99"`   // 99th percentile latency

	// Throughput metrics
	TokensPerSecond    int     `json:"tokens_per_second"`
	ScreenplaysPerHour float64 `json:"screenplays_per_hour"`
	RequestsPerSecond  float64 `json:"requests_per_second"`

	// Resource metrics
	GPUUtilization    float64 `json:"gpu_utilization"`     // Percentage
	MemoryUtilization float64 `json:"memory_utilization"`  // Percentage
	PowerDraw         float64 `json:"power_draw_watts"`

	// Quality metrics
	AccuracyScore     float64 `json:"accuracy_score"`      // Agreement with human readers
	ConsistencyScore  float64 `json:"consistency_score"`   // Test-retest reliability

	// Cost metrics
	CostPerScreenplay float64 `json:"cost_per_screenplay"`
	TokensPerDollar   int     `json:"tokens_per_dollar"`
}

// BenchmarkResult contains results from a benchmark run
type BenchmarkResult struct {
	Suite       string            `json:"suite"`
	Cluster     ClusterArchitecture `json:"cluster"`
	Model       string            `json:"model"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	Iterations  int               `json:"iterations"`
	Screenplays int               `json:"screenplays"`
	Metrics     BenchmarkMetrics  `json:"metrics"`
	Errors      []string          `json:"errors,omitempty"`
	Raw         []BenchmarkMetrics `json:"raw_results,omitempty"`
}

// ExperimentConfig defines an A/B experiment configuration
type ExperimentConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // model, precision, prompt, batch, cluster
	Variants    []string          `json:"variants"`
	Metric      string            `json:"metric"` // accuracy, speed, cost, quality
	Samples     int               `json:"samples"`
	BaselineIdx int               `json:"baseline_idx"`
}

// ExperimentResult contains results from an A/B experiment
type ExperimentResult struct {
	Config   ExperimentConfig           `json:"config"`
	Results  map[string]BenchmarkResult `json:"results"`
	Winner   string                     `json:"winner"`
	Analysis string                     `json:"analysis"`
}

// Predefined benchmark suites
var BenchmarkSuites = map[string]BenchmarkSuite{
	"quick": {
		Name:        "Quick",
		Description: "Fast sanity check (5 minutes)",
		Iterations:  1,
		WarmupRuns:  1,
		Concurrent:  2,
		Screenplays: 5,
		Duration:    5 * time.Minute,
	},
	"standard": {
		Name:        "Standard",
		Description: "Standard benchmark suite (30 minutes)",
		Iterations:  3,
		WarmupRuns:  2,
		Concurrent:  4,
		Screenplays: 25,
		Duration:    30 * time.Minute,
	},
	"extended": {
		Name:        "Extended",
		Description: "Extended benchmark with comprehensive metrics (2 hours)",
		Iterations:  5,
		WarmupRuns:  3,
		Concurrent:  8,
		Screenplays: 100,
		Duration:    2 * time.Hour,
	},
	"production": {
		Name:        "Production",
		Description: "Production-grade benchmark (8 hours)",
		Iterations:  10,
		WarmupRuns:  5,
		Concurrent:  16,
		Screenplays: 500,
		Duration:    8 * time.Hour,
	},
}

// BenchmarkTargets defines target metrics for validation
var BenchmarkTargets = map[string]float64{
	"ttft_ms":               200,    // Max time to first token
	"latency_p99_ms":        2000,   // Max 99th percentile latency
	"throughput_per_hour":   50,     // Min screenplays per hour
	"accuracy_score":        0.80,   // Min accuracy
	"consistency_score":     0.90,   // Min consistency
	"gpu_utilization":       0.80,   // Min GPU utilization
	"cost_per_screenplay":   0.20,   // Max cost per screenplay
}

// CalculatePercentile calculates the nth percentile from a slice of durations
func CalculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	index := int(math.Ceil(percentile/100*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// AggregateMetrics aggregates multiple benchmark metrics
func AggregateMetrics(results []BenchmarkMetrics) BenchmarkMetrics {
	if len(results) == 0 {
		return BenchmarkMetrics{}
	}

	var (
		totalTTFT         time.Duration
		totalTime         time.Duration
		totalTPS          int
		totalSPH          float64
		totalRPS          float64
		totalGPU          float64
		totalMem          float64
		totalPower        float64
		totalAccuracy     float64
		totalConsistency  float64
		totalCostPerSP    float64
		totalTokensDollar int
	)

	latencies := make([]time.Duration, 0, len(results))

	for _, r := range results {
		totalTTFT += r.TTFT
		totalTime += r.TotalTime
		totalTPS += r.TokensPerSecond
		totalSPH += r.ScreenplaysPerHour
		totalRPS += r.RequestsPerSecond
		totalGPU += r.GPUUtilization
		totalMem += r.MemoryUtilization
		totalPower += r.PowerDraw
		totalAccuracy += r.AccuracyScore
		totalConsistency += r.ConsistencyScore
		totalCostPerSP += r.CostPerScreenplay
		totalTokensDollar += r.TokensPerDollar
		latencies = append(latencies, r.TotalTime)
	}

	n := float64(len(results))

	return BenchmarkMetrics{
		TTFT:              time.Duration(float64(totalTTFT) / n),
		TotalTime:         time.Duration(float64(totalTime) / n),
		LatencyP50:        CalculatePercentile(latencies, 50),
		LatencyP90:        CalculatePercentile(latencies, 90),
		LatencyP99:        CalculatePercentile(latencies, 99),
		TokensPerSecond:   int(float64(totalTPS) / n),
		ScreenplaysPerHour: totalSPH / n,
		RequestsPerSecond:  totalRPS / n,
		GPUUtilization:    totalGPU / n,
		MemoryUtilization: totalMem / n,
		PowerDraw:         totalPower / n,
		AccuracyScore:     totalAccuracy / n,
		ConsistencyScore:  totalConsistency / n,
		CostPerScreenplay: totalCostPerSP / n,
		TokensPerDollar:   int(float64(totalTokensDollar) / n),
	}
}

// ValidateMetrics checks if metrics meet targets
func ValidateMetrics(metrics BenchmarkMetrics) (bool, []string) {
	var issues []string
	passed := true

	if float64(metrics.TTFT.Milliseconds()) > BenchmarkTargets["ttft_ms"] {
		issues = append(issues, fmt.Sprintf("TTFT %.0fms exceeds target %.0fms",
			float64(metrics.TTFT.Milliseconds()), BenchmarkTargets["ttft_ms"]))
		passed = false
	}

	if float64(metrics.LatencyP99.Milliseconds()) > BenchmarkTargets["latency_p99_ms"] {
		issues = append(issues, fmt.Sprintf("P99 latency %.0fms exceeds target %.0fms",
			float64(metrics.LatencyP99.Milliseconds()), BenchmarkTargets["latency_p99_ms"]))
		passed = false
	}

	if metrics.ScreenplaysPerHour < BenchmarkTargets["throughput_per_hour"] {
		issues = append(issues, fmt.Sprintf("Throughput %.1f/hr below target %.0f/hr",
			metrics.ScreenplaysPerHour, BenchmarkTargets["throughput_per_hour"]))
		passed = false
	}

	if metrics.AccuracyScore < BenchmarkTargets["accuracy_score"] {
		issues = append(issues, fmt.Sprintf("Accuracy %.2f below target %.2f",
			metrics.AccuracyScore, BenchmarkTargets["accuracy_score"]))
		passed = false
	}

	if metrics.ConsistencyScore < BenchmarkTargets["consistency_score"] {
		issues = append(issues, fmt.Sprintf("Consistency %.2f below target %.2f",
			metrics.ConsistencyScore, BenchmarkTargets["consistency_score"]))
		passed = false
	}

	if metrics.GPUUtilization < BenchmarkTargets["gpu_utilization"] {
		issues = append(issues, fmt.Sprintf("GPU utilization %.1f%% below target %.0f%%",
			metrics.GPUUtilization*100, BenchmarkTargets["gpu_utilization"]*100))
		passed = false
	}

	if metrics.CostPerScreenplay > BenchmarkTargets["cost_per_screenplay"] {
		issues = append(issues, fmt.Sprintf("Cost $%.2f/screenplay exceeds target $%.2f",
			metrics.CostPerScreenplay, BenchmarkTargets["cost_per_screenplay"]))
		passed = false
	}

	return passed, issues
}

// FormatBenchmarkResult formats a benchmark result for display
func FormatBenchmarkResult(result BenchmarkResult) string {
	builder := &strings.Builder{}

	builder.WriteString("═══════════════════════════════════════════════\n")
	builder.WriteString(fmt.Sprintf("    BENCHMARK RESULTS: %s\n", result.Suite))
	builder.WriteString("═══════════════════════════════════════════════\n\n")

	builder.WriteString(fmt.Sprintf("Cluster:     %s\n", result.Cluster))
	builder.WriteString(fmt.Sprintf("Model:       %s\n", result.Model))
	builder.WriteString(fmt.Sprintf("Iterations:  %d\n", result.Iterations))
	builder.WriteString(fmt.Sprintf("Screenplays: %d\n", result.Screenplays))
	builder.WriteString(fmt.Sprintf("Duration:    %s\n\n", result.EndTime.Sub(result.StartTime).Round(time.Second)))

	m := result.Metrics
	builder.WriteString("--- TIMING ---\n")
	builder.WriteString(fmt.Sprintf("  TTFT:        %v\n", m.TTFT.Round(time.Millisecond)))
	builder.WriteString(fmt.Sprintf("  Total Time:  %v\n", m.TotalTime.Round(time.Millisecond)))
	builder.WriteString(fmt.Sprintf("  P50 Latency: %v\n", m.LatencyP50.Round(time.Millisecond)))
	builder.WriteString(fmt.Sprintf("  P90 Latency: %v\n", m.LatencyP90.Round(time.Millisecond)))
	builder.WriteString(fmt.Sprintf("  P99 Latency: %v\n\n", m.LatencyP99.Round(time.Millisecond)))

	builder.WriteString("--- THROUGHPUT ---\n")
	builder.WriteString(fmt.Sprintf("  Tokens/sec:      %d\n", m.TokensPerSecond))
	builder.WriteString(fmt.Sprintf("  Screenplays/hr:  %.1f\n", m.ScreenplaysPerHour))
	builder.WriteString(fmt.Sprintf("  Requests/sec:    %.2f\n\n", m.RequestsPerSecond))

	builder.WriteString("--- RESOURCES ---\n")
	builder.WriteString(fmt.Sprintf("  GPU Utilization: %.1f%%\n", m.GPUUtilization*100))
	builder.WriteString(fmt.Sprintf("  Memory Usage:    %.1f%%\n", m.MemoryUtilization*100))
	builder.WriteString(fmt.Sprintf("  Power Draw:      %.0fW\n\n", m.PowerDraw))

	builder.WriteString("--- QUALITY ---\n")
	builder.WriteString(fmt.Sprintf("  Accuracy:    %.2f\n", m.AccuracyScore))
	builder.WriteString(fmt.Sprintf("  Consistency: %.2f\n\n", m.ConsistencyScore))

	builder.WriteString("--- COST ---\n")
	builder.WriteString(fmt.Sprintf("  Cost/Screenplay: $%.2f\n", m.CostPerScreenplay))
	builder.WriteString(fmt.Sprintf("  Tokens/$:        %d\n\n", m.TokensPerDollar))

	// Validation
	passed, issues := ValidateMetrics(m)
	if passed {
		builder.WriteString("✅ All targets met\n")
	} else {
		builder.WriteString("❌ Targets not met:\n")
		for _, issue := range issues {
			builder.WriteString(fmt.Sprintf("   - %s\n", issue))
		}
	}

	return builder.String()
}

// CompareResults compares benchmark results across clusters
func CompareResults(results []BenchmarkResult) string {
	if len(results) == 0 {
		return "No results to compare"
	}

	builder := &strings.Builder{}

	builder.WriteString("═══════════════════════════════════════════════════════════════════\n")
	builder.WriteString("                    BENCHMARK COMPARISON\n")
	builder.WriteString("═══════════════════════════════════════════════════════════════════\n\n")

	// Header
	builder.WriteString(fmt.Sprintf("| %-12s | %-8s | %-8s | %-8s | %-8s | %-8s |\n",
		"Cluster", "TTFT", "TPS", "SP/Hr", "GPU%", "Cost/SP"))
	builder.WriteString("|--------------|----------|----------|----------|----------|----------|\n")

	// Data rows
	for _, r := range results {
		m := r.Metrics
		builder.WriteString(fmt.Sprintf("| %-12s | %6dms | %8d | %8.1f | %7.1f%% | $%7.2f |\n",
			r.Cluster,
			m.TTFT.Milliseconds(),
			m.TokensPerSecond,
			m.ScreenplaysPerHour,
			m.GPUUtilization*100,
			m.CostPerScreenplay))
	}

	builder.WriteString("\n")

	// Find best in each category
	var bestTTFT, bestTPS, bestSPH, bestCost BenchmarkResult
	bestTTFT = results[0]
	bestTPS = results[0]
	bestSPH = results[0]
	bestCost = results[0]

	for _, r := range results[1:] {
		if r.Metrics.TTFT < bestTTFT.Metrics.TTFT {
			bestTTFT = r
		}
		if r.Metrics.TokensPerSecond > bestTPS.Metrics.TokensPerSecond {
			bestTPS = r
		}
		if r.Metrics.ScreenplaysPerHour > bestSPH.Metrics.ScreenplaysPerHour {
			bestSPH = r
		}
		if r.Metrics.CostPerScreenplay < bestCost.Metrics.CostPerScreenplay {
			bestCost = r
		}
	}

	builder.WriteString("WINNERS:\n")
	builder.WriteString(fmt.Sprintf("  Lowest Latency:    %s (%dms TTFT)\n", bestTTFT.Cluster, bestTTFT.Metrics.TTFT.Milliseconds()))
	builder.WriteString(fmt.Sprintf("  Highest TPS:       %s (%d tokens/sec)\n", bestTPS.Cluster, bestTPS.Metrics.TokensPerSecond))
	builder.WriteString(fmt.Sprintf("  Highest Throughput:%s (%.1f/hr)\n", bestSPH.Cluster, bestSPH.Metrics.ScreenplaysPerHour))
	builder.WriteString(fmt.Sprintf("  Lowest Cost:       %s ($%.2f/screenplay)\n", bestCost.Cluster, bestCost.Metrics.CostPerScreenplay))

	return builder.String()
}
