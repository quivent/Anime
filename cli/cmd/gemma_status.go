package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/ssh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	gemmaStatusRemote bool
	gemmaStatusServer string
)

var gemmaStatusCmd = &cobra.Command{
	Use:   "gemma-status",
	Short: "Show Gemma inference server status",
	Long: `Check status of local and remote inference engines serving Gemma models.

Checks:
  - Ollama: running, loaded models, version
  - vLLM: running, served model, GPU utilization
  - llama.cpp (llama-server): running, model, context size
  - SGLang: running, served model
  - MLX: installed version, cached models

Examples:
  anime gemma-status                # Check local
  anime gemma-status -r -s mybox   # Check remote server`,
	RunE: runGemmaStatus,
}

func init() {
	gemmaStatusCmd.Flags().BoolVarP(&gemmaStatusRemote, "remote", "r", false, "Check remote server")
	gemmaStatusCmd.Flags().StringVarP(&gemmaStatusServer, "server", "s", "", "Server name")
	rootCmd.AddCommand(gemmaStatusCmd)
}

func runGemmaStatus(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🔍 INFERENCE STATUS"))
	fmt.Println()

	if gemmaStatusRemote {
		return runGemmaStatusRemote()
	}
	return runGemmaStatusLocal()
}

// ── local status ────────────────────────────────────────────────────

func runGemmaStatusLocal() error {
	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render("Checking local inference engines..."))

	checkLocalOllama()
	checkLocalVLLM()
	checkLocalLlamaCpp()
	checkLocalSGLang()
	checkLocalMLX()

	return nil
}

