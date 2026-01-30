package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/embeddb"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

const (
	githubTokenKey = "github_token"
	githubUserKey  = "github_user"
	githubHostKey  = "github_host"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Manage GitHub authentication stored in binary",
	Long: `Manage GitHub authentication embedded in the anime binary.

The anime binary can carry its own GitHub token, allowing it to authenticate
with GitHub without relying on the user's gh CLI or environment variables.

Examples:
  anime github import              # Import from current gh auth
  anime github show                # Show stored GitHub user
  anime github delete              # Remove embedded token
  anime github export              # Export token to gh CLI
  anime github status              # Check authentication status
`,
	Run: func(cmd *cobra.Command, args []string) {
		runGitHubStatus(cmd, args)
	},
}

var githubImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import GitHub auth from gh CLI",
	Long:  "Import the current GitHub authentication from the gh CLI into the binary",
	RunE:  runGitHubImport,
}

var githubShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show stored GitHub authentication",
	Long:  "Display the embedded GitHub user and token status",
	Run:   runGitHubShow,
}

var githubDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the embedded GitHub auth",
	Long:  "Remove GitHub authentication from the binary",
	RunE:  runGitHubDelete,
}

var githubExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export token to gh CLI",
	Long:  "Export the embedded GitHub token to authenticate the gh CLI",
	RunE:  runGitHubExport,
}

var githubStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check GitHub authentication status",
	Long:  "Verify the embedded GitHub token is valid",
	RunE:  runGitHubStatusCmd,
}

func init() {
	githubCmd.AddCommand(githubImportCmd)
	githubCmd.AddCommand(githubShowCmd)
	githubCmd.AddCommand(githubDeleteCmd)
	githubCmd.AddCommand(githubExportCmd)
	githubCmd.AddCommand(githubStatusCmd)
	rootCmd.AddCommand(githubCmd)
}

func runGitHubImport(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	// Check if token already exists
	if db.Get(githubTokenKey) != nil {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("GitHub auth already exists in the binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Use 'anime github delete' first to remove it"))
		fmt.Println(theme.DimTextStyle.Render("  Or 'anime github show' to view current auth"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Print(theme.InfoStyle.Render("Importing GitHub auth from gh CLI... "))

	// Get token from gh CLI
	tokenCmd := exec.Command("gh", "auth", "token")
	tokenOutput, err := tokenCmd.Output()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("Could not get token from gh CLI"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Make sure you're logged in:"))
		fmt.Println(theme.HighlightStyle.Render("    gh auth login"))
		fmt.Println()
		return nil
	}
	token := strings.TrimSpace(string(tokenOutput))

	// Get user info
	userCmd := exec.Command("gh", "api", "user", "--jq", ".login")
	userOutput, err := userCmd.Output()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("could not get user info: %w", err)
	}
	user := strings.TrimSpace(string(userOutput))

	// Get host (default to github.com)
	host := "github.com"
	statusCmd := exec.Command("gh", "auth", "status", "--show-token")
	statusOutput, _ := statusCmd.CombinedOutput()
	statusStr := string(statusOutput)
	// Parse host from status output
	for _, line := range strings.Split(statusStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ".com") || strings.HasSuffix(line, ".io") {
			host = line
			break
		}
	}

	// Store in embedded database
	db.Set(githubTokenKey, []byte(token))
	db.Set(githubUserKey, []byte(user))
	db.Set(githubHostKey, []byte(host))

	if err := db.Save(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("done"))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("GitHub authentication imported"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Host:"), theme.HighlightStyle.Render(host))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("User:"), theme.HighlightStyle.Render(user))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Token:"), theme.DimTextStyle.Render(maskToken(token)))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The token is now embedded in the binary"))
	fmt.Println(theme.DimTextStyle.Render("  Use 'anime push' to deploy to other servers"))
	fmt.Println()

	return nil
}

func runGitHubShow(cmd *cobra.Command, args []string) {
	runGitHubStatus(cmd, args)
}

func runGitHubStatus(cmd *cobra.Command, args []string) {
	db, err := embeddb.DB()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("Failed to access embedded database: " + err.Error()))
		return
	}

	tokenData := db.Get(githubTokenKey)
	if tokenData == nil {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("No GitHub auth embedded in binary"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Import from gh CLI:"))
		fmt.Println(theme.HighlightStyle.Render("    anime github import"))
		fmt.Println()
		return
	}

	user := string(db.Get(githubUserKey))
	host := string(db.Get(githubHostKey))
	if host == "" {
		host = "github.com"
	}

	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Embedded GitHub Authentication:"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Host:"), theme.HighlightStyle.Render(host))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("User:"), theme.HighlightStyle.Render(user))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Token:"), theme.DimTextStyle.Render(maskToken(string(tokenData))))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Export to gh CLI on another machine:"))
	fmt.Println(theme.HighlightStyle.Render("    anime github export"))
	fmt.Println()
}

