package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/joshkornreich/anime/internal/gpu"
)

// Wan presets known to the TUI. Order matters: the 'p' key cycles through
// them in this order, smallest VRAM first. Kept in sync with PRESETS in
// embedded/wan-pipeline/wan.py — if you add a preset there, add it here too.
var wanTUIPresets = []string{
	"ti2v-5b",            // ≥12GB VRAM
	"t2v-14b-dual-fast",  // ≥24GB VRAM (default)
	"t2v-14b-dual-maxq",  // ≥48GB VRAM
}

// ─── styling ───
var (
	wanTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	wanAccentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	wanDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wanGoodStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	wanWarnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	wanBadStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	wanBorder      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("213")).Padding(0, 1)
)

func wanStatusStyle(status string) lipgloss.Style {
	switch status {
	case "done":
		return wanGoodStyle
	case "pending", "running":
		return wanWarnStyle
	default:
		return wanBadStyle
	}
}

// ─── model types ───
type wanRender struct {
	ID         int     `json:"id"`
	CreatedAt  string  `json:"created_at"`
	Name       string  `json:"name"`
	Preset     string  `json:"preset"`
	Status     string  `json:"status"`
	Seed       int64   `json:"seed"`
	RenderSecs float64 `json:"render_seconds"`
	FileSize   int64   `json:"file_size"`
	Rating     *int    `json:"rating"`
	Prompt     string  `json:"prompt"`
	OutputURL  string  `json:"output_url"`
}

func (r wanRender) Title() string {
	star := "·····"
	if r.Rating != nil {
		star = strings.Repeat("★", *r.Rating) + strings.Repeat("·", 5-*r.Rating)
	}
	return fmt.Sprintf("#%-3d %s %s %s",
		r.ID,
		wanStatusStyle(r.Status).Render(fmt.Sprintf("%-7s", r.Status)),
		wanAccentStyle.Render(fmt.Sprintf("%-22s", r.Preset)),
		star,
	)
}

func (r wanRender) Description() string {
	dur := "—"
	if r.RenderSecs > 0 {
		dur = fmt.Sprintf("%.0fs", r.RenderSecs)
	}
	sz := "—"
	if r.FileSize > 0 {
		sz = fmt.Sprintf("%.1fM", float64(r.FileSize)/1024/1024)
	}
	p := r.Prompt
	if len(p) > 70 {
		p = p[:67] + "..."
	}
	when := r.CreatedAt
	if len(when) > 16 {
		when = when[:16]
	}
	return wanDimStyle.Render(fmt.Sprintf("  %s · %s · %s · %s", when, dur, sz, p))
}

func (r wanRender) FilterValue() string { return r.Prompt + " " + r.Preset }

// ─── screens ───
type wanScreen int

const (
	scrList wanScreen = iota
	scrDetail
	scrPrompt
	scrPending
)

type wanTUIModel struct {
	scr        wanScreen
	list       list.Model
	input      textinput.Model
	spinner    spinner.Model
	width      int
	height     int
	selected   *wanRender
	pendingMsg string
	flash      string // transient status line shown above hints, cleared on next nav
	preset     string // active preset for the next render (cycled with 'p' / tab)
	gpuLabel   string // cached host blurb for the status bar (e.g. "GH200 · 95GB")
}

const (
	hintsList    = "↑/↓ select · enter detail · n new · p preset · v vary · r resume · 1-5 rate · / filter · q quit"
	hintsDetail  = "v vary · r resume · esc back · q quit"
	hintsPrompt  = "enter submit · tab cycle preset · esc cancel"
	hintsPending = "(detached — ctrl+c to leave; the render keeps going in ComfyUI)"
)

// cyclePreset returns the next preset in wanTUIPresets after `current`. Used
// by the 'p' key (list/detail) and 'tab' key (prompt) to step through.
func cyclePreset(current string) string {
	for i, p := range wanTUIPresets {
		if p == current {
			return wanTUIPresets[(i+1)%len(wanTUIPresets)]
		}
	}
	return wanTUIPresets[0]
}

