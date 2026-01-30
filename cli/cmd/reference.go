package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var referenceCmd = &cobra.Command{
	Use:   "reference",
	Short: "Interactive CLI reference explorer",
	Long: `Launch an interactive TUI to explore the Anime CLI.

Navigate with arrow keys, Enter to select, Esc/q to go back.

Features:
  - Browse all commands by category
  - View detailed help for each command
  - See usage examples
  - Copy commands to clipboard`,
	Run: runReference,
}

func init() {
	rootCmd.AddCommand(referenceCmd)
}

// Styles
var (
	refTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF69B4")).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 2).
			MarginBottom(1)

	refCategoryStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00D9FF"))

	refCommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9"))

	refDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	refSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#50FA7B"))

	refHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)

	refDetailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(1, 2).
			MarginTop(1)

	refExampleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C"))
)

// Reference item types
type refCategory struct {
	name     string
	commands []refCommand
}

type refCommand struct {
	name     string
	short    string
	usage    string
	examples []string
	flags    []string
}

// List item implementation
type refItem struct {
	title       string
	description string
	category    *refCategory
	command     *refCommand
	isCategory  bool
}

func (i refItem) Title() string       { return i.title }
func (i refItem) Description() string { return i.description }
func (i refItem) FilterValue() string { return i.title }

// Model
type refModel struct {
	list        list.Model
	categories  []refCategory
	currentCat  *refCategory
	currentCmd  *refCommand
	viewing     string // "categories", "commands", "detail"
	width       int
	height      int
	quitting    bool
}

