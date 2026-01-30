package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	inferenceHost     string
	inferencePort     int
	inferenceStream   bool
	inferenceMaxTokens int
	inferenceTemp     float64
	inferenceCheck    bool
)

var inferenceCmd = &cobra.Command{
	Use:   "inference",
	Short: "Run inference on local or remote models",
	Long: `Run inference against various AI models.

Supports multiple model backends:
  - llama: Llama 3.3 70B via vLLM
  - (more coming soon)

Examples:
  anime inference llama "What is the meaning of life?"
  anime inference llama --stream "Tell me a story"
  anime inference llama --host 10.0.0.5 "Hello"
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var llamaCmd = &cobra.Command{
	Use:   "llama [prompt]",
	Short: "Run inference on Llama 3.3 70B",
	Long: `Run inference on Llama 3.3 70B model via vLLM.

The model should be running via vLLM on the specified host.
Default assumes local GPU at localhost:8000.

Examples:
  anime inference llama "What is 2+2?"
  anime inference llama --stream "Write a poem"
  anime inference llama --host lambda "Explain quantum computing"
  anime inference llama --max-tokens 500 "Write a story"
`,
	Args: cobra.MinimumNArgs(0),
	RunE: runLlamaInference,
}

func init() {
	// Global inference flags
	inferenceCmd.PersistentFlags().StringVar(&inferenceHost, "host", "localhost", "Model server host (IP, hostname, or anime alias)")
	inferenceCmd.PersistentFlags().IntVar(&inferencePort, "port", 8000, "Model server port")
	inferenceCmd.PersistentFlags().BoolVarP(&inferenceStream, "stream", "s", false, "Stream output tokens")
	inferenceCmd.PersistentFlags().IntVar(&inferenceMaxTokens, "max-tokens", 1024, "Maximum tokens to generate")
	inferenceCmd.PersistentFlags().Float64Var(&inferenceTemp, "temperature", 0.7, "Sampling temperature")
	inferenceCmd.PersistentFlags().BoolVar(&inferenceCheck, "check", false, "Check connectivity without running inference")

	inferenceCmd.AddCommand(llamaCmd)
	rootCmd.AddCommand(inferenceCmd)
}

// LlamaRequest represents a vLLM chat completion request
type LlamaRequest struct {
	Model       string         `json:"model"`
	Messages    []LlamaMessage `json:"messages"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	Stream      bool           `json:"stream"`
}

// LlamaMessage represents a chat message
type LlamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LlamaResponse represents a vLLM chat completion response
type LlamaResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ConnectionValidationResult contains the result of a connection check
type ConnectionValidationResult struct {
	Healthy      bool
	BaseURL      string
	HealthStatus int
	ModelsFound  int
	ErrorMessage string
	ErrorType    string // "network", "health", "models", "none"
}

