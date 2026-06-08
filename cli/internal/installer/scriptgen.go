package installer

import (
	"fmt"
	"strings"
)

// ── shared bash fragments ───────────────────────────────────────────

const bashWaitForDpkg = `wait_for_dpkg() {
    local max_wait=300
    local waited=0
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
          sudo fuser /var/lib/dpkg/lock >/dev/null 2>&1 || \
          sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
        if [ $waited -ge $max_wait ]; then
            echo "Timeout waiting for package manager lock"
            return 1
        fi
        echo "Waiting for other package managers to finish... ($waited/$max_wait s)"
        sleep 5
        waited=$((waited + 5))
    done
    return 0
}
`

const bashFixBrokenPackages = `fix_broken_packages() {
    echo "==> Checking for broken packages..."
    if ! sudo dpkg --audit >/dev/null 2>&1; then
        echo "==> Fixing broken packages..."
        wait_for_dpkg || return 1
        sudo dpkg --configure -a || true
        wait_for_dpkg || return 1
        sudo apt-get install -f -y || true
    fi
}
`

const bashSudo = `SUDO=""; [ "$(id -u)" -ne 0 ] && SUDO=sudo
`

const bashPreamble = "#!/bin/bash\nset -e\n"

// ── Ollama model scripts ────────────────────────────────────────────

type ollamaModel struct {
	displayName string // e.g. "Llama 3.3 70B"
	pullTag     string // e.g. "llama3.3:70b"
}

func (m ollamaModel) script() string {
	return fmt.Sprintf(`#!/bin/bash
set -e
echo "==> Downloading %s via Ollama"
ollama pull %s
echo "==> %s installed successfully"
ollama list
`, m.displayName, m.pullTag, m.displayName)
}

var ollamaModels = map[string]ollamaModel{
	"llama-3.3-70b":      {"Llama 3.3 70B", "llama3.3:70b"},
	"llama-3.3-8b":       {"Llama 3.3 8B", "llama3.3:8b"},
	"mistral":            {"Mistral 7B", "mistral:latest"},
	"mixtral":            {"Mixtral 8x7B", "mixtral:8x7b"},
	"qwen3-235b":         {"Qwen3 235B MoE", "qwen3:235b"},
	"qwen3-32b":          {"Qwen3 32B", "qwen3:32b"},
	"qwen3-30b":          {"Qwen3 30B MoE", "qwen3:30b"},
	"qwen3-14b":          {"Qwen3 14B", "qwen3:14b"},
	"qwen3-8b":           {"Qwen3 8B", "qwen3:8b"},
	"qwen3-4b":           {"Qwen3 4B", "qwen3:4b"},
	"deepseek-coder-33b": {"DeepSeek Coder 33B", "deepseek-coder:33b"},
	"deepseek-v3":        {"DeepSeek V3", "deepseek-v3"},
	"phi-3.5":            {"Phi-3.5 Mini", "phi3.5:latest"},
	"phi-4":              {"Phi-4 14B", "phi4:latest"},
	"deepseek-r1-8b":     {"DeepSeek-R1 8B", "deepseek-r1:8b"},
	"deepseek-r1-70b":    {"DeepSeek-R1 70B", "deepseek-r1:70b"},
	"gemma3-4b":          {"Gemma3 4B", "gemma3:4b"},
	"gemma3-12b":         {"Gemma3 12B", "gemma3:12b"},
	"gemma3-27b":         {"Gemma3 27B", "gemma3:27b"},
	"llama-3.2-1b":       {"Llama 3.2 1B", "llama3.2:1b"},
	"llama-3.2-3b":       {"Llama 3.2 3B", "llama3.2:3b"},
	"qwen3-coder-30b":    {"Qwen3-Coder 30B MoE", "qwen3-coder:30b"},
	"command-r-7b":       {"Command-R 7B", "command-r:7b"},
}

// ── Ollama batch model scripts ──────────────────────────────────────

type ollamaBatch struct {
	displayName string
	sizeClass   string
	models      []string // pull tags
}

