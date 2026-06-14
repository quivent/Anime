//go:build ignore

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshkornreich/anime/internal/theme"
)

// ─── styling ─────────────────────────────────────────────────────
// All styles derive from the project theme. No ad-hoc color numbers.

var (
	// Zone chrome
	tuiStatusBar = lipgloss.NewStyle().
			Background(theme.BgAccent).
			Foreground(theme.TextPrimary).
			Bold(true).
			Padding(0, 1)

	tuiTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.SakuraPink)

	tuiAccent = lipgloss.NewStyle().
			Foreground(theme.ElectricBlue)

	tuiDim = lipgloss.NewStyle().
		Foreground(theme.TextDim)

	tuiMuted = lipgloss.NewStyle().
			Foreground(theme.TextSecondary)

	tuiGood = lipgloss.NewStyle().
		Foreground(theme.MintGreen)

	tuiWarn = lipgloss.NewStyle().
		Foreground(theme.SunsetOrange)

	tuiBad = lipgloss.NewStyle().
		Foreground(theme.ActionRed).
		Bold(true)

	tuiStar = lipgloss.NewStyle().
		Foreground(theme.GoldYellow)

	tuiStarDim = lipgloss.NewStyle().
			Foreground(theme.TextDarkGray)

	tuiLabel = lipgloss.NewStyle().
			Foreground(theme.NeonPurple).
			Bold(true)

	tuiHelpBar = lipgloss.NewStyle().
			Foreground(theme.TextHelp)

	tuiFlash = lipgloss.NewStyle().
			Foreground(theme.TextPrimary).
			Background(theme.BgAccent).
			Padding(0, 1)

	tuiFlashError = lipgloss.NewStyle().
			Foreground(theme.ActionRed).
			Background(theme.BgAccent).
			Bold(true).
			Padding(0, 1)

	tuiBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.NeonPurple).
			Padding(0, 1)

	tuiSelectedRow = lipgloss.NewStyle().
			Foreground(theme.SakuraPink).
			Bold(true)

	tuiNormalRow = lipgloss.NewStyle().
			Foreground(theme.TextPrimary)

	tuiEmptyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.NeonPurple).
			Foreground(theme.TextSecondary).
			Padding(2, 4).
			Align(lipgloss.Center)
)

func tuiRenderStatusStyle(status string) lipgloss.Style {
	switch status {
	case "done":
		return tuiGood
	case "pending", "running":
		return tuiWarn
	case "cancelled":
		return tuiDim
	default:
		return tuiBad
	}
}

// ─── model types ─────────────────────────────────────────────────

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

func (r wanRender) Title() string       { return r.Name }
func (r wanRender) Description() string { return r.Prompt }
func (r wanRender) FilterValue() string { return r.Prompt + " " + r.Preset + " " + r.Name }

func (r wanRender) stars() string {
	if r.Rating == nil {
		return tuiStarDim.Render("-----")
	}
	n := *r.Rating
	return tuiStar.Render(strings.Repeat("*", n)) + tuiStarDim.Render(strings.Repeat("-", 5-n))
}

func (r wanRender) shortPrompt(maxLen int) string {
	p := r.Prompt
	if len(p) > maxLen {
		return p[:maxLen-3] + "..."
	}
	return p
}

func (r wanRender) durationStr() string {
	if r.RenderSecs <= 0 {
		return "--"
	}
	return fmt.Sprintf("%.0fs", r.RenderSecs)
}

func (r wanRender) sizeStr() string {
	if r.FileSize <= 0 {
		return "--"
	}
	return fmt.Sprintf("%.1fM", float64(r.FileSize)/1024/1024)
}

func (r wanRender) shortTime() string {
	if len(r.CreatedAt) > 16 {
		return r.CreatedAt[:16]
	}
	return r.CreatedAt
}

// ─── screens ─────────────────────────────────────────────────────

type wanScreen int

const (
	scrList wanScreen = iota
	scrDetail
	scrPrompt
	scrPending
	scrSystem
)

// ─── flash message with auto-dismiss ─────────────────────────────

type flashMsg struct {
	text    string
	isError bool
}

type flashExpiredMsg struct{}

const flashDuration = 3 * time.Second

func scheduleFlashDismiss() tea.Cmd {
	return tea.Tick(flashDuration, func(time.Time) tea.Msg {
		return flashExpiredMsg{}
	})
}

// ─── auto-refresh tick ───────────────────────────────────────────

type autoRefreshMsg struct{}

const autoRefreshInterval = 5 * time.Second

func scheduleAutoRefresh() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(time.Time) tea.Msg {
		return autoRefreshMsg{}
	})
}

