package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/tui"
	"github.com/spf13/cobra"
)

var (
	workflowInteractive bool
	workflowFromProfile string
	workflowStartDry    bool
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflow profiles for different use cases",
	Long: `Manage named workflow profiles that configure LLM servers, models,
GPU allocation, and optimizations for different use cases.

COMMANDS:
  anime workflow                    Launch workflow TUI
  anime workflow list               List all workflows
  anime workflow use <name>         Switch to a workflow (set active)
  anime workflow start [name]       Start a workflow (load models, run commands)
  anime workflow stop               Stop the current workflow
  anime workflow create <name>      Create new workflow (TUI)
  anime workflow show [name]        Show workflow details
  anime workflow delete <name>      Delete a workflow
  anime workflow clone <src> <dst>  Clone a workflow

EXAMPLES:
  anime workflow                           # Launch TUI
  anime workflow create substrate          # Create 'substrate' workflow in TUI
  anime workflow create training -i        # Create with interactive prompts
  anime workflow use substrate             # Switch to substrate workflow
  anime workflow start                     # Start active workflow
  anime workflow start substrate           # Start specific workflow
  anime workflow clone substrate dev       # Clone substrate as 'dev'
  anime workflow show                      # Show active workflow
  anime workflow list                      # List all workflows`,
	RunE: runWorkflow,
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workflow profiles",
	RunE:  runWorkflowList,
}

var workflowUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a workflow profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowUse,
}

var workflowCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new workflow profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowCreate,
}

var workflowShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show workflow details",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWorkflowShow,
}

var workflowDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a workflow profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowDelete,
}

var workflowCloneCmd = &cobra.Command{
	Use:   "clone <source> <new-name>",
	Short: "Clone an existing workflow",
	Args:  cobra.ExactArgs(2),
	RunE:  runWorkflowClone,
}

var workflowStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a workflow (load models, set env, run commands)",
	Long: `Start a workflow by:
  1. Setting environment variables
  2. Running pre-commands
  3. Starting the LLM server (ollama, vllm, etc.)
  4. Loading configured models
  5. Running post-commands

If no name is provided, starts the active workflow.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWorkflowStart,
}

var workflowStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current workflow (unload models, stop servers)",
	RunE:  runWorkflowStop,
}

func init() {
	workflowCmd.Flags().BoolVarP(&workflowInteractive, "interactive", "i", false, "Use simple interactive prompts instead of TUI")

	workflowCreateCmd.Flags().BoolVarP(&workflowInteractive, "interactive", "i", false, "Use simple interactive prompts instead of TUI")
	workflowCreateCmd.Flags().StringVar(&workflowFromProfile, "from", "", "Clone from existing workflow")

	workflowStartCmd.Flags().BoolVar(&workflowStartDry, "dry-run", false, "Show what would be done without executing")

	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowUseCmd)
	workflowCmd.AddCommand(workflowStartCmd)
	workflowCmd.AddCommand(workflowStopCmd)
	workflowCmd.AddCommand(workflowCreateCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowDeleteCmd)
	workflowCmd.AddCommand(workflowCloneCmd)

	rootCmd.AddCommand(workflowCmd)
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	if workflowInteractive {
		return runWorkflowInteractive()
	}

	// Launch TUI
	m, err := tui.NewWorkflowModel()
	if err != nil {
		return fmt.Errorf("failed to initialize workflow TUI: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

func runWorkflowList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workflows := cfg.ListWorkflows()
	if len(workflows) == 0 {
		fmt.Println(theme.DimTextStyle.Render("No workflows configured."))
		fmt.Println(theme.DimTextStyle.Render("Use 'anime workflow create <name>' to create one."))
		return nil
	}

	fmt.Println(theme.InfoStyle.Render("Workflow Profiles"))
	fmt.Println()

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	for _, w := range workflows {
		isActive := w.Name == cfg.ActiveWorkflow
		marker := "  "
		if isActive {
			marker = activeStyle.Render("* ")
		}

		name := w.Name
		if isActive {
			name = activeStyle.Render(name)
		}

		serverStr := dimStyle.Render(fmt.Sprintf("[%s]", w.Server))
		modelCount := dimStyle.Render(fmt.Sprintf("%d models", len(w.Models)))

		desc := ""
		if w.Description != "" {
			desc = dimStyle.Render(" - " + w.Description)
		}

		fmt.Printf("%s%s %s %s%s\n", marker, name, serverStr, modelCount, desc)
	}

	return nil
}

func runWorkflowUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workflow, err := cfg.GetWorkflow(name)
	if err != nil {
		return err
	}

	if err := cfg.SetActiveWorkflow(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s Switched to workflow: %s\n", theme.SuccessStyle.Render("*"), name)
	fmt.Printf("   Server: %s\n", workflow.Server)
	fmt.Printf("   Models: %d configured\n", len(workflow.Models))

	if workflow.AutoLoad {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Tip: Use 'anime workflow start' to load models and start servers"))
	}

	return nil
}

func runWorkflowCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if workflow already exists
	if _, err := cfg.GetWorkflow(name); err == nil {
		return fmt.Errorf("workflow %s already exists", name)
	}

	// Clone from existing if specified
	if workflowFromProfile != "" {
		if err := cfg.CloneWorkflow(workflowFromProfile, name); err != nil {
			return err
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("%s Created workflow '%s' from '%s'\n",
			theme.SuccessStyle.Render("*"), name, workflowFromProfile)
		return nil
	}

	if workflowInteractive {
		return createWorkflowInteractive(cfg, name)
	}

	// Launch TUI in create mode
	m, err := tui.NewWorkflowModelWithCreate(name)
	if err != nil {
		return fmt.Errorf("failed to initialize workflow TUI: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

func runWorkflowShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var workflow *config.WorkflowProfile
	var name string

	if len(args) > 0 {
		name = args[0]
		workflow, err = cfg.GetWorkflow(name)
		if err != nil {
			return err
		}
	} else {
		// Show active workflow
		workflow, err = cfg.GetActiveWorkflow()
		if err != nil {
			fmt.Println(theme.DimTextStyle.Render("No active workflow set."))
			fmt.Println(theme.DimTextStyle.Render("Use 'anime workflow use <name>' to activate one."))
			return nil
		}
		name = cfg.ActiveWorkflow
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	fmt.Println(titleStyle.Render(fmt.Sprintf("Workflow: %s", name)))
	if workflow.Description != "" {
		fmt.Println(labelStyle.Render(workflow.Description))
	}
	fmt.Println()

	fmt.Printf("%s %s\n", labelStyle.Render("Server:"), valueStyle.Render(string(workflow.Server)))

	if len(workflow.Models) > 0 {
		fmt.Printf("%s\n", labelStyle.Render("Models:"))
		for _, m := range workflow.Models {
			enabledStr := theme.DimTextStyle.Render("disabled")
			if m.Enabled {
				enabledStr = theme.SuccessStyle.Render("enabled")
			}
			gpuStr := ""
			if len(m.GPUs) > 0 {
				gpuStr = fmt.Sprintf(" GPUs: %v", m.GPUs)
			}
			fmt.Printf("  - %s [%s]%s\n", m.ID, enabledStr, gpuStr)
		}
	}

	if workflow.GPUConfig.TotalGPUs > 0 {
		fmt.Printf("%s %d x %s (%dGB each)\n",
			labelStyle.Render("GPUs:"),
			workflow.GPUConfig.TotalGPUs,
			workflow.GPUConfig.GPUType,
			workflow.GPUConfig.GPUMemoryGB)
	}

	// Show optimizations
	opt := workflow.Optimizations
	if opt.FlashAttention || opt.PagedAttention || opt.SpeculativeDecoding {
		fmt.Printf("%s\n", labelStyle.Render("Optimizations:"))
		if opt.FlashAttention {
			fmt.Println("  - Flash Attention")
		}
		if opt.PagedAttention {
			fmt.Println("  - Paged Attention")
		}
		if opt.SpeculativeDecoding {
			fmt.Printf("  - Speculative Decoding (draft: %s)\n", opt.DraftModel)
		}
		if opt.ContinuousBatching {
			fmt.Println("  - Continuous Batching")
		}
	}

	if len(workflow.Tags) > 0 {
		fmt.Printf("%s %s\n", labelStyle.Render("Tags:"), strings.Join(workflow.Tags, ", "))
	}

	if workflow.AutoLoad {
		fmt.Printf("%s %s\n", labelStyle.Render("Auto-load:"), "enabled")
	}

	return nil
}

func runWorkflowDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteWorkflow(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s Deleted workflow: %s\n", theme.SuccessStyle.Render("*"), name)
	return nil
}

func runWorkflowClone(cmd *cobra.Command, args []string) error {
	sourceName := args[0]
	newName := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.CloneWorkflow(sourceName, newName); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s Cloned workflow '%s' to '%s'\n",
		theme.SuccessStyle.Render("*"), sourceName, newName)
	return nil
}

// runWorkflowInteractive provides simple text-based workflow management
func runWorkflowInteractive() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workflows := cfg.ListWorkflows()

	fmt.Println(theme.InfoStyle.Render("Workflow Manager (Interactive)"))
	fmt.Println()

	if len(workflows) == 0 {
		fmt.Println("No workflows configured.")
	} else {
		fmt.Println("Current workflows:")
		for i, w := range workflows {
			active := ""
			if w.Name == cfg.ActiveWorkflow {
				active = " (active)"
			}
			fmt.Printf("  %d. %s [%s] - %d models%s\n",
				i+1, w.Name, w.Server, len(w.Models), active)
		}
	}

	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  1. Create new workflow")
	fmt.Println("  2. Switch workflow")
	fmt.Println("  3. Delete workflow")
	fmt.Println("  4. Exit")
	fmt.Println()

	var choice int
	fmt.Print("Choice: ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		var name string
		fmt.Print("Workflow name: ")
		fmt.Scanln(&name)
		return createWorkflowInteractive(cfg, name)
	case 2:
		var name string
		fmt.Print("Workflow name: ")
		fmt.Scanln(&name)
		if err := cfg.SetActiveWorkflow(name); err != nil {
			return err
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("%s Switched to workflow: %s\n", theme.SuccessStyle.Render("*"), name)
		return nil
	case 3:
		var name string
		fmt.Print("Workflow name to delete: ")
		fmt.Scanln(&name)
		if err := cfg.DeleteWorkflow(name); err != nil {
			return err
		}
		return cfg.Save()
	}

	return nil
}

// createWorkflowInteractive creates a workflow with simple prompts
func createWorkflowInteractive(cfg *config.Config, name string) error {
	workflow := config.WorkflowProfile{
		Name: name,
	}

	// Description
	fmt.Print("Description (optional): ")
	var desc string
	fmt.Scanln(&desc)
	workflow.Description = desc

	// Server type
	fmt.Println("\nServer types: ollama, vllm, tensorrt-llm, llama.cpp, exllamav2")
	fmt.Print("Server type [ollama]: ")
	var server string
	fmt.Scanln(&server)
	if server == "" {
		server = "ollama"
	}
	workflow.Server = config.LLMServerType(server)

	// GPU config
	fmt.Print("Number of GPUs [0]: ")
	var gpus int
	fmt.Scanln(&gpus)
	if gpus > 0 {
		workflow.GPUConfig.TotalGPUs = gpus
		fmt.Print("GPU type (e.g., H100, A100): ")
		var gpuType string
		fmt.Scanln(&gpuType)
		workflow.GPUConfig.GPUType = gpuType
		fmt.Print("VRAM per GPU (GB): ")
		var vram int
		fmt.Scanln(&vram)
		workflow.GPUConfig.GPUMemoryGB = vram
	}

	// Auto-load
	fmt.Print("Auto-load models on activation? [y/N]: ")
	var autoLoad string
	fmt.Scanln(&autoLoad)
	workflow.AutoLoad = strings.ToLower(autoLoad) == "y"

	// Add workflow
	if err := cfg.AddWorkflow(workflow); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n%s Created workflow: %s\n", theme.SuccessStyle.Render("*"), name)
	fmt.Println(theme.DimTextStyle.Render("Use 'anime workflow use " + name + "' to activate it."))
	fmt.Println(theme.DimTextStyle.Render("Use 'anime workflow' to edit in TUI and add models."))

	return nil
}

// runWorkflowStart starts the active or specified workflow
func runWorkflowStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var workflow *config.WorkflowProfile
	var name string

	if len(args) > 0 {
		name = args[0]
		workflow, err = cfg.GetWorkflow(name)
		if err != nil {
			return err
		}
		// Also set as active
		cfg.SetActiveWorkflow(name)
		cfg.Save()
	} else {
		workflow, err = cfg.GetActiveWorkflow()
		if err != nil {
			return fmt.Errorf("no active workflow. Use 'anime workflow start <name>' or 'anime workflow use <name>' first")
		}
		name = cfg.ActiveWorkflow
	}

	fmt.Printf("%s Starting workflow: %s\n", theme.InfoStyle.Render("*"), name)
	fmt.Println()

	// Step 1: Set environment variables
	if len(workflow.Environment) > 0 {
		fmt.Printf("%s Setting environment variables...\n", theme.DimTextStyle.Render("▶"))
		for k, v := range workflow.Environment {
			if workflowStartDry {
				fmt.Printf("   [dry-run] export %s=%s\n", k, v)
			} else {
				os.Setenv(k, v)
				fmt.Printf("   %s=%s\n", k, v)
			}
		}
	}

	// Step 2: Run pre-commands
	if len(workflow.PreCommands) > 0 {
		fmt.Printf("%s Running pre-commands...\n", theme.DimTextStyle.Render("▶"))
		for _, cmdStr := range workflow.PreCommands {
			if workflowStartDry {
				fmt.Printf("   [dry-run] %s\n", cmdStr)
			} else {
				fmt.Printf("   $ %s\n", cmdStr)
				execCmd := exec.Command("bash", "-c", cmdStr)
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				if err := execCmd.Run(); err != nil {
					fmt.Printf("   %s %v\n", theme.WarningStyle.Render("warning:"), err)
				}
			}
		}
	}

	// Step 3: Start LLM server based on type
	fmt.Printf("%s Starting %s server...\n", theme.DimTextStyle.Render("▶"), workflow.Server)
	if err := startLLMServer(workflow, workflowStartDry); err != nil {
		return fmt.Errorf("failed to start LLM server: %w", err)
	}

	// Step 4: Load models
	if len(workflow.Models) > 0 {
		fmt.Printf("%s Loading models...\n", theme.DimTextStyle.Render("▶"))
		for _, model := range workflow.Models {
			if !model.Enabled {
				continue
			}
			if workflowStartDry {
				fmt.Printf("   [dry-run] load %s\n", model.ID)
			} else {
				fmt.Printf("   Loading %s...", model.ID)
				if err := loadModel(workflow.Server, model); err != nil {
					fmt.Printf(" %s\n", theme.WarningStyle.Render("failed: "+err.Error()))
				} else {
					fmt.Printf(" %s\n", theme.SuccessStyle.Render("✓"))
				}
			}
		}
	}

	// Step 5: Run post-commands
	if len(workflow.PostCommands) > 0 {
		fmt.Printf("%s Running post-commands...\n", theme.DimTextStyle.Render("▶"))
		for _, cmdStr := range workflow.PostCommands {
			if workflowStartDry {
				fmt.Printf("   [dry-run] %s\n", cmdStr)
			} else {
				fmt.Printf("   $ %s\n", cmdStr)
				execCmd := exec.Command("bash", "-c", cmdStr)
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				if err := execCmd.Run(); err != nil {
					fmt.Printf("   %s %v\n", theme.WarningStyle.Render("warning:"), err)
				}
			}
		}
	}

	fmt.Println()
	fmt.Printf("%s Workflow '%s' started\n", theme.SuccessStyle.Render("*"), name)

	return nil
}

// runWorkflowStop stops the current workflow
func runWorkflowStop(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	workflow, err := cfg.GetActiveWorkflow()
	if err != nil {
		return fmt.Errorf("no active workflow to stop")
	}

	name := cfg.ActiveWorkflow
	fmt.Printf("%s Stopping workflow: %s\n", theme.InfoStyle.Render("*"), name)
	fmt.Println()

	// Unload models
	if len(workflow.Models) > 0 {
		fmt.Printf("%s Unloading models...\n", theme.DimTextStyle.Render("▶"))
		for _, model := range workflow.Models {
			if !model.Enabled {
				continue
			}
			fmt.Printf("   Unloading %s...", model.ID)
			if err := unloadModel(workflow.Server, model); err != nil {
				fmt.Printf(" %s\n", theme.WarningStyle.Render("skipped"))
			} else {
				fmt.Printf(" %s\n", theme.SuccessStyle.Render("✓"))
			}
		}
	}

	fmt.Println()
	fmt.Printf("%s Workflow '%s' stopped\n", theme.SuccessStyle.Render("*"), name)

	return nil
}

// startLLMServer starts the appropriate LLM server for the workflow
func startLLMServer(workflow *config.WorkflowProfile, dryRun bool) error {
	switch workflow.Server {
	case config.ServerOllama:
		if dryRun {
			fmt.Println("   [dry-run] ollama serve")
			return nil
		}
		// Check if ollama is already running
		checkCmd := exec.Command("pgrep", "-x", "ollama")
		if err := checkCmd.Run(); err == nil {
			fmt.Println("   Ollama already running")
			return nil
		}
		// Start ollama in background
		cmd := exec.Command("ollama", "serve")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start ollama: %w", err)
		}
		fmt.Println("   Ollama started")

	case config.ServerVLLM:
		if dryRun {
			fmt.Println("   [dry-run] vllm serve (would start with model config)")
			return nil
		}
		// vLLM needs to be started with a specific model
		// For now, just check if it's available
		if _, err := exec.LookPath("vllm"); err != nil {
			return fmt.Errorf("vLLM not found in PATH. Install with: pip install vllm")
		}
		fmt.Println("   vLLM available (start manually or use 'anime vllm serve')")

	case config.ServerTensorRT:
		if dryRun {
			fmt.Println("   [dry-run] tensorrt-llm server")
			return nil
		}
		fmt.Println("   TensorRT-LLM (start manually)")

	case config.ServerLlamaCpp:
		if dryRun {
			fmt.Println("   [dry-run] llama.cpp server")
			return nil
		}
		fmt.Println("   llama.cpp (start manually)")

	case config.ServerExllamaV2:
		if dryRun {
			fmt.Println("   [dry-run] exllamav2 server")
			return nil
		}
		fmt.Println("   ExLlamaV2 (start manually)")

	default:
		fmt.Printf("   Unknown server type: %s\n", workflow.Server)
	}

	return nil
}

// loadModel loads a model using the appropriate server
func loadModel(server config.LLMServerType, model config.ModelDeployment) error {
	switch server {
	case config.ServerOllama:
		// Pull model if not present, then run to warm it up
		cmd := exec.Command("ollama", "pull", model.ID)
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()

	case config.ServerVLLM:
		// vLLM loads models when the server starts
		// This is a no-op for now
		return nil

	default:
		return nil
	}
}

// unloadModel unloads a model from the server
func unloadModel(server config.LLMServerType, model config.ModelDeployment) error {
	switch server {
	case config.ServerOllama:
		// Ollama doesn't have an explicit unload, but we can stop the model
		// by sending a request to unload it
		cmd := exec.Command("ollama", "stop", model.ID)
		return cmd.Run()

	default:
		return nil
	}
}