func (b ollamaBatch) script() string {
	var sb strings.Builder
	sb.WriteString(bashPreamble)
	fmt.Fprintf(&sb, "echo \"==> Downloading %s models (%s)\"\n", b.displayName, b.sizeClass)
	for _, m := range b.models {
		fmt.Fprintf(&sb, "ollama pull %s &\n", m)
	}
	sb.WriteString("wait\n")
	fmt.Fprintf(&sb, "echo \"==> %s models downloaded\"\nollama list\n", b.displayName)
	return sb.String()
}

var ollamaBatches = map[string]ollamaBatch{
	"models-small": {"small", "1-8B", []string{
		"llama3.2:1b", "llama3.2:3b", "gemma3:4b", "mistral:latest", "llama3.3:8b", "qwen3:8b",
	}},
	"models-medium": {"medium", "8-34B", []string{
		"deepseek-r1:8b", "phi4:latest", "gemma3:12b", "qwen3:14b", "qwen3-coder:30b", "qwen3:32b", "mixtral:8x7b", "deepseek-coder:33b",
	}},
	"models-large": {"large", "70B+", []string{
		"gemma3:27b", "deepseek-r1:70b", "llama3.3:70b", "qwen3:235b",
	}},
}

// ── HuggingFace download scripts ────────────────────────────────────

type hfDownload struct {
	displayName string
	repo        string   // HF repo id
	targetDir   string   // parent dir to cd into (e.g. ~/ComfyUI/models/checkpoints)
	localDir    string   // --local-dir value (e.g. "svd-xt"); downloaded inside targetDir
	files       []string // specific files (empty = whole repo)
}

func (h hfDownload) script() string {
	var sb strings.Builder
	sb.WriteString(bashPreamble)
	fmt.Fprintf(&sb, "echo \"==> Installing %s\"\n", h.displayName)
	fmt.Fprintf(&sb, "mkdir -p %s\ncd %s\n", h.targetDir, h.targetDir)

	if len(h.files) > 0 {
		for _, f := range h.files {
			fmt.Fprintf(&sb, "huggingface-cli download %s %s --local-dir . --max-workers 8\n", h.repo, f)
		}
	} else {
		fmt.Fprintf(&sb, "huggingface-cli download %s --local-dir %s --max-workers 8\n",
			h.repo, h.localDir)
	}

	fmt.Fprintf(&sb, "echo \"==> %s installed successfully\"\n", h.displayName)
	return sb.String()
}