func checkLocalOllama() {
	fmt.Printf("  %s\n", theme.InfoStyle.Render("Ollama"))

	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var tags struct {
		Models []struct {
			Name       string `json:"name"`
			Size       int64  `json:"size"`
			ModifiedAt string `json:"modified_at"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &tags); err != nil {
		fmt.Printf("    %s Running but bad response\n\n", theme.WarningStyle.Render("●"))
		return
	}

	// Get version
	version := ""
	if vResp, err := http.Get("http://localhost:11434/api/version"); err == nil {
		vData, _ := io.ReadAll(vResp.Body)
		vResp.Body.Close()
		var v struct{ Version string }
		if json.Unmarshal(vData, &v) == nil {
			version = v.Version
		}
	}

	fmt.Printf("    %s Running", theme.SuccessStyle.Render("●"))
	if version != "" {
		fmt.Printf(" (v%s)", version)
	}
	fmt.Println()

	// Check loaded models (ps)
	if psResp, err := http.Get("http://localhost:11434/api/ps"); err == nil {
		psData, _ := io.ReadAll(psResp.Body)
		psResp.Body.Close()
		var ps struct {
			Models []struct {
				Name      string `json:"name"`
				Size      int64  `json:"size"`
				SizeVRAM  int64  `json:"size_vram"`
				ExpiresAt string `json:"expires_at"`
			} `json:"models"`
		}
		if json.Unmarshal(psData, &ps) == nil && len(ps.Models) > 0 {
			fmt.Printf("    Loaded: ")
			for i, m := range ps.Models {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(theme.HighlightStyle.Render(m.Name))
				if m.SizeVRAM > 0 {
					fmt.Printf(" (%s VRAM)", formatBytes(m.SizeVRAM))
				}
			}
			fmt.Println()
		} else {
			fmt.Printf("    Loaded: %s\n", theme.DimTextStyle.Render("none"))
		}
	}

	// List gemma models
	gemmaModels := []string{}
	for _, m := range tags.Models {
		if strings.Contains(strings.ToLower(m.Name), "gemma") {
			gemmaModels = append(gemmaModels, m.Name)
		}
	}
	if len(gemmaModels) > 0 {
		fmt.Printf("    Gemma models: %s\n", theme.HighlightStyle.Render(strings.Join(gemmaModels, ", ")))
	}
	fmt.Printf("    Total models: %d\n\n", len(tags.Models))
}

func checkLocalVLLM() {
	fmt.Printf("  %s\n", theme.InfoStyle.Render("vLLM"))

	resp, err := http.Get("http://localhost:8000/v1/models")
	if err != nil {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &models); err != nil || len(models.Data) == 0 {
		fmt.Printf("    %s Running but no models loaded\n\n", theme.WarningStyle.Render("●"))
		return
	}

	fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
	for _, m := range models.Data {
		fmt.Printf("    Serving: %s\n", theme.HighlightStyle.Render(m.ID))
	}
	fmt.Println()
}

func checkLocalLlamaCpp() {
	fmt.Printf("  %s\n", theme.InfoStyle.Render("llama.cpp (llama-server)"))

	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var health struct {
		Status string `json:"status"`
	}
	json.Unmarshal(data, &health)

	if health.Status == "ok" {
		fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
	} else {
		fmt.Printf("    %s Running (status: %s)\n", theme.WarningStyle.Render("●"), health.Status)
	}

	// Try to get model info from /props
	if propsResp, err := http.Get("http://localhost:8080/props"); err == nil {
		propsData, _ := io.ReadAll(propsResp.Body)
		propsResp.Body.Close()
		var props struct {
			DefaultGenSettings struct {
				Model string `json:"model"`
			} `json:"default_generation_settings"`
		}
		if json.Unmarshal(propsData, &props) == nil && props.DefaultGenSettings.Model != "" {
			fmt.Printf("    Model: %s\n", theme.HighlightStyle.Render(props.DefaultGenSettings.Model))
		}
	}
	fmt.Println()
}

func checkLocalSGLang() {
	fmt.Printf("  %s\n", theme.InfoStyle.Render("SGLang"))

	resp, err := http.Get("http://localhost:30000/v1/models")
	if err != nil {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &models); err != nil || len(models.Data) == 0 {
		fmt.Printf("    %s Running but no models\n\n", theme.WarningStyle.Render("●"))
		return
	}

	fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
	for _, m := range models.Data {
		fmt.Printf("    Serving: %s\n", theme.HighlightStyle.Render(m.ID))
	}
	fmt.Println()
}

func checkLocalMLX() {
	fmt.Printf("  %s\n", theme.InfoStyle.Render("MLX"))

	out, err := exec.Command("python3", "-c", "import mlx_lm; print(mlx_lm.__version__)").Output()
	if err != nil {
		fmt.Printf("    %s Not installed\n\n", theme.DimTextStyle.Render("●"))
		return
	}

	version := strings.TrimSpace(string(out))
	fmt.Printf("    %s Installed (v%s)\n", theme.SuccessStyle.Render("●"), version)

	// Check MLX version
	mlxOut, _ := exec.Command("python3", "-c", "import mlx; print(mlx.__version__)").Output()
	if len(mlxOut) > 0 {
		fmt.Printf("    MLX core: v%s\n", strings.TrimSpace(string(mlxOut)))
	}

	// List cached gemma models
	home, _ := os.UserHomeDir()
	cacheDir := home + "/.cache/huggingface/hub"
	entries, _ := os.ReadDir(cacheDir)
	var gemmaModels []string
	for _, e := range entries {
		name := e.Name()
		if strings.Contains(strings.ToLower(name), "gemma") && strings.HasPrefix(name, "models--") {
			// models--mlx-community--gemma-3-4b-it-4bit -> mlx-community/gemma-3-4b-it-4bit
			clean := strings.TrimPrefix(name, "models--")
			clean = strings.Replace(clean, "--", "/", 1)
			gemmaModels = append(gemmaModels, clean)
		}
	}
	if len(gemmaModels) > 0 {
		fmt.Printf("    Cached Gemma models:\n")
		for _, m := range gemmaModels {
			fmt.Printf("      %s\n", theme.HighlightStyle.Render(m))
		}
	}
	fmt.Println()
}

// ── remote status ───────────────────────────────────────────────────

func runGemmaStatusRemote() error {
	serverName := gemmaStatusServer
	if serverName == "" {
		cfg, err := config.Load()
		if err != nil || len(cfg.Servers) == 0 {
			return fmt.Errorf("no server specified")
		}
		serverName = cfg.Servers[0].Name
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	server, err := cfg.GetServer(serverName)
	if err != nil {
		target := cfg.GetAlias(serverName)
		if target == "" {
			return fmt.Errorf("server %q not found", serverName)
		}
		parts := strings.SplitN(target, "@", 2)
		if len(parts) == 2 {
			server = &config.Server{Name: serverName, User: parts[0], Host: parts[1]}
		} else {
			server = &config.Server{Name: serverName, User: "ubuntu", Host: target}
		}
	}

	fmt.Printf("  %s\n\n", theme.DimTextStyle.Render(fmt.Sprintf("Checking %s@%s...", server.User, server.Host)))

	client, err := ssh.NewClient(server.Host, server.User, server.SSHKey)
	if err != nil {
		return fmt.Errorf("SSH connect: %w", err)
	}
	defer client.Close()

	script := `#!/bin/bash
echo "=== OLLAMA ==="
if curl -s --max-time 2 http://localhost:11434/api/tags 2>/dev/null; then
    echo ""
    echo "OLLAMA_PS:"
    curl -s --max-time 2 http://localhost:11434/api/ps 2>/dev/null
    echo ""
    echo "OLLAMA_VERSION:"
    curl -s --max-time 2 http://localhost:11434/api/version 2>/dev/null
else
    echo "NOT_RUNNING"
fi
echo ""
echo "=== VLLM ==="
if curl -s --max-time 2 http://localhost:8000/v1/models 2>/dev/null; then
    echo ""
else
    echo "NOT_RUNNING"
fi
echo ""
echo "=== LLAMACPP ==="
if curl -s --max-time 2 http://localhost:8080/health 2>/dev/null; then
    echo ""
    echo "LLAMACPP_PROPS:"
    curl -s --max-time 2 http://localhost:8080/props 2>/dev/null
    echo ""
else
    echo "NOT_RUNNING"
fi
echo ""
echo "=== SGLANG ==="
if curl -s --max-time 2 http://localhost:30000/v1/models 2>/dev/null; then
    echo ""
else
    echo "NOT_RUNNING"
fi
echo ""
echo "=== GPU ==="
nvidia-smi --query-gpu=name,memory.used,memory.total,utilization.gpu,temperature.gpu --format=csv,noheader 2>/dev/null || echo "NO_GPU"
`

	output, err := client.RunCommand(script)
	if err != nil {
		return fmt.Errorf("status check failed: %w", err)
	}

	parseRemoteStatus(output)
	return nil
}

func parseRemoteStatus(output string) {
	sections := map[string]string{}
	current := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "=== ") && strings.HasSuffix(line, " ===") {
			current = strings.Trim(line, "= ")
			sections[current] = ""
		} else if current != "" {
			sections[current] += line + "\n"
		}
	}

	// Ollama
	fmt.Printf("  %s\n", theme.InfoStyle.Render("Ollama"))
	ollamaData := strings.TrimSpace(sections["OLLAMA"])
	if ollamaData == "NOT_RUNNING" || ollamaData == "" {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
	} else {
		fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
		// Parse tags
		parts := strings.SplitN(ollamaData, "\nOLLAMA_PS:\n", 2)
		if len(parts) >= 1 {
			var tags struct {
				Models []struct{ Name string } `json:"models"`
			}
			if json.Unmarshal([]byte(strings.TrimSpace(parts[0])), &tags) == nil {
				gemma := []string{}
				for _, m := range tags.Models {
					if strings.Contains(strings.ToLower(m.Name), "gemma") {
						gemma = append(gemma, m.Name)
					}
				}
				if len(gemma) > 0 {
					fmt.Printf("    Gemma: %s\n", theme.HighlightStyle.Render(strings.Join(gemma, ", ")))
				}
				fmt.Printf("    Total models: %d\n", len(tags.Models))
			}
		}
		fmt.Println()
	}

	// vLLM
	fmt.Printf("  %s\n", theme.InfoStyle.Render("vLLM"))
	vllmData := strings.TrimSpace(sections["VLLM"])
	if vllmData == "NOT_RUNNING" || vllmData == "" {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
	} else {
		fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
		var models struct {
			Data []struct{ ID string } `json:"data"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(vllmData)), &models) == nil {
			for _, m := range models.Data {
				fmt.Printf("    Serving: %s\n", theme.HighlightStyle.Render(m.ID))
			}
		}
		fmt.Println()
	}

	// llama.cpp
	fmt.Printf("  %s\n", theme.InfoStyle.Render("llama.cpp"))
	llamaData := strings.TrimSpace(sections["LLAMACPP"])
	if llamaData == "NOT_RUNNING" || llamaData == "" {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
	} else {
		fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
		if propIdx := strings.Index(llamaData, "LLAMACPP_PROPS:"); propIdx >= 0 {
			propsJSON := strings.TrimSpace(llamaData[propIdx+len("LLAMACPP_PROPS:"):])
			var props struct {
				DefaultGenSettings struct {
					Model string `json:"model"`
					NCtx  int    `json:"n_ctx"`
				} `json:"default_generation_settings"`
			}
			if json.Unmarshal([]byte(propsJSON), &props) == nil {
				if props.DefaultGenSettings.Model != "" {
					fmt.Printf("    Model: %s\n", theme.HighlightStyle.Render(props.DefaultGenSettings.Model))
				}
				if props.DefaultGenSettings.NCtx > 0 {
					fmt.Printf("    Context: %d\n", props.DefaultGenSettings.NCtx)
				}
			}
		}
		fmt.Println()
	}

	// SGLang
	fmt.Printf("  %s\n", theme.InfoStyle.Render("SGLang"))
	sgData := strings.TrimSpace(sections["SGLANG"])
	if sgData == "NOT_RUNNING" || sgData == "" {
		fmt.Printf("    %s Not running\n\n", theme.DimTextStyle.Render("●"))
	} else {
		fmt.Printf("    %s Running\n", theme.SuccessStyle.Render("●"))
		var models struct {
			Data []struct{ ID string } `json:"data"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(sgData)), &models) == nil {
			for _, m := range models.Data {
				fmt.Printf("    Serving: %s\n", theme.HighlightStyle.Render(m.ID))
			}
		}
		fmt.Println()
	}

	// GPU
	gpuData := strings.TrimSpace(sections["GPU"])
	if gpuData != "" && gpuData != "NO_GPU" {
		fmt.Printf("  %s\n", theme.InfoStyle.Render("GPU"))
		for _, line := range strings.Split(gpuData, "\n") {
			parts := strings.Split(line, ", ")
			if len(parts) >= 5 {
				fmt.Printf("    %s  %s / %s  util %s  temp %s°C\n",
					theme.HighlightStyle.Render(strings.TrimSpace(parts[0])),
					theme.DimTextStyle.Render(strings.TrimSpace(parts[1])),
					theme.DimTextStyle.Render(strings.TrimSpace(parts[2])),
					theme.InfoStyle.Render(strings.TrimSpace(parts[3])),
					theme.DimTextStyle.Render(strings.TrimSpace(parts[4])))
			}
		}
		fmt.Println()
	}
}

