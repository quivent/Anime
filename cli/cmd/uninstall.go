package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package>",
	Short: "Uninstall packages from remote server",
	Long: `Uninstall packages from a remote server.

Available packages:
  tensorflow    Remove TensorFlow and all related packages (fixes NumPy conflicts)

Examples:
  anime uninstall tensorflow                    # Remove from default server
  anime uninstall tensorflow --server lambda    # Remove from specific server
  anime uninstall tensorflow --local            # Remove from local machine
`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

var (
	uninstallServer string
	uninstallLocal  bool
)

func init() {
	uninstallCmd.Flags().StringVarP(&uninstallServer, "server", "s", "", "Target server (default: lambda)")
	uninstallCmd.Flags().BoolVar(&uninstallLocal, "local", false, "Run on local machine instead of remote")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	pkg := strings.ToLower(args[0])

	switch pkg {
	case "tensorflow", "tf":
		return uninstallTensorFlow()
	default:
		return fmt.Errorf("unknown package: %s\n\nAvailable packages:\n  tensorflow    Remove TensorFlow (fixes NumPy conflicts)", pkg)
	}
}

func uninstallTensorFlow() error {
	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🧹 Removing TensorFlow"))
	fmt.Println()

	script := `
set -e

echo "==> Removing apt packages..."
sudo apt-get remove --purge -y python3-tensorflow 'tensorflow*' 2>/dev/null || true
sudo apt-get autoremove -y 2>/dev/null || true

echo "==> Removing pip packages (user)..."
pip3 uninstall -y \
    tensorflow tensorflow-cpu tensorflow-gpu tensorflow-intel tensorflow-macos \
    tensorflow-io tensorflow-io-gcs-filesystem \
    tf-keras keras \
    tensorboard tensorboard-data-server tensorboard-plugin-wit \
    tensorflow-estimator tensorflow-hub tensorflow-metadata \
    tensorflow-serving-api tensorflow-text tensorflow-addons \
    tensorflow-datasets tensorflow-probability \
    2>/dev/null || true

echo "==> Removing pip packages (system)..."
sudo pip3 uninstall -y tensorflow tensorflow-cpu tensorflow-gpu 2>/dev/null || true

echo "==> Cleaning pip cache..."
pip3 cache purge 2>/dev/null || true

echo "==> Verifying removal..."
if python3 -c "import tensorflow" 2>/dev/null; then
    echo "WARNING: TensorFlow still importable, trying harder..."
    # Find and remove any remaining tensorflow directories
    python3 -c "import tensorflow; print(tensorflow.__file__)" 2>/dev/null | xargs -I{} dirname {} | xargs -I{} rm -rf {} 2>/dev/null || true
    pip3 uninstall -y tensorflow 2>/dev/null || true
fi

# Final check
if python3 -c "import tensorflow" 2>/dev/null; then
    echo "ERROR: TensorFlow still present"
    exit 1
else
    echo "SUCCESS: TensorFlow completely removed"
fi
`

	if uninstallLocal {
		// Run locally
		fmt.Println(theme.DimTextStyle.Render("  Running locally..."))
		fmt.Println()

		shellCmd := exec.Command("bash", "-c", script)
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr
		if err := shellCmd.Run(); err != nil {
			return fmt.Errorf("failed to remove TensorFlow: %w", err)
		}
	} else {
		// Run on remote server
		server := uninstallServer
		if server == "" {
			server = "lambda"
		}

		target, err := parseServerTarget(server)
		if err != nil {
			return err
		}

		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
		fmt.Println()

		args := buildSSHArgs(target, script)
		sshCmd := exec.Command("ssh", args...)
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			return fmt.Errorf("failed to remove TensorFlow: %w", err)
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ TensorFlow removed"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  NumPy 2.x compatible packages (PyTorch, vLLM) should now work correctly."))
	fmt.Println()

	return nil
}
