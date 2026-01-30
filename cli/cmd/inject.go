package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	injectGPUs    string
	injectHost    string
	injectPort    int
	injectVerbose bool
)

var injectCmd = &cobra.Command{
	Use:   "inject [path]",
	Short: "Inject codebase into model KV cache",
	Long: `Inject a codebase into the model's KV cache for context-aware inference.

This pre-loads the entire codebase into the model's memory, enabling
subsequent queries to have full awareness of the code structure.

Files are organized into logical groups for optimal cache access:
  - Core types and interfaces (loaded first for reference)
  - Configuration and constants
  - Utilities and helpers
  - Main implementation files
  - Tests (optional)

If no path is provided, defaults to the anime CLI codebase.

Examples:
  anime inject                          # Inject anime CLI codebase (default)
  anime inject .                        # Inject current directory
  anime inject ~/myproject              # Inject custom project
  anime inject --gpus 0,1,2,3           # Specify GPUs
  anime inject --host lambda            # Use remote model server
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInject,
}

func init() {
	injectCmd.Flags().StringVar(&injectGPUs, "gpus", "", "GPU IDs to use (comma-separated, e.g., 0,1,2,3)")
	injectCmd.Flags().StringVar(&injectHost, "host", "localhost", "Model server host")
	injectCmd.Flags().IntVar(&injectPort, "port", 8000, "Model server port")
	injectCmd.Flags().BoolVarP(&injectVerbose, "verbose", "v", false, "Show detailed progress")
	rootCmd.AddCommand(injectCmd)
}

// FileGroup represents a logical grouping of files for cache optimization
type FileGroup struct {
	Name        string
	Description string
	Priority    int // Lower = loaded first
	Files       []CodeFile
	TotalLines  int
	TotalTokens int
}

// CodeFile represents a source file to inject
type CodeFile struct {
	Path       string
	RelPath    string
	Content    string
	Lines      int
	Tokens     int
	Hash       string // For deduplication
	Group      string
	Imports    []string
	IsTest     bool
	IsGenerated bool
}

// InjectGPU represents a detected GPU for injection
type InjectGPU struct {
	ID          int
	Name        string
	Memory      string
	MemoryUsed  string
	Utilization string
	Selected    bool
}

// InjectionStats tracks injection progress
type InjectionStats struct {
	TotalFiles    int
	TotalLines    int
	TotalTokens   int
	TotalGroups   int
	InjectedFiles int
	DuplicatesSkipped int
	ElapsedTime   time.Duration
}

func runInject(cmd *cobra.Command, args []string) error {
	// Determine codebase path
	var codebasePath string
	if len(args) > 0 {
		codebasePath = args[0]
		// Expand ~ if present
		if strings.HasPrefix(codebasePath, "~") {
			home, _ := os.UserHomeDir()
			codebasePath = filepath.Join(home, codebasePath[1:])
		}
		// Make absolute
		if !filepath.IsAbs(codebasePath) {
			abs, err := filepath.Abs(codebasePath)
			if err == nil {
				codebasePath = abs
			}
		}
	} else {
		// Default to anime source
		if srcDir, err := findSourceDir(); err == nil {
			codebasePath = srcDir
		} else {
			// Try current directory
			codebasePath, _ = os.Getwd()
		}
	}

	// Verify path exists
	if _, err := os.Stat(codebasePath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", codebasePath)
	}

	// Run the injection TUI
	p := tea.NewProgram(newInjectModel(codebasePath), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m := finalModel.(injectModel)
	if m.completed {
		showInjectionSummary(m)
	}

	return nil
}

// injectModel is the bubbletea model for the injection TUI
type injectModel struct {
	codebasePath string
	projectName  string
	gpus         []InjectGPU
	gpuCursor    int
	phase        injectPhase
	spinner      spinner.Model
	progress     progress.Model
	stats        InjectionStats
	groups       []FileGroup
	currentGroup int
	currentFile  int
	seenHashes   map[string]bool
	err          error
	completed    bool
	width        int
	height       int
}

type injectPhase int

const (
	phaseDetectGPUs injectPhase = iota
	phaseSelectGPUs
	phaseScanCode
	phaseOrganize
	phaseInject
	phaseVerify
	phaseDone
)

// Messages
type gpusDetectedMsg struct {
	gpus []InjectGPU
	err  error
}

type codeScannedMsg struct {
	files []CodeFile
	err   error
}

type codeOrganizedMsg struct {
	groups []FileGroup
}

type fileInjectedMsg struct {
	groupIdx int
	fileIdx  int
	skipped  bool
	err      error
}

type injectionCompleteMsg struct{}

func newInjectModel(codebasePath string) injectModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	// Extract project name from path
	projectName := filepath.Base(codebasePath)

	return injectModel{
		codebasePath: codebasePath,
		projectName:  projectName,
		phase:        phaseDetectGPUs,
		spinner:      s,
		progress:     p,
		seenHashes:   make(map[string]bool),
		width:        80,
		height:       24,
	}
}

func (m injectModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		detectInjectGPUsCmd(),
	)
}

func (m injectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.phase == phaseSelectGPUs && m.gpuCursor > 0 {
				m.gpuCursor--
			}
		case "down", "j":
			if m.phase == phaseSelectGPUs && m.gpuCursor < len(m.gpus)-1 {
				m.gpuCursor++
			}
		case " ":
			if m.phase == phaseSelectGPUs && len(m.gpus) > 0 {
				m.gpus[m.gpuCursor].Selected = !m.gpus[m.gpuCursor].Selected
			}
		case "enter":
			if m.phase == phaseSelectGPUs {
				hasSelected := false
				for _, gpu := range m.gpus {
					if gpu.Selected {
						hasSelected = true
						break
					}
				}
				if !hasSelected && len(m.gpus) > 0 {
					for i := range m.gpus {
						m.gpus[i].Selected = true
					}
				}
				m.phase = phaseScanCode
				return m, scanCodebaseCmd(m.codebasePath)
			}
		case "a":
			if m.phase == phaseSelectGPUs {
				for i := range m.gpus {
					m.gpus[i].Selected = true
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case gpusDetectedMsg:
		if msg.err != nil {
			m.phase = phaseScanCode
			return m, scanCodebaseCmd(m.codebasePath)
		}
		m.gpus = msg.gpus
		if len(m.gpus) == 0 {
			m.phase = phaseScanCode
			return m, scanCodebaseCmd(m.codebasePath)
		}
		if injectGPUs != "" {
			m.selectGPUsByID(injectGPUs)
			m.phase = phaseScanCode
			return m, scanCodebaseCmd(m.codebasePath)
		}
		m.phase = phaseSelectGPUs
		return m, nil

	case codeScannedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.phase = phaseOrganize
		return m, organizeCodeCmd(msg.files)

	case codeOrganizedMsg:
		m.groups = msg.groups
		m.stats.TotalGroups = len(msg.groups)
		for _, g := range msg.groups {
			m.stats.TotalFiles += len(g.Files)
			m.stats.TotalLines += g.TotalLines
			m.stats.TotalTokens += g.TotalTokens
		}
		m.phase = phaseInject
		m.currentGroup = 0
		m.currentFile = 0
		if len(m.groups) > 0 && len(m.groups[0].Files) > 0 {
			return m, injectGroupFileCmd(m.groups, 0, 0, m.seenHashes, injectHost, injectPort)
		}
		m.phase = phaseDone
		m.completed = true
		return m, tea.Quit

	case fileInjectedMsg:
		if msg.skipped {
			m.stats.DuplicatesSkipped++
		} else if msg.err == nil {
			m.stats.InjectedFiles++
			// Track hash
			if msg.groupIdx < len(m.groups) && msg.fileIdx < len(m.groups[msg.groupIdx].Files) {
				hash := m.groups[msg.groupIdx].Files[msg.fileIdx].Hash
				m.seenHashes[hash] = true
			}
		}

		// Move to next file
		m.currentFile = msg.fileIdx + 1
		if m.currentFile >= len(m.groups[m.currentGroup].Files) {
			// Move to next group
			m.currentGroup++
			m.currentFile = 0
			if m.currentGroup >= len(m.groups) {
				m.phase = phaseVerify
				return m, verifyInjectionCmd()
			}
		}
		return m, injectGroupFileCmd(m.groups, m.currentGroup, m.currentFile, m.seenHashes, injectHost, injectPort)

	case injectionCompleteMsg:
		m.phase = phaseDone
		m.completed = true
		return m, tea.Quit
	}

	return m, nil
}

func (m *injectModel) selectGPUsByID(gpuList string) {
	ids := strings.Split(gpuList, ",")
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[strings.TrimSpace(id)] = true
	}
	for i := range m.gpus {
		if idMap[fmt.Sprintf("%d", m.gpus[i].ID)] {
			m.gpus[i].Selected = true
		}
	}
}

func (m injectModel) View() string {
	var s strings.Builder

	title := fmt.Sprintf("🧠 INJECTING: %s", strings.ToUpper(m.projectName))
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 2).
		Render(title)

	s.WriteString(header)
	s.WriteString("\n\n")

	switch m.phase {
	case phaseDetectGPUs:
		s.WriteString(fmt.Sprintf("  %s Detecting GPUs...\n", m.spinner.View()))

	case phaseSelectGPUs:
		s.WriteString(m.renderGPUSelection())

	case phaseScanCode:
		s.WriteString(fmt.Sprintf("  %s Scanning codebase...\n", m.spinner.View()))
		s.WriteString(fmt.Sprintf("     %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.codebasePath)))

	case phaseOrganize:
		s.WriteString(fmt.Sprintf("  %s Organizing files for optimal caching...\n", m.spinner.View()))

	case phaseInject:
		s.WriteString(m.renderInjectionProgress())

	case phaseVerify:
		s.WriteString(fmt.Sprintf("  %s Verifying injection...\n", m.spinner.View()))

	case phaseDone:
		s.WriteString(m.renderInjectionComplete())
	}

	s.WriteString("\n")
	if m.phase == phaseSelectGPUs {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"  ↑/↓ navigate • space toggle • a select all • enter confirm • q quit"))
	} else if m.phase != phaseDone {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"  q quit"))
	}

	return s.String()
}

func (m injectModel) renderGPUSelection() string {
	var s strings.Builder
	s.WriteString("  Select GPUs for injection:\n\n")

	for i, gpu := range m.gpus {
		cursor := "  "
		if i == m.gpuCursor {
			cursor = "▶ "
		}
		checkbox := "[ ]"
		if gpu.Selected {
			checkbox = "[✓]"
		}
		style := lipgloss.NewStyle()
		if i == m.gpuCursor {
			style = style.Bold(true).Foreground(lipgloss.Color("205"))
		}
		line := fmt.Sprintf("%s%s GPU %d: %s (%s / %s)",
			cursor, checkbox, gpu.ID, gpu.Name, gpu.MemoryUsed, gpu.Memory)
		s.WriteString(style.Render(line))
		s.WriteString("\n")
	}
	s.WriteString("\n")
	return s.String()
}

func (m injectModel) renderInjectionProgress() string {
	var s strings.Builder

	// Calculate overall progress
	totalFiles := 0
	doneFiles := 0
	for i, g := range m.groups {
		totalFiles += len(g.Files)
		if i < m.currentGroup {
			doneFiles += len(g.Files)
		} else if i == m.currentGroup {
			doneFiles += m.currentFile
		}
	}

	percent := float64(doneFiles) / float64(totalFiles)

	s.WriteString(fmt.Sprintf("  %s Injecting into KV cache...\n\n", m.spinner.View()))

	// Progress bar
	s.WriteString("  ")
	s.WriteString(m.progress.ViewAs(percent))
	s.WriteString("\n\n")

	// Group progress
	s.WriteString("  Groups:\n")
	for i, g := range m.groups {
		status := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		if i < m.currentGroup {
			status = "✓ "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		} else if i == m.currentGroup {
			status = "▶ "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
		}
		line := fmt.Sprintf("  %s%s (%d files)", status, g.Name, len(g.Files))
		s.WriteString(style.Render(line))
		s.WriteString("\n")
	}

	// Stats
	s.WriteString(fmt.Sprintf("\n  Files:    %d / %d\n", doneFiles, totalFiles))
	s.WriteString(fmt.Sprintf("  Injected: %d\n", m.stats.InjectedFiles))
	if m.stats.DuplicatesSkipped > 0 {
		s.WriteString(fmt.Sprintf("  Deduped:  %d\n", m.stats.DuplicatesSkipped))
	}

	// Current file
	if m.currentGroup < len(m.groups) && m.currentFile < len(m.groups[m.currentGroup].Files) {
		f := m.groups[m.currentGroup].Files[m.currentFile]
		s.WriteString(fmt.Sprintf("\n  Current: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(f.RelPath)))
	}

	return s.String()
}

func (m injectModel) renderInjectionComplete() string {
	var s strings.Builder

	checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	s.WriteString(checkStyle.Render("  ✓ Injection complete!\n\n"))

	// Visual KV cache representation
	boxWidth := 55
	s.WriteString(fmt.Sprintf("  ┌%s┐\n", strings.Repeat("─", boxWidth-2)))
	s.WriteString(fmt.Sprintf("  │%s│\n", centerText("🧠 MODEL KV CACHE", boxWidth-2)))
	s.WriteString(fmt.Sprintf("  │%s│\n", centerText(m.projectName, boxWidth-2)))
	s.WriteString(fmt.Sprintf("  ├%s┤\n", strings.Repeat("─", boxWidth-2)))

	// Show groups
	for _, g := range m.groups {
		icon := "📁"
		if strings.Contains(g.Name, "Types") || strings.Contains(g.Name, "Interface") {
			icon = "📐"
		} else if strings.Contains(g.Name, "Config") {
			icon = "⚙️"
		} else if strings.Contains(g.Name, "Util") || strings.Contains(g.Name, "Helper") {
			icon = "🔧"
		} else if strings.Contains(g.Name, "Test") {
			icon = "🧪"
		} else if strings.Contains(g.Name, "Main") || strings.Contains(g.Name, "Core") {
			icon = "💎"
		}
		line := fmt.Sprintf("  %s %-20s %d files, %d lines", icon, g.Name, len(g.Files), g.TotalLines)
		s.WriteString(fmt.Sprintf("  │ %-*s│\n", boxWidth-4, line))
	}

	s.WriteString(fmt.Sprintf("  ├%s┤\n", strings.Repeat("─", boxWidth-2)))

	summary := fmt.Sprintf("Total: %d files, %d lines, ~%dk tokens",
		m.stats.TotalFiles, m.stats.TotalLines, m.stats.TotalTokens/1000)
	s.WriteString(fmt.Sprintf("  │ %-*s│\n", boxWidth-4, summary))

	if m.stats.DuplicatesSkipped > 0 {
		dedup := fmt.Sprintf("Deduplicated: %d files", m.stats.DuplicatesSkipped)
		s.WriteString(fmt.Sprintf("  │ %-*s│\n", boxWidth-4, dedup))
	}

	s.WriteString(fmt.Sprintf("  └%s┘\n", strings.Repeat("─", boxWidth-2)))

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  The model now has full context of "))
	s.WriteString(highlightStyle.Render(m.projectName))
	s.WriteString(dimStyle.Render(".\n"))
	s.WriteString(dimStyle.Render("  Use 'anime inference llama' to query with context.\n"))

	return s.String()
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

// Commands

func detectInjectGPUsCmd() tea.Cmd {
	return func() tea.Msg {
		gpus, err := detectInjectGPUs()
		return gpusDetectedMsg{gpus: gpus, err: err}
	}
}

func detectInjectGPUs() ([]InjectGPU, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,memory.used,utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []InjectGPU
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) >= 5 {
			var id int
			fmt.Sscanf(parts[0], "%d", &id)
			gpus = append(gpus, InjectGPU{
				ID:          id,
				Name:        strings.TrimSpace(parts[1]),
				Memory:      strings.TrimSpace(parts[2]) + " MiB",
				MemoryUsed:  strings.TrimSpace(parts[3]) + " MiB",
				Utilization: strings.TrimSpace(parts[4]) + "%",
			})
		}
	}
	return gpus, nil
}

func scanCodebaseCmd(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := scanAllCode(path)
		return codeScannedMsg{files: files, err: err}
	}
}

func scanAllCode(rootPath string) ([]CodeFile, error) {
	var files []CodeFile

	// Detect language by looking for common files
	isGo := fileExists(filepath.Join(rootPath, "go.mod"))
	isPython := fileExists(filepath.Join(rootPath, "setup.py")) || fileExists(filepath.Join(rootPath, "pyproject.toml"))
	isJS := fileExists(filepath.Join(rootPath, "package.json"))
	isRust := fileExists(filepath.Join(rootPath, "Cargo.toml"))

	// Default extensions based on detected language
	var extensions []string
	if isGo {
		extensions = []string{".go"}
	} else if isPython {
		extensions = []string{".py"}
	} else if isJS {
		extensions = []string{".js", ".ts", ".jsx", ".tsx"}
	} else if isRust {
		extensions = []string{".rs"}
	} else {
		// Default to common source files
		extensions = []string{".go", ".py", ".js", ".ts", ".rs", ".java", ".c", ".cpp", ".h"}
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" ||
				name == "__pycache__" || name == "target" || name == "build" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check extension
		ext := filepath.Ext(path)
		matched := false
		for _, e := range extensions {
			if ext == e {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)
		lines := strings.Count(string(content), "\n") + 1
		tokens := len(content) / 4

		// Compute hash for deduplication
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:8])

		// Parse imports (Go-specific for now)
		var imports []string
		if ext == ".go" {
			imports = parseGoImports(string(content))
		}

		// Detect test files
		isTest := strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_test.py") ||
			strings.Contains(path, "/test/") || strings.Contains(path, "/tests/")

		// Detect generated files
		isGenerated := strings.Contains(string(content), "DO NOT EDIT") ||
			strings.Contains(string(content), "Code generated") ||
			strings.Contains(string(content), "auto-generated")

		files = append(files, CodeFile{
			Path:        path,
			RelPath:     relPath,
			Content:     string(content),
			Lines:       lines,
			Tokens:      tokens,
			Hash:        hashStr,
			Imports:     imports,
			IsTest:      isTest,
			IsGenerated: isGenerated,
		})

		return nil
	})

	return files, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func parseGoImports(content string) []string {
	var imports []string
	inImport := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import (") {
			inImport = true
			continue
		}
		if inImport {
			if line == ")" {
				break
			}
			// Extract import path
			line = strings.Trim(line, `"`)
			if line != "" && !strings.HasPrefix(line, "//") {
				imports = append(imports, line)
			}
		} else if strings.HasPrefix(line, "import ") {
			// Single import
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				imp := strings.Trim(parts[1], `"`)
				imports = append(imports, imp)
			}
		}
	}
	return imports
}

