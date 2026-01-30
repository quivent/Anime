package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var producerBinaryPath = filepath.Join(os.Getenv("HOME"), ".local", "bin", "producer")

var producerCmd = &cobra.Command{
	Use:   "producer [command] [args]",
	Short: "Conversational fine-tuning CLI for Llama 3.3 70B",
	Long: `Producer - Conversational Fine-Tuning CLI for Llama 3.3 70B

Transforms a base Llama model into a specialized "producer" through
structured dialogue, learning to analyze screenplays and provide
production insights.

Architecture:
  8x NVIDIA B200 GPUs (1.536TB vRAM total)
  Multi-LoRA blocks targeting attention and MLP layers
  Three-tier memory system (Active/Warm/Cold)
  Self-generated reward function
  Curriculum-based learning (5 phases)

Commands:
  init          Initialize producer configuration
  wizard        Interactive setup wizard
  config        Configuration TUI
  cluster       Manage the GPU cluster
  model         Model operations
  lora          LoRA block management
  session       Training sessions
  memory        Memory management
  train         Training execution
  monitor       Live monitoring
  guidance      Guidance system
  curriculum    Curriculum management
  learning      Learning detection
  screenplay    Screenplay analysis
  debug         Debugging tools
  cost          Cost tracking
  docs          Show documentation

Examples:
  anime producer init --path /var/producer
  anime producer wizard
  anime producer cluster validate
  anime producer model load
  anime producer session start
  anime producer monitor dashboard`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printProducerHelp()
			return
		}
		runProducerBinary(args)
	},
	DisableFlagParsing: true,
}

func printProducerHelp() {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎬 Producer - Conversational Fine-Tuning CLI"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Fine-tune Llama 3.3 70B through structured dialogue"))
	fmt.Println()

	fmt.Println(theme.HeaderStyle.Render("  Setup & Configuration"))
	fmt.Println()

	setupCmds := []struct {
		name string
		desc string
	}{
		{"init", "Initialize producer configuration"},
		{"wizard", "Interactive setup wizard"},
		{"config", "Configuration TUI"},
		{"docs", "Show documentation"},
	}

	for _, c := range setupCmds {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-14s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Core Operations"))
	fmt.Println()

	coreCmds := []struct {
		name string
		desc string
	}{
		{"cluster", "Manage the GPU cluster"},
		{"model", "Model operations (load, freeze, backup)"},
		{"lora", "LoRA block management"},
		{"session", "Training sessions"},
		{"train", "Training execution"},
	}

	for _, c := range coreCmds {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-14s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Monitoring & Analysis"))
	fmt.Println()

	monitorCmds := []struct {
		name string
		desc string
	}{
		{"monitor", "Live monitoring dashboard"},
		{"memory", "Memory management"},
		{"guidance", "Guidance system"},
		{"curriculum", "Curriculum management"},
		{"learning", "Learning detection"},
		{"screenplay", "Screenplay analysis"},
	}

	for _, c := range monitorCmds {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-14s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Utilities"))
	fmt.Println()

	utilCmds := []struct {
		name string
		desc string
	}{
		{"debug", "Debugging tools"},
		{"cost", "Cost tracking"},
		{"status", "System status"},
	}

	for _, c := range utilCmds {
		fmt.Printf("    %s  %s\n",
			theme.InfoStyle.Render(fmt.Sprintf("%-14s", c.name)),
			theme.DimTextStyle.Render(c.desc))
	}

	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Examples"))
	fmt.Println()
	fmt.Printf("    %s anime producer wizard%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime producer cluster validate%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime producer model load%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime producer session start%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Printf("    %s anime producer monitor dashboard%s\n", theme.DimTextStyle.Render("$"), "")
	fmt.Println()
}

func runProducerBinary(args []string) {
	// Check if binary exists at primary location
	binaryPath := producerBinaryPath
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try alternative locations
		alternatives := []string{
			"/usr/local/bin/producer",
			filepath.Join(os.Getenv("HOME"), "go", "bin", "producer"),
			filepath.Join(os.Getenv("HOME"), "anime", "producer", "build", "producer"),
		}
		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				binaryPath = alt
				found = true
				break
			}
		}
		if !found {
			fmt.Println()
			fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Producer binary not found"))
			fmt.Println()
			fmt.Printf("    Expected: %s\n", theme.DimTextStyle.Render(producerBinaryPath))
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("    Build and install it with:"))
			fmt.Printf("      %s cd ~/anime/producer && make install%s\n", theme.DimTextStyle.Render("$"), "")
			fmt.Println()
			fmt.Println(theme.DimTextStyle.Render("    Or use anime to deploy:"))
			fmt.Printf("      %s anime deploy-producer%s\n", theme.DimTextStyle.Render("$"), "")
			fmt.Println()
			return
		}
	}

	// Execute the producer binary, replacing current process
	cmdArgs := append([]string{binaryPath}, args...)
	env := os.Environ()

	if err := syscall.Exec(binaryPath, cmdArgs, env); err != nil {
		// Fallback to exec.Command if syscall.Exec fails (e.g., on some platforms)
		cmd := exec.Command(binaryPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

func init() {
	rootCmd.AddCommand(producerCmd)
}
