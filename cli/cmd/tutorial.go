package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var tutorialCmd = &cobra.Command{
	Use:   "walkthrough",
	Short: "Interactive walkthrough showing how to use anime",
	Long: `Interactive walkthrough that guides you through the main features of anime step-by-step.

Perfect for first-time users to learn the basics with live demos!`,
	Aliases: []string{"tutorial", "learn", "demo"},
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialTutorialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running walkthrough: %v\n", err)
		}
	},
}

type tutorialStep int

const (
	stepWelcome tutorialStep = iota
	stepOverview
	stepAddServer
	stepAddServerDemo
	stepPackages
	stepPackagesDemo
	stepInstall
	stepInstallDemo
	stepStatus
	stepStatusDemo
	stepComplete
)

type tutorialModel struct {
	step         tutorialStep
	spinner      spinner.Model
	width        int
	height       int
	quitting     bool
	demoRunning  bool
	demoProgress float64
}

func initialTutorialModel() tutorialModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return tutorialModel{
		step:    stepWelcome,
		spinner: s,
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m tutorialModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m tutorialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if !m.demoRunning {
				// Progress to next step
				if m.step < stepComplete {
					m.step++
					m.demoProgress = 0

					// Start demo for certain steps
					if m.step == stepAddServerDemo || m.step == stepPackagesDemo ||
						m.step == stepInstallDemo || m.step == stepStatusDemo {
						m.demoRunning = true
						return m, tickCmd()
					}
				} else {
					m.quitting = true
					return m, tea.Quit
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		if m.demoRunning {
			m.demoProgress += 0.02
			if m.demoProgress >= 1.0 {
				m.demoRunning = false
				m.demoProgress = 0
				m.step++
				return m, nil
			}
			return m, tickCmd()
		}
	}

	return m, nil
}

func (m tutorialModel) View() string {
	if m.quitting {
		return "\n" + theme.SuccessStyle.Render("  ✨ Thanks for completing the tutorial! Run 'anime --help' for more info.") + "\n\n"
	}

	var content string

	switch m.step {
	case stepWelcome:
		content = m.renderWelcome()
	case stepOverview:
		content = m.renderOverview()
	case stepAddServer:
		content = m.renderAddServer()
	case stepAddServerDemo:
		content = m.renderAddServerDemo()
	case stepPackages:
		content = m.renderPackages()
	case stepPackagesDemo:
		content = m.renderPackagesDemo()
	case stepInstall:
		content = m.renderInstall()
	case stepInstallDemo:
		content = m.renderInstallDemo()
	case stepStatus:
		content = m.renderStatus()
	case stepStatusDemo:
		content = m.renderStatusDemo()
	case stepComplete:
		content = m.renderComplete()
	}

	// Add controls at bottom
	controls := "\n\n" + theme.DimTextStyle.Render("  Press Enter to continue • Press q or Esc to quit")

	return "\n" + content + controls + "\n"
}

func (m tutorialModel) renderWelcome() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("🎌 ANIME TUTORIAL"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  Welcome to the anime interactive tutorial!"))
	b.WriteString("\n\n")
	b.WriteString(theme.DimTextStyle.Render("  This walkthrough will guide you through:"))
	b.WriteString("\n\n")
	b.WriteString("    " + theme.SuccessStyle.Render("✓") + " Browsing available packages\n")
	b.WriteString("    " + theme.SuccessStyle.Render("✓") + " Installing packages on your machine\n")
	b.WriteString("    " + theme.SuccessStyle.Render("✓") + " Managing remote servers\n")
	b.WriteString("    " + theme.SuccessStyle.Render("✓") + " Checking system status\n")
	b.WriteString("\n")
	b.WriteString(theme.GlowStyle.Render("  Let's get started!"))

	return b.String()
}

func (m tutorialModel) renderOverview() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("📚 OVERVIEW"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  anime is a package manager for AI/ML software"))
	b.WriteString("\n\n")
	b.WriteString(theme.HighlightStyle.Render("  Main workflows:"))
	b.WriteString("\n\n")

	workflows := []struct {
		emoji   string
		command string
		desc    string
	}{
		{"1️⃣", "anime packages", "Browse available packages"},
		{"2️⃣", "anime install <id>", "Install packages locally"},
		{"3️⃣", "anime server add", "Add remote servers (optional)"},
		{"4️⃣", "anime doctor", "Diagnose installation issues"},
	}

	for _, wf := range workflows {
		b.WriteString(fmt.Sprintf("  %s  %s\n", wf.emoji, theme.SuccessStyle.Render(wf.command)))
		b.WriteString(fmt.Sprintf("      %s\n\n", theme.DimTextStyle.Render(wf.desc)))
	}

	return b.String()
}

func (m tutorialModel) renderAddServer() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("1️⃣ BROWSING PACKAGES"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  First, let's see what packages are available."))
	b.WriteString("\n\n")
	b.WriteString(theme.HighlightStyle.Render("  Command:"))
	b.WriteString("\n\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime packages"))
	b.WriteString("\n\n")
	b.WriteString(theme.DimTextStyle.Render("  This displays:"))
	b.WriteString("\n")
	b.WriteString("    • Core packages (CUDA, Python, Docker)\n")
	b.WriteString("    • AI frameworks (PyTorch, TensorFlow)\n")
	b.WriteString("    • Tools (Ollama, ComfyUI, Claude Code)\n")
	b.WriteString("    • Models (LLMs, video generation)\n")
	b.WriteString("\n")
	b.WriteString(theme.GlowStyle.Render("  Let's see the package list..."))

	return b.String()
}