var hfDownloads = map[string]hfDownload{
	"svd-xt":              {"SVD-XT 1.1", "stabilityai/stable-video-diffusion-img2vid-xt-1-1", "~/ComfyUI/models/checkpoints", "svd-xt", nil},
	"sd3.5-large":         {"SD 3.5 Large", "stabilityai/stable-diffusion-3.5-large", "~/ComfyUI/models/checkpoints", "sd3.5-large", nil},
	"sd3.5-large-turbo":   {"SD 3.5 Large Turbo", "stabilityai/stable-diffusion-3.5-large-turbo", "~/ComfyUI/models/checkpoints", "sd3.5-large-turbo", nil},
	"sd3.5-medium":        {"SD 3.5 Medium", "stabilityai/stable-diffusion-3.5-medium", "~/ComfyUI/models/checkpoints", "sd3.5-medium", nil},
	"sdxl-turbo":          {"SDXL Turbo", "stabilityai/sdxl-turbo", "~/ComfyUI/models/checkpoints", "sdxl-turbo", nil},
	"sdxl-lightning":      {"SDXL Lightning", "ByteDance/SDXL-Lightning", "~/ComfyUI/models/checkpoints", "sdxl-lightning", nil},
	"playground-v2.5":     {"Playground v2.5", "playgroundai/playground-v2.5-1024px-aesthetic", "~/ComfyUI/models/checkpoints", "playground-v2.5", nil},
	"pixart-sigma":        {"PixArt-Sigma", "PixArt-alpha/PixArt-Sigma-XL-2-1024-MS", "~/ComfyUI/models/checkpoints", "pixart-sigma", nil},
	"kandinsky-3":         {"Kandinsky 3", "ai-forever/Kandinsky3.1", "~/ComfyUI/models/checkpoints", "kandinsky-3", nil},
	"kolors":              {"Kolors", "Kwai-Kolors/Kolors", "~/ComfyUI/models/checkpoints", "kolors", nil},
	"aurasr":              {"AuraSR", "fal/AuraSR", "~/ComfyUI/models/upscale_models", "aurasr", nil},
	"supir":               {"SUPIR", "Kijai/SUPIR_pruned", "~/ComfyUI/models/checkpoints", "supir", nil},
	"sd-inpainting":       {"SD 1.5 Inpainting", "runwayml/stable-diffusion-inpainting", "~/ComfyUI/models/checkpoints", "sd-inpainting", nil},
	"sdxl-inpainting":     {"SDXL Inpainting", "diffusers/stable-diffusion-xl-1.0-inpainting-0.1", "~/ComfyUI/models/checkpoints", "sdxl-inpainting", nil},
	"controlnet-canny":    {"ControlNet Canny", "lllyasviel/sd-controlnet-canny", "~/ComfyUI/models/controlnet", "controlnet-canny", nil},
	"controlnet-depth":    {"ControlNet Depth", "lllyasviel/sd-controlnet-depth", "~/ComfyUI/models/controlnet", "controlnet-depth", nil},
	"controlnet-openpose": {"ControlNet OpenPose", "lllyasviel/sd-controlnet-openpose", "~/ComfyUI/models/controlnet", "controlnet-openpose", nil},
	"ip-adapter":          {"IP-Adapter", "h94/IP-Adapter", "~/ComfyUI/models/ipadapter", "ip-adapter", nil},
	"ip-adapter-faceid":   {"IP-Adapter FaceID", "h94/IP-Adapter-FaceID", "~/ComfyUI/models/ipadapter", "ip-adapter-faceid", nil},
	"instantid":           {"InstantID", "InstantX/InstantID", "~/ComfyUI/models/instantid", "instantid", nil},
}

// ── pip+hf one-liner scripts ────────────────────────────────────────

type pipHFDownload struct {
	displayName string
	pipDeps     string // pip packages to install first
	repo        string
	targetDir   string // relative to $HOME
}

func (p pipHFDownload) script() string {
	return fmt.Sprintf(`#!/bin/bash
set -e
echo "==> Installing %s"
pip3 install --upgrade %s
python3 -c "from huggingface_hub import snapshot_download; snapshot_download('%s', local_dir='$HOME/%s')"
echo "==> %s installed successfully"
`, p.displayName, p.pipDeps, p.repo, p.targetDir, p.displayName)
}

var pipHFDownloads = map[string]pipHFDownload{
	"cogvideox-1.5": {"CogVideoX 1.5 5B", "diffusers transformers accelerate", "THUDM/CogVideoX1.5-5B", "models/cogvideox-1.5"},
	"cogvideox-i2v": {"CogVideoX 1.5 I2V", "diffusers transformers accelerate", "THUDM/CogVideoX1.5-5B-I2V", "models/cogvideox-i2v"},
	"hunyuan-video": {"HunyuanVideo", "diffusers transformers accelerate", "tencent/HunyuanVideo", "models/hunyuan-video"},
	"pyramid-flow":  {"Pyramid Flow", "diffusers transformers accelerate", "rain1011/pyramid-flow-miniflux", "models/pyramid-flow"},
}

// ── ComfyUI model download (with mkdir + aria2c/wget fallback) ──────

type comfyUIModel struct {
	displayName string
	subDir      string // under ~/ComfyUI/models/
	url         string // direct download URL
	filename    string // target filename
}

func (c comfyUIModel) script() string {
	return fmt.Sprintf(`#!/bin/bash
set -e
echo "==> Installing %s for ComfyUI"

if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/%s"
cd "$HOME/ComfyUI/models/%s"

if [ -f "%s" ]; then
    echo "%s already installed"
    exit 0
fi

echo "==> Downloading %s with multi-connection download..."
if command -v aria2c &> /dev/null; then
    aria2c -x 16 -s 16 %s
else
    wget -c %s
fi
echo "==> %s installed successfully"
`, c.displayName, c.subDir, c.subDir, c.filename, c.displayName, c.displayName, c.url, c.url, c.displayName)
}

