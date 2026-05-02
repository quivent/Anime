package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// wan studio — boot Comfort (the Wan T2V Atelier) and proxy /api+/ws to ComfyUI.
//
// Layout:
//
//	GET /                   → static (Comfort dist/)
//	GET /api/*              → reverse proxy → http://127.0.0.1:8188/*
//	GET /ws                 → reverse proxy (WS upgrade) → ws://127.0.0.1:8188/ws
//	GET anything else       → static fallback to index.html (SPA routing)
//
// Why fold both into one origin: comfort-ui issues all calls as same-origin
// (`/api/...`, `/ws`) per its CONTRACT.md, so we host them under one server
// instead of teaching it CORS or per-deploy URLs.

func init() {
	wanCmd.AddCommand(&cobra.Command{
		Use:   "studio",
		Short: "Open the Comfort Wan T2V Atelier (web UI) and proxy to ComfyUI",
		Long: `Launch the Comfort Wan studio. By default this is the single
command that gets a fresh CUDA box to a running web studio: it auto-installs
ComfyUI, picks the right PyTorch wheel for the host's driver (cu118 → cu130),
adds sageattention, the Kijai custom-node stack, the Wan 2.2 model set
(~85GB), and the Comfort UI; starts ComfyUI; then serves the SPA and proxies
/api + /ws to it. Works on any CUDA GPU with ≥16GB VRAM (5B preset) or
≥24GB (14B preset). Tested on RTX 4090, A100, L40S, H100, GH200/H200.

Flags:
  --port <int>      Bind port (default: pick a free one starting at 5180)
  --comfy <url>     ComfyUI upstream (default: http://127.0.0.1:8188)
  --dist <path>     Override Comfort dist directory (default: auto-detect)
  --no-open         Do not auto-open the browser
  --dev             Run "npm run dev" instead of serving dist (live reload)
  --public          Bind 0.0.0.0 (reachable from outside the host)
  --check           Print environment status and exit (no install, no serve)
  --skip-install    Don't run any installer phase; fail if anything is missing
  --skip-models     Skip the Wan 2.2 model download (~85GB)
  --yes             Don't prompt for the heavy download phase

Examples:
  anime wan studio              # zero to studio (prompts before the 85GB pull)
  anime wan studio --yes        # unattended bootstrap
  anime wan studio --check      # show what's installed / missing
  anime wan studio --dev        # live-reload from the comfort-ui source

The studio talks to whatever ComfyUI is reachable at --comfy. The default
local flow installs to: ~/ComfyUI, ~/Comfort, ~/.anime/wan-pipeline.db.`,
		DisableFlagParsing: true, // we hand-parse so unknown flags can stream through to the studio
		RunE:               runWanStudio,
	})
}

