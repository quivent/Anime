package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Manage asset collections",
	Long:  "Create, list, and manage collections of images, videos, and other assets for workflow processing",
	Run:   runCollectionHelp,
}

var collectionCreateCmd = &cobra.Command{
	Use:     "create <name> <path>",
	Aliases: []string{"add"},
	Short:   "Create a new asset collection",
	Long: `Create a new asset collection from a directory of files.

The collection type (image/video/mixed) will be auto-detected based on file extensions.

Examples:
  anime collection create photos ~/datasets/photos
  anime collection add photos ~/datasets/photos
  anime collection create renders /mnt/renders
  anime collection create vacation ~/Pictures/vacation-2024`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing collection name"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection create <name> [path]"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Arguments:"))
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("name"), theme.DimTextStyle.Render("Collection name (e.g., 'photos', 'renders')"))
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("path"), theme.DimTextStyle.Render("(Optional) Path to directory - defaults to ./<name>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Examples:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection add mar"))
			fmt.Println(theme.DimTextStyle.Render("    Auto-detects ./mar/ in current directory"))
			fmt.Println()
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection create photos ~/datasets/photos"))
			fmt.Println(theme.DimTextStyle.Render("    Specify custom path"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionCreate,
}

var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all asset collections",
	Long:  "Display all registered asset collections with their paths and types",
	Run:   runCollectionList,
}

var collectionDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an asset collection",
	Long:  "Remove a collection from the registry (does not delete files)",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection delete <name>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Example:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection delete photos"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionDelete,
}

var collectionInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show collection details",
	Long:  "Display detailed information about a collection including file counts and sizes",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection info <name>"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Example:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection info photos"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionInfo,
}

var collectionPushCmd = &cobra.Command{
	Use:   "push <name> [server]",
	Short: "Push collection to lambda server",
	Long: `Push a local collection to a remote server with intelligent transfer optimization.

Uses hybrid strategy for optimal speed:
  • First-time push: tar streaming (fastest for initial large transfers)
  • Incremental push: rsync delta sync (fastest for updates)

The collection will be synced to the server and registered in the remote config.

Examples:
  anime collection push photos                    # Push to default lambda server
  anime collection push photos lambda             # Push to lambda server
  anime collection push photos ubuntu@10.0.0.5    # Push to specific server`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection push <name> [server]"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Examples:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection push photos"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection push photos lambda"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection push photos ubuntu@10.0.0.5"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionPush,
}

var collectionSyncCmd = &cobra.Command{
	Use:   "sync <name> [server]",
	Short: "Sync collection with remote server",
	Long:  "Bidirectional sync of collection between local and remote server",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection sync <name> [server]"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Example:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection sync photos"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionSync,
}

var collectionPullCmd = &cobra.Command{
	Use:   "pull <name> [server]",
	Short: "Pull collection from remote server",
	Long:  "Download collection from remote server to local machine",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render("❌ Missing required argument"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Usage:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection pull <name> [server]"))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Example:"))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection pull photos"))
			fmt.Println()
			return fmt.Errorf("requires collection name")
		}
		return nil
	},
	RunE: runCollectionPull,
}

func init() {
	collectionCmd.AddCommand(collectionCreateCmd)
	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionDeleteCmd)
	collectionCmd.AddCommand(collectionInfoCmd)
	collectionCmd.AddCommand(collectionPushCmd)
	collectionCmd.AddCommand(collectionSyncCmd)
	collectionCmd.AddCommand(collectionPullCmd)
	rootCmd.AddCommand(collectionCmd)
}