// chrome is the lines around the list (title + status bar + hints + spacing).
// Bumped from 5 → 6 when the GPU/preset status bar was added.
const listChrome = 6

func newWanTUIModel() (*wanTUIModel, error) {
	items, _ := loadRenders()
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	l := list.New(items, delegate, 0, 0)
	l.Title = "wan-pipeline · history"
	l.Styles.Title = wanTitleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	ti := textinput.New()
	ti.Placeholder = "type a prompt and press enter to render…"
	ti.CharLimit = 1000
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = wanAccentStyle

	// Auto-pick the right preset for this host so a user pressing 'n' →
	// enter without thinking gets a render that fits their VRAM.
	g := gpu.GetSystemInfo()
	preset, _ := recommendedPreset(g.TotalVRAM)
	if preset == "" {
		preset = "t2v-14b-dual-fast" // pipeline default
	}
	gpuLabel := "no GPU"
	if g.Available && len(g.GPUs) > 0 {
		// "GH200 · 95GB" / "RTX 4090 · 24GB" / "1xH100 · 80GB"
		name := g.GPUs[0].Name
		// Trim "NVIDIA " prefix and 480GB-style memory suffix to keep the
		// status bar tight.
		name = strings.TrimPrefix(name, "NVIDIA ")
		if i := strings.Index(name, " 480GB"); i >= 0 {
			name = name[:i]
		}
		gpuLabel = fmt.Sprintf("%s · %dGB", name, g.TotalVRAM)
		if g.Count > 1 {
			gpuLabel = fmt.Sprintf("%dx %s · %dGB", g.Count, name, g.TotalVRAM)
		}
	}

	return &wanTUIModel{
		scr:      scrList,
		list:     l,
		input:    ti,
		spinner:  sp,
		preset:   preset,
		gpuLabel: gpuLabel,
	}, nil
}

func (m *wanTUIModel) Init() tea.Cmd { return m.spinner.Tick }