func (m tutorialModel) renderAddServerDemo() string {
	var b strings.Builder

	progress := int(m.demoProgress * 100)

	b.WriteString(theme.RenderBanner("📦 PACKAGE BROWSER"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.spinner.View() + " " + theme.InfoStyle.Render("Loading package catalog..."))
	b.WriteString("\n\n")

	if progress > 20 {
		b.WriteString(theme.CategoryStyle("🏗️  Core System:"))
		b.WriteString("\n")
		b.WriteString("  • core - CUDA, Python 3.11, Docker\n")
		b.WriteString("  • python - Python dev tools\n")
		b.WriteString("\n")
	}
	if progress > 40 {
		b.WriteString(theme.CategoryStyle("🤖 AI Frameworks:"))
		b.WriteString("\n")
		b.WriteString("  • pytorch - PyTorch 2.2.0 + AI libraries\n")
		b.WriteString("  • ollama - Ollama LLM server\n")
		b.WriteString("\n")
	}
	if progress > 60 {
		b.WriteString(theme.CategoryStyle("🎨 Creative Tools:"))
		b.WriteString("\n")
		b.WriteString("  • comfyui - Stable Diffusion UI\n")
		b.WriteString("  • mochi - Video generation model\n")
		b.WriteString("\n")
	}
	if progress > 80 {
		b.WriteString(theme.CategoryStyle("💻 Development:"))
		b.WriteString("\n")
		b.WriteString("  • claude - Claude Code CLI\n")
		b.WriteString("\n")
		b.WriteString(theme.GlowStyle.Render("  ✨ 47 packages available!"))
	}

	// Progress bar
	b.WriteString("\n\n  ")
	barWidth := 40
	filled := int(float64(barWidth) * m.demoProgress)
	for i := 0; i < barWidth; i++ {
		if i < filled {
			b.WriteString(theme.SuccessStyle.Render("█"))
		} else {
			b.WriteString(theme.DimTextStyle.Render("░"))
		}
	}
	b.WriteString(fmt.Sprintf(" %d%%", progress))

	return b.String()
}

func (m tutorialModel) renderPackages() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("2️⃣ INSTALLING PACKAGES"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  Now let's install some packages."))
	b.WriteString("\n\n")
	b.WriteString(theme.HighlightStyle.Render("  Command:"))
	b.WriteString("\n\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime install core pytorch ollama"))
	b.WriteString("\n\n")
	b.WriteString(theme.DimTextStyle.Render("  Features:"))
	b.WriteString("\n\n")
	b.WriteString("    • Automatic dependency resolution\n")
	b.WriteString("    • Real-time progress tracking\n")
	b.WriteString("    • Installation time estimates\n")
	b.WriteString("    • Error detection & rollback\n")
	b.WriteString("\n")
	b.WriteString(theme.WarningStyle.Render("  💡 Tip: Use 'anime interactive' for a visual package selector!"))
	b.WriteString("\n\n")
	b.WriteString(theme.GlowStyle.Render("  Let's install some packages..."))

	return b.String()
}

