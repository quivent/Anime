# 🚀 Running Services with anime

## Problem Solved

**Before:** `cd ~/ComfyUI && python main.py` 😫

**Now:** `anime run comfyui` 🎉

## New Command: `anime run`

Quick launcher for installed services with smart defaults and helpful output.

## Usage

```bash
anime run <service> [args]
```

Aliases: `start`, `launch`

## Supported Services

### 1. ComfyUI

Start the ComfyUI web interface:

```bash
anime run comfyui
```

**Output:**
```
🎨 Starting ComfyUI...

  URL: http://127.0.0.1:8188
  Path: ~/ComfyUI

  Press Ctrl+C to stop
```

**Features:**
- Auto-detects `~/ComfyUI` installation
- Finds Python automatically
- Shows helpful URL and path info
- Passes through any additional arguments

**Example with args:**
```bash
anime run comfyui --port 8189
```

### 2. Ollama

Start the Ollama LLM server:

```bash
anime run ollama
# or
anime run ollama serve
```

**Output:**
```
🤖 Starting Ollama server...

  URL: http://127.0.0.1:11434
  Press Ctrl+C to stop
```

**Features:**
- Defaults to `serve` mode
- Can pass any ollama command
- Interactive passthrough

**Examples:**
```bash
anime run ollama           # Start server
anime run ollama run llama2  # Run a model directly
```

### 3. Jupyter

Start Jupyter Lab:

```bash
anime run jupyter
```

**Output:**
```
📓 Starting Jupyter Lab...

  Press Ctrl+C to stop
```

**Features:**
- Defaults to Jupyter Lab
- Opens browser automatically
- Passes through arguments

**Example:**
```bash
anime run jupyter --port 8889
```

### 4. TensorBoard

Start TensorBoard:

```bash
anime run tensorboard
```

**Output:**
```
📊 Starting TensorBoard...

  URL: http://127.0.0.1:6006
  Press Ctrl+C to stop
```

**Features:**
- Defaults to `--logdir=./logs`
- Shows URL
- Custom args supported

**Example:**
```bash
anime run tensorboard --logdir=/path/to/logs
```

## Smart Features

### 1. Auto-Detection
- Finds executables in PATH
- Locates installation directories
- Provides helpful error messages

### 2. Helpful Errors

If a service isn't installed:

```
❌ ComfyUI not found

Install it with:
  anime install comfyui
```

### 3. Clean Output
- Colorized status messages
- URL display for web services
- Path information
- Stop instructions

## Integration with Tutorial

The walkthrough now includes actionable suggestions:

```
✨ TUTORIAL COMPLETE

🚀 Suggested Next Actions:

  1️⃣  anime install core pytorch ollama
      Get started with AI/ML basics

  2️⃣  anime run comfyui
      Launch ComfyUI (after installing)

  3️⃣  anime run ollama
      Start Ollama LLM server

💡 Pro Tips:

    • Use anime run comfyui instead of cd ~/ComfyUI && python main.py
    • Use anime ollama run llama2 for quick LLM access
    • Run anime doctor if you encounter issues
```

## Comparison

### Before 😫

```bash
# ComfyUI
cd ~/ComfyUI
python main.py

# Ollama
ollama serve

# Jupyter
jupyter lab

# TensorBoard
tensorboard --logdir=./logs
```

**Issues:**
- Need to remember paths
- No helpful output
- Different command patterns
- No installation checks

### After 🎉

```bash
# Everything through anime
anime run comfyui
anime run ollama
anime run jupyter
anime run tensorboard
```

**Benefits:**
- ✅ Consistent interface
- ✅ Auto-detection
- ✅ Helpful output
- ✅ Smart defaults
- ✅ Error messages with solutions
- ✅ Works from any directory

## Examples

### Quick Start Workflow

```bash
# 1. Install packages
anime install comfyui ollama

# 2. Start services
anime run comfyui    # Terminal 1
anime run ollama     # Terminal 2

# 3. Use them!
# ComfyUI: http://127.0.0.1:8188
# Ollama: http://127.0.0.1:11434
```

### Advanced Usage

```bash
# ComfyUI with custom settings
anime run comfyui --listen 0.0.0.0 --port 8189

# Ollama with specific model
anime run ollama run qwen3:8b

# Jupyter with custom port
anime run jupyter --port 8889 --no-browser

# TensorBoard with custom logs
anime run tensorboard --logdir=/mnt/data/experiments
```

## Future Services

Planned additions:
- `anime run vscode` - VS Code with AI extensions
- `anime run gradio` - Gradio demos
- `anime run streamlit` - Streamlit apps

---

**Try it now:** `anime run --help`
