package tui

import (
	"dops/internal/catalog"
	"dops/internal/domain"
	"dops/internal/theme"
	"dops/internal/tui/output"
	"dops/internal/tui/palette"
	"dops/internal/tui/sidebar"
	"dops/internal/tui/wizard"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

var appANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func testStyles() *theme.Styles {
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

func testCatalogs() []catalog.CatalogWithRunbooks {
	return []catalog.CatalogWithRunbooks{
		{
			Catalog: domain.Catalog{Name: "default"},
			Runbooks: []domain.Runbook{
				{ID: "default.hello-world", Name: "hello-world", RiskLevel: domain.RiskLow},
				{ID: "default.rotate-tls", Name: "rotate-tls", RiskLevel: domain.RiskMedium},
			},
		},
	}
}

func TestApp_RunbookSelectedMsg(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	rb := domain.Runbook{ID: "default.rotate-tls", Name: "rotate-tls"}
	cat := domain.Catalog{Name: "default"}
	result, _ := m.Update(sidebar.RunbookSelectedMsg{Runbook: rb, Catalog: cat})
	app := result.(App)

	if app.selected == nil {
		t.Fatal("selected should be set after RunbookSelectedMsg")
	}
	if app.selected.ID != "default.rotate-tls" {
		t.Errorf("selected = %q, want default.rotate-tls", app.selected.ID)
	}
}

func TestApp_QuitOnQ(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})
	if cmd == nil {
		t.Fatal("q should produce a quit command")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}
}

func TestApp_ViewNotEmpty(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	// Send WindowSizeMsg so layout has dimensions
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := result.(App)
	view := app.View()
	if view.Content == "" {
		t.Error("View should produce non-empty content")
	}
}

func TestApp_WindowResize(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := result.(App)

	if app.width != 120 || app.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
	}
}

func testCatalogsWithParams() []catalog.CatalogWithRunbooks {
	return []catalog.CatalogWithRunbooks{
		{
			Catalog: domain.Catalog{Name: "default"},
			Runbooks: []domain.Runbook{
				{
					ID:   "default.hello-world",
					Name: "hello-world",
					Parameters: []domain.Parameter{
						{Name: "greeting", Type: domain.ParamString, Required: true, Scope: "global"},
					},
				},
			},
		},
	}
}

func TestApp_ExecuteOpensWizard(t *testing.T) {
	m := NewApp(testCatalogsWithParams(), testStyles())
	m.Init()

	// Sidebar sends RunbookExecuteMsg when Enter is pressed on a runbook
	rb := domain.Runbook{
		ID:   "default.hello-world",
		Name: "hello-world",
		Parameters: []domain.Parameter{
			{Name: "greeting", Type: domain.ParamString, Required: true, Scope: "global"},
		},
	}
	cat := domain.Catalog{Name: "default"}
	result, _ := m.Update(sidebar.RunbookExecuteMsg{Runbook: rb, Catalog: cat})
	app := result.(App)

	if app.state != stateWizard {
		t.Errorf("state = %d, want stateWizard (%d)", app.state, stateWizard)
	}
	if app.wizard == nil {
		t.Error("wizard should be created")
	}
}

func TestApp_WizardCancel(t *testing.T) {
	m := NewApp(testCatalogsWithParams(), testStyles())
	m.Init()

	// Set up and open wizard
	rb := domain.Runbook{
		ID:         "default.hello-world",
		Name:       "hello-world",
		Parameters: []domain.Parameter{{Name: "greeting", Type: domain.ParamString, Required: true, Scope: "global"}},
	}
	m.selected = &rb
	cat := domain.Catalog{Name: "default"}
	m.selCat = &cat
	m.state = stateWizard
	wiz := wizard.New(rb, cat, map[string]string{})
	m.wizard = &wiz

	// Send cancel message
	result, _ := m.Update(wizard.WizardCancelMsg{})
	app := result.(App)

	if app.state != stateNormal {
		t.Errorf("state after cancel = %d, want stateNormal", app.state)
	}
	if app.wizard != nil {
		t.Error("wizard should be nil after cancel")
	}
}

