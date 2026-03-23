package theme

import (
	"dops/internal/domain"
	"encoding/json"
	"testing"
)

func makeThemeFile() *domain.ThemeFile {
	tf := &domain.ThemeFile{
		Name: "test-theme",
		Defs: map[string]string{
			"bg":       "#1a1b26",
			"fg":       "#c0caf5",
			"blue":     "#7aa2f7",
			"green":    "#9ece6a",
			"red":      "#f7768e",
			"dayBg":    "#e1e2e7",
			"dayFg":    "#3760bf",
			"dayBlue":  "#2e7de9",
			"dayGreen": "#587539",
			"dayRed":   "#f52a65",
		},
		Theme: map[string]json.RawMessage{},
	}

	setToken(tf, "background", "bg", "dayBg")
	setToken(tf, "text", "fg", "dayFg")
	setToken(tf, "primary", "blue", "dayBlue")
	setToken(tf, "success", "green", "dayGreen")
	setToken(tf, "error", "red", "dayRed")

	return tf
}

func setToken(tf *domain.ThemeFile, name, dark, light string) {
	data, _ := json.Marshal(domain.ThemeToken{Dark: dark, Light: light})
	tf.Theme[name] = data
}

func TestResolve_Dark(t *testing.T) {
	tf := makeThemeFile()
	resolved, err := Resolve(tf, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	tests := []struct {
		token string
		want  string
	}{
		{"background", "#1a1b26"},
		{"text", "#c0caf5"},
		{"primary", "#7aa2f7"},
		{"success", "#9ece6a"},
		{"error", "#f7768e"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got, ok := resolved.Colors[tt.token]
			if !ok {
				t.Fatalf("token %q not found", tt.token)
			}
			if got != tt.want {
				t.Errorf("token %q = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func TestResolve_Light(t *testing.T) {
	tf := makeThemeFile()
	resolved, err := Resolve(tf, false)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	tests := []struct {
		token string
		want  string
	}{
		{"background", "#e1e2e7"},
		{"text", "#3760bf"},
		{"primary", "#2e7de9"},
		{"success", "#587539"},
		{"error", "#f52a65"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got, ok := resolved.Colors[tt.token]
			if !ok {
				t.Fatalf("token %q not found", tt.token)
			}
			if got != tt.want {
				t.Errorf("token %q = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func TestResolve_NestedTokens(t *testing.T) {
	tf := makeThemeFile()

	// Add nested risk tokens
	risk := map[string]domain.ThemeToken{
		"low":      {Dark: "green", Light: "dayGreen"},
		"critical": {Dark: "red", Light: "dayRed"},
	}
	data, _ := json.Marshal(risk)
	tf.Theme["risk"] = data

	resolved, err := Resolve(tf, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if resolved.Colors["risk.low"] != "#9ece6a" {
		t.Errorf("risk.low = %q, want #9ece6a", resolved.Colors["risk.low"])
	}
	if resolved.Colors["risk.critical"] != "#f7768e" {
		t.Errorf("risk.critical = %q, want #f7768e", resolved.Colors["risk.critical"])
	}
}

func TestResolve_DanglingRef(t *testing.T) {
	tf := &domain.ThemeFile{
		Name: "bad-theme",
		Defs: map[string]string{
			"bg": "#000000",
		},
		Theme: map[string]json.RawMessage{},
	}
	setToken(tf, "background", "bg", "nonexistent")

	_, err := Resolve(tf, false)
	if err == nil {
		t.Error("expected error for dangling def reference")
	}
}

func TestResolve_DirectHexValue(t *testing.T) {
	tf := &domain.ThemeFile{
		Name: "direct-hex",
		Defs: map[string]string{},
		Theme: map[string]json.RawMessage{},
	}
	setToken(tf, "background", "#111111", "#eeeeee")

	resolved, err := Resolve(tf, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if resolved.Colors["background"] != "#111111" {
		t.Errorf("background = %q, want #111111", resolved.Colors["background"])
	}
}

func TestResolve_NoneValue(t *testing.T) {
	tf := &domain.ThemeFile{
		Name: "none-bg",
		Defs: map[string]string{},
		Theme: map[string]json.RawMessage{},
	}
	setToken(tf, "background", "none", "none")

	resolved, err := Resolve(tf, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if resolved.Colors["background"] != "none" {
		t.Errorf("background = %q, want none", resolved.Colors["background"])
	}
}
