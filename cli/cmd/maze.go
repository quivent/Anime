package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

// `anime maze` exposes the dependency-hell that turns a 15-minute "install vLLM"
// into a 7-hour yak-shave on Lambda GH200. Every section here is a question we
// kept needing to answer manually during install/debug; this command makes the
// answers a single invocation.

var mazeCmd = &cobra.Command{
	Use:   "maze",
	Short: "Map the dependency maze that makes installs take hours",
	Long: `Maze pinpoints WHY installs break: driver/CUDA mismatch, torch wheel
shadowing, vllm cu12 vs cu13 wheel availability, numpy ABI gaps,
PIP_CONSTRAINT enforcement, package-version conflicts, etc.

Subcommands:
  lambda    Full diagnostic for Lambda Cloud GH200 (ARM64 / Lambda Stack)
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var (
	mazeOnlyCUDA      bool
	mazeOnlyPackages  bool
	mazeOnlyDeps      bool
	mazeOnlyWheels    bool
	mazeOnlyShadows   bool
	mazeOnlyTorch     bool
	mazeOnlyVLLM      bool
	mazeOnlyHistory   bool
	mazeOnlyAnimePath bool
	mazeNoProbe       bool
)

var mazeLambdaCmd = &cobra.Command{
	Use:   "lambda",
	Short: "Map every dependency issue on Lambda GH200 / ARM64",
	Long: `Walks every dimension of the dependency maze that breaks vLLM-on-Lambda installs:

  driver/cuda          GPU driver vs CUDA toolkit vs vllm wheel CUDA
  torch                System torch vs pip torch vs cu128 vs +cpu shadowing
  vllm                 Which vllm version pairs with which torch + cuda
  wheels               PyPI wheel availability matrix (cu12 vs cu13 vs source)
  packages             Currently installed packages with version conflicts
  dependencies         Transitive deps that silently upgrade torch
  shadows              user-site vs system-site vs venv shadowing
  numpy                numpy 1.x vs 2.x ABI breaks with system pandas/sklearn
  constraints          PIP_CONSTRAINT state and what it pins
  anime-path           Which anime install paths exist + which call which
  history              The git commits that fought these issues

Each section is independent. Pass --only-X to limit. Probes are read-only.

Examples:
  anime maze lambda                # Full report
  anime maze lambda --cuda         # CUDA/driver/wheel matrix only
  anime maze lambda --shadows      # Package shadowing only
  anime maze lambda --no-probe     # Static knowledge only, skip live probes`,
	Run: runMazeLambda,
}

func init() {
	rootCmd.AddCommand(mazeCmd)
	mazeCmd.AddCommand(mazeLambdaCmd)

	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyCUDA, "cuda", false, "Only CUDA/driver/toolkit section")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyTorch, "torch", false, "Only torch shadowing/version section")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyVLLM, "vllm", false, "Only vllm version matrix section")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyWheels, "wheels", false, "Only wheel availability matrix")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyPackages, "packages", false, "Only installed-package conflicts")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyDeps, "dependencies", false, "Only transitive-dep traps")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyShadows, "shadows", false, "Only user/system shadowing")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyHistory, "history", false, "Only fight-the-bug commit history")
	mazeLambdaCmd.Flags().BoolVar(&mazeOnlyAnimePath, "anime-paths", false, "Only anime CLI install path audit")
	mazeLambdaCmd.Flags().BoolVar(&mazeNoProbe, "no-probe", false, "Skip live probes (use static knowledge only)")
}

// ─── probe helpers ───────────────────────────────────────────────────────────

func mazeRun(cmd string, args ...string) string {
	out, _ := exec.Command(cmd, args...).Output()
	return strings.TrimSpace(string(out))
}

func mazePyVer(pkg string) string {
	if mazeNoProbe {
		return "?"
	}
	out, err := exec.Command("python3", "-c",
		fmt.Sprintf("import %s; print(getattr(%s, '__version__', '?'))", pkg, pkg)).Output()
	if err != nil {
		return "(not installed)"
	}
	return strings.TrimSpace(string(out))
}

func mazePyPath(pkg string) string {
	if mazeNoProbe {
		return "?"
	}
	out, err := exec.Command("python3", "-c",
		fmt.Sprintf("import %s; print(%s.__file__)", pkg, pkg)).Output()
	if err != nil {
		return "(not installed)"
	}
	return strings.TrimSpace(string(out))
}

