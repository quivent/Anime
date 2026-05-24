package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/joshkornreich/anime/internal/validate"
	"github.com/spf13/cobra"
)

const (
	lithosHome     = "~/lithos"
	lithosCompiler = "~/lithos/compiler/lithos"
	lithosFontAtlas = "~/lithos/arch/gpu/lithos-font.metallib"
)

var lithosDeployDomain  string
var lithosDeployPort    string
var lithosServePort     string
var lithosDeployNoSSL   bool
var lithosCompileTarget string
var lithosServer        string
var lithosResponse      string

var lithosCmd = &cobra.Command{
	Use:   "lithos",
	Short: "Compile, deploy, and serve Lithos programs",
	Long: `Integration with the Lithos computation language.

Compile .ls programs to native binaries, deploy to servers,
and serve inference endpoints — all from one CLI.

Lithos compiles to ARM64, Metal/AIR, WASM, GLSL — no runtime, no VM.`,
	Run: runLithosHelp,
}

var lithosCompileCmd = &cobra.Command{
	Use:   "compile <file.ls> [output]",
	Short: "Compile a Lithos program",
	Long: `Compile a .ls file to a native binary.

Target is auto-detected from the current platform, or specified with --target.

Examples:
  anime lithos compile mykernel.ls
  anime lithos compile mykernel.ls output.bin
  anime lithos compile mykernel.ls --target arm64
  anime lithos compile mykernel.ls --target wasm`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLithosCompile,
}

var lithosDeployCmd = &cobra.Command{
	Use:   "deploy <file.ls> <server>",
	Short: "Compile for ARM64 Linux, ship to server, set up as service",
	Long: `Full deployment pipeline for Lithos programs:

  1. Compile .ls → ARM64 binary
  2. Ship binary to remote server
  3. Set up as systemd service
  4. Configure nginx + SSL (if --domain given)

Examples:
  anime lithos deploy myapi.ls wings --domain api.example.com
  anime lithos deploy worker.ls wings
  anime lithos deploy inference.ls wings --domain llm.example.com --port 8080`,
	Args: cobra.ExactArgs(2),
	RunE: runLithosDeploy,
}

var lithosServeCmd = &cobra.Command{
	Use:   "serve [server]",
	Short: "Start or deploy a Lithos HTTP server (seed serve.ls response.lion)",
	Long: `Start a Lithos HTTP server locally or deploy to a remote server.

With no arguments, starts locally on port 8787 with a default welcome page.
With a server argument, deploys as a systemd service with optional nginx + SSL.

Use --response to serve your own HTML file instead of the default.

Examples:
  anime lithos serve                                  # Local server on :8787
  anime lithos serve --response mypage.html           # Serve your own page
  anime lithos serve --port 3000                      # Custom port
  anime lithos serve wings --domain api.example.com   # Deploy to remote
  anime lithos serve wings --response ~/app/index.html`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLithosServe,
}

var lithosLightningCmd = &cobra.Command{
	Use:   "lightning <model.lion> <server>",
	Short: "Deploy GPU inference via Lightning",
	Long: `Deploy Lithos Lightning inference to a GPU server.

Ships: lithos compiler, font atlas (.metallib), model config (.lion).
Sets up: systemd service + nginx + SSL.

Examples:
  anime lithos lightning llama.lion wings --domain llm.example.com`,
	Args: cobra.ExactArgs(2),
	RunE: runLithosLightning,
}

var lithosSymbolsCmd = &cobra.Command{
	Use:     "symbols",
	Aliases: []string{"prims", "primitives"},
	Short:   "List all 46 Lithos primitives",
	RunE:    runLithosSymbols,
}

var lithosBenchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run Lithos benchmarks",
	Long: `Run benchmarks comparing Lithos vs native C.

Examples:
  anime lithos bench
  anime lithos bench -s wings   # Benchmark on remote server`,
	RunE: runLithosBench,
}

