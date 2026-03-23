package cli

import (
	"fmt"
	"io"

	lipgloss "charm.land/lipgloss/v2"
)

var (
	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#f7768e")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true)

	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89"))
)

func FormatError(w io.Writer, title, detail string) {
	badge := badgeStyle.Render("ERROR")
	fmt.Fprintf(w, "\n  %s %s\n", badge, titleStyle.Render(title))
	if detail != "" {
		fmt.Fprintf(w, "\n  %s\n", detailStyle.Render(detail))
	}
	fmt.Fprintln(w)
}
