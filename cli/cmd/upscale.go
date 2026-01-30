package cmd

import (
	"fmt"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	upscModel    string
	upscScale    int
	upscParallel int
	upscOutput   string
)

var upscaleCmd = &cobra.Command{
	Use:   "upscale [collection] [scale]",
	Short: "Upscale images to higher resolution",
	Long: `Upscale images using AI models.

Models:
  esrgan    - ESRGAN (default, good quality)
  realesrgan - RealESRGAN (photorealistic)
  gfpgan    - GFPGAN (faces)

Examples:
  anime upscale                        # Browse collections
  anime upscale mar                    # 2x upscale with esrgan
  anime upscale photos 4               # 4x upscale
  anime upscale renders --model gfpgan # Face restoration
  anime upscale mar --parallel 8       # 8 parallel processes`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runUpscale,
}

func init() {
	rootCmd.AddCommand(upscaleCmd)
	
	upscaleCmd.Flags().StringVarP(&upscModel, "model", "m", "esrgan", "Model to use (esrgan, realesrgan, gfpgan)")
	upscaleCmd.Flags().IntVar(&upscScale, "scale", 2, "Upscale factor (2 or 4)")
	upscaleCmd.Flags().IntVarP(&upscParallel, "parallel", "p", 4, "Number of parallel processes")
	upscaleCmd.Flags().StringVarP(&upscOutput, "output", "o", "", "Output directory (default: ./upscaled)")
}

func runUpscale(cmd *cobra.Command, args []string) error {
	// If no args, show browser
	if len(args) == 0 {
		showUpscaleBrowser()
		return nil
	}

	collectionName := args[0]

	// If scale specified as arg, override flag
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%d", &upscScale)
	}

	// Load collection
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	if upscOutput == "" {
		upscOutput = "./upscaled"
	}

	// Show configuration
	fmt.Println()
	fmt.Println(theme.RenderBanner("⬆️  UPSCALING COLLECTION"))
	fmt.Println()
	
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📋 Configuration"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	
	fmt.Printf("  Collection: %s\n", theme.HighlightStyle.Render(collection.Name))
	fmt.Printf("  Model: %s\n", theme.HighlightStyle.Render(upscModel))
	fmt.Printf("  Scale: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%dx", upscScale)))
	fmt.Printf("  Parallel: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", upscParallel)))
	fmt.Printf("  Output: %s\n", theme.DimTextStyle.Render(upscOutput))
	fmt.Println()

	fileCount, _ := getCollectionStats(collection.Path)
	fmt.Printf("  Files to process: %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%d", fileCount)))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("🚀 Starting upscale workflow..."))
	fmt.Println(theme.DimTextStyle.Render("   (This will be implemented with actual execution)"))
	fmt.Println()

	return nil
}
func showUpscaleBrowser() {
	fmt.Println()
	fmt.Println(theme.RenderBanner("⬆️  UPSCALE IMAGES"))
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

	// Show collections
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("📦 Collections with Images"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if len(imageCollections) == 0 {
		fmt.Println(theme.WarningStyle.Render("  No image collections found"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("  💡 Create a collection to get started:"))
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("$ anime collection create photos ~/Pictures"))
		fmt.Println()
	} else {
		for _, col := range imageCollections {
			fileCount, _ := getCollectionStats(col.Path)

			fmt.Printf("  🖼️  %s\n", theme.HighlightStyle.Render(col.Name))
			fmt.Printf("    Path: %s\n", theme.DimTextStyle.Render(col.Path))
			fmt.Printf("    Images: %s\n", theme.InfoStyle.Render(fmt.Sprintf("%d", fileCount)))
			fmt.Println()
		}
	}

	// Show available models
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("🎨 Available Models"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	models := []struct {
		name  string
		desc  string
		use   string
	}{
		{"esrgan", "ESRGAN - General purpose upscaling", "Good for renders, artwork, mixed content"},
		{"realesrgan", "RealESRGAN - Photorealistic images", "Best for photos and realistic images"},
		{"gfpgan", "GFPGAN - Face restoration & upscaling", "Specialized for portraits and faces"},
	}

	for _, m := range models {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(m.name))
		fmt.Printf("    %s\n", theme.InfoStyle.Render(m.desc))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(m.use))
		fmt.Println()
	}

	// Show usage examples
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(theme.InfoStyle.Render("💡 Usage Examples"))
	fmt.Println(theme.SuccessStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if len(imageCollections) > 0 {
		exampleCol := imageCollections[0].Name
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime upscale %s", exampleCol)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Upscale 2x with default model (esrgan)"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime upscale %s 4", exampleCol)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Upscale 4x"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime upscale %s --model realesrgan --scale 4", exampleCol)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Use RealESRGAN for photorealistic 4x upscaling"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("anime upscale %s --model gfpgan --parallel 8", exampleCol)))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Face restoration with 8 parallel processes"))
		fmt.Println()
	} else {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime collection create photos ~/Pictures"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Create a collection first"))
		fmt.Println()
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("anime upscale photos 4"))
		fmt.Printf("  %s\n", theme.DimTextStyle.Render("  Then upscale it"))
		fmt.Println()
	}

	fmt.Println(theme.InfoStyle.Render("💡 Tips:"))
	fmt.Println()
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("2x is faster and uses less VRAM"))
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("4x provides better quality but takes 4x longer"))
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("Use gfpgan specifically for faces and portraits"))
	fmt.Printf("  • %s\n", theme.DimTextStyle.Render("Increase --parallel on multi-GPU systems for speed"))
	fmt.Println()
}
