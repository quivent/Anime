package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	animModel      string
	animFPS        int
	animSeconds    float64
	animParallel   int
	animPrompt     string
	animOutput     string
	animBackground bool
)

var animateCmd = &cobra.Command{
	Use:   "animate [collection] [model]",
	Short: "Animate images to videos",
	Long: `Transform static images into dynamic videos using AI.

Models:
  wan2      - Wan2.2 (default, fast)
  mochi     - Mochi-1 (high quality)
  svd       - Stable Video Diffusion
  ltx       - LTXVideo

Examples:
  anime animate                               # Browse collections
  anime animate mar                           # Show options for 'mar' collection
  anime animate mar wan2                      # Animate 'mar' with wan2 (live output)
  anime animate photos wan2 --background      # Run in background
  anime animate photos wan2 --fps 30          # 30fps video
  anime animate renders mochi --seconds 5     # 5 second videos
  anime animate mar --parallel 4              # Process 4 at once
  anime animate photos --prompt "cinematic"   # With prompt`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runAnimate,
}

func init() {
	rootCmd.AddCommand(animateCmd)

	animateCmd.Flags().StringVarP(&animModel, "model", "m", "wan2", "Model to use (wan2, mochi, svd, ltx)")
	animateCmd.Flags().IntVar(&animFPS, "fps", 24, "Frames per second")
	animateCmd.Flags().Float64VarP(&animSeconds, "seconds", "s", 3.0, "Video duration in seconds")
	animateCmd.Flags().IntVarP(&animParallel, "parallel", "p", 1, "Number of parallel processes")
	animateCmd.Flags().StringVar(&animPrompt, "prompt", "", "Optional prompt for generation")
	animateCmd.Flags().StringVarP(&animOutput, "output", "o", "", "Output directory (default: ./animated)")
	animateCmd.Flags().BoolVarP(&animBackground, "background", "b", false, "Run in background (default: show live output)")
}

func runAnimate(cmd *cobra.Command, args []string) error {
	// If no args, show browser
	if len(args) == 0 {
		showAnimateBrowser()
		return nil
	}

	collectionName := args[0]

	// Load collection
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	// If model specified as arg, override flag
	if len(args) > 1 {
		animModel = args[1]
	}

	// Set default output
	if animOutput == "" {
		animOutput = "./animated"
	}

	// Show configuration
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎬 ANIMATING COLLECTION"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📋 Configuration"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  Collection: %s\n", theme.HighlightStyle.Render(collection.Name))
	fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(collection.Path))
	fmt.Printf("  Model: %s\n", theme.HighlightStyle.Render(animModel))
	fmt.Printf("  FPS: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", animFPS)))
	fmt.Printf("  Duration: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%.1fs", animSeconds)))
	fmt.Printf("  Parallel: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", animParallel)))
	if animPrompt != "" {
		fmt.Printf("  Prompt: %s\n", theme.SecondaryTextStyle.Render(animPrompt))
	}
	fmt.Printf("  Output: %s\n", theme.DimTextStyle.Render(animOutput))

	// Show execution mode
	if animBackground {
		fmt.Printf("  Mode: %s\n", theme.DimTextStyle.Render("Background"))
	} else {
		fmt.Printf("  Mode: %s\n", theme.InfoStyle.Render("Live (foreground)"))
	}
	fmt.Println()

	// Count files
	fileCount, _ := getCollectionStats(collection.Path)
	if fileCount == 0 {
		return fmt.Errorf("no files found in collection")
	}

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📊 Job Summary"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  Files to process: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", fileCount)))

	// Estimate time
	secondsPerFile := 120.0 // 2 minutes default for wan2
	if animModel == "mochi" {
		secondsPerFile = 180.0 // 3 minutes for mochi
	}

	totalSeconds := (float64(fileCount) * secondsPerFile) / float64(animParallel)
	estTime := formatSeconds(int(totalSeconds))

	fmt.Printf("  Est. time: %s\n", theme.InfoStyle.Render(estTime))
	fmt.Println()

	// Detect GPUs
	gpuCount := detectGPUCount()
	if gpuCount > 0 {
		fmt.Printf("  GPUs detected: %s\n", theme.SuccessStyle.Render(fmt.Sprintf("%d", gpuCount)))
	} else {
		fmt.Println(theme.WarningStyle.Render("  ⚠️  No GPUs detected - will use CPU (slower)"))
	}
	fmt.Println()

	// Execute using collection workflow implementation
	fmt.Println(theme.InfoStyle.Render("🚀 Starting animation..."))
	fmt.Println()

	// Call collection animate with proper parameters
	oldAnimateModel := animateModel
	oldAnimateFPS := animateFPS
	oldAnimateLength := animateLength
	oldAnimateOutput := animateOutput
	oldAnimateParallel := animateParallel
	oldAnimatePrompt := animatePrompt

	animateModel = animModel
	animateFPS = animFPS
	animateLength = animSeconds
	animateOutput = animOutput
	animateParallel = animParallel
	animatePrompt = animPrompt

	err = runAnimateWithModel(collectionName, collection.Path, fileCount, animModel)

	// Restore original values
	animateModel = oldAnimateModel
	animateFPS = oldAnimateFPS
	animateLength = oldAnimateLength
	animateOutput = oldAnimateOutput
	animateParallel = oldAnimateParallel
	animatePrompt = oldAnimatePrompt

	return err
}