func organizeCodeCmd(files []CodeFile) tea.Cmd {
	return func() tea.Msg {
		groups := organizeIntoGroups(files)
		return codeOrganizedMsg{groups: groups}
	}
}

func organizeIntoGroups(files []CodeFile) []FileGroup {
	// Categorize files
	var (
		typesFiles    []CodeFile
		configFiles   []CodeFile
		utilFiles     []CodeFile
		coreFiles     []CodeFile
		cmdFiles      []CodeFile
		testFiles     []CodeFile
		otherFiles    []CodeFile
	)

	for _, f := range files {
		// Skip generated files
		if f.IsGenerated {
			continue
		}

		relLower := strings.ToLower(f.RelPath)
		baseName := strings.ToLower(filepath.Base(f.RelPath))

		switch {
		case f.IsTest:
			testFiles = append(testFiles, f)
		case strings.Contains(relLower, "/types") || strings.Contains(baseName, "types") ||
			strings.Contains(baseName, "interface") || strings.Contains(baseName, "model"):
			f.Group = "Types & Interfaces"
			typesFiles = append(typesFiles, f)
		case strings.Contains(relLower, "/config") || strings.Contains(baseName, "config") ||
			strings.Contains(baseName, "const") || strings.Contains(baseName, "settings"):
			f.Group = "Configuration"
			configFiles = append(configFiles, f)
		case strings.Contains(relLower, "/util") || strings.Contains(relLower, "/helper") ||
			strings.Contains(relLower, "/common") || strings.Contains(baseName, "util") ||
			strings.Contains(baseName, "helper"):
			f.Group = "Utilities"
			utilFiles = append(utilFiles, f)
		case strings.Contains(relLower, "/cmd/") || strings.HasPrefix(relLower, "cmd/"):
			f.Group = "Commands"
			cmdFiles = append(cmdFiles, f)
		case strings.Contains(relLower, "/internal/") || strings.Contains(relLower, "/pkg/") ||
			strings.Contains(relLower, "/src/"):
			f.Group = "Core"
			coreFiles = append(coreFiles, f)
		default:
			f.Group = "Other"
			otherFiles = append(otherFiles, f)
		}
	}

	// Sort files within each group by import dependencies (files with fewer deps first)
	sortByDependencies := func(files []CodeFile) {
		sort.Slice(files, func(i, j int) bool {
			return len(files[i].Imports) < len(files[j].Imports)
		})
	}

	sortByDependencies(typesFiles)
	sortByDependencies(configFiles)
	sortByDependencies(utilFiles)
	sortByDependencies(coreFiles)
	sortByDependencies(cmdFiles)

	// Build groups in optimal order for caching
	var groups []FileGroup

	addGroup := func(name, desc string, priority int, files []CodeFile) {
		if len(files) == 0 {
			return
		}
		g := FileGroup{
			Name:        name,
			Description: desc,
			Priority:    priority,
			Files:       files,
		}
		for _, f := range files {
			g.TotalLines += f.Lines
			g.TotalTokens += f.Tokens
		}
		groups = append(groups, g)
	}

	// Order matters for optimal cache access
	addGroup("Types & Interfaces", "Core type definitions loaded first for reference", 1, typesFiles)
	addGroup("Configuration", "Constants and config loaded early", 2, configFiles)
	addGroup("Utilities", "Helper functions available throughout", 3, utilFiles)
	addGroup("Core Implementation", "Main business logic", 4, coreFiles)
	addGroup("Commands", "CLI command handlers", 5, cmdFiles)
	addGroup("Other", "Additional source files", 6, otherFiles)
	// Tests last (or skip entirely for production)
	// addGroup("Tests", "Test files", 7, testFiles)

	return groups
}