func (m *wanTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		w, h := msg.Width-4, msg.Height-listChrome
		if h < 5 {
			h = 5
		}
		m.list.SetSize(w, h)
		m.input.Width = msg.Width - 6

	case tea.KeyMsg:
		// Global quits (but not while typing a prompt — we want to allow 'q' in input)
		if m.scr != scrPrompt {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

		switch m.scr {
		case scrList:
			// Don't intercept keys while filter input is focused
			if m.list.FilterState() == list.Filtering {
				break
			}
			switch msg.String() {
			case "enter":
				if it, ok := m.list.SelectedItem().(wanRender); ok {
					m.selected = &it
					m.flash = ""
					m.scr = scrDetail
				}
			case "n":
				m.input.SetValue("")
				m.input.Focus()
				m.flash = ""
				m.scr = scrPrompt
			case "v":
				if it, ok := m.list.SelectedItem().(wanRender); ok {
					m.pendingMsg = fmt.Sprintf("queueing 1 variation of #%d…", it.ID)
					m.scr = scrPending
					cmds = append(cmds, doWanCmd("vary", fmt.Sprint(it.ID), "-n", "1"))
				}
			case "r":
				if it, ok := m.list.SelectedItem().(wanRender); ok {
					m.pendingMsg = fmt.Sprintf("resuming #%d (seed=%d)…", it.ID, it.Seed)
					m.scr = scrPending
					cmds = append(cmds, doWanCmd("resume", fmt.Sprint(it.ID)))
				}
			case "1", "2", "3", "4", "5":
				if it, ok := m.list.SelectedItem().(wanRender); ok {
					n := msg.String()
					_, _ = runWanCapture("rate", fmt.Sprint(it.ID), n)
					m.flash = wanGoodStyle.Render(fmt.Sprintf("rated #%d %s%s", it.ID, strings.Repeat("★", atoiSafe(n)), strings.Repeat("·", 5-atoiSafe(n))))
					cmds = append(cmds, refreshList())
				}
			case "p":
				m.preset = cyclePreset(m.preset)
				m.flash = wanAccentStyle.Render("preset → ") + m.preset
			case "ctrl+r":
				cmds = append(cmds, refreshList())
			}

		case scrDetail:
			switch msg.String() {
			case "esc", "backspace", "h":
				m.scr = scrList
				m.selected = nil
			case "v":
				if m.selected != nil {
					m.pendingMsg = fmt.Sprintf("queueing 1 variation of #%d…", m.selected.ID)
					m.scr = scrPending
					cmds = append(cmds, doWanCmd("vary", fmt.Sprint(m.selected.ID), "-n", "1"))
				}
			case "r":
				if m.selected != nil {
					m.pendingMsg = fmt.Sprintf("resuming #%d (seed=%d)…", m.selected.ID, m.selected.Seed)
					m.scr = scrPending
					cmds = append(cmds, doWanCmd("resume", fmt.Sprint(m.selected.ID)))
				}
			}

		case scrPrompt:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.scr = scrList
				m.input.Blur()
			case "tab":
				// In-input preset cycling so the user can pick before submitting.
				m.preset = cyclePreset(m.preset)
			case "enter":
				prompt := strings.TrimSpace(m.input.Value())
				if prompt == "" {
					m.scr = scrList
					m.input.Blur()
					break
				}
				m.pendingMsg = fmt.Sprintf("rendering [%s]: %.60s…", m.preset, prompt)
				m.scr = scrPending
				m.input.Blur()
				cmds = append(cmds, doWanCmd("render", prompt, "--preset", m.preset))
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}

		case scrPending:
			// only quit handled above
		}

	case wanCmdDoneMsg:
		m.scr = scrList
		m.flash = msg.summary
		cmds = append(cmds, refreshList())

	case wanRefreshDoneMsg:
		m.list.SetItems(msg.items)
		// keep selected pointer in sync if we're showing detail
		if m.selected != nil {
			for _, it := range msg.items {
				if r, ok := it.(wanRender); ok && r.ID == m.selected.ID {
					rr := r
					m.selected = &rr
					break
				}
			}
		}
	}

	if m.scr == scrList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.scr == scrPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.scr == scrPending {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *wanTUIModel) View() string {
	switch m.scr {
	case scrList:
		body := m.list.View()
		// flash > status bar > hints. Status bar shows GPU + active preset
		// so the user always knows which box they're on and what 'n' will
		// render with.
		statusBar := wanDimStyle.Render("│ ") +
			wanAccentStyle.Render("gpu ") + wanDimStyle.Render(m.gpuLabel) +
			wanDimStyle.Render("  │  ") +
			wanAccentStyle.Render("preset ") + wanGoodStyle.Render(m.preset)
		bottom := statusBar + "\n" + wanDimStyle.Render(hintsList)
		if m.flash != "" {
			bottom = m.flash + "\n" + statusBar + "\n" + wanDimStyle.Render(hintsList)
		}
		return body + "\n" + bottom

	case scrDetail:
		if m.selected == nil {
			return "(no selection)"
		}
		s := m.selected
		dur := "—"
		if s.RenderSecs > 0 {
			dur = fmt.Sprintf("%.0fs", s.RenderSecs)
		}
		sz := "—"
		if s.FileSize > 0 {
			sz = fmt.Sprintf("%.1fMB", float64(s.FileSize)/1024/1024)
		}
		rating := "—"
		if s.Rating != nil {
			rating = strings.Repeat("★", *s.Rating) + strings.Repeat("·", 5-*s.Rating)
		}
		urlLine := wanDimStyle.Render("(no URL)")
		if s.OutputURL != "" {
			urlLine = s.OutputURL
		}
		body := wanBorder.Render(strings.Join([]string{
			wanTitleStyle.Render(fmt.Sprintf("render #%d", s.ID)),
			"",
			wanAccentStyle.Render("preset    ") + s.Preset,
			wanAccentStyle.Render("seed      ") + fmt.Sprint(s.Seed),
			wanAccentStyle.Render("status    ") + wanStatusStyle(s.Status).Render(s.Status),
			wanAccentStyle.Render("when      ") + s.CreatedAt,
			wanAccentStyle.Render("duration  ") + dur,
			wanAccentStyle.Render("size      ") + sz,
			wanAccentStyle.Render("rating    ") + rating,
			"",
			wanAccentStyle.Render("url       ") + urlLine,
			"",
			wanAccentStyle.Render("prompt"),
			wrap(s.Prompt, m.viewWrapWidth()),
		}, "\n"))
		bottom := wanDimStyle.Render(hintsDetail)
		if m.flash != "" {
			bottom = m.flash + "\n" + bottom
		}
		return body + "\n" + bottom

	case scrPrompt:
		body := wanBorder.Render(strings.Join([]string{
			wanTitleStyle.Render("new render"),
			"",
			m.input.View(),
			"",
			wanAccentStyle.Render("preset ") + wanGoodStyle.Render(m.preset) +
				wanDimStyle.Render("   │   gpu ") + wanDimStyle.Render(m.gpuLabel),
			wanDimStyle.Render(hintsPrompt),
		}, "\n"))
		return body

	case scrPending:
		body := wanBorder.Render(strings.Join([]string{
			wanTitleStyle.Render("working"),
			"",
			m.spinner.View() + "  " + m.pendingMsg,
			"",
			wanDimStyle.Render(hintsPending),
		}, "\n"))
		return body
	}
	return ""
}

