package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/claude"
	"github.com/joshkornreich/anime/internal/gh"
	"github.com/joshkornreich/anime/internal/hf"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// EmbeddedFile represents a file that has been embedded
type EmbeddedFile struct {
	OriginalPath string    `json:"original_path"`
	StoredName   string    `json:"stored_name"`
	Size         int64     `json:"size"`
	Hash         string    `json:"hash"`
	EmbeddedAt   time.Time `json:"embedded_at"`
	IsDirectory  bool      `json:"is_directory"`
}

// EmbeddedManifest tracks all embedded files
type EmbeddedManifest struct {
	Files map[string]EmbeddedFile `json:"files"`
}

var embedCmd = &cobra.Command{
	Use:   "embed FILEPATH",
	Short: "Embed a file or directory for deployment",
	Long: `Embed a file or directory into anime's embedded filesystem.

The embedded files are stored in ~/.anime/embedded/ and will be automatically
included when you push to a remote server. You can extract them on the remote
server using 'anime extract --embedded FILEPATH'.

Examples:
  anime embed config.yaml              # Embed a single file
  anime embed ./models                 # Embed a directory
  anime embed ~/datasets/training.csv  # Embed file from home directory

The embedded files are tracked in a manifest and can be listed with:
  anime extract --list
`,
	Args: cobra.ExactArgs(1),
	RunE: runEmbed,
}

func init() {
	rootCmd.AddCommand(embedCmd)
	embedCmd.AddCommand(embedTokenCmd)
	embedTokenCmd.AddCommand(embedTokenListCmd, embedTokenSetCmd, embedTokenClearCmd)
}

// ─── anime embed token ─────────────────────────────────────────────
//
// Bake API tokens (HuggingFace, GitHub, Claude) into the binary by
// rewriting the corresponding `internal/<pkg>/token.go` source file. The
// next `go build` picks them up via the `EmbeddedToken` constant.
//
// Locating the source tree:
//   1. The `BuildDir` ldflag set by `make build` / `make install` (most
//      reliable — points at the original source path).
//   2. Walk up from the current working directory looking for a go.mod
//      with `module github.com/joshkornreich/anime` (handy when the user
//      is `cd`'d inside the repo).
// If neither works we error out with explicit guidance.
//
// Security: this rewrites Go source files. The token then lives in plain
// text in the binary — anyone with the binary can extract it. Suitable
// for self-distribution to your own boxes; not for public release.

var embedTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Embed API tokens (hf | gh | claude) into the anime binary",
	Long: `Bake API tokens into the anime binary so a fresh box just works
without per-host login. Three slots: hf (HuggingFace), gh (GitHub),
claude (Anthropic OAuth). The values are written into Go source
files and picked up on the next ` + "`go build`" + `.

Subcommands:
  anime embed token list                  # show what's embedded (truncated)
  anime embed token set <type> <value>    # set hf | gh | claude
  anime embed token clear <type>          # blank out a slot

Examples:
  anime embed token set hf hf_xxxx...
  anime embed token set gh ghp_xxxx...
  anime embed token set claude sk-ant-oat01-xxxx...
  anime embed token list

WARNING: tokens land in plain text in the compiled binary. Only do this
for binaries you control end-to-end. NEVER for a public release.`,
	Run: func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
}

var embedTokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show currently embedded tokens (truncated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Println(theme.RenderBanner("EMBEDDED TOKENS"))
		fmt.Println()
		row := func(name, val string) {
			label := theme.HighlightStyle.Render(fmt.Sprintf("  %-8s", name))
			if val == "" {
				fmt.Printf("%s  %s\n", label, theme.DimTextStyle.Render("(none — using runtime/env auth only)"))
				return
			}
			fmt.Printf("%s  %s\n", label, theme.SuccessStyle.Render(maskEmbeddedToken(val)))
		}
		row("hf", hf.GetToken())
		row("gh", gh.GetToken())
		row("claude", claude.GetAccessToken())
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Set with: anime embed token set <type> <value>"))
		fmt.Println(theme.DimTextStyle.Render("  Tokens take effect after rebuild: make build (or go build)"))
		fmt.Println()
		return nil
	},
}

var embedTokenSetCmd = &cobra.Command{
	Use:   "set <hf|gh|claude> <value>",
	Short: "Embed a token by rewriting its source file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setEmbeddedToken(args[0], args[1])
	},
}

