package theme

import (
	"testing"
)

func TestBuildStyles_FromBundledTheme(t *testing.T) {
	loader := NewFileLoader(newFakeFS(), "/fake/themes")
	tf, err := loader.Load("tokyonight")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	resolved, err := Resolve(tf, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	styles := BuildStyles(resolved)

	// Verify key styles are populated by checking render doesn't panic
	// and that the style produces non-empty output
	tests := []struct {
		name  string
		style func() string
	}{
		{"Background", func() string { return styles.Background.Render(" ") }},
		{"BackgroundPanel", func() string { return styles.BackgroundPanel.Render(" ") }},
		{"Text", func() string { return styles.Text.Render("test") }},
		{"TextMuted", func() string { return styles.TextMuted.Render("test") }},
		{"Primary", func() string { return styles.Primary.Render("test") }},
		{"Border", func() string { return styles.Border.Render("test") }},
		{"Success", func() string { return styles.Success.Render("test") }},
		{"Warning", func() string { return styles.Warning.Render("test") }},
		{"Error", func() string { return styles.Error.Render("test") }},
		{"RiskLow", func() string { return styles.RiskLow.Render("test") }},
		{"RiskMedium", func() string { return styles.RiskMedium.Render("test") }},
		{"RiskHigh", func() string { return styles.RiskHigh.Render("test") }},
		{"RiskCritical", func() string { return styles.RiskCritical.Render("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.style()
			if out == "" {
				t.Errorf("style %q produced empty output", tt.name)
			}
		})
	}
}

func TestBuildStyles_MissingTokenUsesDefaults(t *testing.T) {
	resolved := &ResolvedTheme{
		Name:   "minimal",
		Colors: map[string]string{
			"text": "#ffffff",
		},
	}

	// Should not panic even with missing tokens
	styles := BuildStyles(resolved)
	out := styles.Text.Render("test")
	if out == "" {
		t.Error("text style produced empty output")
	}
}
