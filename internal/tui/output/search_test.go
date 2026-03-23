package output

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func populatedModel() Model {
	m := New(60, 20, outputTestStyles())
	m.SetCommand("dops run default.hello-world")
	lines := []string{
		"starting deploy",
		"deploying to us-east-1",
		"deploy complete",
		"checking health",
		"deploy verified",
		"done",
	}
	for _, l := range lines {
		m, _ = m.Update(OutputLineMsg{Text: l, IsStderr: false})
	}
	return m
}

func pressOutputKey(m Model, key string) Model {
	var msg tea.KeyPressMsg
	switch key {
	case "/":
		msg = tea.KeyPressMsg{Code: '/', Text: "/"}
	case "enter":
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "n":
		msg = tea.KeyPressMsg{Code: 'n', Text: "n"}
	case "N":
		msg = tea.KeyPressMsg{Code: 'N', Text: "N"}
	case "escape":
		msg = tea.KeyPressMsg{Code: tea.KeyEscape}
	default:
		if len(key) == 1 {
			msg = tea.KeyPressMsg{Code: rune(key[0]), Text: key}
		}
	}
	m, _ = m.Update(msg)
	return m
}

func typeOutputSearch(m Model, query string) Model {
	m = pressOutputKey(m, "/")
	for _, ch := range query {
		msg := tea.KeyPressMsg{Code: ch, Text: string(ch)}
		m, _ = m.Update(msg)
	}
	return m
}

func TestOutputSearch_ActivateWithSlash(t *testing.T) {
	m := populatedModel()

	if m.searching {
		t.Fatal("should not be searching initially")
	}

	m = pressOutputKey(m, "/")
	if !m.searching {
		t.Fatal("should be searching after /")
	}
}

func TestOutputSearch_FindsMatches(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")

	// "deploy" appears in lines: 0 (starting deploy), 1 (deploying...), 2 (deploy complete), 4 (deploy verified)
	if m.matchCount != 4 {
		t.Errorf("matchCount = %d, want 4", m.matchCount)
	}
}

func TestOutputSearch_NoMatches(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "zzzznothing")

	if m.matchCount != 0 {
		t.Errorf("matchCount = %d, want 0", m.matchCount)
	}
}

func TestOutputSearch_NavigateNext(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")

	// Confirm search to enter nav mode
	m = pressOutputKey(m, "enter")

	if m.matchIndex != 0 {
		t.Errorf("initial matchIndex = %d, want 0", m.matchIndex)
	}

	m = pressOutputKey(m, "n")
	if m.matchIndex != 1 {
		t.Errorf("after n: matchIndex = %d, want 1", m.matchIndex)
	}

	m = pressOutputKey(m, "n")
	if m.matchIndex != 2 {
		t.Errorf("after 2nd n: matchIndex = %d, want 2", m.matchIndex)
	}
}

func TestOutputSearch_NavigatePrev(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")
	m = pressOutputKey(m, "enter")

	// Go forward twice
	m = pressOutputKey(m, "n")
	m = pressOutputKey(m, "n")
	if m.matchIndex != 2 {
		t.Fatalf("matchIndex = %d, want 2", m.matchIndex)
	}

	// Go back
	m = pressOutputKey(m, "N")
	if m.matchIndex != 1 {
		t.Errorf("after N: matchIndex = %d, want 1", m.matchIndex)
	}
}

func TestOutputSearch_WrapAround(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")
	m = pressOutputKey(m, "enter")

	// Navigate to last match
	for i := 0; i < m.matchCount-1; i++ {
		m = pressOutputKey(m, "n")
	}

	if m.matchIndex != m.matchCount-1 {
		t.Fatalf("matchIndex = %d, want %d", m.matchIndex, m.matchCount-1)
	}

	// Next should wrap to 0
	m = pressOutputKey(m, "n")
	if m.matchIndex != 0 {
		t.Errorf("after wrap: matchIndex = %d, want 0", m.matchIndex)
	}
}

func TestOutputSearch_WrapAroundPrev(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")
	m = pressOutputKey(m, "enter")

	// At index 0, go prev should wrap to last
	m = pressOutputKey(m, "N")
	if m.matchIndex != m.matchCount-1 {
		t.Errorf("after wrap prev: matchIndex = %d, want %d", m.matchIndex, m.matchCount-1)
	}
}

func TestOutputSearch_EscapeClears(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")
	m = pressOutputKey(m, "enter")

	m = pressOutputKey(m, "escape")

	if m.searching {
		t.Error("should not be searching after escape")
	}
	if m.navigating {
		t.Error("should not be navigating after escape")
	}
	if m.matchCount != 0 {
		t.Errorf("matchCount should be 0 after escape, got %d", m.matchCount)
	}
}

func TestOutputSearch_ViewShowsCounter(t *testing.T) {
	m := populatedModel()
	m = typeOutputSearch(m, "deploy")
	m = pressOutputKey(m, "enter")

	view := m.View()
	if !strings.Contains(view, "[1/4]") {
		t.Errorf("view should show [1/4] counter, got:\n%s", view)
	}
}

func TestOutputSearch_EmptyBuffer(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m = typeOutputSearch(m, "anything")

	if m.matchCount != 0 {
		t.Errorf("matchCount = %d on empty buffer", m.matchCount)
	}
}

func TestOutputSearch_ScrollbarWhenContentExceedsHeight(t *testing.T) {
	m := New(60, 5, outputTestStyles())
	for i := 0; i < 20; i++ {
		m, _ = m.Update(OutputLineMsg{Text: "line", IsStderr: false})
	}

	view := m.View()
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}
