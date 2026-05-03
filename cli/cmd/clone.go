package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/gh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <repo>",
	Short: "Clone a repository from GitHub",
	Long: `Clone a git repository. Supports multiple formats:

  anime clone anime                        # Clone the anime CLI repo itself
  anime clone user/repo                    # GitHub shorthand
  anime clone https://github.com/user/repo # Full URL
  anime clone git@github.com:user/repo.git # SSH URL

The embedded GitHub token is used automatically for HTTPS clones.`,
	Args: cobra.ExactArgs(1),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	repo := args[0]

	// Special case: "anime clone anime" clones this project
	if repo == "anime" {
		repo = "https://github.com/joshkornreich/anime.git"
	} else if !strings.Contains(repo, "://") && !strings.Contains(repo, "@") {
		// GitHub shorthand: user/repo or just repo-name (assumes joshkornreich/)
		if strings.Count(repo, "/") == 1 {
			repo = "https://github.com/" + repo + ".git"
		} else if !strings.Contains(repo, "/") {
			repo = "https://github.com/joshkornreich/" + repo + ".git"
		}
	}

	repoName := filepath.Base(strings.TrimSuffix(repo, ".git"))

	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repository:"), theme.HighlightStyle.Render(repo))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Directory:"), theme.InfoStyle.Render("./"+repoName))
	fmt.Println()

	// Build git clone command with token auth if available
	cloneArgs := []string{"clone", repo}
	cloneExec := exec.Command("git", cloneArgs...)
	cloneExec.Stdout = os.Stdout
	cloneExec.Stderr = os.Stderr

	// Inject GH_TOKEN for HTTPS auth if embedded and not already set
	if strings.Contains(repo, "https://") && gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cloneExec.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}

	if err := cloneExec.Run(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("  Clone failed"))
		return fmt.Errorf("clone failed: %w", err)
	}

	fmt.Println(theme.SuccessStyle.Render("  ✓ Cloned " + repoName))
	fmt.Println()

	return nil
}