func runCollectionHelp(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎨 ASSET COLLECTIONS 🎨"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("📦 Manage collections of images, videos, and assets for AI workflows"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📋 Available Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime collection create <name> <path>", "Create a new collection (alias: add)"},
		{"anime collection list", "List all registered collections"},
		{"anime collection info <name>", "Show detailed collection information"},
		{"anime collection delete <name>", "Remove a collection (keeps files)"},
		{"anime collection push <name>", "Push collection to server (hybrid: tar/rsync)"},
	}

	for _, c := range commands {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(c.cmd))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c.desc))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("✨ Example Workflow"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime collection add photos ~/Pictures"))
	fmt.Println(theme.DimTextStyle.Render("    Create a collection from your photos"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime workflows photos"))
	fmt.Println(theme.DimTextStyle.Render("    See available workflows for this collection"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime workflow photos img2vid"))
	fmt.Println(theme.DimTextStyle.Render("    Run image-to-video workflow with wizard"))
	fmt.Println()
}

func runCollectionCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Auto-detect path if not provided
	var path string
	if len(args) > 1 {
		path = args[1]
	} else {
		// Try current directory + name
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, name)

		// Check if it exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println()
			fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Directory not found: ./%s/", name)))
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("Options:"))
			fmt.Println()
			fmt.Printf("  %s %s\n",
				theme.HighlightStyle.Render("1."),
				theme.DimTextStyle.Render(fmt.Sprintf("Create the directory: mkdir %s", name)))
			fmt.Println()
			fmt.Printf("  %s %s\n",
				theme.HighlightStyle.Render("2."),
				theme.DimTextStyle.Render(fmt.Sprintf("Specify a custom path: anime collection add %s /path/to/folder", name)))
			fmt.Println()
			return fmt.Errorf("directory ./%s/ does not exist in current directory", name)
		}

		fmt.Println()
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("✓ Auto-detected path: ./%s/", name)))
		fmt.Println()
	}

	// Expand path
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path exists (for explicit paths)
	if len(args) > 1 {
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
	}

	// Detect collection type by scanning files
	collectionType := detectCollectionType(absPath)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create collection
	collection := config.Collection{
		Name: name,
		Path: absPath,
		Type: collectionType,
	}

	if err := cfg.AddCollection(collection); err != nil {
		// Check if it already exists
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("ℹ️  Collection '%s' already exists", name)))
			fmt.Println()

			// Get existing collection info
			if existingCol, getErr := cfg.GetCollection(name); getErr == nil {
				fmt.Printf("  Path:  %s\n", theme.DimTextStyle.Render(existingCol.Path))
				fmt.Printf("  Type:  %s\n", theme.DimTextStyle.Render(formatCollectionType(existingCol.Type)))
				fmt.Println()
			}

			fmt.Println(theme.InfoStyle.Render("💡 What you can do:"))
			fmt.Println()
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime collection push %s", name)))
			fmt.Println(theme.DimTextStyle.Render("    Push collection to Lambda server"))
			fmt.Println()
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime collection info %s", name)))
			fmt.Println(theme.DimTextStyle.Render("    View collection details"))
			fmt.Println()
			fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime collection delete %s", name)))
			fmt.Println(theme.DimTextStyle.Render("    Delete and recreate collection"))
			fmt.Println()
			return fmt.Errorf("collection already exists")
		}
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✓ Collection created successfully!"))
	fmt.Println()
	fmt.Printf("  Name:  %s\n", theme.HighlightStyle.Render(name))
	fmt.Printf("  Path:  %s\n", theme.DimTextStyle.Render(absPath))
	fmt.Printf("  Type:  %s\n", theme.InfoStyle.Render(formatCollectionType(collectionType)))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("💡 Next steps:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$ anime %s workflows", name)))
	fmt.Println(theme.DimTextStyle.Render("    View available workflows for this collection"))
	fmt.Println()

	return nil
}

