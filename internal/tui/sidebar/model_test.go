package sidebar

import (
	"dops/internal/catalog"
	"dops/internal/domain"
	"dops/internal/theme"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func sidebarTestStyles() *theme.Styles {
	return theme.BuildStyles(&theme.ResolvedTheme{
		Name: "test",
		Colors: map[string]string{
			"background": "#1a1b26", "backgroundPanel": "#1f2335", "backgroundElement": "#292e42",
			"text": "#c0caf5", "textMuted": "#565f89", "primary": "#7aa2f7",
			"border": "#565f89", "borderActive": "#7aa2f7",
			"success": "#9ece6a", "warning": "#e0af68", "error": "#f7768e",
			"risk.low": "#9ece6a", "risk.medium": "#e0af68", "risk.high": "#f7768e", "risk.critical": "#db4b4b",
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
		{
			Catalog: domain.Catalog{Name: "local"},
			Runbooks: []domain.Runbook{
				{ID: "local.drain-node", Name: "drain-node", RiskLevel: domain.RiskHigh},
			},
		},
	}
}

func pressKey(m Model, key string) (Model, tea.Cmd) {
	var msg tea.KeyPressMsg
	switch key {
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	case "enter":
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "left":
		msg = tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		msg = tea.KeyPressMsg{Code: tea.KeyRight}
	case "/":
		msg = tea.KeyPressMsg{Code: '/', Text: "/"}
	default:
		if len(key) == 1 {
			msg = tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
	}
	return m.Update(msg)
}

func TestSidebar_InitialSelection(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	cmd := m.Init()

	// Cursor starts at 0 (first item = default/ header)
	// But Init should select first runbook
	if cmd == nil {
		t.Fatal("Init should return a command")
	}

	msg := cmd()
	sel, ok := msg.(RunbookSelectedMsg)
	if !ok {
		t.Fatalf("expected RunbookSelectedMsg, got %T", msg)
	}
	if sel.Runbook.ID != "default.hello-world" {
		t.Errorf("initial selection = %q, want default.hello-world", sel.Runbook.ID)
	}
}

func TestSidebar_NavigateDown(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Visible: default/ (0), hello-world (1), rotate-tls (2), local/ (3), drain-node (4)
	// Cursor starts on hello-world (1)
	if sel := m.Selected(); sel == nil || sel.ID != "default.hello-world" {
		t.Errorf("initial: want hello-world, got %v", sel)
	}

	m, _ = pressKey(m, "down") // → rotate-tls
	if sel := m.Selected(); sel == nil || sel.ID != "default.rotate-tls" {
		t.Errorf("after 1 down: want rotate-tls, got %v", sel)
	}

	m, _ = pressKey(m, "down") // → local/ header
	if sel := m.Selected(); sel != nil {
		t.Error("on catalog header, Selected should be nil")
	}

	m, _ = pressKey(m, "down") // → drain-node
	if sel := m.Selected(); sel == nil || sel.ID != "local.drain-node" {
		t.Errorf("after 3 down: want drain-node, got %v", sel)
	}
}

func TestSidebar_NavigateUp(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Go to bottom
	for i := 0; i < 4; i++ {
		m, _ = pressKey(m, "down")
	}

	m, _ = pressKey(m, "up") // → local/ header
	m, _ = pressKey(m, "up") // → rotate-tls
	if sel := m.Selected(); sel == nil || sel.ID != "default.rotate-tls" {
		t.Errorf("want rotate-tls, got %v", sel)
	}

	m, _ = pressKey(m, "up") // → hello-world
	if sel := m.Selected(); sel == nil || sel.ID != "default.hello-world" {
		t.Errorf("want hello-world, got %v", sel)
	}
}

func TestSidebar_CollapseExpand(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Move cursor up to default/ header, then collapse
	m, _ = pressKey(m, "up")
	m, _ = pressKey(m, "enter")

	// default/ should be collapsed — its runbooks hidden
	vis := m.visible()
	for _, idx := range vis {
		e := m.entries[idx]
		if !e.isHeader && e.catalog.Name == "default" {
			t.Error("default runbooks should be hidden when collapsed")
		}
	}

	// Press Enter again to expand
	m, _ = pressKey(m, "enter")

	vis = m.visible()
	found := false
	for _, idx := range vis {
		e := m.entries[idx]
		if !e.isHeader && e.runbook.ID == "default.hello-world" {
			found = true
		}
	}
	if !found {
		t.Error("hello-world should be visible after expand")
	}
}

func TestSidebar_EnterOnRunbook_EmitsExecute(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Cursor starts on hello-world
	_, cmd := pressKey(m, "enter")

	if cmd == nil {
		t.Fatal("enter on runbook should emit a command")
	}

	msg := cmd()
	exec, ok := msg.(RunbookExecuteMsg)
	if !ok {
		t.Fatalf("expected RunbookExecuteMsg, got %T", msg)
	}
	if exec.Runbook.ID != "default.hello-world" {
		t.Errorf("execute runbook = %q", exec.Runbook.ID)
	}
}

func TestSidebar_LeftCollapses(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Cursor starts on hello-world — left jumps to header, left again collapses
	m, _ = pressKey(m, "left") // → default/ header
	m, _ = pressKey(m, "left") // collapse

	if !m.collapsed["default"] {
		t.Error("left on header should collapse catalog")
	}

	// Runbooks should be hidden
	vis := m.visibleRunbooks()
	for _, rb := range vis {
		if rb.ID == "default.hello-world" || rb.ID == "default.rotate-tls" {
			t.Error("default runbooks should be hidden")
		}
	}
}

func TestSidebar_RightExpands(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Move to header, collapse
	m, _ = pressKey(m, "left") // → header
	m, _ = pressKey(m, "left") // collapse
	if !m.collapsed["default"] {
		t.Fatal("should be collapsed")
	}

	// Right arrow expands
	m, _ = pressKey(m, "right")
	if m.collapsed["default"] {
		t.Error("right on collapsed header should expand")
	}
}

func TestSidebar_LeftOnRunbook_JumpsToParent(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Cursor starts on hello-world
	if sel := m.Selected(); sel == nil || sel.ID != "default.hello-world" {
		t.Fatal("should start on hello-world")
	}

	// Left arrow jumps to parent header
	m, _ = pressKey(m, "left")
	if m.Selected() != nil {
		t.Error("should be on header (nil selection)")
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (default/ header)", m.cursor)
	}
}

func TestSidebar_MouseClickRunbook(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Visible: default/ (0), hello-world (1), rotate-tls (2), local/ (3), drain-node (4)
	// yOffset=1 (border), so Y=1 = item 0, Y=3 = item 2 (rotate-tls)
	m, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft})

	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}
	sel := m.Selected()
	if sel == nil || sel.ID != "default.rotate-tls" {
		t.Errorf("selected = %v, want rotate-tls", sel)
	}
	if cmd == nil {
		t.Error("click on runbook should emit selection command")
	}
}

func TestSidebar_MouseClickHeader(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// yOffset=1 (border), so Y=1 = item 0 (default/ header)
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 1, Button: tea.MouseLeft})

	if !m.collapsed["default"] {
		t.Error("click on header should collapse catalog")
	}

	// Click again should expand
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 1, Button: tea.MouseLeft})

	if m.collapsed["default"] {
		t.Error("second click should expand catalog")
	}
}