func runWanStudio(cmd *cobra.Command, args []string) error {
	// Lightweight, ad-hoc flag parsing. The cobra cmd has DisableFlagParsing
	// off (default), but we read os.Args after the cobra.Command's own consumed
	// flags. Simpler: just walk args.
	port := 0
	comfyURL := "http://127.0.0.1:8188"
	distOverride := ""
	open := true
	dev := false
	public := false
	bootstrap := setupOpts{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		take := func() (string, bool) {
			if i+1 >= len(args) {
				return "", false
			}
			i++
			return args[i], true
		}
		switch {
		case a == "--port":
			if v, ok := take(); ok {
				if _, err := fmt.Sscanf(v, "%d", &port); err != nil {
					return fmt.Errorf("invalid --port: %s", v)
				}
			}
		case strings.HasPrefix(a, "--port="):
			if _, err := fmt.Sscanf(strings.TrimPrefix(a, "--port="), "%d", &port); err != nil {
				return fmt.Errorf("invalid --port: %s", a)
			}
		case a == "--comfy":
			if v, ok := take(); ok {
				comfyURL = v
			}
		case strings.HasPrefix(a, "--comfy="):
			comfyURL = strings.TrimPrefix(a, "--comfy=")
		case a == "--dist":
			if v, ok := take(); ok {
				distOverride = v
			}
		case strings.HasPrefix(a, "--dist="):
			distOverride = strings.TrimPrefix(a, "--dist=")
		case a == "--no-open":
			open = false
		case a == "--dev":
			dev = true
		case a == "--public":
			public = true
		case a == "--check", a == "--check-only":
			bootstrap.checkOnly = true
		case a == "--skip-install":
			bootstrap.skipInstall = true
		case a == "--skip-models":
			bootstrap.skipModels = true
		case a == "--yes", a == "-y":
			bootstrap.yes = true
		case a == "-h", a == "--help":
			return cmd.Help()
		default:
			return fmt.Errorf("unknown flag: %s", a)
		}
	}

	// Phase 0: get the host to a state where the studio can actually run.
	// This is idempotent — every phase short-circuits when its check passes.
	// On a fully-set-up box (the common case), this is just a fast probe.
	if err := ensureComfyStudioReady(&bootstrap); err != nil {
		return err
	}
	if bootstrap.checkOnly {
		return nil
	}

	if dev {
		return runStudioDev(distOverride, open, comfyURL)
	}

	dist, err := resolveComfortDist(distOverride)
	if err != nil {
		return err
	}

	upstream, err := url.Parse(comfyURL)
	if err != nil {
		return fmt.Errorf("invalid --comfy URL %q: %w", comfyURL, err)
	}
	if !checkComfyReachable(upstream) {
		fmt.Println(theme.WarningStyle.Render("⚠  ComfyUI not reachable at " + upstream.String()))
		fmt.Println(theme.DimTextStyle.Render("   The studio will load, but renders will fail until ComfyUI starts."))
		fmt.Println(theme.DimTextStyle.Render("   Start it: anime comfyui start"))
		fmt.Println()
	}

	bindHost := "127.0.0.1"
	if public {
		bindHost = "0.0.0.0"
		// Studio + ComfyUI have NO authentication. Binding 0.0.0.0 on a
		// cloud GPU box is a real security risk — anyone who finds the
		// host:port can submit prompts, upload files, and read history.
		// Print a loud warning so the user can't miss it.
		fmt.Println(theme.WarningStyle.Render("⚠  --public binds 0.0.0.0 with NO AUTH"))
		fmt.Println(theme.DimTextStyle.Render("   Anyone who reaches host:port can submit/cancel renders."))
		fmt.Println(theme.DimTextStyle.Render("   Prefer: ssh -L 5180:127.0.0.1:5180 <host>  and drop --public."))
		fmt.Println()
		// Auto-open is meaningless on a headless cloud box (the browser
		// would launch on the H100, not the user's laptop). Disable it so
		// the goroutine doesn't try to xdg-open into the void.
		open = false
	}
	if port == 0 {
		port = pickFreePort(5180, bindHost)
	}
	addr := fmt.Sprintf("%s:%d", bindHost, port)

	srv, err := buildStudioServer(addr, dist, upstream)
	if err != nil {
		return err
	}

	localURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	openURL := localURL
	fmt.Println(theme.GlowStyle.Render("🎬 Wan T2V Atelier"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "local")), theme.SuccessStyle.Render(localURL))
	if public {
		ip := getPublicIPForComfyUI()
		if ip != "" && ip != "127.0.0.1" {
			pubURL := fmt.Sprintf("http://%s:%d", ip, port)
			fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "public")), theme.SuccessStyle.Render(pubURL))
		}
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "binding")),
			theme.WarningStyle.Render("0.0.0.0 — reachable from outside the host"))
	}
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "comfyui")), theme.PrimaryTextStyle.Render(upstream.String()))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "dist")), theme.DimTextStyle.Render(dist))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  Ctrl+C to stop the server."))
	fmt.Println()

	if open {
		go openBrowser(openURL) // package-shared helper from start.go
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-stop:
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("Shutting down studio..."))
		_ = srv.Close()
	}
	return nil
}

// runStudioDev starts `npm run dev` in the comfort-ui directory. Vite handles
// HMR + its own proxy (developer is responsible for the vite.config). We don't
// proxy in dev mode — vite does.
func runStudioDev(distOverride string, openBrowserFlag bool, comfyURL string) error {
	uiDir, err := resolveComfortUIDir(distOverride)
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm not on PATH (run: anime install nodejs)")
	}
	fmt.Println(theme.GlowStyle.Render("🎬 Wan T2V Atelier — dev mode"))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "ui dir")), theme.DimTextStyle.Render(uiDir))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-12s", "comfyui")), theme.PrimaryTextStyle.Render(comfyURL))
	fmt.Println(theme.DimTextStyle.Render("  Ctrl+C to stop."))
	fmt.Println()

	c := exec.Command("npm", "run", "dev")
	c.Dir = uiDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = append(os.Environ(), "COMFY_API="+comfyURL)
	return c.Run()
}

