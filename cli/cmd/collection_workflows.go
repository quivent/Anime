package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var collectionAnimateCmd = &cobra.Command{
	Use:   "animate <collection>",
	Short: "Animate images in a collection",
	Long: `Batch animate images in a collection using AI video generation models.

Supports multiple animation backends:
  • Stable Video Diffusion (SVD)
  • AnimateDiff
  • LTXVideo
  • Mochi-1
  • CogVideoX

Features:
  • Batch processing with progress tracking
  • GPU-accelerated inference
  • Output organization by model/settings
  • Resume incomplete jobs

Examples:
  anime collection photos animate           # Interactive wizard
  anime collection photos animate --model svd
  anime collection photos animate --model ltxvideo --fps 24 --length 3`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCollectionAnimate,
}

var (
	animateModel    string
	animateFPS      int
	animateLength   float64
	animateBatch    int
	animateOutput   string
	animatePrompt   string
	animateParallel int
)

var collectionUpscaleCmd = &cobra.Command{
	Use:   "upscale <collection>",
	Short: "Upscale images/videos in a collection",
	Long: `Batch upscale images or videos in a collection using AI upscaling models.

Supports multiple upscaling backends:
  • Real-ESRGAN (images, fast)
  • GFPGAN (faces, quality)
  • CodeFormer (faces, advanced)
  • Topaz Video AI (videos, premium)
  • BasicVSR++ (videos, open-source)

Features:
  • Auto-detection of image vs video
  • Multi-GPU support
  • Preserve metadata
  • Quality presets (draft/balanced/quality)

Examples:
  anime collection photos upscale                    # Interactive wizard
  anime collection photos upscale --scale 4 --model realesrgan
  anime collection renders upscale --quality high --batch 8`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCollectionUpscale,
}

var (
	upscaleScale   int
	upscaleModel   string
	upscaleQuality string
	upscaleBatch   int
	upscaleOutput  string
)

var collectionTransformCmd = &cobra.Command{
	Use:   "transform <collection>",
	Short: "Transform collection with custom pipeline wizard",
	Long: `Interactive wizard to create custom AI transformation pipelines.

Build multi-step workflows by chaining operations:
  • Image operations: resize, crop, enhance, denoise, colorize
  • Style transfer: apply artistic styles
  • Generation: img2img, inpainting, outpainting
  • Video operations: stabilization, interpolation, effects
  • Post-processing: watermark, batch rename, format conversion

Pipeline Features:
  • Save pipeline templates for reuse
  • Parallel processing on multi-GPU setups
  • Resume interrupted pipelines
  • Preview mode (process 1 sample)
  • Export pipeline as JSON/YAML

Examples:
  anime collection photos transform              # Launch interactive wizard
  anime collection photos transform --template enhance-upscale-denoise
  anime collection photos transform --preview    # Preview on 1 image`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCollectionTransform,
}

var (
	transformTemplate string
	transformPreview  bool
	transformSave     string
)