func injectGroupFileCmd(groups []FileGroup, groupIdx, fileIdx int, seenHashes map[string]bool, host string, port int) tea.Cmd {
	return func() tea.Msg {
		if groupIdx >= len(groups) || fileIdx >= len(groups[groupIdx].Files) {
			return fileInjectedMsg{groupIdx: groupIdx, fileIdx: fileIdx, err: nil}
		}

		file := groups[groupIdx].Files[fileIdx]

		// Check for duplicate
		if seenHashes[file.Hash] {
			return fileInjectedMsg{groupIdx: groupIdx, fileIdx: fileIdx, skipped: true, err: nil}
		}

		err := injectFileToModel(file, groups[groupIdx].Name, host, port)
		return fileInjectedMsg{groupIdx: groupIdx, fileIdx: fileIdx, err: err}
	}
}

func injectFileToModel(file CodeFile, groupName string, host string, port int) error {
	// Create context-aware prompt
	prompt := fmt.Sprintf("[CONTEXT INJECTION - %s]\nFile: %s\nGroup: %s\nLines: %d\n\n```\n%s\n```\n\nACK",
		strings.ToUpper(groupName), file.RelPath, groupName, file.Lines, file.Content)

	req := LlamaRequest{
		Model: "meta-llama/Llama-3.3-70B-Instruct",
		Messages: []LlamaMessage{
			{Role: "system", Content: "You are being loaded with codebase context organized by category. Acknowledge each file with 'OK'. Do not analyze unless asked."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   5,
		Temperature: 0.1,
		Stream:      false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	client := &http.Client{Timeout: 60 * time.Second}

	resp, err := client.Post(
		baseURL+"/v1/chat/completions",
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func verifyInjectionCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond)
		return injectionCompleteMsg{}
	}
}

func showInjectionSummary(m injectModel) {
	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("✨ Codebase injected into model KV cache"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Project:"), theme.HighlightStyle.Render(m.projectName))
	fmt.Printf("  %s %d in %d groups\n", theme.DimTextStyle.Render("Files:"), m.stats.TotalFiles, m.stats.TotalGroups)
	fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Lines:"), m.stats.TotalLines)
	fmt.Printf("  %s ~%dk\n", theme.DimTextStyle.Render("Tokens:"), m.stats.TotalTokens/1000)
	if m.stats.DuplicatesSkipped > 0 {
		fmt.Printf("  %s %d\n", theme.DimTextStyle.Render("Deduped:"), m.stats.DuplicatesSkipped)
	}
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Query with context:"))
	fmt.Println(theme.HighlightStyle.Render("  anime inference llama \"How does X work?\""))
	fmt.Println()
}
