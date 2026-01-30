package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/embeddb"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

const (
	keyPrivate = "ssh_private_key"
	keyPublic  = "ssh_public_key"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage SSH key stored in binary",
	Long: `Manage the SSH key embedded in the anime binary.

The anime binary can carry its own SSH key, allowing it to authenticate
with servers without relying on the user's SSH config.

Examples:
  anime key generate                    # Generate new keypair
  anime key show                        # Show public key
  anime key register lambda             # Register key with server
  anime key register ubuntu@10.0.0.5    # Register with specific server
  anime key export                      # Export private key to file
  anime key delete                      # Remove embedded key
`,
	Run: func(cmd *cobra.Command, args []string) {
		runKeyShow(cmd, args)
	},
}

var keyGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new SSH keypair",
	Long:  "Generate a new Ed25519 SSH keypair and store it in the binary",
	RunE:  runKeyGenerate,
}

var keyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the public key",
	Long:  "Display the embedded public key for copying to servers",
	Run:   runKeyShow,
}

var keyRegisterCmd = &cobra.Command{
	Use:   "register <server>",
	Short: "Register key with a server",
	Long: `Add the embedded public key to a server's authorized_keys.

This requires existing SSH access to the server (password or other key).
After registration, anime can connect using its embedded key.

Examples:
  anime key register lambda              # Use alias
  anime key register ubuntu@10.0.0.5     # Direct address
`,
	Args: cobra.ExactArgs(1),
	RunE: runKeyRegister,
}

var keyExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export private key to file",
	Long:  "Export the embedded private key to a file for backup or external use",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runKeyExport,
}

var keyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the embedded keypair",
	Long:  "Remove the SSH keypair from the binary",
	RunE:  runKeyDelete,
}

var keyImportCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import a private key from file",
	Long:  "Import an existing SSH private key into the binary",
	Args:  cobra.ExactArgs(1),
	RunE:  runKeyImport,
}

func init() {
	keyCmd.AddCommand(keyGenerateCmd)
	keyCmd.AddCommand(keyShowCmd)
	keyCmd.AddCommand(keyRegisterCmd)
	keyCmd.AddCommand(keyExportCmd)
	keyCmd.AddCommand(keyDeleteCmd)
	keyCmd.AddCommand(keyImportCmd)
	rootCmd.AddCommand(keyCmd)
}

func runKeyGenerate(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	// Check if key already exists
	if db.Get(keyPrivate) != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("A key already exists in the binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Use 'anime key delete' first to remove it"))
		fmt.Println(theme.DimTextStyle.Render("  Or 'anime key show' to view the public key"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Print(theme.InfoStyle.Render("Generating Ed25519 keypair... "))

	// Generate Ed25519 keypair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Convert to SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Encode private key to PEM
	privKeyPEM, err := ssh.MarshalPrivateKey(privKey, "anime embedded key")
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Store in embedded database
	db.Set(keyPrivate, pem.EncodeToMemory(privKeyPEM))
	db.Set(keyPublic, ssh.MarshalAuthorizedKey(sshPubKey))

	if err := db.Save(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("done"))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("SSH keypair generated and embedded"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Public key:"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPubKey)))))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Register with a server:"))
	fmt.Println(theme.HighlightStyle.Render("    anime key register <server>"))
	fmt.Println()

	return nil
}

func runKeyShow(cmd *cobra.Command, args []string) {
	db, err := embeddb.DB()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to access embedded database: " + err.Error()))
		return
	}

	pubKeyData := db.Get(keyPublic)
	if pubKeyData == nil {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("No SSH key embedded in binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Generate one with:"))
		fmt.Println(theme.HighlightStyle.Render("    anime key generate"))
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Embedded SSH Public Key:"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(string(pubKeyData))))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Copy this key to ~/.ssh/authorized_keys on your servers"))
	fmt.Println(theme.DimTextStyle.Render("  Or use: ") + theme.HighlightStyle.Render("anime key register <server>"))
	fmt.Println()
}