// ── Flux model scripts (ComfyUI unet dir + HF single-file) ─────────

type fluxModel struct {
	displayName string
	repo        string
	filename    string
}

func (f fluxModel) script() string {
	return fmt.Sprintf(`#!/bin/bash
set -e
echo "==> Installing %s for ComfyUI"

if [ ! -d "$HOME/ComfyUI" ]; then
    echo "Error: ComfyUI not found. Please install 'comfyui' package first: anime install comfyui"
    exit 1
fi

mkdir -p "$HOME/ComfyUI/models/unet"
cd "$HOME/ComfyUI/models/unet"

if [ -f "%s" ]; then
    echo "%s already installed"
    exit 0
fi

echo "==> Downloading %s model with parallel downloads..."
huggingface-cli download %s %s --local-dir . --max-workers 8

echo "==> %s installed successfully"
echo "Model location: ~/ComfyUI/models/unet/%s"
`, f.displayName, f.filename, f.displayName, f.displayName, f.repo, f.filename, f.displayName, f.filename)
}

var fluxModels = map[string]fluxModel{
	"flux-dev":     {"Flux.1 Dev", "black-forest-labs/FLUX.1-dev", "flux1-dev.safetensors"},
	"flux-schnell": {"Flux.1 Schnell", "black-forest-labs/FLUX.1-schnell", "flux1-schnell.safetensors"},
	"flux2":        {"Flux 2 (FP8)", "black-forest-labs/FLUX.2-fp8", "flux2-fp8.safetensors"},
}

// ── SDXL/SD checkpoint downloads (ComfyUI + aria2c/wget) ────────────

var comfyUIDirectModels = map[string]comfyUIModel{
	"sdxl": {
		"Stable Diffusion XL", "checkpoints",
		"https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0/resolve/main/sd_xl_base_1.0.safetensors",
		"sd_xl_base_1.0.safetensors",
	},
	"sd15": {
		"Stable Diffusion 1.5", "checkpoints",
		"https://huggingface.co/runwayml/stable-diffusion-v1-5/resolve/main/v1-5-pruned-emaonly.safetensors",
		"v1-5-pruned-emaonly.safetensors",
	},
	"gfpgan": {
		"GFPGAN", "facerestore_models",
		"https://github.com/TencentARC/GFPGAN/releases/download/v1.3.0/GFPGANv1.4.pth",
		"GFPGANv1.4.pth",
	},
}

// ── Simple tool installs ────────────────────────────────────────────

type toolInstall struct {
	displayName string
	checkCmd    string // command to check if installed
	installBody string // bash to run if not installed
}

func (t toolInstall) script() string {
	return fmt.Sprintf(`#!/bin/bash
set -e
echo "==> Installing %s"
if command -v %s &> /dev/null; then
    echo "%s already installed"
    exit 0
fi
%s
echo "==> %s installed"
`, t.displayName, t.checkCmd, t.displayName, t.installBody, t.displayName)
}

var toolInstalls = map[string]toolInstall{
	"rust": {"Rust Toolchain", "rustc",
		"curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable\nsource \"$HOME/.cargo/env\"\necho \"==> Installing common Rust tools...\"\ncargo install sccache cargo-watch cargo-edit\nrustc --version\ncargo --version"},
	"uv": {"uv (fast Python package manager)", "uv",
		"curl -LsSf https://astral.sh/uv/install.sh | sh\necho \"Note: Restart your shell or run 'source $HOME/.local/bin/env' to update PATH\""},
	"tailscale": {"Tailscale", "tailscale",
		"curl -fsSL https://tailscale.com/install.sh | sh\necho \"Run 'sudo tailscale up' to connect to your tailnet\""},
	"tmux": {"tmux", "tmux", bashSudo + "$SUDO apt-get update -y -qq && $SUDO apt-get install -y -qq tmux\ntmux -V"},
	"htop": {"htop & btop", "htop", bashSudo +
		"$SUDO apt-get update -y -qq\ncommand -v htop &> /dev/null && echo \"htop already installed\" || $SUDO apt-get install -y -qq htop\ncommand -v btop &> /dev/null && echo \"btop already installed\" || $SUDO apt-get install -y -qq btop 2>/dev/null || echo \"btop not in apt, skipping\""},
	"neovim": {"Neovim", "nvim", bashSudo +
		"$SUDO apt-get update -y -qq\n$SUDO apt-get install -y -qq software-properties-common\n$SUDO add-apt-repository -y ppa:neovim-ppa/stable 2>/dev/null || true\n$SUDO apt-get update -y -qq\n$SUDO apt-get install -y -qq neovim\nnvim --version | head -1"},
	"caddy": {"Caddy", "caddy", bashSudo +
		"$SUDO apt-get update -y -qq\n$SUDO apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl\ncurl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | $SUDO gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg\ncurl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | $SUDO tee /etc/apt/sources.list.d/caddy-stable.list\n$SUDO apt-get update -y -qq\n$SUDO apt-get install -y -qq caddy\ncaddy version"},
}

