package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	gemmaBenchRemote  bool
	gemmaBenchServer  string
	gemmaBenchModel   string
	gemmaBenchRounds  int
	gemmaBenchBackend string
	gemmaBenchTokens  int
)

var gemmaBenchCmd = &cobra.Command{
	Use:   "gemma-bench",
	Short: "Benchmark Gemma model TPS",
	Long: `Run inference benchmarks on Gemma models and report tokens/second.

Backends:
  mlx      MLX on Apple Silicon (default on macOS)
  ollama   Ollama API (default on Linux/remote)

Examples:
  anime gemma-bench                                    # MLX, gemma-3-4b-it-4bit, 3 rounds
  anime gemma-bench --model mlx-community/gemma-4-31B-it-OptiQ-4bit
  anime gemma-bench --backend ollama --model gemma3:4b
  anime gemma-bench -r -s mybox --model gemma3:12b     # Remote via Ollama
  anime gemma-bench --rounds 5 --max-tokens 256`,
	RunE: runGemmaBench,
}

func init() {
	gemmaBenchCmd.Flags().BoolVarP(&gemmaBenchRemote, "remote", "r", false, "Run on remote server (uses Ollama)")
	gemmaBenchCmd.Flags().StringVarP(&gemmaBenchServer, "server", "s", "", "Server name for remote bench")
	gemmaBenchCmd.Flags().StringVarP(&gemmaBenchModel, "model", "m", "", "Model to benchmark (default: auto per backend)")
	gemmaBenchCmd.Flags().IntVar(&gemmaBenchRounds, "rounds", 3, "Number of benchmark rounds")
	gemmaBenchCmd.Flags().StringVar(&gemmaBenchBackend, "backend", "", "Backend: mlx or ollama (auto-detected)")
	gemmaBenchCmd.Flags().IntVar(&gemmaBenchTokens, "max-tokens", 128, "Max tokens to generate per round")
	rootCmd.AddCommand(gemmaBenchCmd)
}

const benchPrompt = `Explain the concept of attention mechanisms in transformer architectures. Cover the key-query-value formulation, multi-head attention, and why self-attention scales quadratically with sequence length. Be precise and technical.`

// ── entry point ─────────────────────────────────────────────────────

func runGemmaBench(cmd *cobra.Command, args []string) error {
	// Auto-detect backend
	backend := gemmaBenchBackend
	if backend == "" {
		if gemmaBenchRemote {
			backend = "ollama"
		} else {
			backend = "mlx"
		}
	}

	// Default models per backend
	model := gemmaBenchModel
	if model == "" {
		switch backend {
		case "mlx":
			model = "mlx-community/gemma-3-4b-it-4bit"
		case "ollama":
			model = "gemma3:4b"
		}
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner("⚡ GEMMA BENCHMARK ⚡"))
	fmt.Println()
	fmt.Printf("  Backend:    %s\n", theme.HighlightStyle.Render(backend))
	fmt.Printf("  Model:      %s\n", theme.HighlightStyle.Render(model))
	fmt.Printf("  Rounds:     %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", gemmaBenchRounds)))
	fmt.Printf("  Max tokens: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", gemmaBenchTokens)))

	switch backend {
	case "mlx":
		return runMLXBench(model)
	case "ollama":
		if gemmaBenchRemote {
			return runOllamaBenchRemote(model)
		}
		return runOllamaBenchLocal(model)
	default:
		return fmt.Errorf("unknown backend %q — use mlx or ollama", backend)
	}
}

// ── MLX backend ─────────────────────────────────────────────────────

type benchResult struct {
	Round        int
	PromptTokens int
	GenTokens    int
	PromptTPS    float64
	GenTPS       float64
	PeakMemoryGB float64
	WallTime     time.Duration
}

func runMLXBench(model string) error {
	fmt.Printf("  Runtime:    %s\n", theme.DimTextStyle.Render("mlx_lm (Apple Silicon)"))
	fmt.Println()

	// Verify mlx_lm is available
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 not found")
	}

	var results []benchResult

	fmt.Printf("  %s Running %d rounds...\n\n", theme.SymbolLoading, gemmaBenchRounds)

	for i := 0; i < gemmaBenchRounds; i++ {
		start := time.Now()
		r, err := runMLXRound(i+1, model)
		if err != nil {
			return fmt.Errorf("round %d: %w", i+1, err)
		}
		r.WallTime = time.Since(start)

		fmt.Printf("  Round %d: %s gen, %s prefill, %s peak %s\n",
			i+1,
			theme.HighlightStyle.Render(fmt.Sprintf("%.1f t/s", r.GenTPS)),
			theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", r.PromptTPS)),
			theme.DimTextStyle.Render(fmt.Sprintf("%.2f GB", r.PeakMemoryGB)),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s wall)", r.WallTime.Round(time.Millisecond))))

		results = append(results, r)
	}

	printBenchResults(results, model)
	return nil
}