func TestApp_WizardSubmit(t *testing.T) {
	m := NewApp(testCatalogsWithParams(), testStyles())
	m.Init()

	m.state = stateWizard

	rb := domain.Runbook{ID: "default.hello-world", Name: "hello-world"}
	cat := domain.Catalog{Name: "default"}
	params := map[string]string{"greeting": "world"}

	result, _ := m.Update(wizard.WizardSubmitMsg{Runbook: rb, Catalog: cat, Params: params})
	app := result.(App)

	if app.state != stateNormal {
		t.Errorf("state after submit = %d, want stateNormal", app.state)
	}
	if app.wizard != nil {
		t.Error("wizard should be nil after submit")
	}
}

func TestApp_PaletteCancel(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	// Open palette
	p := palette.New(80)
	m.pal = &p
	m.state = statePalette

	result, _ := m.Update(palette.PaletteCancelMsg{})
	app := result.(App)

	if app.state != stateNormal {
		t.Errorf("state after palette cancel = %d, want stateNormal", app.state)
	}
	if app.pal != nil {
		t.Error("palette should be nil after cancel")
	}
}

func TestApp_PaletteSelect(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	p := palette.New(80)
	m.pal = &p
	m.state = statePalette

	cmd := palette.PaletteCommand{Name: "theme: set"}
	result, _ := m.Update(palette.PaletteSelectMsg{Command: cmd})
	app := result.(App)

	if app.state != stateNormal {
		t.Errorf("state after palette select = %d, want stateNormal", app.state)
	}
	if app.pal != nil {
		t.Error("palette should be nil after select")
	}
}

func TestApp_MouseClickTranslation(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	// Send WindowSizeMsg so layout has dimensions
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := result.(App)

	// Absolute coords for hello-world (visible index 1):
	// Y = layoutMarginTop(3) + borderTop(1) + itemIndex(1) = 5
	// X = layoutMarginLeft(3) + borderLeft(1) + padLeft(1) + some offset = 7
	result, cmd := app.Update(tea.MouseClickMsg{X: 7, Y: 5, Button: tea.MouseLeft})
	_ = result

	if cmd == nil {
		t.Fatal("click on runbook should produce a command")
	}

	msg := cmd()
	sel, ok := msg.(sidebar.RunbookSelectedMsg)
	if !ok {
		t.Fatalf("expected RunbookSelectedMsg, got %T", msg)
	}
	if sel.Runbook.ID != "default.hello-world" {
		t.Errorf("selected = %q, want default.hello-world", sel.Runbook.ID)
	}
}

func TestApp_ClickOutputFooterCopiesLogPath(t *testing.T) {
	m := NewApp(testCatalogs(), testStyles())
	m.Init()

	rb := domain.Runbook{ID: "default.hello-world", Name: "hello-world", Version: "1.0.0"}
	cat := domain.Catalog{Name: "default", Path: "/tmp/default"}
	m.selected = &rb
	m.selCat = &cat
	m.output.SetCommand("dops run default.hello-world")
	m.output, _ = m.output.Update(output.ExecutionDoneMsg{LogPath: "/tmp/test.log"})

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := result.(App)
	view := app.View().Content
	lines := strings.Split(view, "\n")

	clickX, clickY := -1, -1
	for y, line := range lines {
		clean := appANSIPattern.ReplaceAllString(line, "")
		if idx := strings.Index(clean, "Saved to /tmp/test.log"); idx >= 0 {
			clickX = lipgloss.Width(clean[:idx]) + lipgloss.Width("Saved to ")/2
			clickY = y
			break
		}
	}

	if clickY == -1 {
		t.Fatal("footer text not found in rendered app view")
	}

	_, cmd := app.Update(tea.MouseClickMsg{X: clickX, Y: clickY, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatalf("click on rendered footer text at (%d,%d) should produce a copy command", clickX, clickY)
	}
}