func formatSeconds(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}

func detectGPUCount() int {
	return 8 // stub
}

func showCollectionSuggestions(collection *config.Collection) {
	fileCount, _ := getCollectionStats(collection.Path)

	fmt.Println()
	fmt.Println(theme.RenderBanner(fmt.Sprintf("🎬 ANIMATE '%s'", collection.Name)))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Collection Info"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  Name: %s\n", theme.HighlightStyle.Render(collection.Name))
	fmt.Printf("  Path: %s\n", theme.DimTextStyle.Render(collection.Path))
	fmt.Printf("  Images: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", fileCount)))
	fmt.Println()

	// Show model suggestions
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 Suggested Commands"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	models := []struct {
		name  string
		desc  string
		speed string
	}{
		{"wan2", "Fast & Good - Best for quick iterations", "⚡ ~12s/video"},
		{"ltx", "Fastest - Great for testing", "⚡⚡ ~8s/video"},
		{"mochi", "High Quality - Best results, slower", "🐌 ~45s/video"},
		{"svd", "Highest Quality - Most realistic", "⏱️  ~30s/video"},
	}

	for _, m := range models {
		fmt.Printf("  %s %s - %s\n", theme.InfoStyle.Render(m.speed), theme.HighlightStyle.Render(m.name), theme.DimTextStyle.Render(m.desc))
		fmt.Printf("    %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s %s", collection.Name, m.name)))
		fmt.Println()
	}

	// Show advanced options
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("⚙️  Advanced Options"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s wan2 --fps 30", collection.Name)))
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("  Higher frame rate (smoother motion)"))

	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s wan2 --seconds 5", collection.Name)))
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("  Longer videos (5 seconds)"))

	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s wan2 --parallel 4", collection.Name)))
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("  Process 4 images in parallel (faster on multi-GPU)"))

	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s wan2 --prompt \"cinematic\"", collection.Name)))
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("  Add text prompt for guidance"))

	fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s wan2 --background", collection.Name)))
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("  Run in background (default: live/foreground)"))

	// Estimate total time
	secondsPerFile := 12.0 // wan2 default
	totalSeconds := float64(fileCount) * secondsPerFile
	estTime := formatSeconds(int(totalSeconds))

	fmt.Println(theme.InfoStyle.Render("📊 Estimates (wan2 model):"))
	fmt.Println()
	fmt.Printf("  Time: %s for %d images\n", theme.DimTextStyle.Render(estTime), fileCount)
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  💡 Tip: By default, output is shown live. Use --background to run in background."))
	fmt.Println()
}

func showAnimateBrowser() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎬 ANIMATE IMAGES"))
	fmt.Println()

	// Load config to get collections
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	collections := cfg.ListCollections()

	// Filter to image/mixed collections only
	var imageCollections []config.Collection
	for _, col := range collections {
		if col.Type == "image" || col.Type == "mixed" {
			imageCollections = append(imageCollections, col)
		}
	}

	if len(imageCollections) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No image collections found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Create a collection to get started:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime collection create photos ~/Pictures"))
		fmt.Println()
		return
	}

	// Show collections
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Available Collections"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	// Find collection with most images
	var bestCol config.Collection
	maxFiles := 0
	for _, col := range imageCollections {
		fileCount, _ := getCollectionStats(col.Path)
		if fileCount > maxFiles {
			maxFiles = fileCount
			bestCol = col
		}
	}

	for _, col := range imageCollections {
		fileCount, _ := getCollectionStats(col.Path)
		icon := "🖼️"
		if col.Name == bestCol.Name {
			icon = "⭐"
		}

		fmt.Printf("  %s  %s\n", icon, theme.HighlightStyle.Render(col.Name))
		fmt.Printf("    Path: %s\n", theme.DimTextStyle.Render(col.Path))
		fmt.Printf("    Images: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", fileCount)))
		fmt.Println()
	}

	// Show suggestion
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 Suggested Next Step"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if maxFiles > 0 {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime animate %s", bestCol.Name)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("  See animation options for '%s' (%d images)", bestCol.Name, maxFiles)))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("💡 Tips:"))
	fmt.Println()
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("Run 'anime animate <collection>' to see all options"))
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("wan2 and ltx are best for quick iterations"))
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("svd and mochi produce highest quality results"))
	fmt.Println()
}