func mazePyCUDA() (avail, version string) {
	if mazeNoProbe {
		return "?", "?"
	}
	out, err := exec.Command("python3", "-c",
		"import torch; print(torch.cuda.is_available(), torch.version.cuda)").Output()
	if err != nil {
		return "(no torch)", ""
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "?", "?"
}

func mazeNvidiaDriver() string {
	if mazeNoProbe {
		return "?"
	}
	out := mazeRun("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	if out == "" {
		return "(nvidia-smi unavailable)"
	}
	return strings.TrimSpace(out)
}

func mazeNvidiaCUDA() string {
	if mazeNoProbe {
		return "?"
	}
	out := mazeRun("nvidia-smi")
	re := regexp.MustCompile(`CUDA Version: ([\d.]+)`)
	if m := re.FindStringSubmatch(out); len(m) > 1 {
		return m[1]
	}
	return "?"
}

func mazeArch() string {
	return runtime.GOARCH
}

func mazeUserSiteHas(pkg string) (exists bool, location string) {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local/lib/python3.10/site-packages", pkg),
		filepath.Join(home, ".local/lib/python3.10/site-packages", pkg+"-*.dist-info"),
	}
	for _, c := range candidates {
		matches, _ := filepath.Glob(c)
		if len(matches) > 0 {
			return true, matches[0]
		}
	}
	return false, ""
}

func mazeSystemSiteHas(pkg string) (exists bool, location string) {
	candidates := []string{
		filepath.Join("/usr/lib/python3/dist-packages", pkg),
		filepath.Join("/usr/lib/python3.10/dist-packages", pkg),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true, c
		}
	}
	return false, ""
}

func mazePipConstraint() string {
	c := os.Getenv("PIP_CONSTRAINT")
	if c == "" {
		return "(unset)"
	}
	return c
}

func mazeIsLambdaStack() bool {
	out := mazeRun("dpkg", "-l")
	return strings.Contains(out, "lambda-stack")
}

// ─── output helpers ──────────────────────────────────────────────────────────

func mazeBanner() {
	bar := strings.Repeat("━", 78)
	fmt.Println(theme.GlowStyle.Render(bar))
	fmt.Println(theme.GlowStyle.Render("  THE LAMBDA GH200 DEPENDENCY MAZE"))
	fmt.Println(theme.DimTextStyle.Render("  why every install takes 7 hours"))
	fmt.Println(theme.GlowStyle.Render(bar))
	fmt.Println()
}

func mazeSection(title, why string) {
	fmt.Println()
	fmt.Println(theme.HeaderStyle.Render("▶ " + title))
	if why != "" {
		fmt.Println(theme.DimTextStyle.Render("  " + why))
	}
	fmt.Println(theme.DimTextStyle.Render("  "+strings.Repeat("─", 74)))
}

func mazeRow(label, value, note string) {
	lbl := theme.SecondaryTextStyle.Render(fmt.Sprintf("  %-22s", label))
	val := theme.HighlightStyle.Render(value)
	if note != "" {
		fmt.Printf("%s %s  %s\n", lbl, val, theme.DimTextStyle.Render(note))
	} else {
		fmt.Printf("%s %s\n", lbl, val)
	}
}

func mazeAlert(level, msg string) {
	switch level {
	case "fatal":
		fmt.Println("  " + theme.ErrorStyle.Render("✗ FATAL") + "  " + msg)
	case "warn":
		fmt.Println("  " + theme.WarningStyle.Render("⚠ WARN ") + "  " + msg)
	case "ok":
		fmt.Println("  " + theme.SuccessStyle.Render("✓ OK   ") + "  " + msg)
	case "trap":
		fmt.Println("  " + theme.ErrorStyle.Render("💀 TRAP") + "  " + msg)
	default:
		fmt.Println("  " + msg)
	}
}

func mazeFact(s string) {
	fmt.Println("  " + theme.PrimaryTextStyle.Render(s))
}

// ─── sections ────────────────────────────────────────────────────────────────