var embedTokenClearCmd = &cobra.Command{
	Use:   "clear <hf|gh|claude>",
	Short: "Blank out an embedded token (revert to runtime auth)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setEmbeddedToken(args[0], "")
	},
}

// maskEmbeddedToken redacts the middle of a token for safe display.
func maskEmbeddedToken(t string) string {
	if len(t) <= 12 {
		return strings.Repeat("·", len(t))
	}
	return t[:8] + strings.Repeat("·", 16) + t[len(t)-4:]
}

// findSourceTree returns the path to the anime cli/ directory containing
// internal/hf/, internal/gh/, etc. Tries BuildDir ldflag first, then walks
// up from cwd. Returns a clear error if neither finds the tree.
func findSourceTree() (string, error) {
	candidates := []string{}
	if BuildDir != "" {
		candidates = append(candidates, BuildDir)
	}
	if cwd, err := os.Getwd(); err == nil {
		// Walk up looking for go.mod with our module name.
		for d := cwd; d != "/" && d != ""; d = filepath.Dir(d) {
			data, err := os.ReadFile(filepath.Join(d, "go.mod"))
			if err == nil && strings.Contains(string(data), "github.com/joshkornreich/anime") {
				candidates = append(candidates, d)
				break
			}
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "internal", "hf", "token.go")); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf(
		"could not locate the anime source tree.\n" +
			"  BuildDir ldflag was: %q (set via the Makefile)\n" +
			"  Tried: %v\n" +
			"  Fix: cd into the anime cli/ directory and re-run, or build via: make build",
		BuildDir, candidates,
	)
}

// setEmbeddedToken rewrites the appropriate token source file.
//
//	hf      → internal/hf/token.go        : const EmbeddedToken = "..."
//	gh      → internal/gh/token.go        : const EmbeddedToken = "..."
//	claude  → internal/claude/auth.go     : AccessToken: "..."
//
// We do regex-replace rather than AST rewrite because the surface is tiny
// (3 files, one literal each) and a regex is easier to audit + reverse.
func setEmbeddedToken(kind, value string) error {
	root, err := findSourceTree()
	if err != nil {
		return err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)

	var path string
	var pattern *regexp.Regexp
	var replacement string
	var label string

	switch kind {
	case "hf":
		path = filepath.Join(root, "internal", "hf", "token.go")
		pattern = regexp.MustCompile(`const EmbeddedToken = "[^"]*"`)
		replacement = fmt.Sprintf(`const EmbeddedToken = %q`, value)
		label = "HuggingFace"
	case "gh", "github":
		path = filepath.Join(root, "internal", "gh", "token.go")
		pattern = regexp.MustCompile(`const EmbeddedToken = "[^"]*"`)
		replacement = fmt.Sprintf(`const EmbeddedToken = %q`, value)
		label = "GitHub"
	case "claude", "anthropic":
		path = filepath.Join(root, "internal", "claude", "auth.go")
		pattern = regexp.MustCompile(`AccessToken:\s*"[^"]*"`)
		replacement = fmt.Sprintf(`AccessToken:      %q`, value)
		label = "Claude OAuth access token"
	default:
		return fmt.Errorf("unknown token type %q (want: hf, gh, claude)", kind)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if !pattern.Match(data) {
		return fmt.Errorf("could not find the %s token literal in %s — has the file shape changed?", label, path)
	}
	updated := pattern.ReplaceAll(data, []byte(replacement))
	if string(updated) == string(data) {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  %s token already matches; no changes written.", label)))
		return nil
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Println()
	if value == "" {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Cleared %s token in %s", label, path)))
	} else {
		fmt.Printf("%s %s %s\n",
			theme.SuccessStyle.Render("✓ Embedded"),
			theme.HighlightStyle.Render(label),
			theme.DimTextStyle.Render("→ "+maskEmbeddedToken(value)+"  ("+path+")"))
	}
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Rebuild for the change to take effect:"))
	fmt.Println(theme.HighlightStyle.Render("    make build       # or"))
	fmt.Println(theme.HighlightStyle.Render("    go build         # from cli/"))
	fmt.Println()
	return nil
}

func runEmbed(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]

	// Expand home directory if needed
	if sourcePath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		sourcePath = filepath.Join(home, sourcePath[1:])
	}

	// Make path absolute
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	// Check if source exists
	sourceInfo, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(theme.ErrorStyle.Render("❌ File or directory not found"))
			fmt.Println()
			fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(absSourcePath))
			fmt.Println()
			return fmt.Errorf("source not found: %s", absSourcePath)
		}
		return fmt.Errorf("failed to access source: %w", err)
	}

	fmt.Println(theme.InfoStyle.Render("📦 Embedding file for deployment"))
	fmt.Println()
	fmt.Printf("  Source: %s\n", theme.HighlightStyle.Render(absSourcePath))
	if sourceInfo.IsDir() {
		fmt.Printf("  Type:   %s\n", theme.InfoStyle.Render("directory"))
	} else {
		fmt.Printf("  Type:   %s\n", theme.InfoStyle.Render("file"))
		fmt.Printf("  Size:   %s\n", theme.DimTextStyle.Render(formatBytesSize(sourceInfo.Size())))
	}
	fmt.Println()

	// Create embedded directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	embeddedDir := filepath.Join(home, ".anime", "embedded")
	if err := os.MkdirAll(embeddedDir, 0755); err != nil {
		return fmt.Errorf("failed to create embedded directory: %w", err)
	}

	// Generate unique name for the embedded file
	hash := sha256.New()
	hash.Write([]byte(absSourcePath))
	hash.Write([]byte(time.Now().String()))
	hashStr := hex.EncodeToString(hash.Sum(nil))[:16]

	storedName := fmt.Sprintf("%s-%s.tar.gz", filepath.Base(absSourcePath), hashStr)
	tarPath := filepath.Join(embeddedDir, storedName)

	// Create tar.gz archive
	fmt.Print(theme.DimTextStyle.Render("▶ Creating archive... "))
	totalSize, fileHash, err := createTarGz(absSourcePath, tarPath)
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to create archive: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	// Update manifest
	fmt.Print(theme.DimTextStyle.Render("▶ Updating manifest... "))
	if err := updateManifest(absSourcePath, storedName, totalSize, fileHash, sourceInfo.IsDir()); err != nil {
		fmt.Println(theme.ErrorStyle.Render("✗"))
		return fmt.Errorf("failed to update manifest: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓"))

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ File embedded successfully!"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  The file will be automatically included when you push to a server:"))
	fmt.Println(theme.HighlightStyle.Render("    anime push"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Extract on remote server with:"))
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("    anime extract --embedded %s", filepath.Base(absSourcePath))))
	fmt.Println()

	return nil
}