func runCollectionList(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("❌ Failed to load config: " + err.Error()))
		return
	}

	collections := cfg.ListCollections()

	fmt.Println()
	fmt.Println(theme.RenderBanner("📦 ASSET COLLECTIONS 📦"))
	fmt.Println()

	if len(collections) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No collections found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Create your first collection:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime collection create photos ~/Pictures"))
		fmt.Println()
		return
	}

	fmt.Printf("  Total collections: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", len(collections))))
	fmt.Println()

	for _, col := range collections {
		// Get emoji for type
		emoji := "📦"
		switch col.Type {
		case "image":
			emoji = "🖼️"
		case "video":
			emoji = "🎬"
		case "mixed":
			emoji = "🎨"
		}

		fmt.Printf("  %s %s\n", emoji, theme.HighlightStyle.Render(col.Name))
		fmt.Printf("    Type: %s\n", theme.InfoStyle.Render(formatCollectionType(col.Type)))
		fmt.Printf("    Path: %s\n", theme.DimTextStyle.Render(col.Path))

		// Count files
		fileCount, totalSize := getCollectionStats(col.Path)
		fmt.Printf("    Files: %s  |  Size: %s\n",
			theme.SecondaryTextStyle.Render(fmt.Sprintf("%d", fileCount)),
			theme.SecondaryTextStyle.Render(formatBytes(totalSize)))

		if col.Description != "" {
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(col.Description))
		}
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 What to do next:"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection info <name>"))
	fmt.Println(theme.DimTextStyle.Render("    View detailed information about a collection"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection push <name>"))
	fmt.Println(theme.DimTextStyle.Render("    Push collection to Lambda server"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime workflow"))
	fmt.Println(theme.DimTextStyle.Render("    Browse and run workflows on collections"))
	fmt.Println()
	fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection create <name> <path>"))
	fmt.Println(theme.DimTextStyle.Render("    Create a new collection"))
	fmt.Println()
}

func runCollectionDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteCollection(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Collection '%s' deleted", name)))
	fmt.Println(theme.DimTextStyle.Render("  (Files were not deleted)"))
	fmt.Println()

	return nil
}

func runCollectionInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(name)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("📦 %s 📦", strings.ToUpper(name))))
	fmt.Println()

	// Get emoji for type
	emoji := "📦"
	switch collection.Type {
	case "image":
		emoji = "🖼️"
	case "video":
		emoji = "🎬"
	case "mixed":
		emoji = "🎨"
	}

	fmt.Printf("  %s %s\n", emoji, theme.HighlightStyle.Render(collection.Name))
	fmt.Printf("  Type: %s\n", theme.InfoStyle.Render(formatCollectionType(collection.Type)))
	fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(collection.Path))
	fmt.Println()

	// Get detailed stats
	fileCount, totalSize := getCollectionStats(collection.Path)
	imageCount, videoCount := getFileTypeCounts(collection.Path)

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Statistics"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  Total Files:  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", fileCount)))
	fmt.Printf("  Images:       %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", imageCount)))
	fmt.Printf("  Videos:       %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", videoCount)))
	fmt.Printf("  Total Size:   %s\n", theme.HighlightStyle.Render(formatBytes(totalSize)))
	fmt.Println()

	if collection.Description != "" {
		fmt.Println(theme.InfoStyle.Render("📝 Description"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(collection.Description))
		fmt.Println()
	}

	if len(collection.Tags) > 0 {
		fmt.Println(theme.InfoStyle.Render("🏷️  Tags"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(strings.Join(collection.Tags, ", ")))
		fmt.Println()
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🔥 WHAT YOU CAN DO"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Show different actions based on collection type
	fmt.Println(theme.InfoStyle.Render("  ⚡ QUICK START"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflows %s", name)))
	fmt.Println(theme.DimTextStyle.Render("      → Browse all AI workflows for this collection"))
	fmt.Println()

	fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s <name>", name)))
	fmt.Println(theme.DimTextStyle.Render("      → Run a specific workflow (interactive wizard)"))
	fmt.Println()

	if collection.Type == "image" || collection.Type == "mixed" {
		fmt.Println(theme.InfoStyle.Render("  🎨 IMAGE WORKFLOWS"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s img2vid", name)))
		fmt.Println(theme.DimTextStyle.Render("      → Generate videos from images (Stable Video Diffusion)"))
		fmt.Println()

		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s upscale", name)))
		fmt.Println(theme.DimTextStyle.Render("      → AI upscale to 4K/8K resolution"))
		fmt.Println()

		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s style", name)))
		fmt.Println(theme.DimTextStyle.Render("      → Apply artistic styles (anime, oil painting, etc)"))
		fmt.Println()

		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s enhance", name)))
		fmt.Println(theme.DimTextStyle.Render("      → AI enhancement (clarity, colors, details)"))
		fmt.Println()
	}

	if collection.Type == "video" || collection.Type == "mixed" {
		fmt.Println(theme.InfoStyle.Render("  🎬 VIDEO WORKFLOWS"))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s interpolate", name)))
		fmt.Println(theme.DimTextStyle.Render("      → Increase FPS (30fps → 60fps, 60fps → 120fps)"))
		fmt.Println()

		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s extract", name)))
		fmt.Println(theme.DimTextStyle.Render("      → Extract frames to images"))
		fmt.Println()

		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime workflow %s upscale", name)))
		fmt.Println(theme.DimTextStyle.Render("      → Upscale video resolution"))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("  🚀 OTHER OPTIONS"))
	fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime collection push %s", name)))
	fmt.Println(theme.DimTextStyle.Render("      → Push to Lambda server for GPU processing"))
	fmt.Println()

	fmt.Printf("    %s\n", theme.HighlightStyle.Render("anime ui"))
	fmt.Println(theme.DimTextStyle.Render("      → Open ComfyUI web interface"))
	fmt.Println()

	fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime serve %s", collection.Path)))
	fmt.Println(theme.DimTextStyle.Render("      → Share collection over HTTP"))
	fmt.Println()

	return nil
}

// Helper functions

func formatCollectionType(collectionType string) string {
	switch collectionType {
	case "image":
		return "Images"
	case "video":
		return "Videos"
	case "mixed":
		return "Images & Videos"
	default:
		return collectionType
	}
}

func detectCollectionType(path string) string {
	imageCount := 0
	videoCount := 0

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		if isImageExt(ext) {
			imageCount++
		} else if isVideoExt(ext) {
			videoCount++
		}

		// Early exit if we've seen enough files
		if imageCount+videoCount > 100 {
			return filepath.SkipDir
		}
		return nil
	})

	if imageCount > 0 && videoCount == 0 {
		return "image"
	} else if videoCount > 0 && imageCount == 0 {
		return "video"
	}
	return "mixed"
}

func getCollectionStats(path string) (int, int64) {
	fileCount := 0
	var totalSize int64

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		if isImageExt(ext) || isVideoExt(ext) {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	return fileCount, totalSize
}

func getFileTypeCounts(path string) (int, int) {
	imageCount := 0
	videoCount := 0

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		if isImageExt(ext) {
			imageCount++
		} else if isVideoExt(ext) {
			videoCount++
		}
		return nil
	})

	return imageCount, videoCount
}

func isImageExt(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".svg"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isVideoExt(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".flv", ".wmv", ".m4v"}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func runCollectionPush(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	// Determine server target
	serverTarget := "lambda"
	if len(args) > 1 {
		serverTarget = args[1]
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get collection - if it doesn't exist, auto-create it
	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		// Collection doesn't exist - auto-create it
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("ℹ️  Collection '%s' not found - creating automatically...", collectionName)))
		fmt.Println()

		// Use current directory + collection name as path
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		collectionPath := filepath.Join(cwd, collectionName)

		// Check if path exists
		if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s\n\nCreate the directory first or use: anime collection create %s <path>", collectionPath, collectionName)
		}

		// Detect collection type
		collectionType := detectCollectionType(collectionPath)

		// Create and save collection
		newCollection := config.Collection{
			Name: collectionName,
			Path: collectionPath,
			Type: collectionType,
		}

		if err := cfg.AddCollection(newCollection); err != nil {
			return fmt.Errorf("failed to add collection: %w", err)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println(theme.SuccessStyle.Render("✓ Collection created successfully!"))
		fmt.Printf("  Name: %s\n", theme.HighlightStyle.Render(collectionName))
		fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(collectionPath))
		fmt.Printf("  Type: %s\n", theme.InfoStyle.Render(formatCollectionType(collectionType)))
		fmt.Println()

		// Use the newly created collection
		collection = &newCollection
	}

	// Resolve server target
	target, err := resolveServerTarget(cfg, serverTarget)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("📤 PUSHING %s 📤", strings.ToUpper(collectionName))))
	fmt.Println()

	// Get collection stats
	fileCount, totalSize := getCollectionStats(collection.Path)
	fmt.Printf("  Collection: %s\n", theme.HighlightStyle.Render(collectionName))
	fmt.Printf("  Path:       %s\n", theme.DimTextStyle.Render(collection.Path))
	fmt.Printf("  Files:      %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", fileCount)))
	fmt.Printf("  Size:       %s\n", theme.InfoStyle.Render(formatBytes(totalSize)))
	fmt.Printf("  Target:     %s\n", theme.HighlightStyle.Render(target))
	fmt.Println()

	// Check if remote collection exists (for hybrid push strategy)
	remotePath := fmt.Sprintf("~/collections/%s/", collectionName)
	remoteExists := checkRemotePathExists(target, remotePath)

	if !remoteExists {
		// First push: Use tar streaming for fastest initial transfer
		fmt.Println(theme.InfoStyle.Render("🚀 First-time push detected - using tar streaming for speed..."))
		fmt.Println()

		if err := tarStreamToServer(collection.Path, target, remotePath, collectionName); err != nil {
			return fmt.Errorf("tar stream failed: %w", err)
		}

		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("✓ Transfer complete!"))
		fmt.Println()
	} else {
		// Incremental push: Use rsync for delta sync
		fmt.Println(theme.InfoStyle.Render("🔄 Incremental push - using rsync for delta sync..."))
		fmt.Println()

		if err := rsyncCollectionToServer(collection.Path, target, remotePath); err != nil {
			return fmt.Errorf("rsync failed: %w", err)
		}

		fmt.Println()
		fmt.Println(theme.SuccessStyle.Render("✓ Sync complete!"))
		fmt.Println()
	}

	// Register collection on remote server
	fmt.Println(theme.InfoStyle.Render("📝 Registering collection on remote server..."))

	if err := registerRemoteCollection(target, collection, remotePath); err != nil {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠️  Warning: Could not register collection: %s", err.Error())))
		fmt.Println(theme.DimTextStyle.Render("   Collection files are synced but may not appear in remote config"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("✓ Collection registered on remote server!"))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✨ Collection '%s' pushed successfully! ✨", collectionName)))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("💡 Next steps:"))
	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("$ ssh %s", target)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("    Then: anime workflows %s", collectionName)))
	fmt.Println()

	return nil
}