func init() {
	lithosCompileCmd.Flags().StringVarP(&lithosCompileTarget, "target", "t", "", "Target: arm64, air, wasm, glsl (auto-detected if omitted)")

	lithosDeployCmd.Flags().StringVarP(&lithosDeployDomain, "domain", "d", "", "Domain for nginx + SSL")
	lithosDeployCmd.Flags().StringVarP(&lithosDeployPort, "port", "p", "8080", "Service port")
	lithosDeployCmd.Flags().BoolVar(&lithosDeployNoSSL, "no-ssl", false, "Skip SSL setup")

	lithosServeCmd.Flags().StringVarP(&lithosDeployDomain, "domain", "d", "", "Domain for nginx + SSL")
	lithosServeCmd.Flags().StringVarP(&lithosServePort, "port", "p", "9348", "Server port")
	lithosServeCmd.Flags().BoolVar(&lithosDeployNoSSL, "no-ssl", false, "Skip SSL setup")
	lithosServeCmd.Flags().StringVar(&lithosResponse, "response", "", "Custom response.lion file (default: lithos built-in)")

	lithosLightningCmd.Flags().StringVarP(&lithosDeployDomain, "domain", "d", "", "Domain for nginx + SSL")
	lithosLightningCmd.Flags().StringVarP(&lithosDeployPort, "port", "p", "8080", "Inference port")
	lithosLightningCmd.Flags().BoolVar(&lithosDeployNoSSL, "no-ssl", false, "Skip SSL setup")

	lithosBenchCmd.Flags().StringVarP(&lithosServer, "server", "s", "", "Remote server")

	lithosCmd.AddCommand(lithosCompileCmd)
	lithosCmd.AddCommand(lithosDeployCmd)
	lithosCmd.AddCommand(lithosServeCmd)
	lithosCmd.AddCommand(lithosLightningCmd)
	lithosCmd.AddCommand(lithosSymbolsCmd)
	lithosCmd.AddCommand(lithosBenchCmd)
	rootCmd.AddCommand(lithosCmd)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func lithosExists() bool {
	_, err := os.Stat(expandHome(lithosCompiler))
	return err == nil
}

func runLithosHelp(cmd *cobra.Command, args []string) {
	fmt.Println(theme.RenderBanner("⚡ LITHOS ⚡"))
	fmt.Println()

	if !lithosExists() {
		fmt.Println(theme.WarningStyle.Render("  Lithos not found at ~/lithos"))
		fmt.Println(theme.DimTextStyle.Render("  Clone it: git clone <lithos-repo> ~/lithos"))
		fmt.Println()
		return
	}

	fmt.Println(theme.InfoStyle.Render("  GPU computation language — glyph → silicon, no runtime"))
	fmt.Println()

	cmds := []struct{ c, d string }{
		{"anime lithos compile <file.ls>", "Compile a Lithos program"},
		{"anime lithos deploy <file.ls> <server>", "Compile + ship + systemd + nginx + SSL"},
		{"anime lithos serve <server>", "Deploy HTTP server (seed serve.ls response.lion)"},
		{"anime lithos lightning <model.lion> <server>", "Deploy GPU inference via Lightning"},
		{"anime lithos symbols", "List all 46 primitives"},
		{"anime lithos bench", "Run benchmarks"},
	}
	for _, c := range cmds {
		fmt.Printf("  %s\n    %s\n\n", theme.HighlightStyle.Render(c.c), theme.DimTextStyle.Render(c.d))
	}
}

func runLithosCompile(cmd *cobra.Command, args []string) error {
	if !lithosExists() {
		return fmt.Errorf("lithos compiler not found at %s", lithosCompiler)
	}

	source := args[0]
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source not found: %s", source)
	}

	// Determine output name
	output := strings.TrimSuffix(filepath.Base(source), ".ls")
	if len(args) > 1 {
		output = args[1]
	}

	// Auto-detect target
	target := lithosCompileTarget
	if target == "" {
		if runtime.GOARCH == "arm64" {
			target = "arm64"
		} else {
			target = "wasm"
		}
	}

	fmt.Println(theme.RenderBanner("⚡ LITHOS COMPILE ⚡"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.HighlightStyle.Render(source))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Target:"), theme.HighlightStyle.Render(target))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Output:"), theme.HighlightStyle.Render(output))
	fmt.Println()

	compiler := expandHome(lithosCompiler)
	compileCmd := exec.Command(compiler, source, output)
	compileCmd.Stdout = os.Stdout
	compileCmd.Stderr = os.Stderr
	compileCmd.Env = append(os.Environ(), "LITHOS_TARGET="+target)

	fmt.Printf("  %s Compiling...\n", theme.SymbolLoading)
	if err := compileCmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	info, _ := os.Stat(output)
	if info != nil {
		fmt.Printf("  %s %s (%d bytes)\n",
			theme.SuccessStyle.Render(theme.SymbolSuccess),
			output, info.Size())
	} else {
		fmt.Printf("  %s Compiled\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}
	fmt.Println()
	return nil
}

func runLithosDeploy(cmd *cobra.Command, args []string) error {
	if !lithosExists() {
		return fmt.Errorf("lithos compiler not found at %s", lithosCompiler)
	}

	source := args[0]
	server := args[1]

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source not found: %s", source)
	}

	if err := validate.Port(lithosDeployPort); err != nil {
		return err
	}
	if lithosDeployDomain != "" {
		if err := validate.Domain(lithosDeployDomain); err != nil {
			return err
		}
	}

	baseName := strings.TrimSuffix(filepath.Base(source), ".ls")
	binaryName := baseName + "-arm64"

	fmt.Println(theme.RenderBanner("⚡ LITHOS DEPLOY ⚡"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Source:"), theme.HighlightStyle.Render(source))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(server))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Port:"), theme.HighlightStyle.Render(lithosDeployPort))
	if lithosDeployDomain != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Domain:"), theme.HighlightStyle.Render(lithosDeployDomain))
	}
	fmt.Println()

	// Step 1: Cross-compile for ARM64 Linux
	fmt.Printf("  %s Compiling for ARM64...\n", theme.SymbolLoading)
	compiler := expandHome(lithosCompiler)
	compileCmd := exec.Command(compiler, source, binaryName)
	compileCmd.Env = append(os.Environ(), "LITHOS_TARGET=arm64")
	compileCmd.Stderr = os.Stderr
	if err := compileCmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	defer os.Remove(binaryName)
	fmt.Printf("  %s Compiled %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), binaryName)

	// Step 2: Ship binary to server
	fmt.Printf("  %s Shipping to %s...\n", theme.SymbolLoading, server)
	shipCmd := exec.Command(os.Args[0], "ship", binaryName, server)
	shipCmd.Stdout = os.Stdout
	shipCmd.Stderr = os.Stderr
	if err := shipCmd.Run(); err != nil {
		return fmt.Errorf("ship failed: %w", err)
	}

	// Step 3: Set up as systemd service
	fmt.Printf("  %s Setting up systemd service...\n", theme.SymbolLoading)
	serviceScript := fmt.Sprintf(`#!/bin/bash
set -e

# Make executable
chmod +x ~/%s

# Create systemd service
sudo tee /etc/systemd/system/%s.service > /dev/null << 'SVCEOF'
[Unit]
Description=Lithos %s
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER
ExecStart=/home/$USER/%s
Restart=always
RestartSec=3
Environment="PORT=%s"

[Install]
WantedBy=multi-user.target
SVCEOF

sudo systemctl daemon-reload
sudo systemctl enable --now %s
echo "Service %s started on port %s"
`, binaryName, baseName, baseName, binaryName, lithosDeployPort, baseName, baseName, lithosDeployPort)

	output, err := runOnServer(server, serviceScript)
	if err != nil {
		fmt.Printf("  %s Service setup had issues: %s\n", theme.WarningStyle.Render("⚠️"), strings.TrimSpace(output))
	} else {
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	}

	// Step 4: Nginx + SSL
	if lithosDeployDomain != "" {
		fmt.Printf("  %s Configuring nginx...\n", theme.SymbolLoading)
		manageServer = server

		nginxConfig := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`, lithosDeployDomain, lithosDeployPort)

		nginxScript := fmt.Sprintf(`cat > /tmp/nginx_%s << 'NGINXEOF'
%s
NGINXEOF
sudo mv /tmp/nginx_%s /etc/nginx/sites-available/%s
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx && echo "nginx configured"`,
			lithosDeployDomain, nginxConfig, lithosDeployDomain, lithosDeployDomain, lithosDeployDomain, lithosDeployDomain)

		output, err = runOnServer(server, nginxScript)
		if err != nil {
			fmt.Printf("  %s Nginx failed: %s\n", theme.ErrorStyle.Render(theme.SymbolError), strings.TrimSpace(output))
		} else {
			fmt.Printf("  %s Nginx configured\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}

		if !lithosDeployNoSSL {
			fmt.Printf("  %s Setting up SSL...\n", theme.SymbolLoading)
			sslScript := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, lithosDeployDomain)
			output, err = runOnServer(server, sslScript)
			if err != nil {
				fmt.Printf("  %s SSL failed: run 'sudo certbot --nginx -d %s' manually\n",
					theme.WarningStyle.Render("⚠️"), lithosDeployDomain)
			} else {
				fmt.Printf("  %s SSL installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
			}
		}
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	if lithosDeployDomain != "" {
		proto := "https"
		if lithosDeployNoSSL {
			proto = "http"
		}
		fmt.Printf("  %s %s://%s\n",
			theme.SuccessStyle.Render("✨ LITHOS DEPLOYED!"),
			proto, lithosDeployDomain)
	} else {
		fmt.Printf("  %s Running on port %s\n",
			theme.SuccessStyle.Render("✨ LITHOS DEPLOYED!"),
			lithosDeployPort)
	}
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	return nil
}

const lithosDefaultPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Lithos</title>
<style>
  @keyframes pulse { 0%,100%{opacity:0.4} 50%{opacity:1} }
  @keyframes slideIn { from{opacity:0;transform:translateY(20px)} to{opacity:1;transform:translateY(0)} }
  * { margin:0; padding:0; box-sizing:border-box }
  body { font:17px/1.7 'SF Pro Display',-apple-system,system-ui,sans-serif; background:#0a0e14; color:#c8ccd4; min-height:100vh; display:flex; flex-direction:column; align-items:center; justify-content:center; padding:40px 24px }
  .container { max-width:720px; width:100%; animation:slideIn 0.6s ease-out }
  .hero { text-align:center; margin-bottom:48px }
  .glyphs { font-size:4em; letter-spacing:0.3em; color:#f0883e; margin-bottom:16px; font-weight:200 }
  .glyphs span { display:inline-block; animation:pulse 3s ease-in-out infinite }
  .glyphs span:nth-child(2) { animation-delay:0.3s }
  .glyphs span:nth-child(3) { animation-delay:0.6s }
  h1 { font-size:2.4em; font-weight:700; background:linear-gradient(135deg,#58a6ff,#a371f7); -webkit-background-clip:text; -webkit-text-fill-color:transparent; margin-bottom:8px }
  .subtitle { color:#636e7b; font-size:1.05em }
  .card { background:#12161d; border:1px solid #1e2430; border-radius:12px; padding:28px; margin-bottom:20px; transition:border-color 0.2s }
  .card:hover { border-color:#2d333b }
  .card h2 { font-size:1.1em; color:#58a6ff; margin-bottom:12px; display:flex; align-items:center; gap:10px }
  .card h2 .icon { font-size:1.3em }
  .card p, .card li { color:#8b949e; font-size:0.95em }
  .card ul { list-style:none; padding:0 }
  .card li { padding:6px 0; border-bottom:1px solid #1a1f28 }
  .card li:last-child { border-bottom:none }
  .glyph-label { display:inline-flex; align-items:center; gap:8px }
  .glyph-label .g { color:#f0883e; font-size:1.4em; width:28px; text-align:center }
  .glyph-label .arrow { color:#2d333b }
  .glyph-label .desc { color:#c8ccd4 }
  pre { background:#0d1117; border:1px solid #1e2430; border-radius:8px; padding:16px; margin:12px 0; font:0.88em/1.6 'SF Mono',Menlo,monospace; color:#c8ccd4; overflow-x:auto }
  code { background:#161b22; padding:3px 7px; border-radius:5px; font-family:'SF Mono',Menlo,monospace; color:#db6d28; font-size:0.88em }
  .stats { display:grid; grid-template-columns:repeat(3,1fr); gap:16px; margin-top:8px }
  .stat { text-align:center; padding:12px; background:#0d1117; border-radius:8px; border:1px solid #1a1f28 }
  .stat .num { font-size:1.8em; font-weight:700; color:#3fb950 }
  .stat .label { font-size:0.8em; color:#636e7b; margin-top:2px }
  .footer { text-align:center; margin-top:40px; color:#3a3f47; font-size:0.85em }
  .footer code { background:none; color:#484f58 }
</style>
</head>
<body>
<div class="container">

  <div class="hero">
    <div class="glyphs"><span>→</span><span>←</span><span>↻</span></div>
    <h1>Lithos</h1>
    <p class="subtitle">Computation language. Glyph → silicon. No runtime.</p>
  </div>

  <div class="card">
    <h2><span class="icon">⚡</span> What this server is</h2>
    <ul>
      <li><div class="glyph-label"><span class="g">→</span><span class="arrow">—</span><span class="desc">Load: read bytes from the socket</span></div></li>
      <li><div class="glyph-label"><span class="g">←</span><span class="arrow">—</span><span class="desc">Store: write this page to the socket</span></div></li>
      <li><div class="glyph-label"><span class="g">↻</span><span class="arrow">—</span><span class="desc">Loop: accept the next connection</span></div></li>
    </ul>
    <p style="margin-top:12px">Three glyphs. The seed compiler JIT-compiles them + a server preamble into a raw ARM64 socket server. No libc. No runtime. No dependencies.</p>
  </div>

  <div class="card">
    <h2><span class="icon">🚀</span> Quick start</h2>
<pre><span style="color:#636e7b"># Serve locally</span>
anime lithos serve

<span style="color:#636e7b"># Serve your own HTML</span>
anime lithos serve --response mypage.html

<span style="color:#636e7b"># Deploy to a remote server with SSL</span>
anime lithos serve wings --domain api.example.com

<span style="color:#636e7b"># Compile a Lithos program</span>
anime lithos compile mykernel.ls

<span style="color:#636e7b"># Full deploy pipeline</span>
anime lithos deploy api.ls wings --domain api.example.com</pre>
  </div>

  <div class="card">
    <h2><span class="icon">📊</span> By the numbers</h2>
    <div class="stats">
      <div class="stat"><div class="num">46</div><div class="label">primitives</div></div>
      <div class="stat"><div class="num">12</div><div class="label">bytes of source</div></div>
      <div class="stat"><div class="num">&lt;5ms</div><div class="label">cold start</div></div>
    </div>
  </div>

  <div class="card">
    <h2><span class="icon">🎯</span> Compilation targets</h2>
    <ul>
      <li><code>arm64</code> — Apple Silicon, Linux ARM64</li>
      <li><code>air</code> — Metal GPU (M-series)</li>
      <li><code>wasm</code> — WebAssembly</li>
      <li><code>glsl</code> — Fragment shaders</li>
      <li><code>sass</code> — NVIDIA (planned)</li>
    </ul>
  </div>

  <p class="footer">Served by <code>anime lithos serve</code> — the entire server is <code>→ ← ↻</code></p>
</div>
</body>
</html>`

func runLithosServe(cmd *cobra.Command, args []string) error {
	if !lithosExists() {
		return fmt.Errorf("lithos not found at ~/lithos")
	}

	// Local mode if no server argument
	if len(args) == 0 {
		return runLithosServeLocal()
	}

	server := args[0]

	if err := validate.Port(lithosServePort); err != nil {
		return err
	}
	if lithosDeployDomain != "" {
		if err := validate.Domain(lithosDeployDomain); err != nil {
			return err
		}
	}

	fmt.Println(theme.RenderBanner("⚡ LITHOS SERVE ⚡"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(server))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Port:"), theme.HighlightStyle.Render(lithosServePort))
	fmt.Printf("  %s seed serve.ls response.lion → ARM64 socket server\n",
		theme.DimTextStyle.Render("Stack:"))
	fmt.Println()

	// Create a temp directory with the serve bundle
	tmpDir, err := os.MkdirTemp("", "lithos-serve-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	bundleDir := filepath.Join(tmpDir, "lithos-serve")
	os.MkdirAll(bundleDir, 0755)

	// Gather files to ship
	lithosDir := expandHome(lithosHome)
	filesToShip := map[string]string{
		filepath.Join(lithosDir, "cli/kernel/bootstrap/seed"):   "lithos-serve/seed",
		filepath.Join(lithosDir, "cli/kernel/serve/serve.ls"):   "lithos-serve/serve.ls",
		filepath.Join(lithosDir, "cli/kernel/serve/serve.lion"): "lithos-serve/serve.lion",
		filepath.Join(lithosDir, "arch/arm64/arm64-font.s"):     "lithos-serve/arm64-font.s",
	}

	// Use custom response, lithos built-in, or anime default
	if lithosResponse != "" {
		responsePath, err := prepareResponseLion(lithosResponse)
		if err != nil {
			return err
		}
		filesToShip[responsePath] = "lithos-serve/response.lion"
	} else {
		defaultResponse := filepath.Join(lithosDir, "cli/kernel/server/response.lion")
		if _, err := os.Stat(defaultResponse); err == nil {
			filesToShip[defaultResponse] = "lithos-serve/response.lion"
		} else {
			tmpResponse := filepath.Join(bundleDir, "response.lion")
			lionContent := "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nConnection: close\r\n\r\n" + lithosDefaultPage
			os.WriteFile(tmpResponse, []byte(lionContent), 0644)
			filesToShip[tmpResponse] = "lithos-serve/response.lion"
		}
	}

	// Also ship http.ls and programs if they exist
	for _, prog := range []string{"programs/http.ls", "programs/server.ls", "programs/ws.ls"} {
		p := filepath.Join(lithosDir, prog)
		if _, err := os.Stat(p); err == nil {
			filesToShip[p] = "lithos-serve/" + filepath.Base(p)
		}
	}

	fmt.Printf("  %s Bundling serve stack (%d files)...\n", theme.SymbolLoading, len(filesToShip))
	for src, dst := range filesToShip {
		if _, err := os.Stat(src); err != nil {
			fmt.Printf("  %s Skipping %s (not found)\n", theme.DimTextStyle.Render("  "), filepath.Base(src))
			continue
		}
		dstPath := filepath.Join(tmpDir, dst)
		os.MkdirAll(filepath.Dir(dstPath), 0755)
		cpCmd := exec.Command("cp", src, dstPath)
		if err := cpCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy %s: %w", filepath.Base(src), err)
		}
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), filepath.Base(src))
	}

	// Patch port in serve.lion
	serveLion := filepath.Join(bundleDir, "serve.lion")
	if _, err := os.Stat(serveLion); err == nil {
		patchCmd := exec.Command("bash", "-c",
			fmt.Sprintf(`sed -i.bak 's/^port .*/port %s/' "%s" && rm -f "%s.bak"`,
				lithosServePort, serveLion, serveLion))
		patchCmd.Run()
	}

	// Ship the bundle
	fmt.Printf("  %s Shipping to %s...\n", theme.SymbolLoading, server)
	shipCmd := exec.Command(os.Args[0], "ship", bundleDir, server)
	shipCmd.Stdout = os.Stdout
	shipCmd.Stderr = os.Stderr
	if err := shipCmd.Run(); err != nil {
		return fmt.Errorf("ship failed: %w", err)
	}

	// Set up systemd service on remote
	fmt.Printf("  %s Setting up lithos-serve service...\n", theme.SymbolLoading)
	serviceScript := fmt.Sprintf(`#!/bin/bash
set -e
cd ~/lithos-serve
chmod +x seed

# Create systemd service that runs: seed serve.ls response.lion
sudo tee /etc/systemd/system/lithos-serve.service > /dev/null << 'SVCEOF'
[Unit]
Description=Lithos HTTP Server (seed serve.ls)
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=/home/%s/lithos-serve
ExecStart=/home/%s/lithos-serve/seed serve.ls response.lion
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
SVCEOF

sudo systemctl daemon-reload
sudo systemctl enable --now lithos-serve
echo "Lithos serve started on port %s (→ ← ↻)"
`, "$USER", "$USER", "$USER", lithosServePort)

	output, err := runOnServer(server, serviceScript)
	if err != nil {
		fmt.Printf("  %s Service had issues: %s\n", theme.WarningStyle.Render("⚠️"), strings.TrimSpace(output))
	} else {
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	}

	// Nginx + SSL
	lithosSetupNginxSSL(server, lithosServePort)

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Printf("  %s serve.ls → ARM64 socket → port %s\n",
		theme.SuccessStyle.Render("✨ LITHOS SERVING!"), lithosServePort)
	if lithosDeployDomain != "" {
		proto := "https"
		if lithosDeployNoSSL {
			proto = "http"
		}
		fmt.Printf("  Endpoint: %s://%s\n", proto, lithosDeployDomain)
	}
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	return nil
}

func runLithosServeLocal() error {
	// Determine content to serve
	var content string
	if lithosResponse != "" {
		data, err := os.ReadFile(lithosResponse)
		if err != nil {
			return fmt.Errorf("response file not found: %s", lithosResponse)
		}
		content = string(data)
		// Strip HTTP headers if present (response.lion format)
		if strings.HasPrefix(content, "HTTP/") {
			if idx := strings.Index(content, "\n\n"); idx >= 0 {
				content = content[idx+2:]
			} else if idx := strings.Index(content, "\r\n\r\n"); idx >= 0 {
				content = content[idx+4:]
			}
		}
	} else {
		content = lithosDefaultPage
	}

	addr := ":" + lithosServePort
	url := fmt.Sprintf("http://localhost:%s", lithosServePort)

	fmt.Println(theme.RenderBanner("⚡ LITHOS SERVE ⚡"))
	fmt.Println()
	fmt.Printf("  %s %s\n",
		theme.SuccessStyle.Render("→"),
		theme.HighlightStyle.Render(url))
	if lithosResponse != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("  Serving:"), theme.HighlightStyle.Render(lithosResponse))
	}
	fmt.Printf("  %s Ctrl+C to stop\n\n", theme.DimTextStyle.Render("  "))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(content))
	})

	// Open browser after server starts
	go func() {
		time.Sleep(300 * time.Millisecond)
		exec.Command("open", url).Run()
	}()

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Printf("\n  %s Server stopped\n", theme.InfoStyle.Render("→"))
		os.Exit(0)
	}()

	return http.ListenAndServe(addr, nil)
}

