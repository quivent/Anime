package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	queryStream bool
	querySystem string
	queryHost   string
)

var queryCmd = &cobra.Command{
	Use:   "query <model> <prompt>",
	Short: "Query a running Ollama model",
	Long: `Query a running Ollama model using the REST API.

Examples:
  anime query llama3.3 "Explain quantum computing"
  anime query mistral "Write a haiku about GPUs" --no-stream
  anime query qwen3 "Debug this error" --system "You are a helpful coding assistant"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required arguments"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime query <model> <prompt>"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime query llama3.3 \"Explain quantum computing\""))
			fmt.Println(theme.DimTextStyle.Render("  anime query mistral \"Write a haiku about GPUs\" --no-stream"))
			fmt.Println(theme.DimTextStyle.Render("  anime query qwen3 \"Debug this error\" --system \"You are a helpful coding assistant\""))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Tip:"))
			fmt.Println(theme.DimTextStyle.Render("  List available models: ollama list"))
			fmt.Println()
			return fmt.Errorf("query requires a model name and prompt")
		}
		return nil
	},
	RunE: runQuery,
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OllamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

func runQuery(cmd *cobra.Command, args []string) error {
	model := args[0]
	prompt := strings.Join(args[1:], " ")

	// Build messages
	messages := []OllamaMessage{}
	if querySystem != "" {
		messages = append(messages, OllamaMessage{
			Role:    "system",
			Content: querySystem,
		})
	}
	messages = append(messages, OllamaMessage{
		Role:    "user",
		Content: prompt,
	})

	// Build request
	reqBody := OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   queryStream,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request
	url := fmt.Sprintf("%s/api/chat", queryHost)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama at %s: %w\nIs Ollama running? Try: ollama serve", queryHost, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render(fmt.Sprintf("🤖 %s:", model)))
	fmt.Println()

	if queryStream {
		// Stream response
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			var chatResp OllamaChatResponse
			if err := json.Unmarshal(line, &chatResp); err != nil {
				continue
			}
			fmt.Print(chatResp.Message.Content)
		}
		fmt.Println()
		fmt.Println()

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}
	} else {
		// Non-streaming response
		var chatResp OllamaChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
		fmt.Println(theme.InfoStyle.Render(chatResp.Message.Content))
		fmt.Println()
	}

	return nil
}

func init() {
	queryCmd.Flags().BoolVar(&queryStream, "stream", true, "Stream response in real-time")
	queryCmd.Flags().BoolVar(&queryStream, "no-stream", false, "Disable streaming (wait for full response)")
	queryCmd.Flags().StringVar(&querySystem, "system", "", "System prompt to set context")
	queryCmd.Flags().StringVar(&queryHost, "host", "http://localhost:11434", "Ollama API host")

	rootCmd.AddCommand(queryCmd)
}