func (m tutorialModel) renderPackagesDemo() string {
	var b strings.Builder

	progress := int(m.demoProgress * 100)

	b.WriteString(theme.RenderBanner("⚙️ INSTALLING"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.spinner.View() + " " + theme.InfoStyle.Render("Installing packages..."))
	b.WriteString("\n\n")

	b.WriteString("  $ " + theme.HighlightStyle.Render("anime install core pytorch ollama"))
	b.WriteString("\n\n")

	if progress > 20 {
		b.WriteString("  " + theme.InfoStyle.Render("▶") + " Resolving dependencies...\n")
	}
	if progress > 35 {
		b.WriteString("  " + theme.SuccessStyle.Render("✓") + " core → python → pytorch → ollama\n")
		b.WriteString("\n")
	}
	if progress > 50 {
		b.WriteString("  " + theme.InfoStyle.Render("▶") + " Installing core (5m)\n")
		b.WriteString("    Installing CUDA 12.4...\n")
	}
	if progress > 70 {
		b.WriteString("  " + theme.SuccessStyle.Render("✓") + " core complete\n")
		b.WriteString("  " + theme.InfoStyle.Render("▶") + " Installing pytorch (2m)\n")
	}
	if progress > 90 {
		b.WriteString("  " + theme.SuccessStyle.Render("✓") + " pytorch complete\n")
		b.WriteString("  " + theme.SuccessStyle.Render("✓") + " ollama complete\n")
		b.WriteString("\n")
		b.WriteString(theme.GlowStyle.Render("  ✨ Installation complete! 3 packages installed"))
	}

	b.WriteString("\n\n  ")
	barWidth := 40
	filled := int(float64(barWidth) * m.demoProgress)
	for i := 0; i < barWidth; i++ {
		if i < filled {
			b.WriteString(theme.SuccessStyle.Render("█"))
		} else {
			b.WriteString(theme.DimTextStyle.Render("░"))
		}
	}
	b.WriteString(fmt.Sprintf(" %d%%", progress))

	return b.String()
}

func (m tutorialModel) renderInstall() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("3️⃣ INTERACTIVE MODE"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  anime also has an interactive package selector!"))
	b.WriteString("\n\n")
	b.WriteString(theme.HighlightStyle.Render("  Command:"))
	b.WriteString("\n\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime interactive"))
	b.WriteString("\n\n")
	b.WriteString(theme.DimTextStyle.Render("  Features:"))
	b.WriteString("\n\n")
	b.WriteString("    • Visual package selection with checkboxes\n")
	b.WriteString("    • Category-based browsing\n")
	b.WriteString("    • Real-time cost/time estimates\n")
	b.WriteString("    • Keyboard navigation (↑↓ + Space)\n")
	b.WriteString("\n")
	b.WriteString(theme.GlowStyle.Render("  Let's see it in action..."))

	return b.String()
}

func (m tutorialModel) renderInstallDemo() string {
	var b strings.Builder

	progress := int(m.demoProgress * 100)

	b.WriteString(theme.RenderBanner("🎯 INTERACTIVE MODE"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.spinner.View() + " " + theme.InfoStyle.Render("Loading interactive selector..."))
	b.WriteString("\n\n")

	if progress > 20 {
		b.WriteString(theme.CategoryStyle("📦 Core System"))
		b.WriteString("\n")
		b.WriteString("  " + theme.SuccessStyle.Render("[✓]") + " core - CUDA, Python, Docker (5m)\n")
		b.WriteString("  " + theme.DimTextStyle.Render("[ ]") + " python - Python development tools (2m)\n")
		b.WriteString("\n")
	}
	if progress > 40 {
		b.WriteString(theme.CategoryStyle("🤖 AI Frameworks"))
		b.WriteString("\n")
		b.WriteString("  " + theme.SuccessStyle.Render("[✓]") + " pytorch - PyTorch + AI libs (2m)\n")
		b.WriteString("  " + theme.SuccessStyle.Render("[✓]") + " ollama - LLM server (1m)\n")
		b.WriteString("\n")
	}
	if progress > 60 {
		b.WriteString(theme.CategoryStyle("🎨 Creative Tools"))
		b.WriteString("\n")
		b.WriteString("  " + theme.DimTextStyle.Render("[ ]") + " comfyui - Stable Diffusion UI (2m)\n")
		b.WriteString("  " + theme.DimTextStyle.Render("[ ]") + " mochi - Video generation (30m)\n")
		b.WriteString("\n")
	}
	if progress > 80 {
		b.WriteString(theme.InfoStyle.Render("  Selected: 3 packages"))
		b.WriteString("\n")
		b.WriteString(theme.InfoStyle.Render("  Est. time: 8 minutes"))
		b.WriteString("\n\n")
		b.WriteString(theme.GlowStyle.Render("  ✨ Press Enter to install or Esc to cancel"))
	}

	return b.String()
}

func (m tutorialModel) renderStatus() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("4️⃣ DIAGNOSTICS"))
	b.WriteString("\n\n")
	b.WriteString(theme.InfoStyle.Render("  Check your system and diagnose issues."))
	b.WriteString("\n\n")
	b.WriteString(theme.HighlightStyle.Render("  Commands:"))
	b.WriteString("\n\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime models") + " - List downloaded AI models\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime doctor") + " - Diagnose installation issues\n")
	b.WriteString("  " + theme.SuccessStyle.Render("anime tree") + " - View all available commands\n")
	b.WriteString("\n")
	b.WriteString(theme.DimTextStyle.Render("  These help you:"))
	b.WriteString("\n\n")
	b.WriteString("    • Verify installations\n")
	b.WriteString("    • Troubleshoot problems\n")
	b.WriteString("    • Check system resources\n")
	b.WriteString("    • Find missing dependencies\n")
	b.WriteString("\n")
	b.WriteString(theme.GlowStyle.Render("  Let's run the doctor..."))

	return b.String()
}