// prepareResponseLion wraps an HTML file with HTTP headers if needed
func prepareResponseLion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("response file not found: %s", path)
	}

	content := string(data)
	// If it already has HTTP headers, use as-is
	if strings.HasPrefix(content, "HTTP/") {
		return path, nil
	}

	// Wrap raw HTML with HTTP response headers
	tmpFile, err := os.CreateTemp("", "lithos-response-*.lion")
	if err != nil {
		return "", err
	}

	wrapped := "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nConnection: close\r\n\r\n" + content
	tmpFile.WriteString(wrapped)
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func runLithosLightning(cmd *cobra.Command, args []string) error {
	if !lithosExists() {
		return fmt.Errorf("lithos not found at ~/lithos")
	}

	lionFile := args[0]
	server := args[1]

	if _, err := os.Stat(lionFile); err != nil {
		return fmt.Errorf("model config not found: %s", lionFile)
	}

	if err := validate.Port(lithosDeployPort); err != nil {
		return err
	}
	if lithosDeployDomain != "" {
		if err := validate.Domain(lithosDeployDomain); err != nil {
			return err
		}
	}

	fmt.Println(theme.RenderBanner("⚡ LITHOS LIGHTNING ⚡"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Model:"), theme.HighlightStyle.Render(lionFile))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Server:"), theme.HighlightStyle.Render(server))
	fmt.Printf("  %s GPU inference via Metal/AIR font atlas\n", theme.DimTextStyle.Render("Stack:"))
	fmt.Println()

	// Ship compiler + font atlas + lion config + kernel files
	lithosDir := expandHome(lithosHome)
	filesToShip := []string{
		expandHome(lithosCompiler),
		lionFile,
	}

	// Font atlas
	atlasPath := expandHome(lithosFontAtlas)
	if _, err := os.Stat(atlasPath); err == nil {
		filesToShip = append(filesToShip, atlasPath)
	}

	// ARM64 font table
	armFont := filepath.Join(lithosDir, "arch/arm64/arm64-font.s")
	if _, err := os.Stat(armFont); err == nil {
		filesToShip = append(filesToShip, armFont)
	}

	// Lightning kernel files
	lightningDir := filepath.Join(lithosDir, "cli/kernel/lightning")
	if entries, err := os.ReadDir(lightningDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				filesToShip = append(filesToShip, filepath.Join(lightningDir, e.Name()))
			}
		}
	}

	fmt.Printf("  %s Shipping lightning stack (%d files)...\n", theme.SymbolLoading, len(filesToShip))
	for _, f := range filesToShip {
		shipCmd := exec.Command(os.Args[0], "ship", f, server)
		shipCmd.Stderr = os.Stderr
		if err := shipCmd.Run(); err != nil {
			fmt.Printf("  %s Failed: %s\n", theme.WarningStyle.Render("⚠️"), filepath.Base(f))
			continue
		}
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), filepath.Base(f))
	}

	// Set up lightning inference service
	lionBase := filepath.Base(lionFile)
	fmt.Printf("  %s Setting up lightning service...\n", theme.SymbolLoading)

	serviceScript := fmt.Sprintf(`#!/bin/bash
set -e
chmod +x ~/lithos

sudo tee /etc/systemd/system/lithos-lightning.service > /dev/null << 'SVCEOF'
[Unit]
Description=Lithos Lightning Inference
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=/home/%s
ExecStart=/home/%s/lithos lightning infer %s
Restart=always
RestartSec=3
Environment="PORT=%s"

[Install]
WantedBy=multi-user.target
SVCEOF

sudo systemctl daemon-reload
sudo systemctl enable --now lithos-lightning
echo "Lightning inference started with %s on port %s"
`, "$USER", "$USER", "$USER", lionBase, lithosDeployPort, lionBase, lithosDeployPort)

	output, err := runOnServer(server, serviceScript)
	if err != nil {
		fmt.Printf("  %s Service had issues: %s\n", theme.WarningStyle.Render("⚠️"), strings.TrimSpace(output))
	} else {
		fmt.Printf("  %s %s\n", theme.SuccessStyle.Render(theme.SymbolSuccess), strings.TrimSpace(output))
	}

	// Nginx + SSL
	lithosSetupNginxSSL(server, lithosDeployPort)

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println(theme.SuccessStyle.Render("  ✨ LITHOS LIGHTNING DEPLOYED!"))
	if lithosDeployDomain != "" {
		proto := "https"
		if lithosDeployNoSSL {
			proto = "http"
		}
		fmt.Printf("  Endpoint: %s://%s\n", proto, lithosDeployDomain)
	}
	fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()
	return nil
}