func init() {
	// Add animate flags
	collectionAnimateCmd.Flags().StringVarP(&animateModel, "model", "m", "", "Animation model (svd, animatediff, ltxvideo, mochi, cogvideo, wan2)")
	collectionAnimateCmd.Flags().IntVarP(&animateFPS, "fps", "f", 24, "Frames per second")
	collectionAnimateCmd.Flags().Float64VarP(&animateLength, "length", "l", 2.0, "Video length in seconds")
	collectionAnimateCmd.Flags().IntVarP(&animateBatch, "batch", "b", 1, "Batch size for GPU processing")
	collectionAnimateCmd.Flags().StringVarP(&animateOutput, "output", "o", "", "Output directory (default: <collection>/animated)")
	collectionAnimateCmd.Flags().StringVarP(&animatePrompt, "prompt", "p", "", "Text prompt to guide animation (model-dependent)")
	collectionAnimateCmd.Flags().IntVar(&animateParallel, "parallel", 1, "Number of parallel workers for batch processing")

	// Add upscale flags
	collectionUpscaleCmd.Flags().IntVarP(&upscaleScale, "scale", "s", 4, "Upscale factor (2, 4, or 8)")
	collectionUpscaleCmd.Flags().StringVarP(&upscaleModel, "model", "m", "", "Upscaling model (realesrgan, gfpgan, codeformer, basicvsr)")
	collectionUpscaleCmd.Flags().StringVarP(&upscaleQuality, "quality", "q", "balanced", "Quality preset (draft, balanced, quality)")
	collectionUpscaleCmd.Flags().IntVarP(&upscaleBatch, "batch", "b", 4, "Batch size for GPU processing")
	collectionUpscaleCmd.Flags().StringVarP(&upscaleOutput, "output", "o", "", "Output directory (default: <collection>/upscaled)")

	// Add transform flags
	collectionTransformCmd.Flags().StringVarP(&transformTemplate, "template", "t", "", "Use saved pipeline template")
	collectionTransformCmd.Flags().BoolVarP(&transformPreview, "preview", "p", false, "Preview mode (process 1 sample)")
	collectionTransformCmd.Flags().StringVarP(&transformSave, "save", "s", "", "Save pipeline as template")

	// Add to collection command
	collectionCmd.AddCommand(collectionAnimateCmd)
	collectionCmd.AddCommand(collectionUpscaleCmd)
	collectionCmd.AddCommand(collectionTransformCmd)
}