// ── Video model installs (PyTorch prereq + git clone + HF download) ──

type videoModel struct {
	displayName  string
	dirName      string   // under ~/video-models/
	repo         string   // git clone URL
	hfRepo       string   // HF model repo
	hfLocalDir   string   // --local-dir for hf download
	extraPipDeps []string // beyond diffusers/transformers/accelerate
	editable     bool     // pip install -e .
}

func (v videoModel) script() string {
	var sb strings.Builder
	sb.WriteString(bashPreamble)
	fmt.Fprintf(&sb, "echo \"==> Installing %s\"\n\n", v.displayName)

	// PyTorch check
	sb.WriteString("if ! python3 -c \"import torch\" 2>/dev/null; then\n")
	sb.WriteString("    echo \"Error: PyTorch not found. Please install 'pytorch' package first: anime install pytorch\"\n")
	sb.WriteString("    exit 1\nfi\n\n")

	fmt.Fprintf(&sb, "mkdir -p ~/video-models\ncd ~/video-models\n")
	fmt.Fprintf(&sb, "if [ -d \"%s\" ]; then\n    echo \"%s already installed\"\n    exit 0\nfi\n\n", v.dirName, v.displayName)

	// Deps
	sb.WriteString("echo \"==> Checking dependencies...\"\n")
	deps := []string{"diffusers", "transformers", "accelerate"}
	deps = append(deps, v.extraPipDeps...)
	installDeps := "INSTALL_DEPS=\"\"\n"
	for _, d := range deps {
		installDeps += fmt.Sprintf("python3 -c \"import %s\" 2>/dev/null || INSTALL_DEPS=\"$INSTALL_DEPS %s\"\n", pythonImportName(d), d)
	}
	installDeps += "if [ -n \"$INSTALL_DEPS\" ]; then\n"
	installDeps += "    echo \"==> Installing missing dependencies:$INSTALL_DEPS\"\n"
	installDeps += "    pip3 install --upgrade-strategy only-if-needed $INSTALL_DEPS\n"
	installDeps += "else\n    echo \"==> All dependencies already installed\"\nfi\n\n"
	sb.WriteString(installDeps)

	fmt.Fprintf(&sb, "git clone %s %s\ncd %s\n", v.repo, v.dirName, v.dirName)

	if v.editable {
		sb.WriteString("pip3 install -e . --upgrade-strategy only-if-needed\n")
	} else {
		sb.WriteString("if [ -f \"requirements.txt\" ]; then\n")
		sb.WriteString("    echo \"==> Installing additional requirements (excluding torch/cuda to avoid conflicts)...\"\n")
		fmt.Fprintf(&sb, "    grep -v -E \"^torch|^nvidia-|^triton\" requirements.txt > /tmp/%s-requirements-filtered.txt || true\n", v.dirName)
		fmt.Fprintf(&sb, "    if [ -s /tmp/%s-requirements-filtered.txt ]; then\n", v.dirName)
		fmt.Fprintf(&sb, "        pip3 install -r /tmp/%s-requirements-filtered.txt --upgrade-strategy only-if-needed\n", v.dirName)
		sb.WriteString("    else\n        echo \"==> No additional requirements needed (torch/cuda already installed)\"\n    fi\nfi\n")
	}

	fmt.Fprintf(&sb, "echo \"==> Downloading %s model weights with parallel downloads...\"\n", v.displayName)
	fmt.Fprintf(&sb, "huggingface-cli download %s --local-dir %s --max-workers 8\n", v.hfRepo, v.hfLocalDir)
	fmt.Fprintf(&sb, "echo \"==> %s installed successfully\"\n", v.displayName)

	return sb.String()
}

