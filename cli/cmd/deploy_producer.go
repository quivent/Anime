package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var deployProducerCmd = &cobra.Command{
	Use:   "deploy-producer [server]",
	Short: "Deploy Producer CLI to a server",
	Long: `Deploy the Producer CLI to a remote server for conversational fine-tuning.

This command builds Producer for Linux and deploys it to the specified server.
It can optionally set up the required directories for Producer operation.

Examples:
  anime deploy-producer gpu-server-1
  anime deploy-producer gpu-server-1 --full
  anime deploy-producer gpu-server-1 --user root --path /opt/producer/bin`,
	Args: cobra.ExactArgs(1),
	RunE: runDeployProducer,
}

var (
	deployProducerUser string
	deployProducerPath string
	deployProducerFull bool
)

func init() {
	deployProducerCmd.Flags().StringVarP(&deployProducerUser, "user", "u", "ubuntu", "SSH user for deployment")
	deployProducerCmd.Flags().StringVarP(&deployProducerPath, "path", "p", "/usr/local/bin", "Installation path on server")
	deployProducerCmd.Flags().BoolVarP(&deployProducerFull, "full", "f", false, "Full deployment with directory setup")
	rootCmd.AddCommand(deployProducerCmd)
}

func runDeployProducer(cmd *cobra.Command, args []string) error {
	server := args[0]

	// Find producer source directory
	producerDir := filepath.Join(os.Getenv("HOME"), "anime", "producer")
	if _, err := os.Stat(producerDir); os.IsNotExist(err) {
		return fmt.Errorf("producer source not found at %s", producerDir)
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🎬 Deploying Producer CLI"))
	fmt.Println()

	// Step 1: Build for Linux
	fmt.Printf("  %s Building for linux/amd64...\n", theme.SymbolInfo)

	buildDir := filepath.Join(producerDir, "build")
	os.MkdirAll(buildDir, 0755)

	buildCmd := exec.Command("go", "build", "-trimpath",
		"-ldflags", "-X main.Version=0.1.0",
		"-o", filepath.Join(buildDir, "producer-linux-amd64"),
		"./cmd/producer")
	buildCmd.Dir = producerDir
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  %s Build failed: %s\n", theme.SymbolError, string(output))
		return err
	}
	fmt.Printf("  %s Built successfully\n", theme.SymbolSuccess)

	// Step 2: Copy binary to server
	fmt.Printf("  %s Copying to %s@%s...\n", theme.SymbolInfo, deployProducerUser, server)

	binaryPath := filepath.Join(buildDir, "producer-linux-amd64")
	scpCmd := exec.Command("scp", binaryPath, fmt.Sprintf("%s@%s:/tmp/producer", deployProducerUser, server))
	scpCmd.Stdout = os.Stdout
	scpCmd.Stderr = os.Stderr

	if err := scpCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	fmt.Printf("  %s Binary copied\n", theme.SymbolSuccess)

	// Step 3: Install on server
	fmt.Printf("  %s Installing to %s...\n", theme.SymbolInfo, deployProducerPath)

	installCommands := []string{
		fmt.Sprintf("sudo mv /tmp/producer %s/producer", deployProducerPath),
		fmt.Sprintf("sudo chmod +x %s/producer", deployProducerPath),
	}

	if deployProducerFull {
		// Add directory setup commands
		installCommands = append(installCommands,
			"sudo mkdir -p /var/producer/{model,lora,memory,checkpoints,logs,guidance}",
			fmt.Sprintf("sudo chown -R %s:%s /var/producer", deployProducerUser, deployProducerUser),
		)
	}

	sshCmd := exec.Command("ssh", fmt.Sprintf("%s@%s", deployProducerUser, server),
		strings.Join(installCommands, " && "))
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}
	fmt.Printf("  %s Installed successfully\n", theme.SymbolSuccess)

	// Step 4: Verify installation
	fmt.Printf("  %s Verifying installation...\n", theme.SymbolInfo)

	verifyCmd := exec.Command("ssh", fmt.Sprintf("%s@%s", deployProducerUser, server),
		fmt.Sprintf("%s/producer --help | head -5", deployProducerPath))
	verifyOutput, err := verifyCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  %s Verification warning: %v\n", theme.SymbolWarning, err)
	} else {
		fmt.Printf("  %s Verified: producer is accessible\n", theme.SymbolSuccess)
	}

	// Print summary
	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("  Deployment Complete"))
	fmt.Println()
	fmt.Printf("    Server:  %s\n", theme.InfoStyle.Render(server))
	fmt.Printf("    Binary:  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("%s/producer", deployProducerPath)))
	if deployProducerFull {
		fmt.Printf("    Data:    %s\n", theme.DimTextStyle.Render("/var/producer"))
	}
	fmt.Println()

	fmt.Println(theme.HeaderStyle.Render("  Next Steps"))
	fmt.Println()
	fmt.Printf("    %s ssh %s@%s\n", theme.DimTextStyle.Render("$"), deployProducerUser, server)
	if deployProducerFull {
		fmt.Printf("    %s producer wizard\n", theme.DimTextStyle.Render("$"))
	} else {
		fmt.Printf("    %s producer init --path /var/producer\n", theme.DimTextStyle.Render("$"))
	}
	fmt.Printf("    %s producer cluster validate\n", theme.DimTextStyle.Render("$"))
	fmt.Printf("    %s producer model load\n", theme.DimTextStyle.Render("$"))
	fmt.Println()

	// Show verification output if available
	if len(verifyOutput) > 0 {
		fmt.Println(theme.DimTextStyle.Render("  Binary verification:"))
		for _, line := range strings.Split(string(verifyOutput), "\n") {
			if line != "" {
				fmt.Printf("    %s\n", theme.DimTextStyle.Render(line))
			}
		}
		fmt.Println()
	}

	return nil
}