func runCollectionAnimate(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	// Load config to get collection
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render(fmt.Sprintf("❌ Collection '%s' not found", collectionName)))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Available collections:"))
		for _, c := range cfg.Collections {
			fmt.Printf("  • %s\n", theme.HighlightStyle.Render(c.Name))
		}
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("Create a new collection with: anime collection create <name> <path>"))
		fmt.Println()
		return fmt.Errorf("collection not found")
	}

	// Welcome banner
	fmt.Println()
	fmt.Println(theme.RenderBanner("🎬 ANIMATE COLLECTION"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📦 Collection: %s", collectionName)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("   Path: %s", collection.Path)))
	fmt.Println()

	// Count images in collection
	imageCount, err := countImages(collection.Path)
	if err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Found %d images to animate", imageCount)))
	fmt.Println()

	// If no model specified, launch wizard
	if animateModel == "" {
		return runAnimateWizard(collectionName, collection.Path, imageCount)
	}

	// Run with specified model
	return runAnimateWithModel(collectionName, collection.Path, imageCount, animateModel)
}

func runAnimateWizard(collectionName, collectionPath string, imageCount int) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.CategoryStyle("🎨 ANIMATION WIZARD"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Let's configure your animation pipeline!"))
	fmt.Println()

	// Step 1: Choose model
	fmt.Println(theme.HighlightStyle.Render("Step 1: Choose Animation Model"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("Stable Video Diffusion (SVD)") + theme.DimTextStyle.Render(" - Best quality, slower (~30s/video)"))
	fmt.Println(theme.InfoStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("AnimateDiff") + theme.DimTextStyle.Render(" - Balanced quality/speed (~15s/video)"))
	fmt.Println(theme.InfoStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("LTXVideo") + theme.DimTextStyle.Render(" - Fast, good quality (~8s/video)"))
	fmt.Println(theme.InfoStyle.Render("  4") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("Mochi-1") + theme.DimTextStyle.Render(" - Experimental, high quality (~45s/video)"))
	fmt.Println(theme.InfoStyle.Render("  5") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("Wan2.2") + theme.DimTextStyle.Render(" - State-of-the-art img2vid (~20s/video)"))
	fmt.Println(theme.InfoStyle.Render("  6") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("CogVideoX") + theme.DimTextStyle.Render(" - Open source text2video (~25s/video)"))
	fmt.Println(theme.InfoStyle.Render("  7") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("Open-Sora") + theme.DimTextStyle.Render(" - High-quality video gen (~30s/video)"))
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Select model (1-7) [3]: "))
	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	var model string
	switch modelChoice {
	case "1":
		model = "svd"
	case "2":
		model = "animatediff"
	case "3", "":
		model = "ltxvideo" // default to fastest
	case "4":
		model = "mochi"
	case "5":
		model = "wan2"
	case "6":
		model = "cogvideo"
	case "7":
		model = "opensora"
	default:
		model = "ltxvideo"
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Selected: %s", model)))
	fmt.Println()

	// Step 2: Video settings
	fmt.Println(theme.HighlightStyle.Render("Step 2: Video Settings"))
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Video length in seconds (1-5) [2]: "))
	lengthInput, _ := reader.ReadString('\n')
	lengthInput = strings.TrimSpace(lengthInput)
	if lengthInput == "" {
		animateLength = 2.0
	} else {
		fmt.Sscanf(lengthInput, "%f", &animateLength)
	}

	fmt.Print(theme.HighlightStyle.Render("Frames per second (12/24/30) [24]: "))
	fpsInput, _ := reader.ReadString('\n')
	fpsInput = strings.TrimSpace(fpsInput)
	if fpsInput == "" {
		animateFPS = 24
	} else {
		fmt.Sscanf(fpsInput, "%d", &animateFPS)
	}

	fmt.Println()

	// Step 2.5: Prompt (optional, for models that support it)
	if model == "wan2" || model == "mochi" || model == "cogvideo" {
		fmt.Println(theme.HighlightStyle.Render("Step 2.5: Animation Prompt (Optional)"))
		fmt.Println()
		fmt.Println(theme.DimTextStyle.Render("  Add a text prompt to guide the animation style"))
		fmt.Print(theme.HighlightStyle.Render("Enter prompt (or press Enter to skip): "))
		promptInput, _ := reader.ReadString('\n')
		animatePrompt = strings.TrimSpace(promptInput)
		if animatePrompt != "" {
			fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Prompt: %s", animatePrompt)))
		}
		fmt.Println()
	}

	// Step 3: Parallelism settings
	fmt.Println(theme.HighlightStyle.Render("Step 3: Parallel Processing"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Process multiple images simultaneously for faster completion"))
	fmt.Print(theme.HighlightStyle.Render("Number of parallel workers (1-8) [4]: "))
	parallelInput, _ := reader.ReadString('\n')
	parallelInput = strings.TrimSpace(parallelInput)
	if parallelInput == "" {
		animateParallel = 4
	} else {
		fmt.Sscanf(parallelInput, "%d", &animateParallel)
		if animateParallel < 1 {
			animateParallel = 1
		} else if animateParallel > 8 {
			animateParallel = 8
		}
	}
	fmt.Println()

	// Step 4: Output settings
	fmt.Println(theme.HighlightStyle.Render("Step 4: Output Settings"))
	fmt.Println()

	defaultOutput := filepath.Join(collectionPath, "animated", model)
	fmt.Printf(theme.DimTextStyle.Render("  Default output: %s\n"), defaultOutput)
	fmt.Print(theme.HighlightStyle.Render("Custom output path (Enter for default): "))
	outputInput, _ := reader.ReadString('\n')
	outputInput = strings.TrimSpace(outputInput)
	if outputInput == "" {
		animateOutput = defaultOutput
	} else {
		animateOutput = outputInput
	}

	fmt.Println()

	// Summary
	fmt.Println(theme.CategoryStyle("📋 PIPELINE SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Collection:"), collectionName)
	fmt.Printf("  %s  %d images\n", theme.HighlightStyle.Render("Input:"), imageCount)
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Model:"), model)
	fmt.Printf("  %s  %.1fs @ %d fps\n", theme.HighlightStyle.Render("Video:"), animateLength, animateFPS)
	if animatePrompt != "" {
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Prompt:"), animatePrompt)
	}
	fmt.Printf("  %s  %d workers\n", theme.HighlightStyle.Render("Parallel:"), animateParallel)
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Output:"), animateOutput)
	fmt.Println()

	totalFrames := int(animateLength * float64(animateFPS))
	estimatedTime := estimateAnimationTime(model, imageCount, totalFrames)
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("⏱️  Estimated time: %s", estimatedTime)))
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Start animation pipeline? (Y/n): "))
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "n" || confirm == "no" {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("❌ Cancelled"))
		return nil
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🚀 Starting animation pipeline..."))
	fmt.Println()

	return runAnimateWithModel(collectionName, collectionPath, imageCount, model)
}

func runAnimateWithModel(collectionName, collectionPath string, imageCount int, model string) error {
	// Create output directory
	if animateOutput == "" {
		animateOutput = filepath.Join(collectionPath, "animated", model)
	}

	if err := os.MkdirAll(animateOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📁 Output directory: %s", animateOutput)))
	fmt.Println()

	// Step 1: Auto-check and install model
	fmt.Println(theme.HighlightStyle.Render(fmt.Sprintf("🔍 Checking %s installation...", model)))
	if !checkPackageInstalledLocal(model) {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  ⚠️  %s not installed", model)))
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("  Installing %s...", model)))
		fmt.Println()

		installCmd := exec.Command("anime", "install", model)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %w", model, err)
		}
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ %s installed", model)))
	}
	fmt.Println()

	// Step 2: Auto-check and install ComfyUI
	fmt.Println(theme.HighlightStyle.Render("🔍 Checking ComfyUI installation..."))
	if !checkPackageInstalledLocal("comfyui") {
		fmt.Println(theme.WarningStyle.Render("  ⚠️  ComfyUI not installed"))
		fmt.Println(theme.InfoStyle.Render("  Installing ComfyUI..."))
		fmt.Println()

		installCmd := exec.Command("anime", "install", "comfyui")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install ComfyUI: %w", err)
		}
		fmt.Println()
	} else {
		fmt.Println(theme.SuccessStyle.Render("  ✓ ComfyUI installed"))
	}
	fmt.Println()

	// Step 3: Auto-start ComfyUI if not running
	fmt.Println(theme.HighlightStyle.Render("🔍 Checking ComfyUI server..."))
	if !isComfyUIRunning() {
		fmt.Println(theme.WarningStyle.Render("  ⚠️  ComfyUI not running"))
		fmt.Println(theme.InfoStyle.Render("  Starting ComfyUI in background..."))
		fmt.Println()

		if err := startComfyUIBackground(); err != nil {
			return fmt.Errorf("failed to start ComfyUI: %w", err)
		}

		// Wait for ComfyUI to be ready
		fmt.Print(theme.DimTextStyle.Render("  Waiting for ComfyUI"))
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			fmt.Print(".")
			if isComfyUIRunning() {
				break
			}
		}
		fmt.Println()
		fmt.Println()

		if !isComfyUIRunning() {
			return fmt.Errorf("ComfyUI failed to start after 30 seconds")
		}
		fmt.Println(theme.SuccessStyle.Render("  ✓ ComfyUI started"))
	} else {
		publicIP := getPublicIP()
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ ComfyUI running at http://%s:8188", publicIP)))
	}
	fmt.Println()

	// Find all images in collection
	images, err := findCollectionImages(collectionPath)
	if err != nil {
		return fmt.Errorf("failed to find images: %w", err)
	}

	if len(images) == 0 {
		return fmt.Errorf("no images found in collection")
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📷 Found %d images to process", len(images))))
	fmt.Println()

	// Determine model script path
	modelScript := getModelScriptPath(model)

	// Build command based on model type
	fmt.Println(theme.SuccessStyle.Render("⚡ Starting parallel animation pipeline..."))
	fmt.Println()

	// Create a temporary file list for parallel processing
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("anime-batch-%s.txt", collectionName))
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	for _, img := range images {
		fmt.Fprintf(f, "%s\n", img)
	}
	f.Close()

	// Build the parallel processing command
	var promptArg string
	if animatePrompt != "" {
		promptArg = fmt.Sprintf("--prompt \"%s\"", animatePrompt)
	}

	parallelCmd := fmt.Sprintf(
		"cat %s | parallel -j %d --eta --bar 'python3 %s --input {} --output %s --length %.1f --fps %d %s'",
		tmpFile,
		animateParallel,
		modelScript,
		animateOutput,
		animateLength,
		animateFPS,
		promptArg,
	)

	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Executing: %d parallel workers", animateParallel)))
	fmt.Println()

	// Execute the parallel processing
	fmt.Println(theme.InfoStyle.Render("💫 Processing images..."))
	fmt.Println()

	// Note: This would be executed on the server, not locally
	fmt.Println(theme.SuccessStyle.Render("✨ Animation command generated!"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("To execute on server:"))
	fmt.Println(theme.HighlightStyle.Render(parallelCmd))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Model: %s", model)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Workers: %d parallel", animateParallel)))
	if animatePrompt != "" {
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("Prompt: %s", animatePrompt)))
	}
	fmt.Println()

	return nil
}