func mazeSectionCUDA() {
	mazeSection("DRIVER / CUDA / TOOLKIT",
		"the load-bearing version triple — get one wrong, vllm's binary wheel won't load")

	driver := mazeNvidiaDriver()
	smiCUDA := mazeNvidiaCUDA()
	nvccVer := mazeRun("nvcc", "--version")
	nvccLine := ""
	for _, l := range strings.Split(nvccVer, "\n") {
		if strings.Contains(l, "release") {
			nvccLine = strings.TrimSpace(l)
			break
		}
	}

	mazeRow("NVIDIA driver", driver, "")
	mazeRow("Driver max CUDA", smiCUDA, "what nvidia-smi reports — bound by driver, not toolkit")
	mazeRow("nvcc (toolkit)", nvccLine, "what local compiles target — can differ from driver max")
	mazeRow("Arch", mazeArch(), "")
	mazeRow("Lambda Stack?", fmt.Sprintf("%v", mazeIsLambdaStack()),
		"if true, system Python has a curated torch in /usr/lib/python3/dist-packages")

	fmt.Println()
	mazeFact("Hard rule (NVIDIA cuda-compatibility docs):")
	fmt.Println("    CUDA 13.x requires driver >= 580.65.06")
	fmt.Println("    CUDA 12.x works on driver >= 525")
	fmt.Println("    Within-major minor versions ARE binary-compatible.")
	fmt.Println("    Cross-major (12→13) is NOT — libcudart.so.13 ≠ libcudart.so.12.")
	fmt.Println()
	mazeFact("Forward-compat package cuda-compat-13-0 exists for datacenter GPUs")
	fmt.Println("    but the agent research showed it requires driver >= 580 too.")
	fmt.Println("    For driver 570 + cu13 vllm, the package will not bridge.")

	if strings.HasPrefix(driver, "570") {
		fmt.Println()
		mazeAlert("trap", "Driver 570 + modern vllm (≥0.20) cu13 wheel = ImportError: libcudart.so.13")
		mazeAlert("warn", "Either upgrade driver to 580+ OR use vllm <= 0.10.x (cu12 wheels)")
	} else if strings.HasPrefix(driver, "580") || strings.HasPrefix(driver, "590") {
		mazeAlert("ok", "Driver supports CUDA 13 — vllm cu13 wheels should load")
	}
}

func mazeSectionTorch() {
	mazeSection("TORCH SHADOWING",
		"three torches can coexist; the wrong one wins and silently destroys CUDA")

	pyVer := mazePyVer("torch")
	pyPath := mazePyPath("torch")
	avail, cudaVer := mazePyCUDA()

	mazeRow("torch (active)", pyVer, "what python3 -c 'import torch' returns")
	mazeRow("torch.__file__", pyPath, "the file that loaded — tells you which install won")
	mazeRow("torch.cuda.is_available", avail, "False = you have CPU torch shadowing GPU torch")
	mazeRow("torch.version.cuda", cudaVer, "build-time CUDA version of the loaded torch")
	mazeRow("PIP_CONSTRAINT", mazePipConstraint(),
		"if set, pins torch across pip installs; if unset, pip can swap freely")

	fmt.Println()
	mazeFact("Three torch install locations possible on Lambda Stack GH200:")
	fmt.Println("    1. /usr/lib/python3/dist-packages/torch     (apt: Lambda Stack curated ARM+CUDA)")
	fmt.Println("    2. ~/.local/lib/python3.10/site-packages/torch  (pip --user — SHADOWS #1)")
	fmt.Println("    3. /usr/local/lib/python3.10/dist-packages/torch (sudo pip — rare)")
	fmt.Println()
	mazeFact("Python's sys.path puts user-site BEFORE system dist-packages.")
	fmt.Println("    Any `pip install --user X` where X depends on torch will pull torch into user-site")
	fmt.Println("    and the apt Lambda torch becomes invisible — even if it still exists on disk.")
	fmt.Println()
	mazeFact("Default PyPI torch wheel for aarch64 is `torch-X.Y.Z` (CPU-only) NOT a CUDA build.")
	fmt.Println("    To get CUDA torch on aarch64 you must pull from --index-url cu128/cu129/cu130.")
	fmt.Println("    `pip install torch` alone on aarch64 = CPU torch = CUDA gone.")

	sysT, _ := mazeSystemSiteHas("torch")
	userT, userLoc := mazeUserSiteHas("torch")
	if sysT && userT {
		fmt.Println()
		mazeAlert("trap", "Both system and user-site torch exist. User-site at "+userLoc+" wins.")
	} else if userT && !sysT {
		fmt.Println()
		mazeAlert("warn", "Only user-site torch present at "+userLoc)
	} else if sysT && !userT {
		fmt.Println()
		mazeAlert("ok", "Clean: system Lambda Stack torch only")
	}
}

