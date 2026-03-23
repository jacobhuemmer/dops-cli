package metadata

import (
	"dops/internal/domain"
	"dops/internal/theme"
	"strings"
	"testing"
)

func metadataTestStyles() *theme.Styles {
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

func TestRender(t *testing.T) {
	rb := &domain.Runbook{
		ID:          "default.hello-world",
		Name:        "hello-world",
		Description: "Prints a hello world message",
		Version:     "1.0.0",
		RiskLevel:   domain.RiskLow,
	}

	out := Render(rb, 40, metadataTestStyles())

	if !strings.Contains(out, "hello-world") {
		t.Error("output should contain runbook name")
	}
	if !strings.Contains(out, "1.0.0") {
		t.Error("output should contain version")
	}
	if !strings.Contains(out, "low") {
		t.Error("output should contain risk level")
	}
	if !strings.Contains(out, "Prints a hello world message") {
		t.Error("output should contain description")
	}
}

func TestRender_Nil(t *testing.T) {
	out := Render(nil, 40, metadataTestStyles())
	if len(out) == 0 {
		t.Error("nil runbook should still produce output")
	}
}