// ─── the model ───────────────────────────────────────────────────

type wanTUIModel struct {
	scr     wanScreen
	renders []wanRender
	cursor  int // currently highlighted row in the list

	input   textinput.Model
	spinner spinner.Model

	width  int
	height int

	pendingMsg string
	flash      *flashMsg // nil = no flash

	// For filtering
	filterText  string
	isFiltering bool
}

func newWanTUIModel() (*wanTUIModel, error) {
	renders, _ := loadWanRenders()

	ti := textinput.New()
	ti.Placeholder = "type a prompt and press enter to render..."
	ti.CharLimit = 1000
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.SakuraPink)

	return &wanTUIModel{
		scr:     scrList,
		renders: renders,
		cursor:  0,
		input:   ti,
		spinner: sp,
	}, nil
}

func (m *wanTUIModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, scheduleAutoRefresh())
}

// ─── update ──────────────────────────────────────────────────────

func (m *wanTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = msg.Width - 8

	case flashExpiredMsg:
		m.flash = nil

	case autoRefreshMsg:
		// Only auto-refresh when on the list screen and there are pending renders
		hasPending := false
		for _, r := range m.renders {
			if r.Status == "pending" || r.Status == "running" {
				hasPending = true
				break
			}
		}
		if hasPending && m.scr == scrList {
			cmds = append(cmds, refreshWanList())
		}
		cmds = append(cmds, scheduleAutoRefresh())

	case tea.KeyMsg:
		// Global quit -- but not while typing
		if m.scr != scrPrompt && !m.isFiltering {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
		if m.scr == scrPrompt || m.isFiltering {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

		switch m.scr {
		case scrList:
			cmds = append(cmds, m.updateList(msg)...)
		case scrDetail:
			cmds = append(cmds, m.updateDetail(msg)...)
		case scrPrompt:
			cmds = append(cmds, m.updatePrompt(msg)...)
		case scrPending:
			// Only quit handled above; spinner ticks below
		}

	case wanCmdDoneMsg:
		m.scr = scrList
		if msg.summary != "" {
			isErr := strings.Contains(msg.summary, "error") || strings.Contains(msg.summary, "fail")
			m.flash = &flashMsg{text: msg.summary, isError: isErr}
			cmds = append(cmds, scheduleFlashDismiss())
		}
		cmds = append(cmds, refreshWanList())

	case wanRefreshDoneMsg:
		m.renders = msg.renders
		// Clamp cursor
		if m.cursor >= len(m.renders) {
			m.cursor = len(m.renders) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case sysRefreshDoneMsg:
		m.sysPhases = msg.phases

	case sysActionDoneMsg:
		m.sysAction = msg.summary
		cmds = append(cmds, refreshSysPhases())
	}

	// Spinner always ticks
	if m.scr == scrPending {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Text input updates when in prompt or filter mode
	if m.scr == scrPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ─── list screen key handling ────────────────────────────────────

func (m *wanTUIModel) updateList(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	visible := m.visibleRenders()

	switch msg.String() {
	case "j", "down":
		if m.cursor < len(visible)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		if len(visible) > 0 {
			m.cursor = len(visible) - 1
		}

	case "enter":
		if len(visible) > 0 && m.cursor < len(visible) {
			m.scr = scrDetail
		}

	case "r":
		m.input.SetValue("")
		m.input.Placeholder = "type a prompt and press enter to render..."
		m.input.Focus()
		m.scr = scrPrompt

	case "v":
		if sel := m.selectedRender(); sel != nil {
			m.pendingMsg = fmt.Sprintf("queueing variation of #%d...", sel.ID)
			m.scr = scrPending
			cmds = append(cmds, doWanCmd("vary", fmt.Sprint(sel.ID), "-n", "1"))
		}

	case "R":
		if sel := m.selectedRender(); sel != nil {
			m.pendingMsg = fmt.Sprintf("resuming #%d (seed=%d)...", sel.ID, sel.Seed)
			m.scr = scrPending
			cmds = append(cmds, doWanCmd("resume", fmt.Sprint(sel.ID)))
		}

	case "1", "2", "3", "4", "5":
		if sel := m.selectedRender(); sel != nil {
			n, _ := strconv.Atoi(msg.String())
			_, _ = runWanCapture("rate", fmt.Sprint(sel.ID), msg.String())
			// Update in-place immediately
			for i := range m.renders {
				if m.renders[i].ID == sel.ID {
					rating := n
					m.renders[i].Rating = &rating
					break
				}
			}
			starDisplay := tuiStar.Render(strings.Repeat("*", n)) + tuiStarDim.Render(strings.Repeat("-", 5-n))
			m.flash = &flashMsg{text: fmt.Sprintf("Rated #%d  %s", sel.ID, starDisplay)}
			cmds = append(cmds, scheduleFlashDismiss())
		}

	case "/":
		// Toggle filter mode
		m.isFiltering = true
		m.filterText = ""

	case "esc":
		if m.isFiltering {
			m.isFiltering = false
			m.filterText = ""
			m.cursor = 0
		}

	case "backspace":
		if m.isFiltering && len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.cursor = 0
		} else if m.isFiltering {
			m.isFiltering = false
			m.filterText = ""
			m.cursor = 0
		}

	default:
		if m.isFiltering && len(msg.String()) == 1 {
			m.filterText += msg.String()
			m.cursor = 0
		}
	}

	return cmds
}

// ─── detail screen key handling ──────────────────────────────────

func (m *wanTUIModel) updateDetail(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
	case "esc", "backspace", "left", "h":
		m.scr = scrList

	case "v":
		if sel := m.selectedRender(); sel != nil {
			m.pendingMsg = fmt.Sprintf("queueing variation of #%d...", sel.ID)
			m.scr = scrPending
			cmds = append(cmds, doWanCmd("vary", fmt.Sprint(sel.ID), "-n", "1"))
		}

	case "R":
		if sel := m.selectedRender(); sel != nil {
			m.pendingMsg = fmt.Sprintf("resuming #%d (seed=%d)...", sel.ID, sel.Seed)
			m.scr = scrPending
			cmds = append(cmds, doWanCmd("resume", fmt.Sprint(sel.ID)))
		}

	case "1", "2", "3", "4", "5":
		if sel := m.selectedRender(); sel != nil {
			n, _ := strconv.Atoi(msg.String())
			_, _ = runWanCapture("rate", fmt.Sprint(sel.ID), msg.String())
			for i := range m.renders {
				if m.renders[i].ID == sel.ID {
					rating := n
					m.renders[i].Rating = &rating
					break
				}
			}
			starDisplay := tuiStar.Render(strings.Repeat("*", n)) + tuiStarDim.Render(strings.Repeat("-", 5-n))
			m.flash = &flashMsg{text: fmt.Sprintf("Rated #%d  %s", sel.ID, starDisplay)}
			cmds = append(cmds, scheduleFlashDismiss())
		}
	}

	return cmds
}

// ─── prompt screen key handling ──────────────────────────────────

func (m *wanTUIModel) updatePrompt(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
	case "esc":
		m.scr = scrList
		m.input.Blur()
	case "enter":
		prompt := strings.TrimSpace(m.input.Value())
		if prompt == "" {
			m.scr = scrList
			m.input.Blur()
			return nil
		}
		m.input.Blur()
		m.pendingMsg = fmt.Sprintf("rendering: %.60s...", prompt)
		m.scr = scrPending

		// Insert a synthetic pending entry at the top of the list immediately
		synthetic := wanRender{
			ID:        -1, // placeholder
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			Name:      "new render",
			Preset:    "t2v-14b-dual-fast",
			Status:    "pending",
			Prompt:    prompt,
		}
		m.renders = append([]wanRender{synthetic}, m.renders...)
		m.cursor = 0

		cmds = append(cmds, doWanCmd("render", prompt))
	}

	return cmds
}

// ─── helpers for selection and filtering ──────────────────────────

func (m *wanTUIModel) visibleRenders() []wanRender {
	if !m.isFiltering || m.filterText == "" {
		return m.renders
	}
	needle := strings.ToLower(m.filterText)
	var out []wanRender
	for _, r := range m.renders {
		hay := strings.ToLower(r.Prompt + " " + r.Preset + " " + r.Name)
		if strings.Contains(hay, needle) {
			out = append(out, r)
		}
	}
	return out
}

func (m *wanTUIModel) selectedRender() *wanRender {
	visible := m.visibleRenders()
	if m.cursor >= 0 && m.cursor < len(visible) {
		r := visible[m.cursor]
		return &r
	}
	return nil
}

// ─── view ────────────────────────────────────────────────────────

func (m *wanTUIModel) View() string {
	switch m.scr {
	case scrList:
		return m.viewList()
	case scrDetail:
		return m.viewDetail()
	case scrPrompt:
		return m.viewPromptInput()
	case scrPending:
		return m.viewPending()
	}
	return ""
}

// viewList renders the zoned layout:
//   Zone 1: Status bar (top)
//   Zone 2: Flash message (if any)
//   Zone 3: Render list (middle, scrollable)
//   Zone 4: Command bar (bottom)
func (m *wanTUIModel) viewList() string {
	w := m.width
	if w < 20 {
		w = 80
	}

	var sections []string

	// ── Zone 1: Status bar ──
	pendingCount, doneCount, totalCount := 0, 0, len(m.renders)
	for _, r := range m.renders {
		switch r.Status {
		case "pending", "running":
			pendingCount++
		case "done":
			doneCount++
		}
	}
	statusLeft := tuiStatusBar.Render(" wan-pipeline ")
	var statusParts []string
	statusParts = append(statusParts, tuiMuted.Render(fmt.Sprintf("%d renders", totalCount)))
	if pendingCount > 0 {
		statusParts = append(statusParts, tuiWarn.Render(fmt.Sprintf("%d pending", pendingCount)))
	}
	if doneCount > 0 {
		statusParts = append(statusParts, tuiGood.Render(fmt.Sprintf("%d done", doneCount)))
	}
	statusRight := strings.Join(statusParts, tuiDim.Render(" | "))
	statusLine := statusLeft + "  " + statusRight
	if m.isFiltering {
		filterIndicator := tuiAccent.Render("/" + m.filterText)
		statusLine = statusLeft + "  " + filterIndicator + "  " + statusRight
	}
	sections = append(sections, statusLine)

	// ── Zone 2: Flash message ──
	if m.flash != nil {
		style := tuiFlash
		if m.flash.isError {
			style = tuiFlashError
		}
		sections = append(sections, style.Render(m.flash.text))
	}

	// ── Zone 3: Render list or empty state ──
	visible := m.visibleRenders()

	if len(visible) == 0 {
		if len(m.renders) == 0 {
			// True empty: first-time user
			emptyMsg := tuiEmptyBox.Width(minInt(60, w-4)).Render(
				tuiTitle.Render("Welcome to wan-pipeline") + "\n\n" +
					tuiMuted.Render("No renders yet.") + "\n" +
					tuiAccent.Render("Press 'r' to create your first."))
			// Center it vertically in available space
			availH := m.height - 4 // status + help bar + margins
			pad := (availH - strings.Count(emptyMsg, "\n") - 1) / 3
			if pad < 1 {
				pad = 1
			}
			sections = append(sections, strings.Repeat("\n", pad)+emptyMsg)
		} else {
			// Filtering with no results
			sections = append(sections, "\n"+tuiDim.Render("  No renders match '"+m.filterText+"'. Press esc to clear filter."))
		}
	} else {
		// Column header
		header := tuiLabel.Render(fmt.Sprintf("  %-4s %-8s %-22s %-6s %-6s %-7s  %s",
			"ID", "STATUS", "PRESET", "TIME", "SIZE", "RATING", "PROMPT"))
		sections = append(sections, header)

		// How many rows fit?
		usedLines := len(sections) + 2 // +2 for help bar + bottom margin
		availRows := m.height - usedLines
		if availRows < 3 {
			availRows = 3
		}

		// Scroll window
		scrollStart := 0
		if m.cursor >= availRows {
			scrollStart = m.cursor - availRows + 1
		}
		scrollEnd := scrollStart + availRows
		if scrollEnd > len(visible) {
			scrollEnd = len(visible)
		}

		promptWidth := w - 58 // space remaining after fixed columns
		if promptWidth < 10 {
			promptWidth = 10
		}
		if promptWidth > 80 {
			promptWidth = 80
		}

		for i := scrollStart; i < scrollEnd; i++ {
			r := visible[i]
			cursor := "  "
			rowStyle := tuiNormalRow
			if i == m.cursor {
				cursor = tuiSelectedRow.Render("> ")
				rowStyle = tuiSelectedRow
			}

			statusStr := tuiRenderStatusStyle(r.Status).Render(fmt.Sprintf("%-8s", r.Status))
			ratingStr := r.stars()
			promptStr := r.shortPrompt(promptWidth)
			if i != m.cursor {
				promptStr = tuiDim.Render(promptStr)
			}

			line := fmt.Sprintf("%s%s %s %-22s %-6s %-6s %s  %s",
				cursor,
				rowStyle.Render(fmt.Sprintf("%-4d", r.ID)),
				statusStr,
				tuiAccent.Render(truncStr(r.Preset, 22)),
				r.durationStr(),
				r.sizeStr(),
				ratingStr,
				promptStr,
			)
			sections = append(sections, line)
		}

		// Scroll indicator
		if len(visible) > availRows {
			pct := 0
			if len(visible) > 1 {
				pct = m.cursor * 100 / (len(visible) - 1)
			}
			scrollInfo := tuiDim.Render(fmt.Sprintf("  [%d/%d  %d%%]", m.cursor+1, len(visible), pct))
			sections = append(sections, scrollInfo)
		}
	}

	// ── Zone 4: Command bar (bottom) ──
	help := tuiHelpBar.Render("  j/k navigate  enter detail  r render  v vary  R resume  1-5 rate  / filter  q quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

// viewDetail renders the detail pane for the selected render.
func (m *wanTUIModel) viewDetail() string {
	sel := m.selectedRender()
	if sel == nil {
		return tuiDim.Render("(no selection)")
	}
	s := sel

	rating := tuiStarDim.Render("-----")
	if s.Rating != nil {
		n := *s.Rating
		rating = tuiStar.Render(strings.Repeat("*", n)) + tuiStarDim.Render(strings.Repeat("-", 5-n))
	}

	urlLine := tuiDim.Render("(not available)")
	if s.OutputURL != "" {
		urlLine = tuiAccent.Render(s.OutputURL)
	}

	wrapW := m.width - 10
	if wrapW < 40 {
		wrapW = 40
	}
	if wrapW > 100 {
		wrapW = 100
	}

	rows := []string{
		tuiTitle.Render(fmt.Sprintf("render #%d", s.ID)),
		"",
		tuiLabel.Render("  preset    ") + s.Preset,
		tuiLabel.Render("  seed      ") + fmt.Sprint(s.Seed),
		tuiLabel.Render("  status    ") + tuiRenderStatusStyle(s.Status).Render(s.Status),
		tuiLabel.Render("  when      ") + s.shortTime(),
		tuiLabel.Render("  duration  ") + s.durationStr(),
		tuiLabel.Render("  size      ") + s.sizeStr(),
		tuiLabel.Render("  rating    ") + rating,
		"",
		tuiLabel.Render("  url       ") + urlLine,
		"",
		tuiLabel.Render("  prompt"),
		"  " + wrap(s.Prompt, wrapW),
	}

	body := tuiBorder.Width(minInt(m.width-4, 100)).Render(strings.Join(rows, "\n"))

	var sections []string
	sections = append(sections, body)

	if m.flash != nil {
		style := tuiFlash
		if m.flash.isError {
			style = tuiFlashError
		}
		sections = append(sections, style.Render(m.flash.text))
	}

	help := tuiHelpBar.Render("  esc/h back  v vary  R resume  1-5 rate  q quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

// viewPromptInput renders the new-render prompt input.
func (m *wanTUIModel) viewPromptInput() string {
	rows := []string{
		tuiTitle.Render("new render"),
		"",
		m.input.View(),
		"",
		tuiDim.Render("default preset: t2v-14b-dual-fast"),
	}
	body := tuiBorder.Width(minInt(m.width-4, 90)).Render(strings.Join(rows, "\n"))
	help := tuiHelpBar.Render("  enter submit  esc cancel")
	return body + "\n" + help
}

// viewPending renders the in-progress spinner.
func (m *wanTUIModel) viewPending() string {
	rows := []string{
		tuiTitle.Render("working"),
		"",
		m.spinner.View() + "  " + m.pendingMsg,
		"",
		tuiDim.Render("Detached -- the render continues in ComfyUI. ctrl+c to leave."),
	}
	body := tuiBorder.Width(minInt(m.width-4, 70)).Render(strings.Join(rows, "\n"))
	return body
}

// ─── bubbletea cmds ──────────────────────────────────────────────

type wanCmdDoneMsg struct{ summary string }
type wanRefreshDoneMsg struct{ renders []wanRender }

func doWanCmd(args ...string) tea.Cmd {
	return func() tea.Msg {
		out, err := runWanCapture(args...)
		if err != nil {
			return wanCmdDoneMsg{summary: "error: " + wanTruncate(out, 80)}
		}
		summary := ""
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "done") || strings.Contains(line, "url:") {
				summary = strings.TrimSpace(line)
				break
			}
		}
		if summary == "" {
			summary = "complete"
		}
		return wanCmdDoneMsg{summary: summary}
	}
}

func refreshWanList() tea.Cmd {
	return func() tea.Msg {
		renders, _ := loadWanRenders()
		return wanRefreshDoneMsg{renders: renders}
	}
}

// ─── history fetch ───────────────────────────────────────────────

func loadWanRenders() ([]wanRender, error) {
	out, err := runWanCapture("history", "-n", "200", "--json")
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "[") {
		return nil, nil
	}
	var rows []wanRender
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ─── helpers ─────────────────────────────────────────────────────

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
	return s[:n-1] + "..."
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-2] + ".."
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── entrypoint ──────────────────────────────────────────────────

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
