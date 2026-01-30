# 🌸 ANIME Desktop

**A**utomated **N**eural **I**nfrastructure **M**anagement **E**nvironment

Beautiful anime-themed desktop application for deploying and managing Lambda GH200 AI infrastructure.

## ✨ Features

- 🎨 **Anime Sakura Theme**: Beautiful pink and purple gradients with floating sakura petals
- 📦 **Package Management**: Visual package selection with category filtering
- 🔮 **Configuration Wizard**: Guided setup for inference, training, art, and development nodes
- 🖥️ **Server Management**: SSH-based remote deployment and monitoring
- ⚡ **Real-time Progress**: Live installation progress tracking
- 🎯 **Smart Dependencies**: Automatic dependency resolution

## 🚀 Quick Start

### Prerequisites

- **Node.js** 18+ and npm
- **Rust** 1.70+
- **Operating System**: macOS, Linux, or Windows

### Installation

```bash
cd anime-desktop

# Install dependencies
npm install

# Run in development mode with hot-reload
npm run tauri:dev

# Build for production
npm run tauri:build
```

## 📦 Packages Available

### Foundation
- **Core System**: NVIDIA drivers, CUDA 12.4, Node.js, Docker
- **Python & AI Libs**: Python 3.11+, numpy, scipy, pandas

### ML Framework
- **PyTorch Stack**: PyTorch, transformers, diffusers, accelerate

### LLM Runtime
- **Ollama**: LLM server with systemd integration

### Models
- **Small Models** (7-8B): Mistral, Llama 3.3, Qwen 2.5
- **Medium Models** (14-34B): Qwen 14B, Mixtral, DeepSeek Coder
- **Large Models** (70B+): Llama 70B, Qwen 72B, DeepSeek V3

### Video Generation
- **Mochi-1**: 10B parameter video generation
- **Stable Video Diffusion**: Image-to-video for ComfyUI
- **AnimateDiff**: SD animation module
- **CogVideoX-5B**: Text-to-video
- **Open-Sora 2.0**: High-quality video gen
- **LTXVideo**: Fast latent transformer video

### Applications
- **ComfyUI**: Node-based UI with custom nodes
- **Claude Code**: Official Anthropic CLI

## 🎨 Theme & Design

The app features a custom anime-inspired sakura theme:

- **Sakura Pink** (#FF69B4): Primary accent color
- **Electric Blue** (#00D9FF): Interactive elements
- **Neon Purple** (#BD93F9): Secondary accents
- **Mint Green** (#50FA7B): Success states
- **Sunset Orange** (#FFB86C): Warnings

Floating sakura petals animate across the background for that authentic anime aesthetic.

## 🏗️ Architecture

```
anime-desktop/
├── src/                      # React frontend
│   ├── components/           # UI components
│   │   ├── PackageGrid.tsx  # Package selection grid
│   │   ├── WizardFlow.tsx   # Configuration wizard
│   │   ├── ServerManager.tsx # SSH server management
│   │   └── InstallProgress.tsx # Real-time progress
│   ├── store/               # Zustand state management
│   ├── types/               # TypeScript type definitions
│   └── App.tsx              # Main application
│
└── src-tauri/               # Rust backend
    ├── src/
    │   ├── packages.rs      # Package definitions
    │   ├── installer.rs     # Installation logic
    │   ├── ssh.rs           # SSH connection handling
    │   └── main.rs          # Tauri commands
    └── Cargo.toml           # Rust dependencies
```

## 🔧 Development

### Frontend Development

```bash
# Run Vite dev server only
npm run dev

# TypeScript type checking
npm run build
```

### Backend Development

```bash
# Build Rust backend
cd src-tauri
cargo build

# Run tests
cargo test

# Format code
cargo fmt
```

### IPC Commands

The frontend communicates with the Rust backend via Tauri commands:

```typescript
// Get all packages
const packages = await invoke('get_packages_command')

// Resolve dependencies
const resolved = await invoke('resolve_dependencies_command', {
  packageIds: ['pytorch', 'ollama']
})

// Install packages
await invoke('install_packages', {
  packageIds: ['core', 'python']
})
```

## 🎯 Roadmap

### Phase 1: MVP (Current)
- [x] Project scaffolding
- [x] Anime theme system
- [x] Package definitions
- [x] Basic UI components
- [ ] SSH connection handling
- [ ] Installation execution

### Phase 2: Core Features
- [ ] Full wizard implementation
- [ ] Real-time progress tracking
- [ ] Server persistence
- [ ] Installation logs viewer
- [ ] Error handling and retry

### Phase 3: Advanced
- [ ] Multi-server orchestration
- [ ] Cluster coordination
- [ ] Auto-updater
- [ ] System tray integration
- [ ] Notification system

### Phase 4: Polish
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Binary size reduction
- [ ] Code signing
- [ ] Distribution packages

## 📝 Usage

1. **Select Packages**: Browse and select packages from the grid
2. **Configure Node**: Use the wizard to configure your deployment
3. **Add Servers**: Configure SSH connection to Lambda servers
4. **Install**: Click "Install Selected" and monitor progress
5. **Deploy**: Packages will be installed on remote servers via SSH

## 🛠️ Building for Production

```bash
# Build optimized release
npm run tauri:build

# Output locations:
# macOS: src-tauri/target/release/bundle/macos/ANIME.app
# Linux: src-tauri/target/release/bundle/appimage/anime_0.1.0_amd64.AppImage
# Windows: src-tauri/target/release/bundle/msi/ANIME_0.1.0_x64_en-US.msi
```

## 🤝 Contributing

This project was scaffolded by Claude Code based on the ANIME CLI project. Contributions welcome!

## 📄 License

[Your License Here]

## 🙏 Acknowledgments

- Built with [Tauri](https://tauri.app/) - Lightweight desktop framework
- UI powered by [React](https://react.dev/) and [Tailwind CSS](https://tailwindcss.com/)
- Backend in [Rust](https://www.rust-lang.org/) for performance and safety
- Inspired by anime aesthetics and sakura blossoms 🌸

---

**ANIME v0.1.0** - Automated Neural Infrastructure Management Environment
*Lambda GH200 deployment made beautiful* ⚡🌸
