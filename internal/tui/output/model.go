package output

import (
	"fmt"
	"image/color"
	"strings"

	"dops/internal/theme"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type Model struct {
	command     string
	lines       []OutputLineMsg
	logPath     string
	width       int
	height      int
	offset      int
	searching   bool
	navigating  bool
	searchQuery string
	matchLines  []int // line indices that contain matches
	matchCount  int
	matchIndex  int
	styles      *theme.Styles
}

func New(width, height int, styles *theme.Styles) Model {
	return Model{width: width, height: height, styles: styles}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Clear() {
	m.command = ""
	m.lines = nil
	m.logPath = ""
	m.clearSearch()
}

func (m *Model) SetCommand(cmd string) {
	m.command = cmd
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case OutputLineMsg:
		m.lines = append(m.lines, msg)
		// Auto-scroll to bottom when new output arrives and not searching
		if !m.searching && !m.navigating {
			m.scrollToBottom()
		}
		return m, nil

	case ExecutionDoneMsg:
		m.logPath = msg.LogPath
		return m, nil

	case tea.KeyPressMsg:
		if m.navigating {
			return m.updateNavigating(msg), nil
		}
		if m.searching {
			return m.updateSearching(msg), nil
		}
		return m.updateNormal(msg), nil
	}

	return m, nil
}

func (m Model) updateNormal(msg tea.KeyPressMsg) Model {
	switch {
	case msg.Text == "/" || msg.String() == "/":
		m.searching = true
		m.searchQuery = ""
	case msg.Code == tea.KeyUp:
		if m.offset > 0 {
			m.offset--
		}
	case msg.Code == tea.KeyDown:
		m.offset++
		m.clampOffset()
	}
	return m
}

func (m Model) updateSearching(msg tea.KeyPressMsg) Model {
	switch {
	case msg.Code == tea.KeyEscape:
		m.clearSearch()
	case msg.Code == tea.KeyEnter:
		if m.matchCount > 0 {
			m.navigating = true
			m.searching = false
			m.matchIndex = 0
			m.scrollToMatch()
		}
	case msg.Code == tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.applySearch()
		}
	default:
		if msg.Text != "" {
			m.searchQuery += msg.Text
			m.applySearch()
		}
	}
	return m
}

func (m Model) updateNavigating(msg tea.KeyPressMsg) Model {
	switch {
	case msg.Code == tea.KeyEscape:
		m.clearSearch()
	case msg.Text == "n":
		if m.matchCount > 0 {
			m.matchIndex = (m.matchIndex + 1) % m.matchCount
			m.scrollToMatch()
		}
	case msg.Text == "N":
		if m.matchCount > 0 {
			m.matchIndex = (m.matchIndex - 1 + m.matchCount) % m.matchCount
			m.scrollToMatch()
		}
	}
	return m
}

func (m *Model) applySearch() {
	m.matchLines = nil
	m.matchCount = 0
	m.matchIndex = 0

	if m.searchQuery == "" {
		return
	}

	q := strings.ToLower(m.searchQuery)
	for i, line := range m.lines {
		if strings.Contains(strings.ToLower(line.Text), q) {
			m.matchLines = append(m.matchLines, i)
			m.matchCount++
		}
	}
}

func (m *Model) clearSearch() {
	m.searching = false
	m.navigating = false
	m.searchQuery = ""
	m.matchLines = nil
	m.matchCount = 0
	m.matchIndex = 0
}

func (m *Model) scrollToMatch() {
	if m.matchIndex < 0 || m.matchIndex >= len(m.matchLines) {
		return
	}
	lineIdx := m.matchLines[m.matchIndex]
	bodyHeight := m.bodyHeight()
	if bodyHeight <= 0 {
		return
	}
	if lineIdx < m.offset {
		m.offset = lineIdx
	}
	if lineIdx >= m.offset+bodyHeight {
		m.offset = lineIdx - bodyHeight + 1
	}
}

func (m *Model) scrollToBottom() {
	bodyH := m.bodyHeight()
	if bodyH <= 0 {
		return
	}
	if len(m.lines) > bodyH {
		m.offset = len(m.lines) - bodyH
	}
}

func (m *Model) clampOffset() {
	bodyH := m.bodyHeight()
	if bodyH <= 0 {
		return
	}
	maxOffset := len(m.lines) - bodyH
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m Model) bodyHeight() int {
	h := m.height - 2 // header + footer reserve
	if h < 1 {
		h = 1
	}
	return h
}