// validateVLLMConnection performs comprehensive pre-flight connectivity checks
func validateVLLMConnection(baseURL string, timeout time.Duration) (*ConnectionValidationResult, error) {
	result := &ConnectionValidationResult{
		BaseURL: baseURL,
	}

	client := &http.Client{Timeout: timeout}

	// Step 1: Check health endpoint
	healthURL := baseURL + "/health"
	healthResp, err := client.Get(healthURL)
	if err != nil {
		result.ErrorType = "network"
		result.ErrorMessage = fmt.Sprintf("Cannot reach vLLM server at %s", baseURL)

		// Provide more specific network error details
		if strings.Contains(err.Error(), "connection refused") {
			result.ErrorMessage += " (connection refused - server may not be running)"
		} else if strings.Contains(err.Error(), "timeout") {
			result.ErrorMessage += " (connection timeout - check network or firewall)"
		} else if strings.Contains(err.Error(), "no such host") {
			result.ErrorMessage += " (host not found - check hostname/IP)"
		}

		return result, fmt.Errorf("%s: %w", result.ErrorMessage, err)
	}
	defer healthResp.Body.Close()

	result.HealthStatus = healthResp.StatusCode

	if healthResp.StatusCode != http.StatusOK {
		result.ErrorType = "health"
		body, _ := io.ReadAll(healthResp.Body)
		result.ErrorMessage = fmt.Sprintf("vLLM server unhealthy (status %d)", healthResp.StatusCode)
		if len(body) > 0 {
			result.ErrorMessage += fmt.Sprintf(": %s", string(body))
		}
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	// Step 2: Check models endpoint
	modelsURL := baseURL + "/v1/models"
	modelsResp, err := client.Get(modelsURL)
	if err != nil {
		result.ErrorType = "models"
		result.ErrorMessage = "Health check passed but cannot query models endpoint"
		return result, fmt.Errorf("%s: %w", result.ErrorMessage, err)
	}
	defer modelsResp.Body.Close()

	if modelsResp.StatusCode != http.StatusOK {
		result.ErrorType = "models"
		result.ErrorMessage = fmt.Sprintf("Models endpoint returned status %d", modelsResp.StatusCode)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	var modelsData struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(modelsResp.Body).Decode(&modelsData); err != nil {
		result.ErrorType = "models"
		result.ErrorMessage = "Failed to decode models response"
		return result, fmt.Errorf("%s: %w", result.ErrorMessage, err)
	}

	result.ModelsFound = len(modelsData.Data)
	if result.ModelsFound == 0 {
		result.ErrorType = "models"
		result.ErrorMessage = "No models available on vLLM server"
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	// All checks passed
	result.Healthy = true
	result.ErrorType = "none"
	return result, nil
}

// displayConnectionCheck shows detailed connection validation results
func displayConnectionCheck(result *ConnectionValidationResult) {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Connection Validation Results"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(result.BaseURL))

	if result.Healthy {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Status:"), theme.SuccessStyle.Render("✓ Healthy"))
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Models Available:"), result.ModelsFound)
		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("Connection successful! Ready for inference."))
	} else {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Status:"), theme.ErrorStyle.Render("✗ Unhealthy"))
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Error Type:"), result.ErrorType)
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("Error: " + result.ErrorMessage))
		fmt.Println()

		// Provide troubleshooting suggestions based on error type
		fmt.Println(theme.InfoStyle.Render("Troubleshooting:"))
		switch result.ErrorType {
		case "network":
			fmt.Println(theme.DimTextStyle.Render("  1. Check if vLLM is running on the target host"))
			fmt.Println(theme.DimTextStyle.Render("  2. Verify network connectivity and firewall rules"))
			fmt.Println(theme.DimTextStyle.Render("  3. Confirm the host and port are correct"))
		case "health":
			fmt.Println(theme.DimTextStyle.Render("  1. vLLM server is reachable but reports unhealthy status"))
			fmt.Println(theme.DimTextStyle.Render("  2. Check vLLM server logs for errors"))
			fmt.Println(theme.DimTextStyle.Render("  3. Restart vLLM service if needed"))
		case "models":
			fmt.Println(theme.DimTextStyle.Render("  1. Ensure vLLM was started with a model"))
			fmt.Println(theme.DimTextStyle.Render("  2. Example: vllm serve meta-llama/Llama-3.3-70B-Instruct"))
			fmt.Println(theme.DimTextStyle.Render("  3. Check vLLM startup logs for model loading errors"))
		}
	}
	fmt.Println()
}

func runLlamaInference(cmd *cobra.Command, args []string) error {
	// Resolve host (could be an anime alias)
	host := resolveInferenceHost(inferenceHost)
	baseURL := fmt.Sprintf("http://%s:%d", host, inferencePort)

	// Perform pre-flight validation with configurable timeout
	validationTimeout := 10 * time.Second
	validationResult, validationErr := validateVLLMConnection(baseURL, validationTimeout)

	// If --check flag is set, display results and exit
	if inferenceCheck {
		displayConnectionCheck(validationResult)
		if validationErr != nil {
			return fmt.Errorf("connection check failed")
		}
		return nil
	}

	// If validation failed, display detailed error and exit
	if validationErr != nil {
		displayConnectionCheck(validationResult)
		return fmt.Errorf("cannot proceed with inference - connection validation failed")
	}

	// Get prompt from args or stdin
	var prompt string
	if len(args) > 0 {
		prompt = strings.Join(args, " ")
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Read from pipe
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			prompt = strings.TrimSpace(string(data))
		} else {
			// Interactive mode
			fmt.Print(theme.InfoStyle.Render("Enter prompt: "))
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			prompt = strings.TrimSpace(input)
		}
	}

	if prompt == "" {
		return fmt.Errorf("no prompt provided")
	}

	// Detect which model to use (validation already confirmed models exist)
	modelName, err := detectLlamaModel(baseURL)
	if err != nil {
		// This shouldn't happen since validation passed, but handle it gracefully
		return fmt.Errorf("failed to detect model: %w", err)
	}

	// Run inference
	if inferenceStream {
		return runStreamingInference(baseURL, modelName, prompt)
	}
	return runBatchInference(baseURL, modelName, prompt)
}

