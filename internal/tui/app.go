package tui

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dops/internal/adapters"
	"dops/internal/catalog"
	"dops/internal/config"
	"dops/internal/domain"
	"dops/internal/executor"
	"dops/internal/theme"
	"dops/internal/tui/footer"
	"dops/internal/tui/metadata"
	"dops/internal/tui/output"
	"dops/internal/tui/palette"
	"dops/internal/tui/sidebar"
	"dops/internal/tui/wizard"
	"dops/internal/vars"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type viewState int

const (
	stateNormal viewState = iota
	stateWizard
	statePalette
)

type executionResultMsg struct {
	Lines   []output.OutputLineMsg
	LogPath string
}

type AppDeps struct {
	Styles    *theme.Styles
	Store     config.ConfigStore
	Runner    executor.Runner
	LogWriter *adapters.LogWriter
	Config    *domain.Config
	Catalogs  []catalog.CatalogWithRunbooks
	AltScreen bool
}

type App struct {
	sidebar  sidebar.Model
	output   output.Model
	wizard   *wizard.Model
	pal      *palette.Model
	selected *domain.Runbook
	selCat   *domain.Catalog
	deps     AppDeps
	state    viewState
	width    int
	height   int
}

func NewApp(catalogs []catalog.CatalogWithRunbooks, styles *theme.Styles) App {
	return NewAppWithDeps(AppDeps{
		Styles:   styles,
		Catalogs: catalogs,
	})
}

func NewAppWithDeps(deps AppDeps) App {
	return App{
		sidebar: sidebar.New(deps.Catalogs, 20, deps.Styles),
		output:  output.New(60, 20, deps.Styles),
		deps:    deps,
		// width/height start at 0 — View() returns empty until WindowSizeMsg arrives
	}
}

func (m *App) SetConfig(cfg *domain.Config) {
	m.deps.Config = cfg
}

func (m App) Init() tea.Cmd {
	return m.sidebar.Init()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Recalculate sidebar dimensions for new terminal size
		panelRows := m.height - 2 // separator + footer
		sidebarContentH := panelRows - 2 // border top+bottom
		if sidebarContentH < 3 {
			sidebarContentH = 3
		}
		m.sidebar.SetHeight(sidebarContentH)
		return m, nil

	case tea.KeyPressMsg:
		if m.state == stateNormal {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "ctrl+shift+p":
				return m.openPalette()
			}
		}

	case sidebar.RunbookSelectedMsg:
		rb := msg.Runbook
		cat := msg.Catalog
		m.selected = &rb
		m.selCat = &cat
		m.output.Clear()
		return m, nil

	case sidebar.RunbookExecuteMsg:
		rb := msg.Runbook
		cat := msg.Catalog
		m.selected = &rb
		m.selCat = &cat
		return m.openWizard()

	case executionResultMsg:
		for _, line := range msg.Lines {
			m.output, _ = m.output.Update(line)
		}
		m.output, _ = m.output.Update(output.ExecutionDoneMsg{LogPath: msg.LogPath})
		return m, nil

	case output.OutputLineMsg:
		m.output, _ = m.output.Update(msg)
		return m, nil

	case output.ExecutionDoneMsg:
		m.output, _ = m.output.Update(msg)
		return m, nil

	case wizard.WizardSubmitMsg:
		m.state = stateNormal
		m.wizard = nil
		return m.startExecution(msg.Runbook, msg.Catalog, msg.Params)

	case wizard.WizardCancelMsg:
		m.state = stateNormal
		m.wizard = nil
		return m, nil

	case palette.PaletteSelectMsg:
		m.state = stateNormal
		m.pal = nil
		return m, nil

	case palette.PaletteCancelMsg:
		m.state = stateNormal
		m.pal = nil
		return m, nil
	}

	// Route to focused component
	switch m.state {
	case stateNormal:
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		return m, cmd

	case stateWizard:
		if m.wizard != nil {
			var cmd tea.Cmd
			wiz := *m.wizard
			wiz, cmd = wiz.Update(msg)
			m.wizard = &wiz
			return m, cmd
		}

	case statePalette:
		if m.pal != nil {
			var cmd tea.Cmd
			p := *m.pal
			p, cmd = p.Update(msg)
			m.pal = &p
			return m, cmd
		}
	}

	return m, nil
}