func runCollectionSync(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	serverTarget := "lambda"
	if len(args) > 1 {
		serverTarget = args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return err
	}

	target, err := resolveServerTarget(cfg, serverTarget)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("🔄 SYNCING %s 🔄", strings.ToUpper(collectionName))))
	fmt.Println()

	// Bidirectional sync
	fmt.Println(theme.InfoStyle.Render("📤 Push: Local → Remote"))
	remotePath := fmt.Sprintf("~/collections/%s/", collectionName)
	if err := rsyncCollectionToServer(collection.Path, target, remotePath); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓ Push complete"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("📥 Pull: Remote → Local"))
	if err := rsyncFromServer(target, remotePath, collection.Path); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}
	fmt.Println(theme.SuccessStyle.Render("✓ Pull complete"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Collection '%s' synced!", collectionName)))
	fmt.Println()

	return nil
}

func runCollectionPull(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	serverTarget := "lambda"
	if len(args) > 1 {
		serverTarget = args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return err
	}

	target, err := resolveServerTarget(cfg, serverTarget)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("📥 PULLING %s 📥", strings.ToUpper(collectionName))))
	fmt.Println()

	fmt.Printf("  Collection: %s\n", theme.HighlightStyle.Render(collectionName))
	fmt.Printf("  Source:     %s\n", theme.InfoStyle.Render(target))
	fmt.Printf("  Dest:       %s\n", theme.DimTextStyle.Render(collection.Path))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🚀 Starting rsync..."))
	fmt.Println()

	remotePath := fmt.Sprintf("~/collections/%s/", collectionName)
	if err := rsyncFromServer(target, remotePath, collection.Path); err != nil {
		return fmt.Errorf("rsync failed: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Collection '%s' pulled successfully!", collectionName)))
	fmt.Println()

	return nil
}