// viewWrapWidth is the width to wrap detail text — based on current terminal.
func (m *wanTUIModel) viewWrapWidth() int {
	w := m.width - 8 // border + padding
	if w < 40 {
		return 40
	}
	if w > 100 {
		return 100
	}
	return w
}

// ─── bubbletea cmds ───

type wanCmdDoneMsg struct{ summary string }
type wanRefreshDoneMsg struct{ items []list.Item }

func doWanCmd(args ...string) tea.Cmd {
	return func() tea.Msg {
		out, err := runWanCapture(args...)
		if err != nil {
			return wanCmdDoneMsg{summary: wanBadStyle.Render("✗ ") + wanTruncate(out, 80)}
		}
		summary := ""
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "✓ done") || strings.Contains(line, "url:") {
				summary = strings.TrimSpace(line)
				break
			}
		}
		if summary == "" {
			summary = "complete"
		}
		return wanCmdDoneMsg{summary: wanGoodStyle.Render("✓ ") + summary}
	}
}

func refreshList() tea.Cmd {
	// Run the (potentially slow) Python history fetch in the bubbletea goroutine,
	// not the main Update goroutine.
	return func() tea.Msg {
		items, _ := loadRenders()
		return wanRefreshDoneMsg{items: items}
	}
}

// ─── history fetch (delegates to wan.py history --json) ───

func loadRenders() ([]list.Item, error) {
	out, err := runWanCapture("history", "-n", "200", "--json")
	if err != nil {
		return []list.Item{}, nil
	}
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "[") {
		return []list.Item{}, nil
	}
	var rows []wanRender
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, err
	}
	items := make([]list.Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, r)
	}
	return items, nil
}

// ─── helpers ───

func wrap(s string, w int) string {
	if w <= 0 {
		return s
	}
	var out []string
	for len(s) > w {
		cut := w
		for i := w; i > w/2; i-- {
			if s[i] == ' ' {
				cut = i
				break
			}
		}
		out = append(out, s[:cut])
		s = strings.TrimLeft(s[cut:], " ")
	}
	if s != "" {
		out = append(out, s)
	}
	return strings.Join(out, "\n")
}

func wanTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}


// ─── entrypoint ───

func runWanTUI() error {
	if _, err := extractWanScript(); err != nil {
		return err
	}
	m, err := newWanTUIModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, runErr := p.Run()
	return runErr
}
