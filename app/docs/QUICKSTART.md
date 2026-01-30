# 🚀 ANIME Desktop - Quick Start Guide

Get up and running in 5 minutes!

## Prerequisites

Make sure you have:
- **Node.js** 18+ → `node --version`
- **Rust** 1.70+ → `rustc --version`
- **npm** or **yarn**

Don't have Rust? Install it:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

## Installation Steps

### 1. Navigate to Project

```bash
cd /Users/joshkornreich/lambda/anime-desktop
```

### 2. Install Dependencies

```bash
# Install Node.js dependencies
npm install
```

This will install:
- React 19
- Tauri 2.1
- Tailwind CSS
- Zustand (state management)
- TypeScript

### 3. Run in Development Mode

```bash
npm run tauri:dev
```

This will:
1. Start the Vite dev server on port 7890
2. Compile the Rust backend
3. Launch the desktop app with hot-reload

**First launch takes 2-3 minutes** as Rust compiles all dependencies.

### 4. Explore the App!

The app should now be running with:
- 🌸 Beautiful sakura-themed UI
- 📦 Package grid with 15+ packages
- 🎨 Floating sakura petals animation
- ⚡ Responsive sidebar navigation

## Project Structure

```
anime-desktop/
├── src/                     # React frontend
│   ├── components/          # UI components
│   ├── types/               # TypeScript types
│   ├── store/               # State management
│   └── App.tsx              # Main app
├── src-tauri/              # Rust backend
│   ├── src/                # Rust source
│   │   ├── main.rs         # Tauri commands
│   │   ├── packages.rs     # Package definitions
│   │   ├── installer.rs    # Install logic
│   │   └── ssh.rs          # SSH handling
│   └── Cargo.toml          # Rust deps
└── package.json            # Node deps
```

## Common Commands

```bash
# Development with hot-reload
npm run tauri:dev

# Build frontend only
npm run build

# Build production app
npm run tauri:build

# Format Rust code
cd src-tauri && cargo fmt

# Run Rust tests
cd src-tauri && cargo test
```

## Building for Production

```bash
npm run tauri:build
```

**Output locations:**
- **macOS**: `src-tauri/target/release/bundle/macos/ANIME.app`
- **Linux**: `src-tauri/target/release/bundle/appimage/anime_0.1.0_amd64.AppImage`
- **Windows**: `src-tauri/target/release/bundle/msi/ANIME_0.1.0_x64_en-US.msi`

## Troubleshooting

### "Command not found: tauri"
```bash
npm install
```

### Rust compilation errors
```bash
cd src-tauri
cargo clean
cargo build
```

### Port 7890 already in use
Edit `vite.config.ts` and change the port number.

### Hot-reload not working
Restart with:
```bash
killall anime-desktop
npm run tauri:dev
```

## Next Steps

1. **Explore Components**: Check out `src/components/` for the UI
2. **Add Features**: Modify `src-tauri/src/main.rs` to add Tauri commands
3. **Customize Theme**: Edit `tailwind.config.js` for color changes
4. **Port Go Logic**: Copy more functionality from the ANIME CLI

## Development Tips

- **Fast Rust Recompiles**: Use `cargo-watch` for auto-rebuild
- **React DevTools**: Install the browser extension
- **Tauri DevTools**: Press `Cmd+Shift+I` (macOS) or `Ctrl+Shift+I` (Windows/Linux)
- **Console Logs**: Rust `println!` appears in terminal, JS `console.log` in DevTools

## Getting Help

- Tauri Docs: https://tauri.app/
- React Docs: https://react.dev/
- Tailwind Docs: https://tailwindcss.com/

---

Ready to deploy? The CLI version is in `/Users/joshkornreich/lambda` - this desktop app is a beautiful GUI wrapper! 🌸⚡