// Helper functions

func resolveServerTarget(cfg *config.Config, target string) (string, error) {
	// Check if it's an alias
	if alias := cfg.GetAlias(target); alias != "" {
		// If alias doesn't have user@, add ubuntu@
		if !strings.Contains(alias, "@") {
			return "ubuntu@" + alias, nil
		}
		return alias, nil
	}

	// Check if it's a server config
	if server, err := cfg.GetServer(target); err == nil {
		return fmt.Sprintf("%s@%s", server.User, server.Host), nil
	}

	// Check if it looks like user@host - and auto-configure lambda if target was "lambda"
	if strings.Contains(target, "@") {
		return target, nil
	}

	// Check if it looks like an IP/hostname - auto-configure as lambda
	if strings.Contains(target, ".") {
		resolvedTarget := "ubuntu@" + target

		// If we're trying to use this as lambda, auto-configure it
		if target != "lambda" {
			fmt.Println()
			fmt.Println(theme.InfoStyle.Render("ℹ️  Auto-configuring Lambda server..."))
			cfg.SetAlias("lambda", target)
			if err := cfg.Save(); err != nil {
				fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("⚠️  Warning: Could not save config: %s", err.Error())))
			} else {
				fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Lambda server set to: %s", target)))
			}
			fmt.Println()
		}

		return resolvedTarget, nil
	}

	// Special handling for "lambda" not configured
	if target == "lambda" {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("❌ Lambda server not configured"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("💡 Options:"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("1. Provide the IP directly in this command:"))
		fmt.Printf("     %s\n", theme.HighlightStyle.Render("anime collection push <name> 209.20.159.132"))
		fmt.Println(theme.DimTextStyle.Render("     (Will auto-configure lambda for future use)"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("2. Configure lambda first:"))
		fmt.Printf("     %s\n", theme.HighlightStyle.Render("anime set lambda 209.20.159.132"))
		fmt.Println()
		return "", fmt.Errorf("lambda server not configured")
	}

	return "", fmt.Errorf("could not resolve server target: %s", target)
}

func rsyncCollectionToServer(localPath, target, remotePath string) error {
	// Ensure local path ends with /
	if !strings.HasSuffix(localPath, "/") {
		localPath = localPath + "/"
	}

	// Build rsync command
	rsyncCmd := exec.Command("rsync",
		"-avz",
		"--progress",
		localPath,
		target+":"+remotePath,
	)

	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	return rsyncCmd.Run()
}

func rsyncFromServer(target, remotePath, localPath string) error {
	// Ensure local path exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return err
	}

	// Ensure local path ends with /
	if !strings.HasSuffix(localPath, "/") {
		localPath = localPath + "/"
	}

	// Build rsync command
	rsyncCmd := exec.Command("rsync",
		"-avz",
		"--progress",
		target+":"+remotePath,
		localPath,
	)

	rsyncCmd.Stdout = os.Stdout
	rsyncCmd.Stderr = os.Stderr

	return rsyncCmd.Run()
}