func initialRefModel() refModel {
	categories := getRefCategories()

	items := make([]list.Item, len(categories))
	for i, cat := range categories {
		items[i] = refItem{
			title:       cat.name,
			description: fmt.Sprintf("%d commands", len(cat.commands)),
			category:    &categories[i],
			isCategory:  true,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = refSelectedStyle
	delegate.Styles.SelectedDesc = refDescStyle

	l := list.New(items, delegate, 60, 20)
	l.Title = "Anime CLI Reference"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = refTitleStyle

	return refModel{
		list:       l,
		categories: categories,
		viewing:    "categories",
	}
}

func (m refModel) Init() tea.Cmd {
	return nil
}

func (m refModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-6)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			if m.viewing == "categories" {
				m.quitting = true
				return m, tea.Quit
			}
			return m.goBack(), nil

		case "esc", "backspace":
			return m.goBack(), nil

		case "enter":
			return m.selectItem(), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m refModel) goBack() refModel {
	switch m.viewing {
	case "detail":
		m.viewing = "commands"
		m.currentCmd = nil
	case "commands":
		m.viewing = "categories"
		m.currentCat = nil
		m.list.Title = "Anime CLI Reference"
		items := make([]list.Item, len(m.categories))
		for i, cat := range m.categories {
			items[i] = refItem{
				title:       cat.name,
				description: fmt.Sprintf("%d commands", len(cat.commands)),
				category:    &m.categories[i],
				isCategory:  true,
			}
		}
		m.list.SetItems(items)
	}
	return m
}

func (m refModel) selectItem() refModel {
	selected := m.list.SelectedItem()
	if selected == nil {
		return m
	}

	item := selected.(refItem)

	if item.isCategory {
		m.currentCat = item.category
		m.viewing = "commands"
		m.list.Title = item.category.name

		items := make([]list.Item, len(item.category.commands))
		for i, cmd := range item.category.commands {
			items[i] = refItem{
				title:       cmd.name,
				description: cmd.short,
				command:     &item.category.commands[i],
				isCategory:  false,
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
	} else if item.command != nil {
		m.currentCmd = item.command
		m.viewing = "detail"
	}

	return m
}

func (m refModel) View() string {
	if m.quitting {
		return ""
	}

	if m.viewing == "detail" && m.currentCmd != nil {
		return m.detailView()
	}

	help := refHelpStyle.Render("↑/↓ Navigate • Enter Select • Esc Back • q Quit • / Filter")
	return m.list.View() + "\n" + help
}

func (m refModel) detailView() string {
	cmd := m.currentCmd

	var b strings.Builder

	// Title
	b.WriteString(refTitleStyle.Render(cmd.name))
	b.WriteString("\n\n")

	// Description
	b.WriteString(refCategoryStyle.Render("Description"))
	b.WriteString("\n")
	b.WriteString(cmd.short)
	b.WriteString("\n\n")

	// Usage
	b.WriteString(refCategoryStyle.Render("Usage"))
	b.WriteString("\n")
	b.WriteString(refCommandStyle.Render(cmd.usage))
	b.WriteString("\n\n")

	// Examples
	if len(cmd.examples) > 0 {
		b.WriteString(refCategoryStyle.Render("Examples"))
		b.WriteString("\n")
		for _, ex := range cmd.examples {
			b.WriteString(refExampleStyle.Render("  " + ex))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Flags
	if len(cmd.flags) > 0 {
		b.WriteString(refCategoryStyle.Render("Flags"))
		b.WriteString("\n")
		for _, f := range cmd.flags {
			b.WriteString(refDescStyle.Render("  " + f))
			b.WriteString("\n")
		}
	}

	content := refDetailStyle.Render(b.String())
	help := refHelpStyle.Render("Esc/Backspace Go back • q Quit")

	return content + "\n" + help
}

func runReference(cmd *cobra.Command, args []string) {
	p := tea.NewProgram(initialRefModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// Key bindings
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Filter key.Binding
}

func getRefCategories() []refCategory {
	return []refCategory{
		{
			name: "Installer",
			commands: []refCommand{
				{
					name:  "packages",
					short: "Browse available packages",
					usage: "anime packages",
					examples: []string{
						"anime packages",
					},
				},
				{
					name:  "interactive",
					short: "Interactive package selector",
					usage: "anime interactive",
					examples: []string{
						"anime interactive",
					},
				},
				{
					name:  "install",
					short: "Install packages with dependency resolution",
					usage: "anime install [packages...] [flags]",
					examples: []string{
						"anime install core",
						"anime install python pytorch",
						"anime install -y claude",
						"anime install --dry-run comfyui",
						"anime install -r -s alice python",
					},
					flags: []string{
						"-y, --yes        Skip confirmation",
						"--dry-run        Preview installation",
						"-r, --remote     Install on remote server",
						"-s, --server     Server name for remote install",
						"--phased         Install with phase confirmations",
					},
				},
			},
		},
		{
			name: "Source Control",
			commands: []refCommand{
				{
					name:  "source push",
					short: "Push local changes to remote",
					usage: "anime source push [path]",
					examples: []string{
						"anime source push",
						"anime source push myproject",
						"anime source push -n",
					},
					flags: []string{
						"-n, --dry-run    Preview without changes",
						"-s, --server     Override default server",
					},
				},
				{
					name:  "source pull",
					short: "Pull remote changes to local",
					usage: "anime source pull [path]",
					examples: []string{
						"anime source pull",
						"anime source pull org/project",
					},
					flags: []string{
						"-n, --dry-run    Preview without changes",
						"-s, --server     Override default server",
					},
				},
				{
					name:  "source clone",
					short: "Clone remote repo into new folder",
					usage: "anime source clone <path>",
					examples: []string{
						"anime source clone myproject",
						"anime source clone org/project",
						"anime source clone -f org/project",
					},
					flags: []string{
						"-f, --force      Force overwrite",
						"-n, --dry-run    Preview without changes",
					},
				},
				{
					name:  "source status",
					short: "Show sync status between local and remote",
					usage: "anime source status [path]",
					examples: []string{
						"anime source status",
					},
				},
				{
					name:  "source sync",
					short: "Bidirectional sync (newer files win)",
					usage: "anime source sync [path]",
					examples: []string{
						"anime source sync",
						"anime source sync -n",
					},
				},
				{
					name:  "source link",
					short: "Link current directory to remote path",
					usage: "anime source link <path>",
					examples: []string{
						"anime source link myproject",
						"anime source link org/project",
					},
				},
				{
					name:  "source init",
					short: "Initialize and push new repo",
					usage: "anime source init <name>",
					examples: []string{
						"anime source init myproject",
					},
				},
				{
					name:  "source list",
					short: "List remote repositories",
					usage: "anime source list [path]",
					examples: []string{
						"anime source list",
						"anime source list org",
					},
				},
				{
					name:  "source tree",
					short: "Tree view of remote repos",
					usage: "anime source tree [path]",
					examples: []string{
						"anime source tree",
					},
				},
				{
					name:  "source history",
					short: "Show push/pull history",
					usage: "anime source history <path>",
					examples: []string{
						"anime source history myproject",
					},
				},
				{
					name:  "source rename",
					short: "Rename/move remote repo",
					usage: "anime source rename <old> <new>",
					examples: []string{
						"anime source rename oldname newname",
						"anime source rename proj org/proj",
					},
				},
				{
					name:  "source delete",
					short: "Delete remote repo",
					usage: "anime source delete <path>",
					examples: []string{
						"anime source delete myproject",
					},
				},
			},
		},
		{
			name: "Package Manager",
			commands: []refCommand{
				{
					name:  "pkg init",
					short: "Create a new cpm.json file",
					usage: "anime pkg init [name]",
					examples: []string{
						"anime pkg init",
						"anime pkg init mypackage",
					},
				},
				{
					name:  "pkg publish",
					short: "Publish package to registry",
					usage: "anime pkg publish",
					examples: []string{
						"anime pkg publish",
						"anime pkg publish -n",
					},
					flags: []string{
						"-n, --dry-run    Preview without publishing",
						"-f, --force      Force overwrite existing version",
					},
				},
				{
					name:  "pkg republish",
					short: "Update published version in place",
					usage: "anime pkg republish",
					examples: []string{
						"anime pkg republish",
					},
				},
				{
					name:  "pkg install",
					short: "Install a package",
					usage: "anime pkg install <package[@version]>",
					examples: []string{
						"anime pkg install mypackage",
						"anime pkg install mypackage@1.0.0",
						"anime pkg install -g mypackage",
						"anime pkg install -f mypackage",
					},
					flags: []string{
						"-g, --global     Install globally",
						"-f, --force      Force overwrite",
					},
				},
				{
					name:  "pkg uninstall",
					short: "Remove installed package",
					usage: "anime pkg uninstall <package>",
					examples: []string{
						"anime pkg uninstall mypackage",
						"anime pkg uninstall -g mypackage",
					},
					flags: []string{
						"-g, --global     Uninstall global package",
					},
				},
				{
					name:  "pkg search",
					short: "Search for packages",
					usage: "anime pkg search <query>",
					examples: []string{
						"anime pkg search utils",
						"anime pkg search json parser",
					},
				},
				{
					name:  "pkg info",
					short: "Show package information",
					usage: "anime pkg info <package[@version]>",
					examples: []string{
						"anime pkg info mypackage",
						"anime pkg info mypackage@1.0.0",
					},
				},
				{
					name:  "pkg versions",
					short: "List available versions",
					usage: "anime pkg versions <package>",
					examples: []string{
						"anime pkg versions mypackage",
					},
				},
				{
					name:  "pkg update",
					short: "Update installed packages",
					usage: "anime pkg update [package]",
					examples: []string{
						"anime pkg update",
						"anime pkg update mypackage",
						"anime pkg update -g",
					},
					flags: []string{
						"-g, --global     Update global packages",
					},
				},
				{
					name:  "pkg list",
					short: "List installed packages",
					usage: "anime pkg list",
					examples: []string{
						"anime pkg list",
						"anime pkg list -g",
					},
					flags: []string{
						"-g, --global     List global packages",
					},
				},
			},
		},
		{
			name: "Server Management",
			commands: []refCommand{
				{
					name:  "add",
					short: "Add a server",
					usage: "anime add <name> <ip>",
					examples: []string{
						"anime add alice 192.168.1.100",
					},
				},
				{
					name:  "set",
					short: "Set/update server (auto-detect IP)",
					usage: "anime set <name> [ip]",
					examples: []string{
						"anime set alice",
						"anime set alice 192.168.1.101",
					},
				},
				{
					name:  "status",
					short: "Show server status",
					usage: "anime status [server]",
					examples: []string{
						"anime status",
						"anime status alice",
					},
				},
				{
					name:  "remove",
					short: "Remove a server",
					usage: "anime remove <name>",
					examples: []string{
						"anime remove alice",
					},
				},
				{
					name:  "ssh",
					short: "SSH into server",
					usage: "anime ssh <name> [command]",
					examples: []string{
						"anime ssh alice",
						"anime ssh alice \"nvidia-smi\"",
					},
				},
				{
					name:  "deploy",
					short: "Deploy to server",
					usage: "anime deploy",
					examples: []string{
						"anime deploy",
					},
				},
			},
		},
		{
			name: "LLM & AI",
			commands: []refCommand{
				{
					name:  "query",
					short: "Query Ollama models",
					usage: "anime query <model> \"<prompt>\"",
					examples: []string{
						"anime query llama3 \"Hello\"",
						"anime query deepseek-coder \"Write fizzbuzz\"",
					},
				},
				{
					name:  "prompt",
					short: "AI-interpreted natural language commands",
					usage: "anime prompt \"<natural language>\"",
					examples: []string{
						"anime prompt \"install python and pytorch\"",
						"anime prompt \"show server status\"",
					},
				},
				{
					name:  "models",
					short: "List downloaded models",
					usage: "anime models",
					examples: []string{
						"anime models",
					},
				},
			},
		},
		{
			name: "Help & Documentation",
			commands: []refCommand{
				{
					name:  "docs",
					short: "Display comprehensive documentation",
					usage: "anime docs [section]",
					examples: []string{
						"anime docs",
						"anime docs installer",
						"anime docs source",
						"anime docs packages",
						"anime docs all",
					},
				},
				{
					name:  "usage",
					short: "Quick usage examples",
					usage: "anime usage [command]",
					examples: []string{
						"anime usage",
						"anime usage install",
						"anime usage source",
						"anime usage all",
					},
				},
				{
					name:  "reference",
					short: "Interactive CLI explorer (this)",
					usage: "anime reference",
					examples: []string{
						"anime reference",
					},
				},
				{
					name:  "tree",
					short: "Full command tree",
					usage: "anime tree",
					examples: []string{
						"anime tree",
					},
				},
			},
		},
		{
			name: "Configuration",
			commands: []refCommand{
				{
					name:  "config",
					short: "Show/set configuration",
					usage: "anime config [set|get] [key] [value]",
					examples: []string{
						"anime config",
						"anime config set server alice",
						"anime config get server",
					},
				},
				{
					name:  "doctor",
					short: "Run diagnostics",
					usage: "anime doctor",
					examples: []string{
						"anime doctor",
						"anime doctor --check",
					},
				},
			},
		},
		{
			name: "Legacy (CPM)",
			commands: []refCommand{
				{
					name:  "cpm",
					short: "Legacy CPM commands (backwards compatible)",
					usage: "anime cpm <command>",
					examples: []string{
						"anime cpm push",
						"anime cpm pull",
						"anime cpm publish",
						"anime cpm install mypackage",
					},
				},
			},
		},
	}
}