func TestSidebar_DoubleClickExecutes(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Single click on Y=2 (hello-world) — selects
	m, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 2, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("single click should emit selection")
	}
	if _, ok := cmd().(RunbookExecuteMsg); ok {
		t.Error("single click should NOT execute")
	}

	// Second click on same Y immediately — double-click executes
	m, cmd = m.Update(tea.MouseClickMsg{X: 5, Y: 2, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("double click should emit a command")
	}
	msg := cmd()
	exec, ok := msg.(RunbookExecuteMsg)
	if !ok {
		t.Fatalf("double click should emit RunbookExecuteMsg, got %T", msg)
	}
	if exec.Runbook.ID != "default.hello-world" {
		t.Errorf("executed = %q, want default.hello-world", exec.Runbook.ID)
	}
}

func TestSidebar_MouseHover(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Hover over Y=2 (item 1 = hello-world)
	m, _ = m.Update(tea.MouseMotionMsg{X: 5, Y: 2})

	if m.hoverIdx != 1 {
		t.Errorf("hoverIdx = %d, want 1", m.hoverIdx)
	}

	// Hover outside bounds
	m, _ = m.Update(tea.MouseMotionMsg{X: 5, Y: 100})
	if m.hoverIdx != -1 {
		t.Errorf("hoverIdx = %d, want -1 (out of bounds)", m.hoverIdx)
	}
}

func TestSidebar_KeyboardClearsHover(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	// Set hover
	m, _ = m.Update(tea.MouseMotionMsg{X: 5, Y: 2})
	if m.hoverIdx != 1 {
		t.Fatal("hover should be set")
	}

	// Keyboard input clears hover
	m, _ = pressKey(m, "down")
	if m.hoverIdx != -1 {
		t.Errorf("hoverIdx = %d, want -1 after keyboard", m.hoverIdx)
	}
}

func TestSidebar_EmptyCatalogs(t *testing.T) {
	m := New(nil, 20, sidebarTestStyles())
	m.Init()

	if m.Selected() != nil {
		t.Error("expected nil selection with no catalogs")
	}

	m, _ = pressKey(m, "down")
	m, _ = pressKey(m, "up")
}

func TestSidebar_ViewNotEmpty(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	view := m.View()
	if len(view) == 0 {
		t.Error("View returned empty string")
	}
}

func TestSidebar_ViewShowsCollapseIndicator(t *testing.T) {
	m := New(testCatalogs(), 20, sidebarTestStyles())
	m.Init()

	view := m.View()
	if !containsStr(view, "▾") {
		t.Error("expanded catalog should show ▾")
	}

	m, _ = pressKey(m, "up")    // → default/ header
	m, _ = pressKey(m, "enter") // collapse default/

	view = m.View()
	if !containsStr(view, "▸") {
		t.Error("collapsed catalog should show ▸")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && findStr(s, sub)
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
