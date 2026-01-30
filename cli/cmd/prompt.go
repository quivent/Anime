package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	promptYes bool
)

var promptCmd = &cobra.Command{
	Use:                "prompt <natural language command>",
	Short:              "Execute anime commands using natural language",
	Long:               `Use natural language to describe what you want anime to do, and it will figure out the right command.`,
	Aliases:            []string{"do", "please"},
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true, // Allow any text
	RunE:               runPrompt,
}

func runPrompt(cmd *cobra.Command, args []string) error {
	// Join all args into the natural language prompt
	userPrompt := strings.Join(args, " ")

	// Check if Ollama is running
	if !isOllamaRunning() {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Ollama not running"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Please start Ollama first:"))
		fmt.Println("  " + theme.HighlightStyle.Render("anime run ollama"))
		fmt.Println()
		return fmt.Errorf("Ollama server not available")
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🤔 Interpreting command..."))
	fmt.Println()

	// Build the system prompt with available commands
	systemPrompt := `You are an AI assistant that translates natural language requests into anime CLI commands.

Available anime commands:
- anime install <package> - Install packages (core, python, pytorch, ollama, comfyui, claude, svd, animatediff, mochi, cogvideo, ltxvideo, wan2, opensora)
- anime packages - List all available packages
- anime packages status - Show installation status
- anime status - Check server status
- anime metrics - View GPU metrics and costs
- anime query <model> <prompt> - Query Ollama models
- anime run <service> - Start services (comfyui, ollama, jupyter)
- anime animate <collection> [model] - Animate images to videos (models: wan2, mochi, svd, ltx). Flags: --seconds N, --fps N, --prompt "text", --parallel N
- anime upscale <collection> [scale] - Upscale images (models: esrgan, realesrgan, gfpgan). Flags: --model name, --scale N, --parallel N
- anime collection create <name> <path> - Create collection
- anime collection list - List collections
- anime collection info <name> - Show collection details
- anime ssh - SSH to server
- anime push - Build and push anime to server
- anime tree - View all commands
- anime interactive - Interactive package selector
- anime models - List AI models
- anime doctor - Analyze installation issues
- anime workstation - Launch monitoring dashboard

Respond with ONLY the anime command that matches the user's intent. Do not include explanations, markdown, or code blocks.
If the request is ambiguous, pick the most likely command.
If you cannot map it to a command, respond with: UNCLEAR: <reason>

Examples:
User: "install pytorch"
Response: anime install pytorch

User: "check the status"
Response: anime status

User: "show me all packages"
Response: anime packages

User: "query llama about quantum physics"
Response: anime query llama3.3 "explain quantum physics"

User: "start comfyui"
Response: anime run comfyui

User: "animate all photos in mar to 3 second clips"
Response: anime animate mar --seconds 3

User: "animate mar with mochi model in cinematic style"
Response: anime animate mar mochi --prompt "cinematic"

User: "upscale images in photos collection to 4x"
Response: anime upscale photos 4`

	// Query Ollama to interpret the command
	interpretedCmd, err := queryOllamaForCommand(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to interpret command: %w", err)
	}

	interpretedCmd = strings.TrimSpace(interpretedCmd)

	// Check for unclear response
	if strings.HasPrefix(interpretedCmd, "UNCLEAR:") {
		fmt.Println(theme.WarningStyle.Render("❓ Could not interpret command"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render(strings.TrimPrefix(interpretedCmd, "UNCLEAR:")))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Try:"))
		fmt.Println("  " + theme.HighlightStyle.Render("anime tree") + theme.DimTextStyle.Render(" - View all commands"))
		fmt.Println("  " + theme.HighlightStyle.Render("anime packages") + theme.DimTextStyle.Render(" - List packages"))
		fmt.Println()
		return nil
	}

	// Remove "anime" prefix if present
	interpretedCmd = strings.TrimPrefix(interpretedCmd, "anime ")
	interpretedCmd = strings.TrimSpace(interpretedCmd)

	// Show what will be executed
	fmt.Println(theme.InfoStyle.Render("📝 Interpreted command:"))
	fmt.Println("  " + theme.HighlightStyle.Render("anime "+interpretedCmd))
	fmt.Println()

	// Execute the command
	if !promptYes {
		fmt.Print(theme.HighlightStyle.Render("Execute this command? (Y/n): "))
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "n" || response == "no" {
			fmt.Println()
			fmt.Println(theme.WarningStyle.Render("❌ Cancelled"))
			return nil
		}
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🚀 Executing..."))
	fmt.Println()

	// Parse and execute the command
	cmdParts := strings.Fields(interpretedCmd)
	if len(cmdParts) == 0 {
		return fmt.Errorf("empty command")
	}

	execCmd := exec.Command("anime", cmdParts...)
	execCmd.Stdout = cmd.OutOrStdout()
	execCmd.Stderr = cmd.ErrOrStderr()
	execCmd.Stdin = cmd.InOrStdin()

	return execCmd.Run()
}

func queryOllamaForCommand(systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "llama3.3",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Create request with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Show progress while waiting
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		count := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				count++
				if count == 1 {
					fmt.Print(theme.DimTextStyle.Render("  Waiting for Ollama response"))
				} else if count <= 10 {
					fmt.Print(theme.DimTextStyle.Render("."))
				} else if count == 11 {
					fmt.Println()
					fmt.Println(theme.WarningStyle.Render("  (This is taking a while - model may need to be pulled first)"))
					fmt.Print(theme.DimTextStyle.Render("  Still waiting"))
				} else {
					fmt.Print(theme.DimTextStyle.Render("."))
				}
			}
		}
	}()

	client := &http.Client{}
	resp, err := client.Do(req)
	close(done)
	fmt.Println() // New line after progress dots

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Request timed out after 30 seconds"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 This usually means:"))
			fmt.Println(theme.DimTextStyle.Render("  1. The model 'llama3.3' is not pulled yet"))
			fmt.Println(theme.DimTextStyle.Render("  2. Ollama is still starting up"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Try pulling the model first:"))
			fmt.Println(theme.HighlightStyle.Render("  ollama pull llama3.3"))
			fmt.Println()
			return "", fmt.Errorf("timeout waiting for Ollama")
		}
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	message, ok := result["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return content, nil
}

func isOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func init() {
	promptCmd.Flags().BoolVarP(&promptYes, "yes", "y", false, "Execute without confirmation")
	rootCmd.AddCommand(promptCmd)
}