// buildStudioServer wires up: static SPA + /api proxy + /ws proxy.
func buildStudioServer(addr, dist string, upstream *url.URL) (*http.Server, error) {
	mux := http.NewServeMux()

	apiProxy := httputil.NewSingleHostReverseProxy(upstream)
	apiProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "ComfyUI unreachable: "+err.Error(), http.StatusBadGateway)
	}
	// Forward /api/* and /ws (and bare /view, which ComfyUI serves at root).
	for _, prefix := range []string{"/api/", "/ws", "/view", "/prompt", "/queue", "/history", "/upload/", "/object_info", "/system_stats", "/interrupt"} {
		mux.Handle(prefix, apiProxy)
	}

	// Static SPA: serve files under dist/, fall back to index.html for unknown
	// paths so client-side routing works.
	staticFS := http.Dir(dist)
	indexPath := filepath.Join(dist, "index.html")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		full := filepath.Join(dist, path)
		if path == "" || isDir(full) || !exists(full) {
			http.ServeFile(w, r, indexPath)
			return
		}
		http.FileServer(staticFS).ServeHTTP(w, r)
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}, nil
}

// resolveComfortDist picks the dist directory using (in order):
//
//	override > $COMFORT_DIST > ~/.anime/comfort-path's dist > ~/Comfort/comfort-ui/dist
func resolveComfortDist(override string) (string, error) {
	candidates := []string{}
	if override != "" {
		candidates = append(candidates, override)
	}
	if env := os.Getenv("COMFORT_DIST"); env != "" {
		candidates = append(candidates, env)
	}
	home, _ := os.UserHomeDir()
	if data, err := os.ReadFile(filepath.Join(home, ".anime", "comfort-path")); err == nil {
		ui := strings.TrimSpace(string(data))
		if ui != "" {
			candidates = append(candidates, filepath.Join(ui, "dist"))
		}
	}
	candidates = append(candidates,
		filepath.Join(home, "Comfort", "comfort-ui", "dist"),
		filepath.Join(home, "comfort", "comfort-ui", "dist"),
	)

	for _, c := range candidates {
		if isDir(c) && exists(filepath.Join(c, "index.html")) {
			return c, nil
		}
	}
	return "", fmt.Errorf("Comfort dist not found in any of: %s\n\n  Run: anime install comfort\n  Or:  anime wan studio --dist /path/to/dist",
		strings.Join(candidates, ", "))
}

// resolveComfortUIDir finds the comfort-ui dir (parent of dist).
func resolveComfortUIDir(distOverride string) (string, error) {
	if distOverride != "" {
		return filepath.Dir(distOverride), nil
	}
	dist, err := resolveComfortDist("")
	if err != nil {
		// dev mode tolerates a missing dist if the source tree exists
		home, _ := os.UserHomeDir()
		for _, c := range []string{
			filepath.Join(home, "Comfort", "comfort-ui"),
			filepath.Join(home, "comfort", "comfort-ui"),
		} {
			if isDir(c) && exists(filepath.Join(c, "package.json")) {
				return c, nil
			}
		}
		return "", err
	}
	return filepath.Dir(dist), nil
}

// pickFreePort returns the first available port at or after start, probing
// on the same host the server will actually bind to. A port can be free on
// 127.0.0.1 yet busy on 0.0.0.0 (e.g., another service bound to a specific
// interface), so the probe must match the bind.
func pickFreePort(start int, host string) int {
	for p := start; p < start+200; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, p))
		if err == nil {
			_ = ln.Close()
			return p
		}
	}
	return start // best-effort; ListenAndServe will surface the real error
}

func checkComfyReachable(u *url.URL) bool {
	cli := &http.Client{Timeout: 2 * time.Second}
	resp, err := cli.Get(u.String() + "/system_stats")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