func runMLXRound(round int, model string) (benchResult, error) {
	// Use Python to call mlx_lm and extract metrics from GenerationResponse
	script := fmt.Sprintf(`
import mlx_lm, json, sys

model, tokenizer = mlx_lm.load('%s')
last = None
for resp in mlx_lm.stream_generate(model, tokenizer, prompt='''%s''', max_tokens=%d):
    last = resp

if last is None:
    print('{"error":"no output"}')
    sys.exit(1)

print(json.dumps({
    "prompt_tokens": last.prompt_tokens,
    "prompt_tps": last.prompt_tps,
    "generation_tokens": last.generation_tokens,
    "generation_tps": last.generation_tps,
    "peak_memory_gb": last.peak_memory,
}))
`, model, strings.ReplaceAll(benchPrompt, "'", "\\'"), gemmaBenchTokens)

	cmd := exec.Command("python3", "-c", script)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return benchResult{}, fmt.Errorf("mlx_lm failed: %w", err)
	}

	var data struct {
		PromptTokens int     `json:"prompt_tokens"`
		PromptTPS    float64 `json:"prompt_tps"`
		GenTokens    int     `json:"generation_tokens"`
		GenTPS       float64 `json:"generation_tps"`
		PeakMemGB    float64 `json:"peak_memory_gb"`
		Error        string  `json:"error"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return benchResult{}, fmt.Errorf("parse mlx output: %w\n%s", err, string(out))
	}
	if data.Error != "" {
		return benchResult{}, fmt.Errorf("mlx error: %s", data.Error)
	}

	return benchResult{
		Round:        round,
		PromptTokens: data.PromptTokens,
		GenTokens:    data.GenTokens,
		PromptTPS:    data.PromptTPS,
		GenTPS:       data.GenTPS,
		PeakMemoryGB: data.PeakMemGB,
	}, nil
}

// ── Ollama backend ──────────────────────────────────────────────────

type ollamaGenerateResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

func runOllamaBenchLocal(model string) error {
	host := "http://localhost:11434"
	fmt.Printf("  Target:     %s\n", theme.DimTextStyle.Render(host))
	fmt.Println()

	results, err := runOllamaRounds(host, model)
	if err != nil {
		return err
	}
	printBenchResults(results, model)
	return nil
}

func runOllamaBenchRemote(model string) error {
	serverName := gemmaBenchServer
	if serverName == "" {
		cfg, err := config.Load()
		if err != nil || len(cfg.Servers) == 0 {
			return fmt.Errorf("no server specified and no default configured")
		}
		serverName = cfg.Servers[0].Name
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	server, err := cfg.GetServer(serverName)
	if err != nil {
		target := cfg.GetAlias(serverName)
		if target == "" {
			return fmt.Errorf("server %q not found", serverName)
		}
		parts := strings.SplitN(target, "@", 2)
		if len(parts) == 2 {
			server = &config.Server{Name: serverName, User: parts[0], Host: parts[1]}
		} else {
			server = &config.Server{Name: serverName, User: "ubuntu", Host: target}
		}
	}

	fmt.Printf("  Target:     %s\n", theme.DimTextStyle.Render(fmt.Sprintf("%s@%s", server.User, server.Host)))
	fmt.Println()

	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("SSH connect: %w", err)
	}
	defer client.Close()

	escapedPrompt := strings.ReplaceAll(benchPrompt, `"`, `\"`)
	benchScript := fmt.Sprintf(`#!/bin/bash
set -e
for i in $(seq 1 %d); do
    curl -s http://localhost:11434/api/generate -d '{"model":"%s","prompt":"%s","stream":false,"options":{"num_predict":%d}}'
    echo ""
done
`, gemmaBenchRounds, model, escapedPrompt, gemmaBenchTokens)

	fmt.Printf("  %s Running %d rounds...\n\n", theme.SymbolLoading, gemmaBenchRounds)
	output, err := client.RunCommand(benchScript)
	if err != nil {
		return fmt.Errorf("benchmark failed: %w\n%s", err, output)
	}

	var results []benchResult
	for i, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var resp ollamaGenerateResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		r := parseOllamaResponse(i+1, &resp)
		results = append(results, r)
	}

	if len(results) == 0 {
		return fmt.Errorf("no valid results — is %s installed? Try: ollama pull %s", model, model)
	}

	printBenchResults(results, model)
	return nil
}

