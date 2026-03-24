package footer

import (
	"dops/internal/theme"

	lipgloss "charm.land/lipgloss/v2"
)

type State int

const (
	StateNormal  State = iota
	StateWizard
	StateRunning
	StatePalette
	StateConfirm
	StateHelp
)

type binding struct {
	key  string
	desc string
}

func Render(state State, width int, styles *theme.Styles) string {
	var bindings []binding

	switch state {
	case StateNormal:
		bindings = []binding{
			{"↑↓", "navigate"},
			{"enter", "run"},
			{"/", "search"},
			{"ctrl+shift+p", "palette"},
			{"q", "quit"},
		}
	case StateWizard:
		bindings = []binding{
			{"tab", "next"},
			{"shift+tab", "prev"},
			{"enter", "submit"},
			{"esc", "cancel"},
		}
	case StateRunning:
		bindings = []binding{
			{"", "Running..."},
			{"esc", "cancel"},
		}
	case StatePalette:
		bindings = []binding{
			{"↑↓", "select"},
			{"enter", "confirm"},
			{"esc", "close"},
		}
	case StateConfirm:
		bindings = []binding{
			{"enter", "confirm"},
			{"esc", "cancel"},
		}
	case StateHelp:
		bindings = []binding{
			{"?", "close"},
			{"esc", "close"},
		}
	}

	keyStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle()
	barStyle := lipgloss.NewStyle().Width(width)

	if styles != nil {
		keyStyle = styles.Primary
		descStyle = styles.TextMuted
	}

	var content string
	for i, b := range bindings {
		if i > 0 {
			content += descStyle.Render(" • ")
		}
		if b.key != "" {
			content += keyStyle.Render(b.key) + " " + descStyle.Render(b.desc)
		} else {
			content += descStyle.Render(b.desc)
		}
	}

	return barStyle.Render("  " + content)
}