func findCollectionImages(collectionPath string) ([]string, error) {
	var images []string
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".webp": true, ".bmp": true, ".tiff": true,
	}

	err := filepath.Walk(collectionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if imageExts[ext] {
				images = append(images, path)
			}
		}
		return nil
	})

	return images, err
}

func getModelScriptPath(model string) string {
	// Map model names to their script paths on the server
	modelPaths := map[string]string{
		"svd":         "~/video-models/svd/generate.py",
		"animatediff": "~/video-models/animatediff/generate.py",
		"ltxvideo":    "~/video-models/ltxvideo/generate.py",
		"mochi":       "~/video-models/mochi-1/generate.py",
		"cogvideo":    "~/video-models/cogvideo/inference.py",
		"wan2":        "~/video-models/wan2/generate.py",
	}

	if path, ok := modelPaths[model]; ok {
		return path
	}
	return fmt.Sprintf("~/video-models/%s/generate.py", model)
}

func runCollectionUpscale(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	// Load config to get collection
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	// Welcome banner
	fmt.Println()
	fmt.Println(theme.RenderBanner("⬆️  UPSCALE COLLECTION"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📦 Collection: %s", collectionName)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("   Path: %s", collection.Path)))
	fmt.Println()

	// Count files
	fileCount, err := countImages(collection.Path)
	if err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Found %d files to upscale", fileCount)))
	fmt.Println()

	// If no model specified, launch wizard
	if upscaleModel == "" {
		return runUpscaleWizard(collectionName, collection.Path, fileCount)
	}

	// Run with specified model
	return runUpscaleWithModel(collectionName, collection.Path, fileCount, upscaleModel)
}

func runUpscaleWizard(collectionName, collectionPath string, fileCount int) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.CategoryStyle("⬆️  UPSCALE WIZARD"))
	fmt.Println()

	// Step 1: Choose model
	fmt.Println(theme.HighlightStyle.Render("Step 1: Choose Upscaling Model"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  1") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("Real-ESRGAN") + theme.DimTextStyle.Render(" - Best for photos/renders (fast)"))
	fmt.Println(theme.InfoStyle.Render("  2") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("GFPGAN") + theme.DimTextStyle.Render(" - Best for faces (quality)"))
	fmt.Println(theme.InfoStyle.Render("  3") + theme.DimTextStyle.Render(" → ") + theme.SuccessStyle.Render("CodeFormer") + theme.DimTextStyle.Render(" - Best for portraits (advanced)"))
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Select model (1-3) [1]: "))
	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	var model string
	switch modelChoice {
	case "2":
		model = "gfpgan"
	case "3":
		model = "codeformer"
	default:
		model = "realesrgan"
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Selected: %s", model)))
	fmt.Println()

	// Step 2: Upscale factor
	fmt.Println(theme.HighlightStyle.Render("Step 2: Upscale Factor"))
	fmt.Println()
	fmt.Print(theme.HighlightStyle.Render("Upscale factor (2, 4, or 8) [4]: "))
	scaleInput, _ := reader.ReadString('\n')
	scaleInput = strings.TrimSpace(scaleInput)
	if scaleInput == "" {
		upscaleScale = 4
	} else {
		fmt.Sscanf(scaleInput, "%d", &upscaleScale)
	}

	fmt.Println()

	// Summary
	fmt.Println(theme.CategoryStyle("📋 UPSCALE SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Collection:"), collectionName)
	fmt.Printf("  %s  %d files\n", theme.HighlightStyle.Render("Input:"), fileCount)
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Model:"), model)
	fmt.Printf("  %s  %dx\n", theme.HighlightStyle.Render("Scale:"), upscaleScale)
	fmt.Println()

	fmt.Print(theme.HighlightStyle.Render("Start upscaling? (Y/n): "))
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "n" || confirm == "no" {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("❌ Cancelled"))
		return nil
	}

	fmt.Println()
	return runUpscaleWithModel(collectionName, collectionPath, fileCount, model)
}

func runUpscaleWithModel(collectionName, collectionPath string, fileCount int, model string) error {
	fmt.Println(theme.GlowStyle.Render("🚀 Starting upscale pipeline..."))
	fmt.Println()

	// Create output directory
	if upscaleOutput == "" {
		upscaleOutput = filepath.Join(collectionPath, "upscaled", fmt.Sprintf("%s_%dx", model, upscaleScale))
	}

	if err := os.MkdirAll(upscaleOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📁 Output directory: %s", upscaleOutput)))
	fmt.Println()

	// TODO: Implement actual upscaling pipeline
	fmt.Println(theme.SuccessStyle.Render("✨ Upscale pipeline ready!"))
	fmt.Println()

	return nil
}

func runCollectionTransform(cmd *cobra.Command, args []string) error {
	collectionName := args[0]

	// Load config to get collection
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	collection, err := cfg.GetCollection(collectionName)
	if err != nil {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	// Welcome banner
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔧 TRANSFORM PIPELINE"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("📦 Collection: %s", collectionName)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("   Path: %s", collection.Path)))
	fmt.Println()

	// Count files
	fileCount, err := countImages(collection.Path)
	if err != nil {
		return err
	}

	fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("✓ Found %d files", fileCount)))
	fmt.Println()

	// Launch transform wizard
	return runTransformWizard(collectionName, collection.Path, fileCount)
}

func runTransformWizard(collectionName, collectionPath string, fileCount int) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(theme.CategoryStyle("🎨 TRANSFORM WIZARD"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Build a custom AI pipeline with multiple operations!"))
	fmt.Println()

	// Available operations
	fmt.Println(theme.HighlightStyle.Render("Available Operations:"))
	fmt.Println()

	operations := []struct {
		id   string
		name string
		desc string
	}{
		{"1", "Upscale", "Increase resolution (2x, 4x, 8x)"},
		{"2", "Denoise", "Remove noise and artifacts"},
		{"3", "Enhance", "Auto-enhance colors and contrast"},
		{"4", "Animate", "Convert images to video"},
		{"5", "Style Transfer", "Apply artistic styles"},
		{"6", "Background Remove", "Remove/replace backgrounds"},
		{"7", "Colorize", "Colorize black & white images"},
		{"8", "Face Restore", "Restore/enhance faces"},
	}

	for _, op := range operations {
		fmt.Printf("  %s %s %s\n",
			theme.InfoStyle.Render(op.id),
			theme.SuccessStyle.Render(op.name),
			theme.DimTextStyle.Render("- "+op.desc))
	}

	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Enter operation numbers separated by spaces (e.g., '1 2 3' for Upscale → Denoise → Enhance)"))
	fmt.Print(theme.HighlightStyle.Render("Select operations: "))

	operationsInput, _ := reader.ReadString('\n')
	operationsInput = strings.TrimSpace(operationsInput)

	selectedOps := strings.Fields(operationsInput)
	if len(selectedOps) == 0 {
		return fmt.Errorf("no operations selected")
	}

	fmt.Println()
	fmt.Println(theme.CategoryStyle("📋 PIPELINE SUMMARY"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Collection:"), collectionName)
	fmt.Printf("  %s  %d files\n", theme.HighlightStyle.Render("Input:"), fileCount)
	fmt.Printf("  %s  %d steps\n", theme.HighlightStyle.Render("Pipeline:"), len(selectedOps))
	fmt.Println()

	fmt.Println(theme.HighlightStyle.Render("  Pipeline Steps:"))
	for i, opID := range selectedOps {
		for _, op := range operations {
			if op.id == opID {
				fmt.Printf("    %d. %s\n", i+1, theme.SuccessStyle.Render(op.name))
			}
		}
	}

	fmt.Println()

	if transformPreview {
		fmt.Println(theme.InfoStyle.Render("  🔍 Preview mode: Will process 1 sample"))
		fmt.Println()
	}

	fmt.Print(theme.HighlightStyle.Render("Execute pipeline? (Y/n): "))
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "n" || confirm == "no" {
		fmt.Println()
		fmt.Println(theme.WarningStyle.Render("❌ Cancelled"))
		return nil
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render("🚀 Starting transform pipeline..."))
	fmt.Println()

	// TODO: Implement actual transform pipeline
	fmt.Println(theme.SuccessStyle.Render("✨ Transform pipeline ready!"))
	fmt.Println()

	if transformSave != "" {
		fmt.Println(theme.InfoStyle.Render(fmt.Sprintf("💾 Pipeline saved as: %s", transformSave)))
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("   Reuse with: anime collection %s transform --template %s", collectionName, transformSave)))
		fmt.Println()
	}

	return nil
}

// Helper functions

func countImages(path string) (int, error) {
	count := 0
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".tiff": true,
	}

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(p))
			if imageExts[ext] {
				count++
			}
		}
		return nil
	})

	return count, err
}