func runOllamaRounds(host, model string) ([]benchResult, error) {
	var results []benchResult
	fmt.Printf("  %s Running %d rounds...\n\n", theme.SymbolLoading, gemmaBenchRounds)

	for i := 0; i < gemmaBenchRounds; i++ {
		body := fmt.Sprintf(`{"model":"%s","prompt":"%s","stream":false,"options":{"num_predict":%d}}`,
			model, strings.ReplaceAll(benchPrompt, `"`, `\"`), gemmaBenchTokens)

		start := time.Now()
		resp, err := http.Post(host+"/api/generate", "application/json", strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("round %d: ollama not reachable at %s: %w", i+1, host, err)
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		wallTime := time.Since(start)

		var gen ollamaGenerateResponse
		if err := json.Unmarshal(data, &gen); err != nil {
			return nil, fmt.Errorf("round %d: bad response: %w\n%s", i+1, err, string(data))
		}

		if !gen.Done {
			return nil, fmt.Errorf("round %d: generation incomplete — is %s installed? Try: ollama pull %s", i+1, model, model)
		}

		r := parseOllamaResponse(i+1, &gen)
		r.WallTime = wallTime
		fmt.Printf("  Round %d: %s gen, %s prefill %s\n",
			i+1,
			theme.HighlightStyle.Render(fmt.Sprintf("%.1f t/s", r.GenTPS)),
			theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", r.PromptTPS)),
			theme.DimTextStyle.Render(fmt.Sprintf("(%s wall)", wallTime.Round(time.Millisecond))))
		results = append(results, r)
	}
	return results, nil
}

func parseOllamaResponse(round int, resp *ollamaGenerateResponse) benchResult {
	r := benchResult{
		Round:        round,
		PromptTokens: resp.PromptEvalCount,
		GenTokens:    resp.EvalCount,
	}
	if resp.PromptEvalDuration > 0 {
		r.PromptTPS = float64(resp.PromptEvalCount) / (float64(resp.PromptEvalDuration) / 1e9)
	}
	if resp.EvalDuration > 0 {
		r.GenTPS = float64(resp.EvalCount) / (float64(resp.EvalDuration) / 1e9)
	}
	return r
}

// ── shared results display ──────────────────────────────────────────

func printBenchResults(results []benchResult, model string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("📊 RESULTS"))
	fmt.Println()

	var minGen, maxGen, sumGen float64
	var minPrompt, maxPrompt, sumPrompt float64
	var totalTokens int
	var peakMem float64

	minGen = 1e9
	minPrompt = 1e9

	for _, r := range results {
		if r.GenTPS < minGen {
			minGen = r.GenTPS
		}
		if r.GenTPS > maxGen {
			maxGen = r.GenTPS
		}
		sumGen += r.GenTPS

		if r.PromptTPS < minPrompt {
			minPrompt = r.PromptTPS
		}
		if r.PromptTPS > maxPrompt {
			maxPrompt = r.PromptTPS
		}
		sumPrompt += r.PromptTPS
		totalTokens += r.GenTokens

		if r.PeakMemoryGB > peakMem {
			peakMem = r.PeakMemoryGB
		}
	}

	n := float64(len(results))
	avgGen := sumGen / n
	avgPrompt := sumPrompt / n

	fmt.Printf("  Model: %s  |  Rounds: %d  |  Tokens generated: %d\n",
		theme.HighlightStyle.Render(model), len(results), totalTokens)
	if peakMem > 0 {
		fmt.Printf("  Peak memory: %s\n", theme.DimTextStyle.Render(fmt.Sprintf("%.2f GB", peakMem)))
	}
	fmt.Println()

	fmt.Printf("  %s\n", theme.InfoStyle.Render("Generation (decode):"))
	fmt.Printf("    avg  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.1f t/s", avgGen)))
	fmt.Printf("    min  %s    max  %s\n",
		theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", minGen)),
		theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", maxGen)))
	fmt.Println()

	fmt.Printf("  %s\n", theme.InfoStyle.Render("Prompt eval (prefill):"))
	fmt.Printf("    avg  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%.1f t/s", avgPrompt)))
	fmt.Printf("    min  %s    max  %s\n",
		theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", minPrompt)),
		theme.DimTextStyle.Render(fmt.Sprintf("%.1f t/s", maxPrompt)))
	fmt.Println()

	if len(results) > 1 {
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("Per-round detail:"))
		for _, r := range results {
			mem := ""
			if r.PeakMemoryGB > 0 {
				mem = fmt.Sprintf("  %.2fGB", r.PeakMemoryGB)
			}
			wall := ""
			if r.WallTime > 0 {
				wall = fmt.Sprintf("  %s wall", r.WallTime.Round(time.Millisecond))
			}
			fmt.Printf("    #%d  gen %5.1f t/s  prefill %5.1f t/s  %d tok%s%s\n",
				r.Round, r.GenTPS, r.PromptTPS, r.GenTokens, mem, wall)
		}
		fmt.Println()
	}

	if avgGen >= 100 {
		fmt.Printf("  %s\n", theme.SuccessStyle.Render("🔥 Excellent — production-grade throughput"))
	} else if avgGen >= 40 {
		fmt.Printf("  %s\n", theme.SuccessStyle.Render("✓ Good — solid interactive performance"))
	} else if avgGen >= 15 {
		fmt.Printf("  %s\n", theme.WarningStyle.Render("⚠ Moderate — usable but not fast"))
	} else {
		fmt.Printf("  %s\n", theme.ErrorStyle.Render("✗ Slow — check GPU utilization"))
	}
	fmt.Println()
}