// lithosSetupNginxSSL configures nginx + SSL for lithos services
func lithosSetupNginxSSL(server, port string) {
	if lithosDeployDomain == "" {
		return
	}

	fmt.Printf("  %s Configuring nginx for %s...\n", theme.SymbolLoading, lithosDeployDomain)

	nginxConfig := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`, lithosDeployDomain, port)

	nginxScript := fmt.Sprintf(`cat > /tmp/nginx_%s << 'NGINXEOF'
%s
NGINXEOF
sudo mv /tmp/nginx_%s /etc/nginx/sites-available/%s
sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s
sudo nginx -t 2>&1 && sudo systemctl reload nginx && echo "nginx configured"`,
		lithosDeployDomain, nginxConfig, lithosDeployDomain, lithosDeployDomain, lithosDeployDomain, lithosDeployDomain)

	output, err := runOnServer(server, nginxScript)
	if err != nil {
		fmt.Printf("  %s Nginx failed: %s\n", theme.ErrorStyle.Render(theme.SymbolError), strings.TrimSpace(output))
	} else {
		fmt.Printf("  %s Nginx configured\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
	}

	if !lithosDeployNoSSL {
		fmt.Printf("  %s SSL...\n", theme.SymbolLoading)
		sslScript := fmt.Sprintf(`sudo certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1`, lithosDeployDomain)
		output, err = runOnServer(server, sslScript)
		if err != nil {
			fmt.Printf("  %s SSL failed: run 'sudo certbot --nginx -d %s' manually\n",
				theme.WarningStyle.Render("⚠️"), lithosDeployDomain)
		} else {
			fmt.Printf("  %s SSL installed\n", theme.SuccessStyle.Render(theme.SymbolSuccess))
		}
	}
}

func runLithosSymbols(cmd *cobra.Command, args []string) error {
	// Try running lithos symbols if available
	if lithosExists() {
		lithosSymCmd := exec.Command(expandHome("~/lithos/cli/lithos"), "symbols")
		lithosSymCmd.Stdout = os.Stdout
		lithosSymCmd.Stderr = os.Stderr
		if err := lithosSymCmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback: inline primitive table
	fmt.Println(theme.RenderBanner("⚡ LITHOS PRIMITIVES ⚡"))
	fmt.Println()

	categories := []struct {
		name  string
		prims []string
	}{
		{"Arithmetic", []string{
			"*  +  -  /        (scalar)",
			"** ++ -- //       (vector)",
			"*** +++ --- ///   (matrix)",
			"**** ++++ ---- //// (tensor)",
		}},
		{"Reductions", []string{
			"Σ  sum",
			"△  max",
			"▽  min",
			"#  index prefix",
		}},
		{"Rank Ops", []string{
			"·  inner product",
			"⊗  outer product",
			"×  cross product (3D)",
			"|  shift",
			"⊚  scale",
			"⟲  conjugate",
		}},
		{"Scalar Math", []string{
			"√  sqrt       ⅟  reciprocal",
			"log₂          ln  natural log",
			"∿  sine       ∾  cosine",
			"^  power      <  floor     >  ceil",
		}},
		{"Memory/Control", []string{
			"→  load        ←  store",
			"↑  read reg    ↓  write reg",
			"?  conditional",
			"↻  loop",
		}},
		{"Constants", []string{
			"e  Euler's     π  pi     i  imaginary",
		}},
		{"Named Compositions", []string{
			"σ    ⇌  sigmoid (reciprocal of decay)",
			"η    ⇌  RMSNorm (normalize + scale)",
			"ς    ⇌  softplus",
			"silu ⇌  amplify by survival rate (⊛ ⅟λ)",
			"⊛    ⇌  amplify (element-wise scale)",
		}},
	}

	for _, cat := range categories {
		fmt.Printf("  %s\n", theme.InfoStyle.Render(cat.name))
		for _, p := range cat.prims {
			fmt.Printf("    %s\n", theme.DimTextStyle.Render(p))
		}
		fmt.Println()
	}

	fmt.Printf("  %s 46 primitives, 8 categories, one construct: name ⇌ body\n",
		theme.DimTextStyle.Render("Total:"))
	fmt.Println()
	return nil
}

func runLithosBench(cmd *cobra.Command, args []string) error {
	fmt.Println(theme.RenderBanner("⚡ LITHOS BENCH ⚡"))
	fmt.Println()

	if lithosServer != "" {
		// Remote benchmark
		fmt.Printf("  %s Running benchmarks on %s...\n\n", theme.SymbolLoading, lithosServer)
		script := `cd ~/lithos 2>/dev/null && make bench 2>&1 || echo "Lithos not found at ~/lithos"`
		output, err := runOnServer(lithosServer, script)
		if err != nil && output == "" {
			return fmt.Errorf("benchmark failed: %w", err)
		}
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			fmt.Printf("  %s\n", theme.DimTextStyle.Render(line))
		}
		fmt.Println()
		return nil
	}

	// Local benchmark
	lithosDir := expandHome(lithosHome)
	if _, err := os.Stat(lithosDir); err != nil {
		return fmt.Errorf("lithos not found at ~/lithos")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("  %s Lithos vs GCC -O2 (ARM64 primitives)\n", theme.HighlightStyle.Render("1"))
	fmt.Printf("  %s Full benchmark suite\n", theme.HighlightStyle.Render("2"))
	fmt.Println()
	fmt.Print("  Choice [1]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	target := "test"
	if choice == "2" {
		target = "bench"
	}

	fmt.Printf("\n  %s Running...\n\n", theme.SymbolLoading)
	benchCmd := exec.Command("make", target)
	benchCmd.Dir = lithosDir
	benchCmd.Stdout = os.Stdout
	benchCmd.Stderr = os.Stderr
	return benchCmd.Run()
}
