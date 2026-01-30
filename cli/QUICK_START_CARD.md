# 🎌 anime Quick Start Card

## First Time User?

```bash
anime walkthrough
```
**↑ Start here!** Interactive 2-minute tutorial with live demos.

---

## Essential Commands

### 📦 Browse & Install
```bash
anime packages              # Browse available packages
anime interactive           # Visual package selector
anime install core pytorch  # Install packages
```

### 🚀 Run Services
```bash
anime run comfyui          # Start ComfyUI (instead of cd ~/ComfyUI && python main.py)
anime run ollama           # Start Ollama server
anime run jupyter          # Start Jupyter Lab
anime run tensorboard      # Start TensorBoard
```

### 🏥 Diagnostics
```bash
anime doctor               # Diagnose issues
anime models               # List downloaded models
anime tree                 # View all commands
```

### ⚙️ Setup
```bash
anime wizard               # Guided setup for your node
```

### 🎬 Collection Workflows
```bash
anime collection list              # List asset collections
anime collection photos animate    # Batch animate images to video
anime collection photos upscale    # Batch upscale images/videos
anime collection photos transform  # Custom multi-step pipelines
```

---

## Quick Workflows

### Get Started with AI/ML
```bash
anime install core pytorch ollama
anime run ollama
```

### Image/Video Generation
```bash
anime install comfyui
anime run comfyui
# Open: http://127.0.0.1:8188
```

### Development Setup
```bash
anime wizard               # Choose "Development" option
anime run jupyter
```

### Batch Process Collections
```bash
anime collection create renders ~/my-images
anime collection renders upscale --scale 4
anime collection renders animate --model svd
```

---

## Pro Tips 💡

1. **Launch services easily:**
   - Use `anime run <service>` instead of manual commands
   - Works from any directory
   - Shows helpful URLs

2. **Quick LLM access:**
   ```bash
   anime ollama run llama2  # Run models directly
   ```

3. **Troubleshooting:**
   ```bash
   anime doctor             # Diagnoses most issues
   ```

4. **Discover commands:**
   ```bash
   anime tree               # Visual command tree
   ```

---

## Common Tasks

| Task | Command |
|------|---------|
| See what's available | `anime packages` |
| Install something | `anime install <name>` |
| Start ComfyUI | `anime run comfyui` |
| Start Ollama | `anime run ollama` |
| Check what's installed | `anime models` |
| Fix problems | `anime doctor` |
| Configure system | `anime wizard` |
| Learn anime | `anime walkthrough` |
| List collections | `anime collection list` |
| Animate images | `anime collection <name> animate` |
| Upscale images/videos | `anime collection <name> upscale` |
| Custom pipelines | `anime collection <name> transform` |

---

## Getting Help

```bash
anime --help              # Main help
anime <command> --help    # Command-specific help
anime tree                # All commands visual
anime walkthrough         # Interactive tutorial
```

---

## Aliases

These work too:
- `anime tutorial` → `anime walkthrough`
- `anime start` → `anime run`
- `anime launch` → `anime run`

---

**Need more?** Run `anime walkthrough` for a complete interactive guide! 🚀
