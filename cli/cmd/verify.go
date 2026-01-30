package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <package>",
	Short: "Verify if a package is installed",
	Long: `Check if a specific package is installed and working.

Examples:
  anime verify vllm
  anime verify pytorch
  anime verify ollama
  anime verify comfyui`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime verify <package>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Examples:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime verify vllm"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime verify pytorch"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime verify ollama"))
			fmt.Println()
			return fmt.Errorf("requires package name")
		}
		return nil
	},
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

// Package verification rules
type verifyCheck struct {
	checkType string // "python", "command", "dir", "ollama"
	target    string // module name, command name, or directory path
}

// Map packages to their verification checks
var packageChecks = map[string]verifyCheck{
	// Python packages
	"pytorch":    {checkType: "python", target: "torch"},
	"vllm":       {checkType: "python", target: "vllm"},
	"flash-attn": {checkType: "python", target: "flash_attn"},
	"python":     {checkType: "command", target: "python3"},

	// Commands
	"ollama":    {checkType: "command", target: "ollama"},
	"docker":    {checkType: "command", target: "docker"},
	"nodejs":    {checkType: "command", target: "node"},
	"go":        {checkType: "command", target: "go"},
	"claude":    {checkType: "command", target: "claude"},
	"gh":        {checkType: "command", target: "gh"},
	"make":      {checkType: "command", target: "make"},
	"nvidia":    {checkType: "command", target: "nvidia-smi"},
	"comfy-cli": {checkType: "command", target: "comfy"},

	// Directories
	"comfyui":     {checkType: "dir", target: "$HOME/ComfyUI"},
	"mochi":       {checkType: "dir", target: "$HOME/video-models/mochi-1"},
	"cogvideo":    {checkType: "dir", target: "$HOME/video-models/cogvideo"},
	"opensora":    {checkType: "dir", target: "$HOME/video-models/open-sora"},
	"ltxvideo":    {checkType: "dir", target: "$HOME/video-models/ltxvideo"},
	"wan2":        {checkType: "dir", target: "$HOME/video-models/wan2"},
	"comfyui-wan2": {checkType: "dir", target: "$HOME/ComfyUI/custom_nodes/ComfyUI-WanWrapper"},

	// Ollama models
	"llama-3.3-70b":     {checkType: "ollama", target: "llama3.3:70b"},
	"llama-3.3-8b":      {checkType: "ollama", target: "llama3.3:8b"},
	"mistral":           {checkType: "ollama", target: "mistral:latest"},
	"mixtral":           {checkType: "ollama", target: "mixtral:8x7b"},
	"qwen3-235b":        {checkType: "ollama", target: "qwen3:235b"},
	"qwen3-32b":         {checkType: "ollama", target: "qwen3:32b"},
	"qwen3-30b":         {checkType: "ollama", target: "qwen3:30b"},
	"qwen3-14b":         {checkType: "ollama", target: "qwen3:14b"},
	"qwen3-8b":          {checkType: "ollama", target: "qwen3:8b"},
	"qwen3-4b":          {checkType: "ollama", target: "qwen3:4b"},
	"deepseek-coder-33b": {checkType: "ollama", target: "deepseek-coder:33b"},
	"deepseek-v3":       {checkType: "ollama", target: "deepseek-v3"},
	"phi-3.5":           {checkType: "ollama", target: "phi3.5:latest"},
	"phi-4":             {checkType: "ollama", target: "phi4:latest"},
	"deepseek-r1-8b":    {checkType: "ollama", target: "deepseek-r1:8b"},
	"deepseek-r1-70b":   {checkType: "ollama", target: "deepseek-r1:70b"},
	"gemma3-4b":         {checkType: "ollama", target: "gemma3:4b"},
	"gemma3-12b":        {checkType: "ollama", target: "gemma3:12b"},
	"gemma3-27b":        {checkType: "ollama", target: "gemma3:27b"},
	"llama-3.2-1b":      {checkType: "ollama", target: "llama3.2:1b"},
	"llama-3.2-3b":      {checkType: "ollama", target: "llama3.2:3b"},
	"qwen3-coder-30b":   {checkType: "ollama", target: "qwen3-coder:30b"},
	"command-r-7b":      {checkType: "ollama", target: "command-r:7b"},

	// ComfyUI models (check for directory/file existence)
	"sdxl":         {checkType: "file", target: "$HOME/ComfyUI/models/checkpoints/sd_xl_base_1.0.safetensors"},
	"sd15":         {checkType: "file", target: "$HOME/ComfyUI/models/checkpoints/v1-5-pruned-emaonly.safetensors"},
	"flux-dev":     {checkType: "file", target: "$HOME/ComfyUI/models/unet/flux1-dev.safetensors"},
	"flux-schnell": {checkType: "file", target: "$HOME/ComfyUI/models/unet/flux1-schnell.safetensors"},
	"flux2":        {checkType: "file", target: "$HOME/ComfyUI/models/unet/flux2-fp8.safetensors"},
	"svd":          {checkType: "dir", target: "$HOME/ComfyUI/custom_nodes/ComfyUI-VideoHelperSuite"},
	"animatediff":  {checkType: "dir", target: "$HOME/ComfyUI/custom_nodes/ComfyUI-AnimateDiff-Evolved"},
}