func mazeSectionVLLM() {
	mazeSection("VLLM VERSION ↔ TORCH ↔ CUDA MATRIX",
		"pick the wrong vllm version and you need a driver upgrade")

	vllmVer := mazePyVer("vllm")
	vllmPath := mazePyPath("vllm")
	mazeRow("vllm (active)", vllmVer, "")
	mazeRow("vllm.__file__", vllmPath, "")

	canLoadC := "?"
	if !mazeNoProbe {
		if err := exec.Command("python3", "-c", "import vllm._C").Run(); err == nil {
			canLoadC = "yes"
		} else {
			canLoadC = "NO (libcudart mismatch likely)"
		}
	}
	mazeRow("vllm._C loads?", canLoadC, "the C extension's CUDA must match driver's max CUDA")

	fmt.Println()
	mazeFact("vllm version → torch pin → CUDA wheel target (per agent research):")
	fmt.Println()
	rows := []struct{ vllm, torch, cuda, wheel string }{
		{"0.6.6.post1", "==2.5.1", "cu12", "no aarch64 wheel — source-only"},
		{"0.7.3", "==2.5.1", "cu12", "no aarch64 wheel — source-only"},
		{"0.8.5.post1", "==2.6.0", "cu12", "no aarch64 wheel — source-only"},
		{"0.9.0 / 0.9.2", "==2.7.0", "cu12", "no aarch64 wheel — source-only"},
		{"0.10.0 / 0.10.1.1", "==2.7.1", "cu12", "aarch64 cu128 wheel exists ← SAFE for driver 570"},
		{"0.10.2", "==2.8.0", "cu12", "aarch64 wheel ships, torch 2.8.0+cu128 may be gone from index"},
		{"0.11.0+", "==2.8.0", "cu12 → cu13 transition", "PyPI default is cu13 binary"},
		{"0.20.0+", "==2.11.0", "cu13 by default", "needs driver 580+; cu128 wheel via uv --torch-backend=cu128"},
		{"0.21.0 (latest)", "==2.11.0", "cu13 default / cu128 alt", "use uv --torch-backend=cu128 for cu12.x drivers"},
	}
	for _, r := range rows {
		fmt.Printf("    %s  torch%s  %s  %s\n",
			theme.HighlightStyle.Render(fmt.Sprintf("%-18s", r.vllm)),
			theme.SecondaryTextStyle.Render(fmt.Sprintf("%-10s", r.torch)),
			theme.InfoStyle.Render(fmt.Sprintf("%-22s", r.cuda)),
			theme.DimTextStyle.Render(r.wheel))
	}
	fmt.Println()
	mazeFact("Critical: vllm METADATA hard-pins one exact torch version per release.")
	fmt.Println("    Without --no-deps or PIP_CONSTRAINT, vllm install REPLACES your torch.")
}