func (m App) startExecution(rb domain.Runbook, cat domain.Catalog, params map[string]string) (tea.Model, tea.Cmd) {
	m.output.Clear()
	cmdStr := wizard.BuildCommand(rb, params)
	m.output.SetCommand(cmdStr)

	if m.deps.Store != nil && m.deps.Config != nil {
		for _, p := range rb.Parameters {
			val, ok := params[p.Name]
			if !ok {
				continue
			}
			var keyPath string
			switch p.Scope {
			case "global":
				keyPath = fmt.Sprintf("vars.global.%s", p.Name)
			case "catalog":
				keyPath = fmt.Sprintf("vars.catalog.%s.%s", cat.Name, p.Name)
			case "runbook":
				keyPath = fmt.Sprintf("vars.catalog.%s.runbooks.%s.%s", cat.Name, rb.Name, p.Name)
			default:
				keyPath = fmt.Sprintf("vars.global.%s", p.Name)
			}
			config.Set(m.deps.Config, keyPath, val)
		}
		m.deps.Store.Save(m.deps.Config)
	}

	if m.deps.Runner == nil {
		return m, nil
	}

	catPath := expandTilde(cat.Path)
	scriptPath := filepath.Join(catPath, rb.Name, rb.Script)

	var logPath string
	if m.deps.LogWriter != nil {
		lp, err := m.deps.LogWriter.Create(cat.Name, rb.Name, time.Now())
		if err == nil {
			logPath = lp
		}
	}

	env := make(map[string]string)
	for k, v := range params {
		env[strings.ToUpper(k)] = v
	}

	runner := m.deps.Runner
	lw := m.deps.LogWriter
	finalLogPath := logPath

	return m, func() tea.Msg {
		lines, errs := runner.Run(context.Background(), scriptPath, env)
		var collected []output.OutputLineMsg
		for line := range lines {
			if lw != nil {
				lw.WriteLine(line.Text)
			}
			collected = append(collected, output.OutputLineMsg{
				Text:     line.Text,
				IsStderr: line.IsStderr,
			})
		}
		if lw != nil {
			lw.Close()
		}
		<-errs
		return executionResultMsg{
			Lines:   collected,
			LogPath: finalLogPath,
		}
	}
}

func (m App) openPalette() (tea.Model, tea.Cmd) {
	p := palette.New(m.width)
	m.pal = &p
	m.state = statePalette
	return m, nil
}

func (m App) openWizard() (tea.Model, tea.Cmd) {
	if m.selected == nil || m.selCat == nil {
		return m, nil
	}

	resolved := m.resolveVars()

	if wizard.ShouldSkip(m.selected.Parameters, resolved) {
		return m.startExecution(*m.selected, *m.selCat, resolved)
	}

	wiz := wizard.New(*m.selected, *m.selCat, resolved)
	m.wizard = &wiz
	m.state = stateWizard
	return m, wiz.Init()
}

func (m App) resolveVars() map[string]string {
	if m.deps.Config == nil || m.selected == nil || m.selCat == nil {
		return make(map[string]string)
	}
	resolver := vars.NewDefaultResolver()
	return resolver.Resolve(m.deps.Config, m.selCat.Name, m.selected.Name, m.selected.Parameters)
}

