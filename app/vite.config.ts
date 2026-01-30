import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],

  // Tauri expects a fixed port, will fail if it's in use
  server: {
    port: 7890,
    strictPort: true,
  },

  // Tauri uses a custom protocol for serving files
  build: {
    // Tauri supports es2021
    target: ['es2021', 'chrome100', 'safari13'],
    // Don't minify for debug builds
    minify: !process.env.TAURI_DEBUG ? 'esbuild' : false,
    // Produce sourcemaps for debug builds
    sourcemap: !!process.env.TAURI_DEBUG,
  },

  // Clear screen on build
  clearScreen: false,

  // Configure server host
  envPrefix: ['VITE_', 'TAURI_'],
})