func registerRemoteCollection(target string, collection *config.Collection, remotePath string) error {
	// Parse target to get user and host
	var user, host string
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = target
	}

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return err
	}
	defer sshClient.Close()

	// Expand remote path
	expandedPath, err := sshClient.RunCommand(fmt.Sprintf("echo %s", remotePath))
	if err != nil {
		return err
	}
	expandedPath = strings.TrimSpace(expandedPath)

	// Create collection registration command
	registerCmd := fmt.Sprintf(`
# Ensure anime config exists
mkdir -p ~/.config/anime

# Check if collection already exists
if ! grep -q "name: %s" ~/.config/anime/config.yaml 2>/dev/null; then
  # Add collection to config
  cat >> ~/.config/anime/config.yaml <<'EOF'
collections:
  - name: %s
    path: %s
    type: %s
    description: "Pushed from local machine"
EOF
  echo "Collection registered"
else
  echo "Collection already exists in config"
fi
`, collection.Name, collection.Name, expandedPath, collection.Type)

	_, err = sshClient.RunCommand(registerCmd)
	return err
}

func checkRemotePathExists(target, remotePath string) bool {
	// Parse target to get user and host
	var user, host string
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		user = "ubuntu"
		host = target
	}

	// Create SSH client
	sshClient, err := ssh.NewClient(host, user, "")
	if err != nil {
		return false
	}
	defer sshClient.Close()

	// Check if path exists
	checkCmd := fmt.Sprintf("test -d %s && echo 'exists' || echo 'not_exists'", remotePath)
	output, err := sshClient.RunCommand(checkCmd)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == "exists"
}

func tarStreamToServer(localPath, target, remotePath, collectionName string) error {
	// Parse target to get user and host for logging
	var targetHost string
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		targetHost = parts[1]
	} else {
		targetHost = target
	}

	// Ensure local path doesn't end with / for tar
	localPath = strings.TrimSuffix(localPath, "/")

	// Get parent directory and collection directory name
	parentDir := filepath.Dir(localPath)
	dirName := filepath.Base(localPath)

	// Create remote directory first
	mkdirCmd := exec.Command("ssh", target, fmt.Sprintf("mkdir -p %s", remotePath))
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Build tar streaming command:
	// cd to parent dir, tar the directory, stream via ssh, untar on remote
	// Using pigz if available for parallel compression (much faster)
	tarCmd := fmt.Sprintf(
		"cd %s && (tar -cf - %s | pigz -c 2>/dev/null || tar -czf - %s) | ssh %s 'cd %s && tar -xzf -'",
		shellQuote(parentDir),
		shellQuote(dirName),
		shellQuote(dirName),
		target,
		remotePath,
	)

	fmt.Printf("  Streaming to %s...\n", theme.DimTextStyle.Render(targetHost))

	// Execute tar streaming command
	streamCmd := exec.Command("sh", "-c", tarCmd)
	streamCmd.Stdout = os.Stdout
	streamCmd.Stderr = os.Stderr

	return streamCmd.Run()
}

func shellQuote(s string) string {
	// Simple shell quoting - wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