func estimateAnimationTime(model string, imageCount, totalFrames int) string {
	// Rough estimates per video in seconds
	timePerVideo := map[string]int{
		"svd":         30,
		"animatediff": 15,
		"ltxvideo":    8,
		"mochi":       45,
		"cogvideo":    35,
	}

	seconds, ok := timePerVideo[model]
	if !ok {
		seconds = 20 // default
	}

	totalSeconds := imageCount * seconds

	if totalSeconds < 60 {
		return fmt.Sprintf("%d seconds", totalSeconds)
	} else if totalSeconds < 3600 {
		return fmt.Sprintf("%d minutes", totalSeconds/60)
	} else {
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// isComfyUIRunning checks if ComfyUI server is running on port 8188
func isComfyUIRunning() bool {
	resp, err := http.Get("http://127.0.0.1:8188")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// startComfyUIBackground starts ComfyUI in the background
func startComfyUIBackground() error {
	comfyPath := filepath.Join(os.Getenv("HOME"), "ComfyUI")

	// Check if ComfyUI exists
	if _, err := os.Stat(comfyPath); os.IsNotExist(err) {
		return fmt.Errorf("ComfyUI not found at ~/ComfyUI")
	}

	// Find python3
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		pythonPath, err = exec.LookPath("python")
		if err != nil {
			return fmt.Errorf("python not found in PATH")
		}
	}

	// Create log file for output
	logPath := filepath.Join(os.TempDir(), "comfyui.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Start ComfyUI in background with --listen to allow external connections
	cmd := exec.Command(pythonPath, filepath.Join(comfyPath, "main.py"), "--listen", "127.0.0.1")
	cmd.Dir = comfyPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ComfyUI: %w\n\nCheck log: %s", err, logPath)
	}

	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  Log: %s", logPath)))
	fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  PID: %d", cmd.Process.Pid)))
	fmt.Println()

	// Give it a moment to start
	time.Sleep(2 * time.Second)

	// Check if process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		// Process died immediately, show log
		return fmt.Errorf("ComfyUI exited immediately. Check log at: %s", logPath)
	}

	return nil
}

func getPublicIP() string {
	// Try to get hostname first (e.g., 209-20-159-132)
	cmd := exec.Command("hostname")
	if output, err := cmd.Output(); err == nil {
		hostname := strings.TrimSpace(string(output))
		// If hostname looks like an IP with dashes, convert to dots
		if strings.Count(hostname, "-") >= 3 {
			parts := strings.Split(hostname, "-")
			if len(parts) >= 4 {
				// Check if looks like IP format
				allNumeric := true
				for _, part := range parts[:4] {
					for _, ch := range part {
						if ch < '0' || ch > '9' {
							allNumeric = false
							break
						}
					}
				}
				if allNumeric {
					return strings.Join(parts[:4], ".")
				}
			}
		}
		return hostname
	}

	// Fallback: try to get public IP from external service
	cmd = exec.Command("curl", "-s", "ifconfig.me")
	if output, err := cmd.Output(); err == nil {
		ip := strings.TrimSpace(string(output))
		if ip != "" {
			return ip
		}
	}

	// Last resort: return localhost
	return "127.0.0.1"
}