var videoModels = map[string]videoModel{
	"mochi": {
		displayName: "Mochi-1 Video Generation", dirName: "mochi-1",
		repo: "https://github.com/genmoai/mochi", hfRepo: "genmo/mochi-1-preview", hfLocalDir: "./weights",
		extraPipDeps: []string{"einops"},
	},
	"cogvideo": {
		displayName: "CogVideoX-5B", dirName: "cogvideo",
		repo: "https://github.com/THUDM/CogVideo", hfRepo: "THUDM/CogVideoX-5b", hfLocalDir: "./weights",
	},
	"ltxvideo": {
		displayName: "LTXVideo", dirName: "ltxvideo",
		repo: "https://github.com/Lightricks/LTX-Video", hfRepo: "Lightricks/LTX-Video", hfLocalDir: "./checkpoints",
	},
	"wan2": {
		displayName: "Wan2.2 (Image-to-Video)", dirName: "wan2",
		repo: "https://github.com/alibaba/Wan.git", hfRepo: "Alibaba-PAI/wan2.2", hfLocalDir: "./checkpoints",
		extraPipDeps: []string{"einops", "imageio", "imageio-ffmpeg"},
	},
	"opensora": {
		displayName: "Open-Sora 2.0", dirName: "open-sora",
		repo: "https://github.com/hpcaitech/Open-Sora", hfRepo: "hpcai-tech/OpenSora-STDiT-v3", hfLocalDir: "./pretrained_models",
		editable: true,
	},
}

// ── ComfyUI custom node installs ────────────────────────────────────

type comfyUINode struct {
	displayName     string
	dirName         string // under ~/ComfyUI/custom_nodes/
	repo            string
	hasRequirements bool   // install requirements.txt
	runInstall      bool   // run install.py after pip
	modelURL        string // optional model download URL
	modelDir        string // optional model download target dir (relative to ~/ComfyUI/)
}

func (n comfyUINode) script() string {
	var sb strings.Builder
	sb.WriteString(bashPreamble)
	fmt.Fprintf(&sb, "echo \"==> Installing %s for ComfyUI\"\n", n.displayName)

	// Prerequisite: ComfyUI must be installed
	sb.WriteString("COMFY_DIR=\"$HOME/ComfyUI\"\n")
	sb.WriteString("if [ ! -d \"$COMFY_DIR\" ]; then\n")
	sb.WriteString("    echo \"Error: ComfyUI not found. Install comfyui first.\"\n")
	sb.WriteString("    exit 1\nfi\n")

	sb.WriteString("cd \"$COMFY_DIR/custom_nodes\"\n")

	// Idempotency: exit early if already installed
	fmt.Fprintf(&sb, "if [ -d \"%s\" ]; then\n    echo \"%s already installed\"\n    exit 0\nfi\n", n.dirName, n.displayName)
	fmt.Fprintf(&sb, "git clone %s\n", n.repo)

	if n.hasRequirements {
		fmt.Fprintf(&sb, "cd %s\nif [ -f requirements.txt ]; then pip3 install -r requirements.txt; fi\n", n.dirName)
	}
	if n.runInstall {
		if !n.hasRequirements {
			fmt.Fprintf(&sb, "cd %s\n", n.dirName)
		}
		sb.WriteString("python3 install.py\n")
	}
	if n.modelURL != "" {
		fmt.Fprintf(&sb, "mkdir -p \"$COMFY_DIR/%s\"\ncd \"$COMFY_DIR/%s\"\n", n.modelDir, n.modelDir)
		fmt.Fprintf(&sb, "if command -v aria2c &> /dev/null; then\n    aria2c -x 16 -s 16 %s\nelse\n    wget -c %s\nfi\n", n.modelURL, n.modelURL)
	}
	fmt.Fprintf(&sb, "echo \"==> %s installed successfully\"\n", n.displayName)
	return sb.String()
}