func (m tutorialModel) renderStatusDemo() string {
	var b strings.Builder

	progress := int(m.demoProgress * 100)

	b.WriteString(theme.RenderBanner("🏥 SYSTEM DIAGNOSTICS"))
	b.WriteString("\n\n")

	if progress < 30 {
		b.WriteString("  " + m.spinner.View() + " " + theme.InfoStyle.Render("Running diagnostics..."))
	} else {
		b.WriteString("  " + theme.SuccessStyle.Render("✓") + " " + theme.InfoStyle.Render("System check complete"))
	}

	b.WriteString("\n\n")

	if progress > 30 {
		b.WriteString(theme.HighlightStyle.Render("  System Information:"))
		b.WriteString("\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " OS: Ubuntu 22.04.3 LTS\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " GPU: NVIDIA GH200 480GB\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " CUDA: 12.4\n")
		b.WriteString("\n")
	}

	if progress > 55 {
		b.WriteString(theme.HighlightStyle.Render("  Installed Packages:"))
		b.WriteString("\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " core (Python 3.11.8, Docker 24.0.7)\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " pytorch (PyTorch 2.2.0+cu124)\n")
		b.WriteString("    " + theme.SuccessStyle.Render("✓") + " ollama (Ollama 0.1.29)\n")
		b.WriteString("\n")
	}

	if progress > 75 {
		b.WriteString(theme.HighlightStyle.Render("  Running Services:"))
		b.WriteString("\n")
		b.WriteString("    " + theme.SuccessStyle.Render("●") + " ollama (port 11434)\n")
		b.WriteString("    " + theme.SuccessStyle.Render("●") + " docker daemon\n")
		b.WriteString("\n")
	}

	if progress > 90 {
		b.WriteString(theme.GlowStyle.Render("  ✨ System is healthy! No issues detected."))
	}

	return b.String()
}

func (m tutorialModel) renderComplete() string {
	var b strings.Builder

	b.WriteString(theme.RenderBanner("✨ TUTORIAL COMPLETE"))
	b.WriteString("\n\n")
	b.WriteString(theme.GlowStyle.Render("  🎉 Congratulations! You've learned the basics of anime!"))
	b.WriteString("\n\n")

	// Suggested next actions
	b.WriteString(theme.CategoryStyle("🚀 Suggested Next Actions:"))
	b.WriteString("\n\n")

	actions := []struct {
		emoji string
		cmd   string
		desc  string
	}{
		{"1️⃣", "anime install core pytorch ollama", "Get started with AI/ML basics"},
		{"2️⃣", "anime run comfyui", "Launch ComfyUI (after installing)"},
		{"3️⃣", "anime run ollama", "Start Ollama LLM server"},
		{"4️⃣", "anime wizard", "Configure your node with guided setup"},
		{"5️⃣", "anime interactive", "Browse and select packages visually"},
	}

	for _, action := range actions {
		b.WriteString(fmt.Sprintf("  %s  %s\n", action.emoji, theme.SuccessStyle.Render(action.cmd)))
		b.WriteString(fmt.Sprintf("      %s\n\n", theme.DimTextStyle.Render(action.desc)))
	}

	b.WriteString(theme.CategoryStyle("📚 Quick Reference:"))
	b.WriteString("\n\n")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"anime packages", "Browse available packages"},
		{"anime install <id>", "Install packages locally"},
		{"anime run <service>", "Start ComfyUI, Ollama, Jupyter"},
		{"anime models", "List downloaded AI models"},
		{"anime doctor", "Diagnose installation issues"},
		{"anime tree", "View all available commands"},
	}

	for _, cmd := range commands {
		b.WriteString(fmt.Sprintf("  %s - %s\n",
			theme.HighlightStyle.Render(cmd.cmd),
			theme.DimTextStyle.Render(cmd.desc)))
	}

	b.WriteString("\n")
	b.WriteString(theme.InfoStyle.Render("  💡 Pro Tips:"))
	b.WriteString("\n\n")
	b.WriteString("    • Use " + theme.HighlightStyle.Render("anime run comfyui") + " instead of cd ~/ComfyUI && python main.py\n")
	b.WriteString("    • Use " + theme.HighlightStyle.Render("anime ollama run llama2") + " for quick LLM access\n")
	b.WriteString("    • Run " + theme.HighlightStyle.Render("anime doctor") + " if you encounter issues\n")
	b.WriteString("\n")
	b.WriteString(theme.GlowStyle.Render("  Happy hacking! 🚀"))

	return b.String()
}

func init() {
	rootCmd.AddCommand(tutorialCmd)
}