func runKeyRegister(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	pubKeyData := db.Get(keyPublic)
	if pubKeyData == nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("No SSH key embedded in binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Generate one first:"))
		fmt.Println(theme.HighlightStyle.Render("    anime key generate"))
		fmt.Println()
		return nil
	}

	server := args[0]

	// Resolve server alias
	target, err := parseServerTarget(server)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%s Registering key with %s...\n", theme.InfoStyle.Render(""), theme.HighlightStyle.Render(target))
	fmt.Println()

	// SSH command to add key to authorized_keys
	pubKey := strings.TrimSpace(string(pubKeyData))
	registerScript := fmt.Sprintf(`
		mkdir -p ~/.ssh
		chmod 700 ~/.ssh
		touch ~/.ssh/authorized_keys
		chmod 600 ~/.ssh/authorized_keys

		# Check if key already exists
		if grep -qF "%s" ~/.ssh/authorized_keys 2>/dev/null; then
			echo "KEY_EXISTS"
		else
			echo "%s" >> ~/.ssh/authorized_keys
			echo "KEY_ADDED"
		fi
	`, pubKey, pubKey)

	sshCmd := exec.Command("ssh", target, registerScript)
	sshCmd.Stdin = os.Stdin
	output, err := sshCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to register key"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Error: " + err.Error()))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  Make sure you can SSH to the server:"))
		fmt.Println(theme.HighlightStyle.Render("    ssh " + target))
		fmt.Println()
		return nil
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "KEY_EXISTS") {
		fmt.Println(theme.SuccessStyle.Render("Key already registered with this server"))
	} else if strings.Contains(outputStr, "KEY_ADDED") {
		fmt.Println(theme.SuccessStyle.Render("Key registered successfully"))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The anime binary can now SSH to this server using its embedded key"))
	fmt.Println()

	return nil
}

func runKeyExport(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	privKeyData := db.Get(keyPrivate)
	if privKeyData == nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("No SSH key embedded in binary"))
		fmt.Println()
		return nil
	}

	// Determine output path
	outputPath := "anime_key"
	if len(args) > 0 {
		outputPath = args[0]
	}

	// Write private key
	if err := os.WriteFile(outputPath, privKeyData, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key
	pubKeyData := db.Get(keyPublic)
	if pubKeyData != nil {
		pubPath := outputPath + ".pub"
		os.WriteFile(pubPath, pubKeyData, 0644)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Key exported"))
	fmt.Println()
	fmt.Printf("  Private: %s\n", theme.HighlightStyle.Render(outputPath))
	fmt.Printf("  Public:  %s\n", theme.HighlightStyle.Render(outputPath+".pub"))
	fmt.Println()

	return nil
}

func runKeyDelete(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	if db.Get(keyPrivate) == nil {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("No SSH key to delete"))
		fmt.Println()
		return nil
	}

	db.Delete(keyPrivate)
	db.Delete(keyPublic)

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("SSH key deleted from binary"))
	fmt.Println()

	return nil
}

func runKeyImport(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	// Check if key already exists
	if db.Get(keyPrivate) != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("A key already exists in the binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Use 'anime key delete' first to remove it"))
		fmt.Println()
		return nil
	}

	keyPath := args[0]

	// Read private key
	privKeyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse to validate
	privKey, err := ssh.ParsePrivateKey(privKeyData)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// Generate public key
	pubKey := privKey.PublicKey()
	pubKeyData := ssh.MarshalAuthorizedKey(pubKey)

	// Store
	db.Set(keyPrivate, privKeyData)
	db.Set(keyPublic, pubKeyData)

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("SSH key imported"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Public key:"))
	fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.TrimSpace(string(pubKeyData))))
	fmt.Println()

	return nil
}

// GetEmbeddedSSHKeyPath writes the embedded key to a temp file and returns its path
// This is used by push and other commands that need to SSH with the embedded key
func GetEmbeddedSSHKeyPath() (string, func(), error) {
	db, err := embeddb.DB()
	if err != nil {
		return "", nil, err
	}

	privKeyData := db.Get(keyPrivate)
	if privKeyData == nil {
		return "", nil, fmt.Errorf("no embedded SSH key")
	}

	// Write to temp file
	tmpDir := os.TempDir()
	keyPath := filepath.Join(tmpDir, "anime_ssh_key")

	if err := os.WriteFile(keyPath, privKeyData, 0600); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		os.Remove(keyPath)
	}

	return keyPath, cleanup, nil
}

// HasEmbeddedSSHKey returns true if the binary has an embedded SSH key
func HasEmbeddedSSHKey() bool {
	db, err := embeddb.DB()
	if err != nil {
		return false
	}
	return db.Get(keyPrivate) != nil
}