func resolveInferenceHost(host string) string {
	// If it's already an IP or has dots, use as-is
	if strings.Contains(host, ".") || host == "localhost" {
		return host
	}

	// Try to resolve as anime alias
	target, err := parseServerTarget(host)
	if err == nil && target != "" {
		// Extract just the host part (remove user@)
		parts := strings.Split(target, "@")
		if len(parts) == 2 {
			return parts[1]
		}
		return target
	}

	return host
}

func detectLlamaModel(baseURL string) (string, error) {
	// Query vLLM for available models
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/v1/models")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return "", err
	}

	// Look for Llama model
	for _, model := range modelsResp.Data {
		if strings.Contains(strings.ToLower(model.ID), "llama") {
			return model.ID, nil
		}
	}

	// Return first model if no Llama found
	if len(modelsResp.Data) > 0 {
		return modelsResp.Data[0].ID, nil
	}

	return "", fmt.Errorf("no models available")
}

func runBatchInference(baseURL, model, prompt string) error {
	fmt.Println()
	fmt.Printf("%s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(model))
	fmt.Println()

	req := LlamaRequest{
		Model: model,
		Messages: []LlamaMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   inferenceMaxTokens,
		Temperature: inferenceTemp,
		Stream:      false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Post(
		baseURL+"/v1/chat/completions",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var result LlamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	elapsed := time.Since(start)

	// Print response
	if len(result.Choices) > 0 {
		content := result.Choices[0].Message.Content
		fmt.Println(content)
		fmt.Println()

		// Stats
		tokensPerSec := float64(result.Usage.CompletionTokens) / elapsed.Seconds()
		fmt.Printf("%s %d prompt + %d completion = %d total (%.1f tok/s)\n",
			theme.DimTextStyle.Render("Tokens:"),
			result.Usage.PromptTokens,
			result.Usage.CompletionTokens,
			result.Usage.TotalTokens,
			tokensPerSec,
		)
		fmt.Printf("%s %s\n", theme.DimTextStyle.Render("Time:"), elapsed.Round(time.Millisecond))
	}
	fmt.Println()

	return nil
}

func runStreamingInference(baseURL, model, prompt string) error {
	fmt.Println()
	fmt.Printf("%s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(model))
	fmt.Println()

	req := LlamaRequest{
		Model: model,
		Messages: []LlamaMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   inferenceMaxTokens,
		Temperature: inferenceTemp,
		Stream:      true,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Post(
		baseURL+"/v1/chat/completions",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	// Stream SSE response
	reader := bufio.NewReader(resp.Body)
	var totalTokens int

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var chunk LlamaResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					fmt.Print(content)
					totalTokens++
				}
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Println()

	// Stats
	tokensPerSec := float64(totalTokens) / elapsed.Seconds()
	fmt.Printf("%s ~%d tokens (%.1f tok/s)\n",
		theme.DimTextStyle.Render("Generated:"),
		totalTokens,
		tokensPerSec,
	)
	fmt.Printf("%s %s\n", theme.DimTextStyle.Render("Time:"), elapsed.Round(time.Millisecond))
	fmt.Println()

	return nil
}

// checkVLLMStatus checks if vLLM is running and returns GPU info
func checkVLLMStatus(host string, port int) error {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return fmt.Errorf("vLLM not responding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vLLM unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// startVLLM attempts to start vLLM with the Llama model
func startVLLM(model string, gpuCount int) error {
	args := []string{
		"serve", model,
		"--tensor-parallel-size", fmt.Sprintf("%d", gpuCount),
		"--host", "0.0.0.0",
		"--port", "8000",
	}

	cmd := exec.Command("vllm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}
