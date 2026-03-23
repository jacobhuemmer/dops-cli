package footer

import (
	"dops/internal/theme"
	"strings"
	"testing"
)

func footerTestStyles() *theme.Styles {
	return theme.BuildStyles(&theme.ResolvedTheme{
		Name: "test",
		Colors: map[string]string{
			"background":        "#1a1b26",
			"backgroundPanel":   "#1f2335",
			"backgroundElement": "#292e42",
			"text":              "#c0caf5",
			"textMuted":         "#565f89",
			"primary":           "#7aa2f7",
			"border":            "#3b4261",
			"borderActive":      "#7aa2f7",
			"success":           "#9ece6a",
			"warning":           "#e0af68",
			"error":             "#f7768e",
			"risk.low":          "#9ece6a",
			"risk.medium":       "#e0af68",
			"risk.high":         "#f7768e",
			"risk.critical":     "#db4b4b",
		},
	})
}

func TestRender_Normal(t *testing.T) {
	out := Render(StateNormal, 60, footerTestStyles())
	if !strings.Contains(out, "enter") {
		t.Error("normal state should show enter keybind")
	}
	if !strings.Contains(out, "q") {
		t.Error("normal state should show quit keybind")
	}
}

func TestRender_Running(t *testing.T) {
	out := Render(StateRunning, 60, footerTestStyles())
	if !strings.Contains(out, "running") && !strings.Contains(out, "Running") {
		t.Error("running state should indicate execution")
	}
}