var comfyUINodes = map[string]comfyUINode{
	"svd": {
		displayName: "Stable Video Diffusion", dirName: "ComfyUI-VideoHelperSuite",
		repo:     "https://github.com/Kosinkadink/ComfyUI-VideoHelperSuite.git",
		modelURL: "https://huggingface.co/stabilityai/stable-video-diffusion-img2vid-xt/resolve/main/svd_xt.safetensors",
		modelDir: "models/checkpoints",
	},
	"animatediff": {
		displayName: "AnimateDiff", dirName: "ComfyUI-AnimateDiff-Evolved",
		repo:     "https://github.com/Kosinkadink/ComfyUI-AnimateDiff-Evolved.git",
		modelURL: "https://huggingface.co/guoyww/animatediff/resolve/main/mm_sd_v15_v2.ckpt",
		modelDir: "models/animatediff_models",
	},
	"rife": {
		displayName: "RIFE", dirName: "ComfyUI-Frame-Interpolation",
		repo: "https://github.com/Fannovel16/ComfyUI-Frame-Interpolation.git",
		hasRequirements: true, runInstall: true,
	},
	"film": {
		displayName: "FILM", dirName: "ComfyUI-Frame-Interpolation",
		repo: "https://github.com/Fannovel16/ComfyUI-Frame-Interpolation.git",
		hasRequirements: true, runInstall: true,
	},
	"i2v-adapter": {
		displayName: "I2V-Adapter", dirName: "I2V-Adapter",
		repo: "https://github.com/KlingTeam/I2V-Adapter.git",
		hasRequirements: true,
	},
}

// ── Multi-tool installs (jq+yq, bat+eza, glow, fzf, lazygit) ───────
// These have enough conditional logic that a table doesn't help.
// Kept as literal scripts below.

