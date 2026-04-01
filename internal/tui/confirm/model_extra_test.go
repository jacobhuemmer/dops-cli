package confirm

import (
	"strings"
	"testing"

	"dops/internal/domain"

	tea "charm.land/bubbletea/v2"
)

func newModel(risk domain.RiskLevel) Model {
	return New(Params{
		Runbook: domain.Runbook{
			ID:        "infra.deploy-app",
			Name:      "deploy-app",
			RiskLevel: risk,
		},
		Catalog:  domain.Catalog{Name: "infra"},
		Resolved: map[string]string{"env": "prod"},
		Width:    80,
	})
}

func TestNew_Defaults(t *testing.T) {
	m := newModel(domain.RiskHigh)

	if m.runbook.ID != "infra.deploy-app" {
		t.Errorf("runbook.ID = %q, want infra.deploy-app", m.runbook.ID)
	}
	if m.catalog.Name != "infra" {
		t.Errorf("catalog.Name = %q, want infra", m.catalog.Name)
	}
	if m.risk != domain.RiskHigh {
		t.Errorf("risk = %q, want high", m.risk)
	}
	if m.cursor != 1 {
		t.Error("cursor should default to 1 (No)")
	}
	if m.input != "" {
		t.Error("input should default to empty")
	}
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.params["env"] != "prod" {
		t.Errorf("params[env] = %q, want prod", m.params["env"])
	}
}

func TestInit_ReturnsNil(t *testing.T) {
	m := newModel(domain.RiskHigh)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// --- High-risk toggle tests ---

func TestHighRisk_LeftArrowTogglesToYes(t *testing.T) {
	m := newModel(domain.RiskHigh)
	if m.cursor != 1 {
		t.Fatal("precondition: cursor should start at 1")
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.cursor != 0 {
		t.Error("left arrow should move cursor to 0 (Yes)")
	}
}

func TestHighRisk_RightArrowTogglesToNo(t *testing.T) {
	m := newModel(domain.RiskHigh)
	m.cursor = 0
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.cursor != 1 {
		t.Error("right arrow should move cursor to 1 (No)")
	}
}

func TestHighRisk_TabTogglesToYes(t *testing.T) {
	m := newModel(domain.RiskHigh)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.cursor != 0 {
		t.Error("tab should move cursor to 0 (Yes)")
	}
}

func TestHighRisk_H_TogglesToYes(t *testing.T) {
	m := newModel(domain.RiskHigh)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	if m.cursor != 0 {
		t.Error("h should move cursor to 0 (Yes)")
	}
}

func TestHighRisk_L_TogglesToNo(t *testing.T) {
	m := newModel(domain.RiskHigh)
	m.cursor = 0
	m, _ = m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	if m.cursor != 1 {
		t.Error("l should move cursor to 1 (No)")
	}
}

func TestHighRisk_UpperY_Accepts(t *testing.T) {
	m := newModel(domain.RiskHigh)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'Y', Text: "Y"})
	if cmd == nil {
		t.Fatal("Y should produce a command")
	}
	if _, ok := cmd().(AcceptMsg); !ok {
		t.Fatal("expected AcceptMsg")
	}
}

func TestHighRisk_UpperN_Cancels(t *testing.T) {
	m := newModel(domain.RiskHigh)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})
	if cmd == nil {
		t.Fatal("N should produce a command")
	}
	if _, ok := cmd().(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestHighRisk_EnterOnYes_Accepts(t *testing.T) {
	m := newModel(domain.RiskHigh)
	m.cursor = 0
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on Yes should produce a command")
	}
	msg := cmd()
	a, ok := msg.(AcceptMsg)
	if !ok {
		t.Fatalf("expected AcceptMsg, got %T", msg)
	}
	if a.Runbook.ID != "infra.deploy-app" {
		t.Errorf("AcceptMsg.Runbook.ID = %q", a.Runbook.ID)
	}
	if a.Catalog.Name != "infra" {
		t.Errorf("AcceptMsg.Catalog.Name = %q", a.Catalog.Name)
	}
	if a.Params["env"] != "prod" {
		t.Errorf("AcceptMsg.Params[env] = %q", a.Params["env"])
	}
}

