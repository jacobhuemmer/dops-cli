package confirm

import (
	"fmt"
	"strings"

	"dops/internal/domain"
	"dops/internal/theme"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type Model struct {
	runbook domain.Runbook
	catalog domain.Catalog
	params  map[string]string
	risk    domain.RiskLevel
	input   string
	width   int
	styles  *theme.Styles
}

func New(rb domain.Runbook, cat domain.Catalog, params map[string]string, width int, styles *theme.Styles) Model {
	return Model{
		runbook: rb,
		catalog: cat,
		params:  params,
		risk:    rb.RiskLevel,
		width:   width,
		styles:  styles,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case msg.Code == tea.KeyEscape:
			return m, func() tea.Msg { return ConfirmCancelMsg{} }

		case msg.Code == tea.KeyEnter:
			if m.isConfirmed() {
				return m, func() tea.Msg {
					return ConfirmAcceptMsg{
						Runbook: m.runbook,
						Catalog: m.catalog,
						Params:  m.params,
					}
				}
			}

		case msg.Code == tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			if m.risk == domain.RiskHigh {
				// High: y/N single key confirmation.
				if msg.Text == "y" || msg.Text == "Y" {
					return m, func() tea.Msg {
						return ConfirmAcceptMsg{
							Runbook: m.runbook,
							Catalog: m.catalog,
							Params:  m.params,
						}
					}
				}
				if msg.Text == "n" || msg.Text == "N" {
					return m, func() tea.Msg { return ConfirmCancelMsg{} }
				}
			} else if msg.Text != "" {
				// Critical: accumulate typed input.
				m.input += msg.Text
			}
		}
	}
	return m, nil
}

func (m Model) isConfirmed() bool {
	switch m.risk {
	case domain.RiskHigh:
		return false // high uses y/N, not Enter
	case domain.RiskCritical:
		return strings.TrimSpace(m.input) == m.runbook.ID
	default:
		return true
	}
}

func (m Model) View() string {
	var warningFg, mutedFg, textFg, errorFg lipgloss.Style
	if m.styles != nil {
		warningFg = m.styles.Warning
		mutedFg = m.styles.TextMuted
		textFg = m.styles.Text
		errorFg = m.styles.Error
	}

	riskLabel := strings.ToUpper(string(m.risk))
	var riskStyle lipgloss.Style
	switch m.risk {
	case domain.RiskHigh:
		riskStyle = warningFg
	case domain.RiskCritical:
		riskStyle = errorFg.Bold(true)
	default:
		riskStyle = textFg
	}

	w := m.width
	if w < 30 {
		w = 30
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, riskStyle.Render(fmt.Sprintf("  ⚠  %s RISK", riskLabel)))
	lines = append(lines, "")
	lines = append(lines, mutedFg.Render(fmt.Sprintf("  Runbook: %s", m.runbook.ID)))
	lines = append(lines, "")

	switch m.risk {
	case domain.RiskHigh:
		lines = append(lines, textFg.Render("  Confirm execution? (y/N)"))
	case domain.RiskCritical:
		lines = append(lines, textFg.Render("  Type the runbook ID to confirm:"))
		lines = append(lines, mutedFg.Render(fmt.Sprintf("  %s", m.runbook.ID)))
		lines = append(lines, "")
		lines = append(lines, textFg.Render(fmt.Sprintf("  > %s▎", m.input)))
	}

	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(w).Render(content)
}
