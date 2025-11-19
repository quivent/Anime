package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/installer"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [server-name]",
	Short: "Check status of a Lambda server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		serverName := args[0]
		server, err := cfg.GetServer(serverName)
		if err != nil {
			return fmt.Errorf("server %s not found", serverName)
		}

		fmt.Printf("Connecting to %s...\n", server.Name)

		client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		inst := installer.New(client)

		fmt.Println("\nSystem Information:")
		info, err := inst.GetSystemInfo()
		if err != nil {
			return err
		}

		fmt.Printf("  OS: %s\n", info["os"])
		fmt.Printf("  Architecture: %s\n", info["arch"])
		fmt.Printf("  Kernel: %s\n", info["kernel"])
		fmt.Printf("  GPU: %s\n", info["gpu"])
		fmt.Printf("  Free Disk: %s\n", info["disk_free"])
		fmt.Printf("  Free Memory: %s\n", info["mem_free"])

		fmt.Println("\nInstalled Components:")

		// Check for various components
		checks := map[string]string{
			"Python":     "python3 --version 2>/dev/null",
			"Node.js":    "node --version 2>/dev/null",
			"Docker":     "docker --version 2>/dev/null",
			"NVIDIA":     "nvidia-smi --version 2>/dev/null | head -1",
			"CUDA":       "nvcc --version 2>/dev/null | grep release",
			"PyTorch":    "python3 -c 'import torch; print(torch.__version__)' 2>/dev/null",
			"Ollama":     "ollama --version 2>/dev/null",
			"ComfyUI":    "[ -d ~/ComfyUI ] && echo 'Installed' || echo 'Not found'",
			"Claude Code": "claude-code --version 2>/dev/null",
		}

		for name, cmd := range checks {
			output, err := client.RunCommand(cmd)
			if err != nil || output == "Not found" {
				fmt.Printf("  %s: ✗ Not installed\n", name)
			} else {
				fmt.Printf("  %s: ✓ %s\n", name, output)
			}
		}

		// List Ollama models if installed
		output, err := client.RunCommand("ollama list 2>/dev/null")
		if err == nil && output != "" {
			fmt.Println("\nOllama Models:")
			fmt.Println(output)
		}

		return nil
	},
}