func runVerify(cmd *cobra.Command, args []string) error {
	pkg := args[0]

	fmt.Println()
	fmt.Printf("  Package: %s\n", theme.HighlightStyle.Render(pkg))
	fmt.Println()

	// Check if it's a known package
	_, isKnown := installer.Scripts[pkg]
	if !isKnown {
		fmt.Println(theme.WarningStyle.Render("  Unknown package"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Use 'anime install --list' to see available packages"))
		fmt.Println()
		return nil
	}

	// Get verification check for this package
	check, hasCheck := packageChecks[pkg]
	if !hasCheck {
		// No specific check defined, try to infer
		fmt.Println(theme.WarningStyle.Render("  No verification check defined for this package"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  The package is known but verification is not implemented"))
		fmt.Println()
		return nil
	}

	// Run the appropriate check
	installed := false
	var details string

	switch check.checkType {
	case "python":
		installed, details = checkPythonModule(check.target)
	case "command":
		installed, details = checkCommand(check.target)
	case "dir":
		installed, details = checkDirectory(check.target)
	case "file":
		installed, details = checkFile(check.target)
	case "ollama":
		installed, details = checkOllamaModel(check.target)
	}

	if installed {
		fmt.Println(theme.SuccessStyle.Render("  Installed"))
		if details != "" {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(details))
		}
	} else {
		fmt.Println(theme.ErrorStyle.Render("  Not installed"))
		if details != "" {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(details))
		}
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  To install:"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime install %s", pkg)))
	}
	fmt.Println()

	return nil
}

func checkPythonModule(module string) (bool, string) {
	cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s; print(%s.__version__)", module, module))
	output, err := cmd.Output()
	if err != nil {
		// Try without version
		cmd2 := exec.Command("python3", "-c", fmt.Sprintf("import %s", module))
		if err2 := cmd2.Run(); err2 != nil {
			return false, ""
		}
		return true, "Version: unknown"
	}
	return true, fmt.Sprintf("Version: %s", strings.TrimSpace(string(output)))
}

func checkCommand(command string) (bool, string) {
	path, err := exec.LookPath(command)
	if err != nil {
		return false, ""
	}

	// Try to get version
	cmd := exec.Command(command, "--version")
	output, err := cmd.Output()
	if err != nil {
		return true, fmt.Sprintf("Path: %s", path)
	}

	// Get first line of version output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return true, lines[0]
	}
	return true, fmt.Sprintf("Path: %s", path)
}

func checkDirectory(dir string) (bool, string) {
	// Expand $HOME
	dir = os.ExpandEnv(dir)
	if strings.HasPrefix(dir, "$HOME") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, strings.TrimPrefix(dir, "$HOME"))
	}

	info, err := os.Stat(dir)
	if err != nil {
		return false, ""
	}
	if !info.IsDir() {
		return false, "Path exists but is not a directory"
	}
	return true, fmt.Sprintf("Path: %s", dir)
}

func checkFile(file string) (bool, string) {
	// Expand $HOME
	file = os.ExpandEnv(file)
	if strings.HasPrefix(file, "$HOME") {
		home, _ := os.UserHomeDir()
		file = filepath.Join(home, strings.TrimPrefix(file, "$HOME"))
	}

	info, err := os.Stat(file)
	if err != nil {
		return false, ""
	}
	if info.IsDir() {
		return false, "Path exists but is a directory, not a file"
	}

	// Format size
	size := info.Size()
	var sizeStr string
	switch {
	case size >= 1<<30:
		sizeStr = fmt.Sprintf("%.2f GB", float64(size)/(1<<30))
	case size >= 1<<20:
		sizeStr = fmt.Sprintf("%.2f MB", float64(size)/(1<<20))
	default:
		sizeStr = fmt.Sprintf("%d bytes", size)
	}

	return true, fmt.Sprintf("Path: %s (%s)", file, sizeStr)
}

func checkOllamaModel(model string) (bool, string) {
	// First check if ollama is installed
	if _, err := exec.LookPath("ollama"); err != nil {
		return false, "Ollama not installed"
	}

	// Check if model is downloaded
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, "Could not list Ollama models"
	}

	// Parse output to find model
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			// Model name is first field, may include :tag
			modelName := fields[0]
			if modelName == model || strings.HasPrefix(modelName, model+":") {
				// Found the model, get size if available
				if len(fields) >= 3 {
					return true, fmt.Sprintf("Model: %s, Size: %s", modelName, fields[2])
				}
				return true, fmt.Sprintf("Model: %s", modelName)
			}
		}
	}

	return false, ""
}