func (m Model) ViewWithSize(width, height int) string {
	m.width = width
	m.height = height
	return m.View()
}

func (m Model) View() string {
	innerW := m.width - 4 // border + padding
	if innerW < 10 {
		innerW = 10
	}

	headerStyle := lipgloss.NewStyle().Width(innerW).Padding(0, 1)
	footerLogStyle := lipgloss.NewStyle().Width(innerW).Padding(0, 1)
	stderrStyle := lipgloss.NewStyle()
	placeholderStyle := lipgloss.NewStyle()
	var borderColor color.Color = lipgloss.NoColor{}

	if m.styles != nil {
		bgElem := m.styles.BackgroundElem.GetForeground()
		headerStyle = headerStyle.
			Background(bgElem).
			Foreground(m.styles.Text.GetForeground())
		footerLogStyle = footerLogStyle.
			Background(bgElem).
			Foreground(m.styles.TextMuted.GetForeground())
		stderrStyle = m.styles.Error
		placeholderStyle = m.styles.TextMuted
		borderColor = m.styles.Border.GetForeground()
	}

	var b strings.Builder

	// Header
	if m.command != "" {
		b.WriteString(headerStyle.Render("$ "+m.command) + "\n")
	}

	// Body
	if len(m.lines) == 0 && m.command == "" {
		// Center placeholder vertically and horizontally
		padTop := m.height / 3
		for i := 0; i < padTop; i++ {
			b.WriteString("\n")
		}
		placeholder := placeholderStyle.Render("Press enter to run a runbook")
		centered := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, placeholder)
		b.WriteString(centered + "\n")
	} else if len(m.lines) == 0 {
		padTop := m.height / 3
		for i := 0; i < padTop; i++ {
			b.WriteString("\n")
		}
		running := placeholderStyle.Render("Running...")
		centered := lipgloss.PlaceHorizontal(innerW, lipgloss.Center, running)
		b.WriteString(centered + "\n")
	} else {
		bodyH := m.bodyHeight()
		start := m.offset
		if start > len(m.lines) {
			start = len(m.lines)
		}
		end := start + bodyH
		if end > len(m.lines) {
			end = len(m.lines)
		}

		matchSet := m.matchLineSet()
		needsScrollbar := len(m.lines) > bodyH

		for i := start; i < end; i++ {
			line := m.lines[i]
			prefix := "  "
			if matchSet[i] {
				prefix = "» "
			}
			lineText := prefix + line.Text
			if line.IsStderr {
				lineText = stderrStyle.Render(lineText)
			}
			if needsScrollbar {
				lineText += " " + scrollbarChar(i-start, len(m.lines), bodyH)
			}
			b.WriteString(lineText + "\n")
		}
	}

	// Footer — log path or search status
	if m.logPath != "" && !m.searching && !m.navigating {
		b.WriteString(footerLogStyle.Render("Saved to "+m.logPath) + "\n")
	}

	if m.searching {
		b.WriteString(fmt.Sprintf("  /%s", m.searchQuery))
	}

	if m.navigating && m.matchCount > 0 {
		b.WriteString(fmt.Sprintf("  [%d/%d]", m.matchIndex+1, m.matchCount))
	}

	// Wrap in rounded border — height fills allocated space
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2). // -2 for left+right border chars
		Height(m.height)

	return borderStyle.Render(b.String())
}

func (m Model) matchLineSet() map[int]bool {
	set := make(map[int]bool, len(m.matchLines))
	for _, idx := range m.matchLines {
		set[idx] = true
	}
	return set
}

func scrollbarChar(lineIdx, total, visible int) string {
	if total <= visible {
		return " "
	}
	thumbSize := visible * visible / total
	if thumbSize < 1 {
		thumbSize = 1
	}
	thumbStart := lineIdx * total / visible
	_ = thumbStart
	// Simple proportional thumb
	pos := lineIdx * total / visible
	thumbPos := (total - visible) * lineIdx / visible
	_ = pos
	_ = thumbPos
	// Simplified: just show block for thumb region
	ratio := float64(lineIdx) / float64(visible)
	scrollRatio := float64(total-visible) / float64(total)
	_ = scrollRatio
	if ratio < float64(thumbSize)/float64(visible) {
		return "█"
	}
	return "░"
}
