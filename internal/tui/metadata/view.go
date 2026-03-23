package metadata

import (
	"fmt"
	"strings"

	"dops/internal/domain"
	"dops/internal/theme"

	lipgloss "charm.land/lipgloss/v2"
)

// Render returns the metadata content WITHOUT a border.
// The parent layout wraps it in a border for consistent alignment.
func Render(rb *domain.Runbook, width int, styles *theme.Styles) string {
	if rb == nil {
		return "  No runbook selected"
	}

	nameStyle := lipgloss.NewStyle().Bold(true)
	descStyle := lipgloss.NewStyle()
	labelStyle := lipgloss.NewStyle()

	if styles != nil {
		nameStyle = styles.Text.Bold(true)
		descStyle = styles.TextMuted
		labelStyle = styles.TextMuted
	}

	var b strings.Builder
	fmt.Fprintf(&b, " %s\n", nameStyle.Render(rb.Name))
	fmt.Fprintf(&b, " %s\n", descStyle.Render(rb.Description))
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, " %s %s\n", labelStyle.Render("Version:   "), rb.Version)
	fmt.Fprintf(&b, " %s %s\n", labelStyle.Render("Risk Level:"), riskBadge(rb.RiskLevel, styles))
	fmt.Fprintf(&b, " %s %s", labelStyle.Render("ID:        "), rb.ID)

	return b.String()
}

func riskBadge(level domain.RiskLevel, styles *theme.Styles) string {
	label := string(level)
	if styles == nil {
		return label
	}
	switch level {
	case domain.RiskLow:
		return styles.RiskLow.Render(label)
	case domain.RiskMedium:
		return styles.RiskMedium.Render(label)
	case domain.RiskHigh:
		return styles.RiskHigh.Render(label)
	case domain.RiskCritical:
		return styles.RiskCritical.Render(label)
	default:
		return label
	}
}