func createTarGz(sourcePath, tarPath string) (int64, string, error) {
	// Create tar.gz file
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return 0, "", err
	}
	defer tarFile.Close()

	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	var totalSize int64
	hash := sha256.New()

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return 0, "", err
	}

	if sourceInfo.IsDir() {
		// Add directory
		err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip hidden files and directories
			if strings.HasPrefix(info.Name(), ".") && path != sourcePath {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			// Calculate relative path
			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.Join(filepath.Base(sourcePath), relPath)

			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// Copy file content
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			n, err := io.Copy(io.MultiWriter(tarWriter, hash), file)
			totalSize += n
			return err
		})
	} else {
		// Add single file
		header, err := tar.FileInfoHeader(sourceInfo, "")
		if err != nil {
			return 0, "", err
		}
		header.Name = filepath.Base(sourcePath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return 0, "", err
		}

		file, err := os.Open(sourcePath)
		if err != nil {
			return 0, "", err
		}
		defer file.Close()

		totalSize, err = io.Copy(io.MultiWriter(tarWriter, hash), file)
		if err != nil {
			return 0, "", err
		}
	}

	return totalSize, hex.EncodeToString(hash.Sum(nil)), err
}

func updateManifest(originalPath, storedName string, size int64, hash string, isDir bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(home, ".anime", "embedded", "manifest.json")

	// Load existing manifest or create new one
	manifest := &EmbeddedManifest{
		Files: make(map[string]EmbeddedFile),
	}

	if data, err := os.ReadFile(manifestPath); err == nil {
		if err := json.Unmarshal(data, manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}
	}

	// Add or update entry
	manifest.Files[filepath.Base(originalPath)] = EmbeddedFile{
		OriginalPath: originalPath,
		StoredName:   storedName,
		Size:         size,
		Hash:         hash,
		EmbeddedAt:   time.Now(),
		IsDirectory:  isDir,
	}

	// Save manifest
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

func formatBytesSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
