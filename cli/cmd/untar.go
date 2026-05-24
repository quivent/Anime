package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var untarCmd = &cobra.Command{
	Use:   "untar <archive> [destination]",
	Short: "Extract a tar archive",
	Long: `Extract a tar archive (.tar, .tar.gz, .tgz, .tar.bz2, .tar.xz, .tar.zst).

If no destination is given, extracts to the current directory.

Examples:
  anime untar backup.tar.gz
  anime untar deploy.tgz ./output/
  anime untar data.tar.xz /tmp/data`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUntar,
}

func init() {
	rootCmd.AddCommand(untarCmd)
}

func runUntar(cmd *cobra.Command, args []string) error {
	archive := args[0]
	dest := "."
	if len(args) > 1 {
		dest = args[1]
	}

	// Validate archive exists
	info, err := os.Stat(archive)
	if err != nil {
		return fmt.Errorf("archive not found: %s", archive)
	}

	// Determine tar flags from extension
	flags := tarFlags(archive)
	if flags == nil {
		return fmt.Errorf("unsupported archive format: %s", filepath.Ext(archive))
	}

	// Create destination if needed
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("cannot create destination: %w", err)
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	fmt.Printf("  %s %s %s\n",
		theme.SymbolLoading,
		theme.HighlightStyle.Render(filepath.Base(archive)),
		theme.DimTextStyle.Render(fmt.Sprintf("(%.1f MB)", sizeMB)))
	fmt.Printf("  %s %s\n",
		theme.DimTextStyle.Render("→"),
		theme.InfoStyle.Render(dest))

	tarArgs := append(flags, archive, "-C", dest)
	tarCmd := exec.Command("tar", tarArgs...)
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr

	start := time.Now()
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	elapsed := time.Since(start)

	fmt.Printf("  %s Extracted in %s\n",
		theme.SuccessStyle.Render(theme.SymbolSuccess),
		theme.DimTextStyle.Render(elapsed.Round(time.Millisecond).String()))
	return nil
}

func tarFlags(archive string) []string {
	lower := strings.ToLower(archive)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return []string{"-xzf"}
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return []string{"-xjf"}
	case strings.HasSuffix(lower, ".tar.xz"), strings.HasSuffix(lower, ".txz"):
		return []string{"-xJf"}
	case strings.HasSuffix(lower, ".tar.zst"), strings.HasSuffix(lower, ".tzst"):
		return []string{"--zstd", "-xf"}
	case strings.HasSuffix(lower, ".tar"):
		return []string{"-xf"}
	default:
		return nil
	}
}