func mazeSectionWheels() {
	mazeSection("WHEEL AVAILABILITY MAZE",
		"which index has what — most install hours are spent here")

	mazeFact("PyPI default (pypi.org):")
	fmt.Println("    torch-X.Y.Z aarch64 = CPU-only (NO CUDA). Trap #1 on GH200.")
	fmt.Println("    vllm aarch64 wheels exist from 0.10.2+. They are cu13 from 0.20+.")
	fmt.Println("    nvidia-*-cu12 packages: latest aarch64 wheels available (12.9.x).")
	fmt.Println("    nvidia-*-cu13 packages: most are 0.0.1 placeholders pointing elsewhere.")
	fmt.Println()
	mazeFact("PyTorch CUDA index (https://download.pytorch.org/whl/cuXYZ):")
	fmt.Println("    cu128 aarch64 torch: 2.7.0, 2.7.1, 2.9.0, 2.9.1, 2.10.0, 2.11.0")
	fmt.Println("    cu129 aarch64 torch: 2.10+, 2.11+ (newer line)")
	fmt.Println("    cu130 aarch64 torch: 2.11+ — needs driver 580+")
	fmt.Println("    NOTE: 2.8.0 is missing from cu128. Skip from 2.7.1 → 2.9.0.")
	fmt.Println()
	mazeFact("NVIDIA NGC containers (nvcr.io/nvidia/vllm):")
	fmt.Println("    Started at 25.09 with CUDA 13.0. Every tag since is cu13.")
	fmt.Println("    Driver requirement floor: 580+. No cu12 NGC vllm image exists.")
	fmt.Println()
	mazeFact("Docker Hub (vllm/vllm-openai):")
	fmt.Println("    v0.8.x → v0.10.x = cu128 (works on driver 570)")
	fmt.Println("    v0.11+ = cu129 / cu130 (needs driver 580+)")
	fmt.Println("    drikster80/vllm-gh200-openai = stale (~1 year, vllm 0.6.x)")
	fmt.Println()
	mazeFact("`uv --torch-backend=cuXYZ` flag:")
	fmt.Println("    Only affects the TORCH wheel uv picks. Does NOT change vllm's wheel.")
	fmt.Println("    Won't save you from cu13 vllm + cu12 driver mismatch.")
}

func mazeSectionPackages() {
	mazeSection("INSTALLED PACKAGE CONFLICTS",
		"the live state of dependency conflicts in user-site")

	if mazeNoProbe {
		fmt.Println("  (skipped: --no-probe)")
		return
	}
	out, _ := exec.Command("pip3", "check").CombinedOutput()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && (lines[0] == "" || strings.Contains(lines[0], "No broken")) {
		mazeAlert("ok", "pip3 check reports no conflicts")
	} else {
		mazeAlert("warn", fmt.Sprintf("%d conflicts:", len(lines)))
		for _, l := range lines {
			if l != "" {
				fmt.Println("    " + theme.WarningStyle.Render(l))
			}
		}
	}

	fmt.Println()
	mazeFact("Common chronic conflicts on this stack:")
	fmt.Println("    • pandas (system) requires tzdata>=2022.7 — apt has older")
	fmt.Println("    • pandas (system) requires python-dateutil>=2.8.2 — apt has 2.8.1")
	fmt.Println("    • flashinfer-python pins nvidia-cutlass-dsl>=4.4.2 — often not installed")
	fmt.Println("    • opencv-python wants numpy>=2 — but system pandas C-ext needs numpy<2")
	fmt.Println("    • flatbuffers system pkg has invalid version '1.12.1-git...' (apt warning, harmless)")
}

func mazeSectionDeps() {
	mazeSection("TRANSITIVE DEPENDENCY TRAPS",
		"packages that silently replace your torch when pip-installed naively")

	fmt.Println()
	mazeFact("Packages with hard torch dependencies that trigger upgrade:")
	traps := []struct{ pkg, why string }{
		{"xformers", "Declares torch>=N. Without --no-deps, pulls latest torch — usually wrong CUDA."},
		{"vllm", "Hard-pins torch==EXACT_VERSION in METADATA. Will replace any other version."},
		{"flashinfer-python", "Pulls nvidia-*-cu13 packages as runtime deps."},
		{"bitsandbytes", "Bundles its own CUDA libs; can conflict with system torch's CUDA."},
		{"torchaudio", "Tightly couples to torch version. Mismatched versions = symbol errors."},
		{"transformers", "Soft torch dep, but its dep chain (tokenizers, etc.) sometimes pulls torch."},
	}
	for _, t := range traps {
		fmt.Printf("    %s\n", theme.WarningStyle.Render("• "+t.pkg))
		fmt.Printf("      %s\n", theme.DimTextStyle.Render(t.why))
	}
	fmt.Println()
	mazeFact("Defenses (deployed in anime as of this commit):")
	fmt.Println("    • PIP_CONSTRAINT — file pinning torch==<current>, exported globally")
	fmt.Println("    • --no-deps for vllm + xformers installs")
	fmt.Println("    • --no-build-isolation for source builds (uses existing torch)")
	fmt.Println("    • Post-install assert that torch.cuda.is_available() still works")
}

