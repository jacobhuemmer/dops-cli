package theme

import "charm.land/lipgloss/v2"

type Styles struct {
	Background      lipgloss.Style
	BackgroundPanel lipgloss.Style
	BackgroundElem  lipgloss.Style
	Text            lipgloss.Style
	TextMuted       lipgloss.Style
	Primary         lipgloss.Style
	Border          lipgloss.Style
	BorderActive    lipgloss.Style
	Success         lipgloss.Style
	Warning         lipgloss.Style
	Error           lipgloss.Style
	RiskLow         lipgloss.Style
	RiskMedium      lipgloss.Style
	RiskHigh        lipgloss.Style
	RiskCritical    lipgloss.Style
}

func BuildStyles(rt *ResolvedTheme) *Styles {
	return &Styles{
		Background:      styleWithFg(rt, "background"),
		BackgroundPanel: styleWithFg(rt, "backgroundPanel"),
		BackgroundElem:  styleWithFg(rt, "backgroundElement"),
		Text:            styleWithFg(rt, "text"),
		TextMuted:       styleWithFg(rt, "textMuted"),
		Primary:         styleWithFg(rt, "primary"),
		Border:          styleWithFg(rt, "border"),
		BorderActive:    styleWithFg(rt, "borderActive"),
		Success:         styleWithFg(rt, "success"),
		Warning:         styleWithFg(rt, "warning"),
		Error:           styleWithFg(rt, "error"),
		RiskLow:         styleWithFg(rt, "risk.low"),
		RiskMedium:      styleWithFg(rt, "risk.medium"),
		RiskHigh:        styleWithFg(rt, "risk.high"),
		RiskCritical:    styleWithFg(rt, "risk.critical"),
	}
}

func styleWithFg(rt *ResolvedTheme, token string) lipgloss.Style {
	hex, ok := rt.Colors[token]
	if !ok || hex == "" || hex == "none" {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}