func TestHighRisk_EnterOnNo_Cancels(t *testing.T) {
	m := newModel(domain.RiskHigh)
	// cursor defaults to 1 (No)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on No should produce a command")
	}
	if _, ok := cmd().(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

// --- Critical-risk tests ---

func TestCritical_Backspace(t *testing.T) {
	m := newModel(domain.RiskCritical)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	if m.input != "ab" {
		t.Fatalf("input = %q, want ab", m.input)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.input != "a" {
		t.Errorf("after backspace input = %q, want a", m.input)
	}
}

func TestCritical_BackspaceOnEmpty(t *testing.T) {
	m := newModel(domain.RiskCritical)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.input != "" {
		t.Errorf("backspace on empty should keep input empty, got %q", m.input)
	}
}

func TestCritical_EscapeCancels(t *testing.T) {
	m := newModel(domain.RiskCritical)
	// Type some input then escape
	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("escape should produce a command")
	}
	if _, ok := cmd().(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestCritical_WhitespaceAroundIDAccepted(t *testing.T) {
	m := newModel(domain.RiskCritical)
	// Type " infra.deploy-app " with leading/trailing space
	text := " infra.deploy-app "
	for _, ch := range text {
		m, _ = m.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter with correct ID (whitespace-padded) should accept")
	}
	if _, ok := cmd().(AcceptMsg); !ok {
		t.Fatal("expected AcceptMsg")
	}
}

// --- Low/medium (default) risk tests ---

func TestLowRisk_EnterImmediatelyAccepts(t *testing.T) {
	m := newModel(domain.RiskLow)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on low risk should immediately accept")
	}
	if _, ok := cmd().(AcceptMsg); !ok {
		t.Fatal("expected AcceptMsg")
	}
}

func TestMediumRisk_EnterImmediatelyAccepts(t *testing.T) {
	m := newModel(domain.RiskMedium)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on medium risk should immediately accept")
	}
	if _, ok := cmd().(AcceptMsg); !ok {
		t.Fatal("expected AcceptMsg")
	}
}

func TestLowRisk_EscapeCancels(t *testing.T) {
	m := newModel(domain.RiskLow)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("escape should produce a command")
	}
	if _, ok := cmd().(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

// --- Non-key messages are no-ops ---

func TestUpdate_IgnoresNonKeyMsg(t *testing.T) {
	m := newModel(domain.RiskHigh)
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if cmd != nil {
		t.Error("non-key message should not produce a command")
	}
}

// --- View tests ---

func TestView_HighRisk_ContainsElements(t *testing.T) {
	m := newModel(domain.RiskHigh)
	v := m.View()

	checks := []string{
		"dops run infra.deploy-app",
		"HIGH RISK",
		"Runbook: infra.deploy-app",
		"Confirm execution?",
		"Yes",
		"No",
		"toggle",
		"esc cancel",
	}
	for _, want := range checks {
		if !strings.Contains(v, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestView_CriticalRisk_ContainsElements(t *testing.T) {
	m := newModel(domain.RiskCritical)
	// Type partial input
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	v := m.View()

	checks := []string{
		"dops run infra.deploy-app",
		"CRITICAL RISK",
		"Runbook: infra.deploy-app",
		"Type the runbook ID to confirm",
		"infra.deploy-app",
		"> a",
		"enter confirm",
		"esc cancel",
	}
	for _, want := range checks {
		if !strings.Contains(v, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestView_LowRisk_NoRiskWarning(t *testing.T) {
	m := newModel(domain.RiskLow)
	v := m.View()

	if strings.Contains(v, "RISK") {
		t.Error("low risk View() should not show a RISK warning")
	}
	if !strings.Contains(v, "dops run infra.deploy-app") {
		t.Error("View() should show the command header")
	}
}

// --- isConfirmed tests ---

func TestIsConfirmed_HighRisk_AlwaysFalse(t *testing.T) {
	m := newModel(domain.RiskHigh)
	if m.isConfirmed() {
		t.Error("high risk isConfirmed() should always return false")
	}
}

func TestIsConfirmed_LowRisk_AlwaysTrue(t *testing.T) {
	m := newModel(domain.RiskLow)
	if !m.isConfirmed() {
		t.Error("low risk isConfirmed() should return true")
	}
}

func TestIsConfirmed_Critical_MatchesID(t *testing.T) {
	m := newModel(domain.RiskCritical)
	if m.isConfirmed() {
		t.Error("empty input should not confirm")
	}
	m.input = "infra.deploy-app"
	if !m.isConfirmed() {
		t.Error("matching ID should confirm")
	}
	m.input = "wrong"
	if m.isConfirmed() {
		t.Error("wrong input should not confirm")
	}
}