func runGitHubStatusCmd(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	tokenData := db.Get(githubTokenKey)
	if tokenData == nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("No GitHub auth embedded in binary"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Print(theme.InfoStyle.Render("Verifying GitHub token... "))

	// Verify token by making an API call
	token := string(tokenData)
	verifyCmd := exec.Command("gh", "api", "user", "--jq", ".login")
	verifyCmd.Env = append(os.Environ(), "GH_TOKEN="+token)
	output, err := verifyCmd.Output()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("invalid"))
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("Token is invalid or expired"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Re-import with:"))
		fmt.Println(theme.HighlightStyle.Render("    anime github delete && anime github import"))
		fmt.Println()
		return nil
	}

	user := strings.TrimSpace(string(output))
	fmt.Println(theme.SuccessStyle.Render("valid"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Authenticated as:"), theme.HighlightStyle.Render(user))
	fmt.Println()

	return nil
}

func runGitHubDelete(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	if db.Get(githubTokenKey) == nil {
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("No GitHub auth to delete"))
		fmt.Println()
		return nil
	}

	db.Delete(githubTokenKey)
	db.Delete(githubUserKey)
	db.Delete(githubHostKey)

	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("GitHub auth deleted from binary"))
	fmt.Println()

	return nil
}

func runGitHubExport(cmd *cobra.Command, args []string) error {
	db, err := embeddb.DB()
	if err != nil {
		return fmt.Errorf("failed to access embedded database: %w", err)
	}

	tokenData := db.Get(githubTokenKey)
	if tokenData == nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("No GitHub auth embedded in binary"))
		fmt.Println()
		return nil
	}

	token := string(tokenData)
	host := string(db.Get(githubHostKey))
	if host == "" {
		host = "github.com"
	}

	fmt.Println()
	fmt.Print(theme.InfoStyle.Render("Exporting GitHub auth to gh CLI... "))

	// Use gh auth login with token via stdin
	loginCmd := exec.Command("gh", "auth", "login", "--with-token", "--hostname", host)
	loginCmd.Stdin = strings.NewReader(token)
	output, err := loginCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Error: " + string(output)))
		fmt.Println()
		return nil
	}

	fmt.Println(theme.SuccessStyle.Render("done"))
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("GitHub CLI is now authenticated"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Verify with:"))
	fmt.Println(theme.HighlightStyle.Render("    gh auth status"))
	fmt.Println()

	return nil
}

// maskToken returns a masked version of the token for display
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// GetEmbeddedGitHubToken returns the embedded GitHub token if available
func GetEmbeddedGitHubToken() (string, error) {
	db, err := embeddb.DB()
	if err != nil {
		return "", err
	}

	tokenData := db.Get(githubTokenKey)
	if tokenData == nil {
		return "", fmt.Errorf("no embedded GitHub token")
	}

	return string(tokenData), nil
}

// HasEmbeddedGitHubToken returns true if the binary has an embedded GitHub token
func HasEmbeddedGitHubToken() bool {
	db, err := embeddb.DB()
	if err != nil {
		return false
	}
	return db.Get(githubTokenKey) != nil
}

// SetupGitHubEnv returns environment variables with the embedded GitHub token
func SetupGitHubEnv() []string {
	token, err := GetEmbeddedGitHubToken()
	if err != nil {
		return os.Environ()
	}

	env := os.Environ()
	env = append(env, "GH_TOKEN="+token)
	env = append(env, "GITHUB_TOKEN="+token)
	return env
}

// WriteGitHubNetrc writes a .netrc file with GitHub credentials for git operations
func WriteGitHubNetrc() (string, func(), error) {
	token, err := GetEmbeddedGitHubToken()
	if err != nil {
		return "", nil, err
	}

	db, _ := embeddb.DB()
	user := string(db.Get(githubUserKey))
	if user == "" {
		user = "git"
	}

	host := string(db.Get(githubHostKey))
	if host == "" {
		host = "github.com"
	}

	// Write to temp file
	tmpDir := os.TempDir()
	netrcPath := filepath.Join(tmpDir, "anime_netrc")

	content := fmt.Sprintf("machine %s login %s password %s\n", host, user, token)
	if err := os.WriteFile(netrcPath, []byte(content), 0600); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		os.Remove(netrcPath)
	}

	return netrcPath, cleanup, nil
}