func (m App) View() tea.View {
	// Guard: before WindowSizeMsg arrives, width/height are defaults (80x24).
	// In alt screen, BubbleTea sends WindowSizeMsg on startup, but View()
	// may be called first. Return minimal content to avoid broken layout.
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("")
		v.AltScreen = m.deps.AltScreen
		return v
	}

	var v tea.View

	if m.state == stateWizard && m.wizard != nil {
		v = m.viewWizardOverlay()
	} else if m.state == statePalette && m.pal != nil {
		v = m.viewPaletteOverlay()
	} else {
		v = m.viewNormal()
	}

	v.AltScreen = m.deps.AltScreen
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m App) viewNormal() tea.View {
	// --- Layout variables ---
	marginLeft := 3 // space from left terminal edge
	marginTop  := 3 // space from top terminal edge
	footerH    := 1 // footer height
	gap        := 1 // space between sidebar and right panel
	borderSize := 2 // top + bottom (or left + right) border chars

	// --- Dimension budget ---
	innerW    := clamp(m.width - marginLeft, 1)
	sw        := sidebarWidth(innerW)
	rightW    := clamp(innerW - sw - borderSize - gap, 1)
	contentW  := clamp(rightW - borderSize, 1) // content width inside bordered panels
	panelRows := clamp(m.height - marginTop - footerH - marginLeft, 1)

	// --- Theme colors ---
	var borderColor color.Color = lipgloss.NoColor{}
	var panelBg color.Color = lipgloss.NoColor{}
	if m.deps.Styles != nil {
		borderColor = m.deps.Styles.Border.GetForeground()
		panelBg = m.deps.Styles.BackgroundPanel.GetForeground()
	}

	// --- Sidebar: anchor panel ---
	sidebarContentH := clamp(panelRows - borderSize - 1, 3) // -1 accounts for border rendering offset
	m.sidebar.SetHeight(sidebarContentH)
	m.sidebar.SetYOffset(1)

	sidebarView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBg).
		Width(sw).
		Height(sidebarContentH).
		Render(m.sidebar.View())

	sidebarRenderedH := lipgloss.Height(sidebarView)

	// --- Metadata: bordered, auto-height ---
	metaContent := metadata.Render(m.selected, contentW, m.deps.Styles)
	metaView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(panelBg).
		Width(contentW).
		Render(metaContent)

	metaRenderedH := lipgloss.Height(metaView)

	// --- Output: bordered, fills remaining to match sidebar ---
	outputH := clamp(sidebarRenderedH - metaRenderedH, 3)
	outputView := m.output.ViewWithSize(rightW, outputH)

	// --- Compose panels ---
	rightPanel := lipgloss.JoinVertical(lipgloss.Left, metaView, outputView)

	body := lipgloss.NewStyle().
		MarginLeft(marginLeft).
		MarginTop(marginTop).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			sidebarView,
			strings.Repeat(" ", gap),
			rightPanel,
		))

	// --- Footer ---
	footerView := lipgloss.NewStyle().
		MarginLeft(marginLeft - 1).
		Render(footer.Render(appFooterState(m.state), m.width-marginLeft, m.deps.Styles))

	// --- Outer container: enforce exact terminal dimensions ---
	content := lipgloss.JoinVertical(lipgloss.Left, body, footerView)
	content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content)

	return tea.NewView(content)
}

func (m App) viewWizardOverlay() tea.View {
	wizView := m.wizard.View()

	overlayW := m.width * 2 / 3
	if overlayW < 50 {
		overlayW = 50
	}

	overlay := lipgloss.NewStyle().
		Width(overlayW).
		Render(wizView)

	content := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlay,
	)

	footerView := footer.Render(footer.StateWizard, m.width, m.deps.Styles)
	content = lipgloss.JoinVertical(lipgloss.Left, content, footerView)

	return tea.NewView(content)
}

func (m App) viewPaletteOverlay() tea.View {
	palView := m.pal.View()

	overlay := lipgloss.NewStyle().
		Width(m.width).
		Render(palView)

	footerView := footer.Render(footer.StatePalette, m.width, m.deps.Styles)
	content := lipgloss.JoinVertical(lipgloss.Left, overlay, footerView)

	return tea.NewView(content)
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// clamp returns v if v >= min, otherwise min.
func clamp(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func sidebarWidth(totalWidth int) int {
	w := totalWidth / 3
	if w < 30 {
		w = 30
	}
	if w > 50 {
		w = 50
	}
	return w
}

func appFooterState(s viewState) footer.State {
	switch s {
	case stateWizard:
		return footer.StateWizard
	case statePalette:
		return footer.StatePalette
	default:
		return footer.StateNormal
	}
}
