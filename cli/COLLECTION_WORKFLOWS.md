# 🎬 Collection Workflow Commands

## Overview

Three new batch processing commands for asset collections:

1. **`animate`** - Convert images to videos with AI models
2. **`upscale`** - Enhance resolution of images/videos
3. **`transform`** - Build custom multi-step pipelines

## Quick Start

```bash
# List your collections
anime collection list

# Animate images to video
anime collection photos animate

# Upscale images or videos
anime collection photos upscale

# Custom transformation pipeline
anime collection photos transform
```

---

## 1. anime collection animate

Batch animate images in a collection using AI video generation models.

### Supported Models

| Model | Best For | Speed | Quality |
|-------|----------|-------|---------|
| **SVD** (Stable Video Diffusion) | General purpose | Fast | High |
| **AnimateDiff** | Character animation | Medium | Very High |
| **LTXVideo** | Cinematic motion | Fast | High |
| **Mochi-1** | Smooth transitions | Medium | High |
| **CogVideoX** | Text-guided motion | Slow | Very High |

### Usage

#### Interactive Wizard (Recommended)
```bash
anime collection photos animate
```

The wizard will guide you through:
1. **Model Selection** - Choose animation backend
2. **Video Settings** - FPS and length
3. **Output Options** - Directory and organization
4. **Confirmation** - Review before processing

#### Direct Command
```bash
# Use specific model
anime collection photos animate --model svd

# Custom settings
anime collection photos animate --model ltxvideo --fps 24 --length 3

# Batch processing
anime collection photos animate --model animatediff --batch 4
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-m, --model` | string | (wizard) | Animation model (svd, animatediff, ltxvideo, mochi, cogvideo) |
| `-f, --fps` | int | 24 | Frames per second |
| `-l, --length` | float | 2.0 | Video length in seconds |
| `-b, --batch` | int | 1 | Batch size for GPU processing |
| `-o, --output` | string | `<collection>/animated` | Output directory |

### Output Structure

```
photos/
├── animated/
│   ├── svd_24fps_2s/
│   │   ├── image1_animated.mp4
│   │   ├── image2_animated.mp4
│   │   └── metadata.json
│   └── ltxvideo_30fps_3s/
│       ├── image1_animated.mp4
│       └── metadata.json
```

### Examples

```bash
# Quick start with SVD
anime collection renders animate --model svd

# High-quality character animation
anime collection characters animate --model animatediff --fps 30 --length 4

# Cinematic output
anime collection scenes animate --model ltxvideo --fps 24 --length 5 --batch 2

# Text-guided animation (requires prompts)
anime collection concepts animate --model cogvideo
```

---

## 2. anime collection upscale

Batch upscale images or videos in a collection using AI upscaling models.

### Supported Models

| Model | Type | Best For | Speed | Max Scale |
|-------|------|----------|-------|-----------|
| **Real-ESRGAN** | Image | General images | Fast | 8x |
| **GFPGAN** | Image | Faces/portraits | Medium | 4x |
| **CodeFormer** | Image | Face restoration | Slow | 4x |
| **BasicVSR++** | Video | Video upscaling | Slow | 4x |

### Usage

#### Interactive Wizard (Recommended)
```bash
anime collection photos upscale
```

The wizard will guide you through:
1. **Model Selection** - Choose upscaling backend
2. **Scale Factor** - 2x, 4x, or 8x
3. **Quality Preset** - Draft, Balanced, or Quality
4. **Output Options** - Directory and batch size
5. **Confirmation** - Review settings

#### Direct Command
```bash
# 4x upscale with Real-ESRGAN
anime collection photos upscale --scale 4 --model realesrgan

# High-quality face restoration
anime collection portraits upscale --model gfpgan --quality high

# Video upscaling
anime collection videos upscale --model basicvsr --scale 2 --batch 2
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-m, --model` | string | (wizard) | Upscaling model (realesrgan, gfpgan, codeformer, basicvsr) |
| `-s, --scale` | int | 4 | Upscale factor (2, 4, or 8) |
| `-q, --quality` | string | balanced | Quality preset (draft, balanced, quality) |
| `-b, --batch` | int | 4 | Batch size for GPU processing |
| `-o, --output` | string | `<collection>/upscaled` | Output directory |

### Quality Presets

| Preset | Speed | Quality | VRAM | Best For |
|--------|-------|---------|------|----------|
| **draft** | Very Fast | Good | Low | Quick previews |
| **balanced** | Fast | Very Good | Medium | Most use cases |
| **quality** | Slow | Excellent | High | Final output |

### Output Structure

```
photos/
├── upscaled/
│   ├── 4x_realesrgan_balanced/
│   │   ├── image1_4x.png
│   │   ├── image2_4x.png
│   │   └── metadata.json
│   └── 8x_realesrgan_quality/
│       ├── image1_8x.png
│       └── metadata.json
```

### Examples

```bash
# Quick 2x upscale
anime collection photos upscale --scale 2

# Maximum quality 8x
anime collection artwork upscale --scale 8 --quality high

# Face-focused restoration
anime collection portraits upscale --model codeformer --quality high

# Batch video upscaling
anime collection footage upscale --model basicvsr --batch 4 --scale 2
```

---

## 3. anime collection transform

Interactive wizard to create custom AI transformation pipelines.

### Available Operations

| Category | Operations |
|----------|------------|
| **Enhancement** | Upscale, Denoise, Enhance, Face Restore |
| **Generation** | Animate, Style Transfer, Colorize |
| **Editing** | Background Remove, Crop, Resize |
| **Post** | Watermark, Rename, Format Convert |

### Usage

#### Interactive Wizard (Recommended)
```bash
anime collection photos transform
```