func mazeSectionShadows() {
	mazeSection("PACKAGE SHADOWING (user-site vs system-site)",
		"two installs of the same package — Python picks the wrong one")

	pkgs := []string{"torch", "torchvision", "torchaudio", "numpy", "pandas", "sklearn",
		"vllm", "transformers", "flash_attn", "xformers"}
	fmt.Println()
	fmt.Printf("  %-18s %-12s %-12s %s\n",
		theme.SecondaryTextStyle.Render("PACKAGE"),
		theme.SecondaryTextStyle.Render("USER-SITE"),
		theme.SecondaryTextStyle.Render("SYSTEM-SITE"),
		theme.SecondaryTextStyle.Render("WINS"))
	for _, p := range pkgs {
		us, _ := mazeUserSiteHas(p)
		ss, _ := mazeSystemSiteHas(p)
		userMark := theme.DimTextStyle.Render("--")
		sysMark := theme.DimTextStyle.Render("--")
		winner := theme.DimTextStyle.Render("(not installed)")
		if us {
			userMark = theme.SuccessStyle.Render("yes")
		}
		if ss {
			sysMark = theme.SuccessStyle.Render("yes")
		}
		switch {
		case us && ss:
			winner = theme.WarningStyle.Render("user-site (shadows system!)")
		case us:
			winner = theme.HighlightStyle.Render("user-site")
		case ss:
			winner = theme.HighlightStyle.Render("system-site")
		}
		fmt.Printf("  %-18s %-12s %-12s %s\n", p, userMark, sysMark, winner)
	}
	fmt.Println()
	mazeFact("Python sys.path order on this system:")
	out, _ := exec.Command("python3", "-c",
		"import sys; [print('   ', p) for p in sys.path if p]").Output()
	fmt.Print(string(out))
}

func mazeSectionNumpy() {
	mazeSection("NUMPY 1.x vs 2.x ABI BREAK",
		"numpy 2.x is a hard ABI break for C-extension packages compiled against 1.x")

	ver := mazePyVer("numpy")
	mazeRow("numpy (active)", ver, "")

	fmt.Println()
	mazeFact("Lambda Stack ships system pandas, scipy, sklearn compiled against numpy 1.x.")
	fmt.Println("    Any install that bumps numpy → 2.x silently breaks these:")
	fmt.Println("      'ValueError: numpy.dtype size changed, may indicate binary incompatibility'")
	fmt.Println()
	mazeFact("Things that upgrade numpy without asking:")
	fmt.Println("    • opencv-python >= 4.13 requires numpy>=2")
	fmt.Println("    • xformers builds against latest numpy")
	fmt.Println("    • vllm 0.20+ requires numpy>=2 indirectly")
	fmt.Println()
	mazeFact("Defense:")
	fmt.Println("    pip install 'numpy<2' --user  (forces 1.26.x — last 1.x line)")
	fmt.Println("    Add 'numpy<2' to PIP_CONSTRAINT to make it sticky.")
}

func mazeSectionConstraints() {
	mazeSection("PIP_CONSTRAINT STATE",
		"the single mechanism that prevents transitive deps from replacing torch")

	c := mazePipConstraint()
	mazeRow("PIP_CONSTRAINT env", c, "")
	if c != "(unset)" && c != "" {
		if data, err := os.ReadFile(c); err == nil {
			fmt.Println()
			mazeFact("Contents of " + c + ":")
			for _, l := range strings.Split(string(data), "\n") {
				if l != "" {
					fmt.Println("    " + theme.HighlightStyle.Render(l))
				}
			}
		}
	} else {
		mazeAlert("warn", "PIP_CONSTRAINT unset — pip can freely upgrade torch via any transitive dep")
		fmt.Println()
		mazeFact("anime sets PIP_CONSTRAINT automatically on every command (PersistentPreRun")
		fmt.Println("    in cmd/pip_constraint.go). If unset here, you ran pip outside anime.")
	}
}

