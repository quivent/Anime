package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <service>",
	Short: "Run services locally (comfyui, ollama, jupyter)",
	Long: `Quick launcher for locally installed services.

Available services:
  comfyui   - Start ComfyUI web interface
  ollama    - Start Ollama server (interactive mode)
  jupyter   - Start Jupyter Lab
  tensorboard - Start TensorBoard

Examples:
  anime run comfyui       # Start ComfyUI on http://127.0.0.1:8188
  anime run ollama serve  # Start Ollama server
  anime run jupyter       # Start Jupyter Lab

Note: Use 'anime start' to run services on a remote server with port forwarding.`,
	Aliases: []string{"launch"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Service name required"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("📖 Usage:"))
			fmt.Println(theme.HighlightStyle.Render("  anime run <service> [args...]"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Available Services:"))
			fmt.Println(theme.DimTextStyle.Render("  comfyui       - Start ComfyUI web interface (http://127.0.0.1:8188)"))
			fmt.Println(theme.DimTextStyle.Render("  ollama        - Start Ollama server (http://127.0.0.1:11434)"))
			fmt.Println(theme.DimTextStyle.Render("  jupyter       - Start Jupyter Lab"))
			fmt.Println(theme.DimTextStyle.Render("  tensorboard   - Start TensorBoard"))
			fmt.Println()
			fmt.Println(theme.SuccessStyle.Render("✨ Examples:"))
			fmt.Println(theme.DimTextStyle.Render("  anime run comfyui"))
			fmt.Println(theme.DimTextStyle.Render("  anime run ollama serve"))
			fmt.Println(theme.DimTextStyle.Render("  anime run jupyter"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Tip:"))
			fmt.Println(theme.DimTextStyle.Render("  Use 'anime start' for remote server with port forwarding"))
			fmt.Println()
			return fmt.Errorf("run requires a service name")
		}
		return nil
	},
	RunE: runService,
}

func runService(cmd *cobra.Command, args []string) error {
	service := args[0]
	serviceArgs := args[1:]

	switch service {
	case "comfyui":
		return runServiceComfyUI(serviceArgs)
	case "comfyui-log", "comfyui-logs":
		return showComfyUILog()
	case "ollama":
		return runOllama(serviceArgs)
	case "jupyter":
		return runJupyter(serviceArgs)
	case "tensorboard":
		return runTensorBoard(serviceArgs)
	default:
		return fmt.Errorf("unknown service: %s\n\nAvailable services: comfyui, comfyui-log, ollama, jupyter, tensorboard", service)
	}
}

func showComfyUILog() error {
	logPath := filepath.Join(os.TempDir(), "comfyui.log")

	// Check if log exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("❌ ComfyUI log not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Log path: ") + theme.DimTextStyle.Render(logPath))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  ComfyUI may not have been started yet"))
		fmt.Println(theme.DimTextStyle.Render("  Try: anime run comfyui"))
		fmt.Println()
		return nil
	}

	// Show log with tail
	tailCmd := exec.Command("tail", "-50", logPath)
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("📄 ComfyUI Log (last 50 lines)"))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  %s", logPath)))
	fmt.Println()

	return tailCmd.Run()
}

func runServiceComfyUI(args []string) error {
	comfyPath := filepath.Join(os.Getenv("HOME"), "ComfyUI")

	// Check if ComfyUI exists
	if _, err := os.Stat(comfyPath); os.IsNotExist(err) {
		fmt.Println(theme.ErrorStyle.Render("❌ ComfyUI not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Install it with:"))
		fmt.Println("  " + theme.HighlightStyle.Render("anime install comfyui"))
		return fmt.Errorf("ComfyUI not installed at ~/ComfyUI")
	}

	// Find python3
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		pythonPath, err = exec.LookPath("python")
		if err != nil {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Python not found"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("💡 Install Python:"))
			fmt.Println(theme.DimTextStyle.Render("  macOS:   brew install python3"))
			fmt.Println(theme.DimTextStyle.Render("  Ubuntu:  sudo apt install python3"))
			fmt.Println(theme.DimTextStyle.Render("  Or use:  anime install core"))
			fmt.Println()
			return fmt.Errorf("python not found in PATH")
		}
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎨 Starting ComfyUI..."))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  URL: ") + theme.HighlightStyle.Render("http://127.0.0.1:8188"))
	fmt.Println(theme.DimTextStyle.Render("  Path: ~/ComfyUI"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Press Ctrl+C to stop"))
	fmt.Println()

	// Build command
	cmdArgs := []string{filepath.Join(comfyPath, "main.py")}
	cmdArgs = append(cmdArgs, args...)

	// Create the command
	comfyCmd := exec.Command(pythonPath, cmdArgs...)
	comfyCmd.Dir = comfyPath
	comfyCmd.Stdin = os.Stdin
	comfyCmd.Stdout = os.Stdout
	comfyCmd.Stderr = os.Stderr

	// Run ComfyUI
	if err := comfyCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}

func runOllama(args []string) error {
	ollamaPath, err := exec.LookPath("ollama")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Ollama not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Install it with:"))
		fmt.Println("  " + theme.HighlightStyle.Render("anime install ollama"))
		return fmt.Errorf("ollama not found in PATH")
	}

	// Default to 'serve' if no args provided
	if len(args) == 0 {
		args = []string{"serve"}
		fmt.Println()
		fmt.Println(theme.GlowStyle.Render("🤖 Starting Ollama server..."))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  URL: ") + theme.HighlightStyle.Render("http://127.0.0.1:11434"))
		fmt.Println(theme.DimTextStyle.Render("  Press Ctrl+C to stop"))
		fmt.Println()
	}

	// Create the command
	ollamaCmd := exec.Command(ollamaPath, args...)
	ollamaCmd.Stdin = os.Stdin
	ollamaCmd.Stdout = os.Stdout
	ollamaCmd.Stderr = os.Stderr

	// Run Ollama
	if err := ollamaCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}

func runJupyter(args []string) error {
	jupyterPath, err := exec.LookPath("jupyter")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Jupyter not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Install it with:"))
		fmt.Println("  " + theme.HighlightStyle.Render("pip install jupyterlab"))
		return fmt.Errorf("jupyter not found in PATH")
	}

	// Default to 'lab' if no args provided
	if len(args) == 0 {
		args = []string{"lab"}
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("📓 Starting Jupyter Lab..."))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Press Ctrl+C to stop"))
	fmt.Println()

	// Create the command
	jupyterCmd := exec.Command(jupyterPath, args...)
	jupyterCmd.Stdin = os.Stdin
	jupyterCmd.Stdout = os.Stdout
	jupyterCmd.Stderr = os.Stderr

	// Run Jupyter
	if err := jupyterCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}

func runTensorBoard(args []string) error {
	tensorboardPath, err := exec.LookPath("tensorboard")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ TensorBoard not found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Install it with:"))
		fmt.Println("  " + theme.HighlightStyle.Render("pip install tensorboard"))
		return fmt.Errorf("tensorboard not found in PATH")
	}

	// Default to current directory logs if no args
	if len(args) == 0 {
		args = []string{"--logdir=./logs"}
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("📊 Starting TensorBoard..."))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  URL: ") + theme.HighlightStyle.Render("http://127.0.0.1:6006"))
	fmt.Println(theme.DimTextStyle.Render("  Press Ctrl+C to stop"))
	fmt.Println()

	// Create the command
	tbCmd := exec.Command(tensorboardPath, args...)
	tbCmd.Stdin = os.Stdin
	tbCmd.Stdout = os.Stdout
	tbCmd.Stderr = os.Stderr

	// Run TensorBoard
	if err := tbCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