The wizard will guide you through:
1. **Operation Selection** - Choose operations to chain
2. **Operation Order** - Arrange processing sequence
3. **Settings** - Configure each operation
4. **Pipeline Preview** - Test on 1 sample
5. **Template Save** - Save for reuse
6. **Execution** - Batch process

#### Using Templates
```bash
# Save pipeline
anime collection photos transform --save enhance-upscale

# Reuse pipeline
anime collection videos transform --template enhance-upscale

# Preview before running
anime collection renders transform --template studio-polish --preview
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-t, --template` | string | - | Use saved pipeline template |
| `-s, --save` | string | - | Save pipeline as template |
| `-p, --preview` | bool | false | Preview mode (process 1 sample) |

### Example Pipelines

#### Workflow 1: Studio Enhancement
```
1. Denoise (remove artifacts)
2. Upscale (4x with Real-ESRGAN)
3. Enhance (auto color/contrast)
4. Face Restore (CodeFormer)
```

#### Workflow 2: Artistic Transform
```
1. Upscale (2x for detail)
2. Style Transfer (apply artistic style)
3. Enhance (fine-tune colors)
```

#### Workflow 3: Video Production
```
1. Denoise (clean frames)
2. Upscale (2x resolution)
3. Animate (convert to video)
4. Watermark (add branding)
```

### Output Structure

```
photos/
├── transformed/
│   ├── pipeline_enhance-upscale/
│   │   ├── step1_denoise/
│   │   ├── step2_upscale/
│   │   ├── step3_enhance/
│   │   ├── final/
│   │   │   ├── image1_final.png
│   │   │   └── image2_final.png
│   │   └── pipeline.yaml
│   └── pipeline.yaml (saved template)
```

### Template Format

```yaml
name: enhance-upscale
description: Denoise, upscale, and enhance
operations:
  - type: denoise
    settings:
      strength: 0.8
  - type: upscale
    settings:
      model: realesrgan
      scale: 4
      quality: balanced
  - type: enhance
    settings:
      auto_color: true
      auto_contrast: true
```

### Examples

```bash
# Create custom workflow
anime collection photos transform

# Quick enhancement pipeline
anime collection renders transform --template quick-enhance

# Test before processing all
anime collection large-set transform --template studio --preview

# Save new workflow
anime collection photos transform --save my-workflow
```

---

## Common Workflows

### Workflow: Image to High-Quality Video

```bash
# Step 1: Upscale images
anime collection photos upscale --scale 4 --model realesrgan

# Step 2: Animate upscaled images
anime collection photos-upscaled animate --model animatediff --fps 30
```

### Workflow: Face Photo Restoration

```bash
# Use transform pipeline
anime collection old-photos transform

# Select operations:
# 1. Denoise
# 2. Face Restore (CodeFormer)
# 3. Upscale (4x GFPGAN)
# 4. Enhance
```

### Workflow: Batch Style Transfer

```bash
# Transform with style transfer
anime collection artwork transform

# Select:
# 1. Upscale (2x for detail)
# 2. Style Transfer (pick style)
# 3. Enhance (finalize)
```

---

## Performance Tips

### GPU Optimization

```bash
# Increase batch size for better GPU utilization
anime collection photos upscale --batch 8

# For large collections, use quality presets
anime collection huge-set upscale --quality draft  # Fast preview
anime collection huge-set upscale --quality quality  # Final pass
```

### Pipeline Efficiency

1. **Order matters**: Put fast operations first (denoise, resize)
2. **Use preview mode**: Test on 1 sample before full batch
3. **Save templates**: Reuse successful pipelines
4. **Incremental processing**: Process in stages, review between

### VRAM Management

| Task | VRAM | Recommended Batch |
|------|------|-------------------|
| Upscale 2x | 4GB | 8 |
| Upscale 4x | 8GB | 4 |
| Upscale 8x | 12GB | 2 |
| Animate (SVD) | 12GB | 2 |
| Animate (AnimateDiff) | 16GB | 1 |
| Face Restore | 6GB | 4 |

---

## Troubleshooting

### Command Not Found

```bash
# Verify installation
anime --version

# Rebuild if needed
cd ~/lambda/anime-cli
make build
```

### Collection Not Found

```bash
# List collections
anime collection list

# Create if missing
anime collection create photos /path/to/images
```

### Model Not Available

```bash
# Check installed models
anime models

# Install missing model
anime install <model-package>
```

### Out of Memory

```bash
# Reduce batch size
anime collection photos upscale --batch 2

# Use draft quality
anime collection photos upscale --quality draft

# Process smaller collections
anime collection subset upscale
```

---

## Integration Examples

### With ComfyUI

```bash
# Generate images in ComfyUI, then:
anime collection comfy-outputs upscale --scale 4
anime collection comfy-outputs animate --model svd
```

### With Ollama

```bash
# Generate image prompts with LLM
anime run ollama run llama2 "Generate 10 scene descriptions"

# Create images, then transform
anime collection generated-scenes transform
```

### With Workflows

```bash
# List available workflows
anime workflow

# Run workflow that includes collection processing
anime workflow run enhance-and-animate --collection photos
```

---

## Future Enhancements

Planned features:

- [ ] **Real-time preview** - Live preview during processing
- [ ] **Parallel GPU** - Multi-GPU support for faster processing
- [ ] **Cloud processing** - Offload to Lambda instances
- [ ] **Progress resume** - Resume interrupted pipelines
- [ ] **Metadata preservation** - Keep EXIF/metadata
- [ ] **Audio support** - Add audio to animated videos
- [ ] **Comparison view** - Before/after visualization
- [ ] **Cost estimation** - Estimate processing time/cost

---

**Built with:** v1.0.113
**Date:** November 21, 2025
**Status:** ✅ Complete and ready to use!

Try it now:
```bash
anime collection list
anime collection photos animate
```