func mazeSectionAnimePaths() {
	mazeSection("ANIME INSTALL PATH AUDIT",
		"every code path that ends up calling pip install — must all share defenses")

	paths := []struct{ where, what string }{
		{"internal/installer/scripts.go: pytorch", "CANONICAL. PIP_CONSTRAINT + xformers --no-deps + post-verify."},
		{"internal/installer/scripts.go: vllm", "CANONICAL. PIP_CONSTRAINT + vllm --no-deps + source-build for aarch64+cu12."},
		{"internal/tui/vllm.go: install flow", "DEDUPED. Now calls installer.GetScript() — no inline pip."},
		{"cmd/vllm.go: runVLLMStart preflight", "Invokes vllm_doctor. Doctor may run fixes."},
		{"cmd/vllm_doctor.go: 6 FixCommand entries", "PATCHED. All vllm reinstalls use --no-deps."},
		{"cmd/pip_constraint.go", "NEW. PersistentPreRun writes constraint + os.Setenv(PIP_CONSTRAINT)."},
		{"cmd/embedded/sky/*", "EMBEDDED TOOL. Separate codepath, not audited."},
		{"internal/protocol/coverage.go", "Mentions 'vllm>=0.6.0' as install target. Reference only."},
	}
	for _, p := range paths {
		fmt.Printf("  %s\n", theme.HighlightStyle.Render("• "+p.where))
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(p.what))
	}
	fmt.Println()
	mazeFact("Doctor's FixCommand runs via exec.Command with no cmd.Env — inherits os.Environ().")
	fmt.Println("    PersistentPreRun's os.Setenv(PIP_CONSTRAINT) propagates automatically.")
	fmt.Println()
	mazeFact("Untouched paths (potential trap surface):")
	fmt.Println("    • internal/installer/scripts.go: cogvideo, opensora, ltxvideo all use")
	fmt.Println("      'pip3 install --upgrade-strategy only-if-needed diffusers transformers")
	fmt.Println("      accelerate' — same flaw as the old pytorch script but only fires on")
	fmt.Println("      `anime install <video-model>`. Not in the vllm-start blast radius.")
}

func mazeSectionHistory() {
	mazeSection("FIGHT-THE-BUG COMMIT HISTORY",
		"every git commit that fought a dimension of this maze")

	if mazeNoProbe {
		fmt.Println("  (skipped: --no-probe)")
		return
	}
	cmd := exec.Command("git", "-C", "/home/ubuntu/Anime", "log",
		"--all", "--oneline", "--grep=vllm\\|llamaserve\\|flash_attn\\|torch\\|cuda\\|GH200\\|aarch64\\|Lambda Stack", "-20")
	out, err := cmd.Output()
	if err != nil {
		mazeAlert("warn", "git log failed: "+err.Error())
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	commits := []string{}
	for scanner.Scan() {
		commits = append(commits, scanner.Text())
	}
	sort.Strings(commits)
	for _, c := range commits {
		fmt.Printf("    %s\n", theme.DimTextStyle.Render(c))
	}
	fmt.Println()
	mazeFact(fmt.Sprintf("%d commits in this repo fought this exact maze.", len(commits)))
}

// ─── orchestration ───────────────────────────────────────────────────────────

func runMazeLambda(cmd *cobra.Command, args []string) {
	mazeBanner()

	only := mazeOnlyCUDA || mazeOnlyTorch || mazeOnlyVLLM || mazeOnlyWheels ||
		mazeOnlyPackages || mazeOnlyDeps || mazeOnlyShadows || mazeOnlyHistory || mazeOnlyAnimePath

	if !only || mazeOnlyCUDA {
		mazeSectionCUDA()
	}
	if !only || mazeOnlyTorch {
		mazeSectionTorch()
	}
	if !only || mazeOnlyVLLM {
		mazeSectionVLLM()
	}
	if !only || mazeOnlyWheels {
		mazeSectionWheels()
	}
	if !only || mazeOnlyPackages {
		mazeSectionPackages()
	}
	if !only || mazeOnlyDeps {
		mazeSectionDeps()
	}
	if !only || mazeOnlyShadows {
		mazeSectionShadows()
	}
	if !only {
		mazeSectionNumpy()
		mazeSectionConstraints()
	}
	if !only || mazeOnlyAnimePath {
		mazeSectionAnimePaths()
	}
	if !only || mazeOnlyHistory {
		mazeSectionHistory()
	}

	fmt.Println()
	fmt.Println(theme.GlowStyle.Render(strings.Repeat("━", 78)))
	fmt.Println(theme.SuccessStyle.Render("  Maze mapped. Now you know why."))
	fmt.Println(theme.GlowStyle.Render(strings.Repeat("━", 78)))
}