var multiToolScripts = map[string]string{
	"jq": bashPreamble + `echo "==> Installing jq & yq"
` + bashSudo + `$SUDO apt-get update -y -qq
command -v jq &> /dev/null && echo "jq already installed" || $SUDO apt-get install -y -qq jq
if ! command -v yq &> /dev/null; then
    ARCH=$(dpkg --print-architecture 2>/dev/null || echo amd64)
    $SUDO curl -fsSL "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${ARCH}" -o /usr/local/bin/yq
    $SUDO chmod +x /usr/local/bin/yq
fi
echo "==> jq & yq installed"
`,
	"fzf": bashPreamble + `echo "==> Installing fzf"
if command -v fzf &> /dev/null; then
    echo "fzf already installed: $(fzf --version)"
    exit 0
fi
` + bashSudo + `$SUDO apt-get update -y -qq && $SUDO apt-get install -y -qq fzf 2>/dev/null || {
    git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
    ~/.fzf/install --all --no-bash --no-zsh --no-fish
}
echo "==> fzf installed"
`,
	"bat": bashPreamble + `echo "==> Installing bat & eza"
` + bashSudo + `$SUDO apt-get update -y -qq
command -v bat &> /dev/null || command -v batcat &> /dev/null && echo "bat already installed" || {
    $SUDO apt-get install -y -qq bat 2>/dev/null || $SUDO apt-get install -y -qq batcat 2>/dev/null
    [ -f /usr/bin/batcat ] && ! [ -f /usr/local/bin/bat ] && $SUDO ln -s /usr/bin/batcat /usr/local/bin/bat
}
if ! command -v eza &> /dev/null; then
    $SUDO apt-get install -y -qq eza 2>/dev/null || {
        ARCH=$(dpkg --print-architecture 2>/dev/null || echo amd64)
        curl -fsSL "https://github.com/eza-community/eza/releases/latest/download/eza_${ARCH}-unknown-linux-gnu.tar.gz" | tar xz -C /tmp
        $SUDO mv /tmp/eza /usr/local/bin/eza
    }
fi
echo "==> bat & eza installed"
`,
	"glow": bashPreamble + `echo "==> Installing Glow (terminal markdown renderer)"
if command -v glow &> /dev/null; then
    echo "glow already installed: $(glow --version 2>/dev/null || echo ok)"
    exit 0
fi
` + bashSudo + `if command -v go &> /dev/null; then
    go install github.com/charmbracelet/glow@latest
elif command -v brew &> /dev/null; then
    brew install glow
else
    mkdir -p /tmp/glow-install && cd /tmp/glow-install
    ARCH=$(dpkg --print-architecture 2>/dev/null || echo amd64)
    curl -fsSL "https://github.com/charmbracelet/glow/releases/latest/download/glow_Linux_${ARCH}.tar.gz" | tar xz
    $SUDO mv glow /usr/local/bin/glow
    rm -rf /tmp/glow-install
fi
echo "==> Glow installed"
`,
	"real-esrgan": bashPreamble + `echo "==> Installing Real-ESRGAN for ComfyUI"
mkdir -p ~/ComfyUI/models/upscale_models
cd ~/ComfyUI/models/upscale_models
wget -nc https://github.com/xinntao/Real-ESRGAN/releases/download/v0.1.0/RealESRGAN_x4plus.pth || true
wget -nc https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.2.4/RealESRGAN_x4plus_anime_6B.pth || true
echo "==> Real-ESRGAN installed successfully"
`,
	"lazygit": bashPreamble + `echo "==> Installing lazygit"
if command -v lazygit &> /dev/null; then
    echo "lazygit already installed: $(lazygit --version | head -1)"
    exit 0
fi
` + bashSudo + `ARCH=$(dpkg --print-architecture 2>/dev/null || echo amd64)
[ "$ARCH" = "amd64" ] && ARCH="x86_64" || ARCH="arm64"
LAZYGIT_VERSION=$(curl -fsSL "https://api.github.com/repos/jesseduffield/lazygit/releases/latest" | grep '"tag_name"' | cut -d'"' -f4 | tr -d v)
curl -fsSL "https://github.com/jesseduffield/lazygit/releases/download/v${LAZYGIT_VERSION}/lazygit_${LAZYGIT_VERSION}_Linux_${ARCH}.tar.gz" | tar xz -C /tmp lazygit
$SUDO mv /tmp/lazygit /usr/local/bin/lazygit
echo "==> lazygit installed"
lazygit --version | head -1
`,
}

// ── helpers ─────────────────────────────────────────────────────────

// pythonImportName maps pip package names to their python import names
// where they differ.
func pythonImportName(pkg string) string {
	switch pkg {
	case "imageio-ffmpeg":
		return "imageio_ffmpeg"
	default:
		return pkg
	}
}

// GenerateScripts builds the complete Scripts map from all generators
// plus the retained complex script literals.
func GenerateScripts() map[string]string {
	scripts := make(map[string]string, 92)

	// Ollama individual models
	for id, m := range ollamaModels {
		scripts[id] = m.script()
	}

	// Ollama batch downloads
	for id, b := range ollamaBatches {
		scripts[id] = b.script()
	}

	// HuggingFace downloads (simple)
	for id, h := range hfDownloads {
		scripts[id] = h.script()
	}

	// pip + HF one-liners
	for id, p := range pipHFDownloads {
		scripts[id] = p.script()
	}

	// Flux models
	for id, f := range fluxModels {
		scripts[id] = f.script()
	}

	// ComfyUI direct download models
	for id, c := range comfyUIDirectModels {
		scripts[id] = c.script()
	}

	// Video models (git clone + HF)
	for id, v := range videoModels {
		scripts[id] = v.script()
	}

	// ComfyUI custom nodes
	for id, n := range comfyUINodes {
		scripts[id] = n.script()
	}

	// Simple tool installs
	for id, t := range toolInstalls {
		scripts[id] = t.script()
	}

	// Multi-tool scripts (literal)
	for id, s := range multiToolScripts {
		scripts[id] = s
	}

	// Complex scripts (retained as literals in scripts_complex.go)
	for id, s := range complexScripts {
		scripts[id] = s
	}

	return scripts
}
