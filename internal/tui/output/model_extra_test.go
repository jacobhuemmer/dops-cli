package output

import (
	"dops/internal/testutil"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

func TestSetSize(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m.SetSize(100, 50)
	// No panic, just ensure it sets internal state.
	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestSetFocused(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m.SetFocused(true)
	if !m.focused {
		t.Error("should be focused")
	}
	m.SetFocused(false)
	if m.focused {
		t.Error("should not be focused")
	}
}

func TestCopyFlash(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	if m.CopyFlash() {
		t.Error("should not flash initially")
	}
	m.SetCopyFlash(true)
	if !m.CopyFlash() {
		t.Error("should flash after SetCopyFlash(true)")
	}
	m.SetCopyFlash(false)
	if m.CopyFlash() {
		t.Error("should not flash after SetCopyFlash(false)")
	}
}

func TestTryCopy(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	if !m.TryCopy() {
		t.Error("first TryCopy should succeed")
	}
	if !m.CopyFlash() {
		t.Error("TryCopy should enable flash")
	}
	if m.TryCopy() {
		t.Error("second TryCopy should fail (lock held)")
	}
	// Release lock
	m.SetCopyFlash(false)
	if !m.TryCopy() {
		t.Error("TryCopy after release should succeed")
	}
}

func TestTryLock(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	if !m.TryLock() {
		t.Error("first TryLock should succeed")
	}
	if m.CopyFlash() {
		t.Error("TryLock should NOT enable flash")
	}
	if m.TryLock() {
		t.Error("second TryLock should fail (lock held)")
	}
}

func TestSelection(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	sel := m.Selection()
	if sel.Active {
		t.Error("selection should not be active initially")
	}
}

func TestSetCopiedHeaderFooter(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m.SetCopiedHeader(true)
	if !m.copiedHeader {
		t.Error("copiedHeader should be true")
	}
	m.SetCopiedFooter(true)
	if !m.copiedFooter {
		t.Error("copiedFooter should be true")
	}
}

func TestClear(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m.SetCommand("test")
	m, _ = m.Update(OutputLineMsg{Text: "line1"})
	m.Clear()
	if m.Command() != "" {
		t.Error("command should be empty after clear")
	}
	if len(m.Lines()) != 0 {
		t.Error("lines should be empty after clear")
	}
}

func TestHasSession(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	if m.HasSession() {
		t.Error("should not have session initially")
	}
	m.SetCommand("test")
	if !m.HasSession() {
		t.Error("should have session after SetCommand")
	}
}

func TestTruncateLine(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	// truncateLine needs a line and maxWidth
	// Test with a short line that doesn't need truncation.
	result := m.truncateLine("hello", 10)
	if result != "hello" {
		t.Errorf("short line should not be truncated, got %q", result)
	}

	// Test with a line that needs truncation.
	long := "this is a very long line that should be truncated"
	result = m.truncateLine(long, 10)
	if len(result) > 15 { // some slack for ANSI
		t.Errorf("long line should be truncated, got len %d", len(result))
	}
}

func TestRenderHeader_WithCommand(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.SetCommand("dops run default.hello-world")

	c := m.resolveColors()
	header := m.renderHeader(70, c)
	if header == "" {
		t.Error("header should not be empty when command is set")
	}
}

func TestRenderHeader_WithCopiedBadge(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.SetCommand("dops run test")
	m.SetCopiedHeader(true)

	c := m.resolveColors()
	header := m.renderHeader(70, c)
	if header == "" {
		t.Error("header should not be empty")
	}
}

func TestRenderHeader_WithDoneBadge(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.SetCommand("dops run test")
	m, _ = m.Update(ExecutionDoneMsg{LogPath: "/tmp/test.log"})

	c := m.resolveColors()
	header := m.renderHeader(70, c)
	if header == "" {
		t.Error("header should not be empty after execution done")
	}
}

func TestVisibleLineTexts(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m, _ = m.Update(OutputLineMsg{Text: "line1"})
	m, _ = m.Update(OutputLineMsg{Text: "line2"})
	m, _ = m.Update(OutputLineMsg{Text: "line3"})

	texts := m.visibleLineTexts()
	// visibleLineTexts returns viewport-visible lines, which may include empty padding
	if len(texts) == 0 {
		t.Error("should have visible lines")
	}
}

func TestUpdate_WindowResize(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	// Just ensure no panic
}

func TestUpdate_MouseWheel(t *testing.T) {
	m := New(60, 20, testutil.TestStyles())
	m.SetCommand("test")
	m, _ = m.Update(OutputLineMsg{Text: "line"})
	// Send mouse wheel messages (no panic)
	m, _ = m.Update(tea.MouseWheelMsg{X: 10, Y: 5})
	_ = m
}

func TestHandleClick_HeaderRegion(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.SetCommand("dops run test")

	// Click at very top of output area — should be header
	copyText, region := m.HandleClick(10, 0, 70, 25)
	// May or may not match — just ensure no panic
	_ = copyText
	_ = region
}

func TestHandleClick_FooterRegion(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.SetCommand("dops run test")
	m, _ = m.Update(ExecutionDoneMsg{LogPath: "/tmp/test.log"})

	// Click at bottom of output area — should be footer
	_, region := m.HandleClick(10, 24, 70, 25)
	_ = region
}

func TestView_ContentWidthRespectsPadding(t *testing.T) {
	// View() uses contentWidth = max(1, m.width - padX*2) where padX=1.
	// If the negation is inverted (width+2 instead of width-2), the rendered
	// output would exceed the model's width.
	m := New(40, 20, testutil.TestStyles())
	m.SetCommand("echo hello")
	m, _ = m.Update(OutputLineMsg{Text: "test output line"})

	view := m.View()
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w > 40 {
			t.Errorf("line %d width %d exceeds model width 40: %q", i, w, line)
		}
	}
}

func TestRenderHeader_LongCommandWraps(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	// Command with --param boundaries that is wider than contentWidth.
	longCmd := "dops run default.deploy --param region=us-east-1 --param env=staging --param replicas=3 --param dry_run=true"
	m.SetCommand(longCmd)

	c := m.resolveColors()
	header := m.renderHeader(40, c)

	// Header should contain at least 2 lines (wrapped at --param boundary).
	lines := strings.Split(header, "\n")
	if len(lines) < 2 {
		t.Errorf("long command should wrap, got %d lines", len(lines))
	}
	// Each line should respect the width (accounting for ANSI codes).
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w > 42 { // 40 + 2 for "$ " prefix
			t.Errorf("header line %d width %d exceeds content width: %q", i, w, line)
		}
	}
}

func TestRenderSearchBar_Navigating_NoMatches(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.navigating = true
	m.matchCount = 0
	m.searchQuery = "test"

	contentStyle := lipgloss.NewStyle()
	successStyle := lipgloss.NewStyle()
	result := m.renderSearchBar(60, contentStyle, successStyle)

	// With matchCount=0, no search bar should be rendered.
	if result != nil {
		t.Errorf("navigating with matchCount=0 should return nil, got %d lines", len(result))
	}
}

func TestRenderSearchBar_Navigating_WithMatches(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.navigating = true
	m.matchCount = 3
	m.matchIndex = 1
	m.searchQuery = "error"

	contentStyle := lipgloss.NewStyle()
	successStyle := lipgloss.NewStyle()
	result := m.renderSearchBar(60, contentStyle, successStyle)

	if result == nil || len(result) != 2 {
		t.Fatalf("navigating with matches should return 2 lines, got %v", result)
	}
	if !strings.Contains(result[0], "[2/3]") {
		t.Errorf("search bar should show match position [2/3], got %q", result[0])
	}
}

func TestRenderSearchBar_Searching(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.searching = true
	m.searchQuery = "foo"

	contentStyle := lipgloss.NewStyle()
	result := m.renderSearchBar(60, contentStyle, contentStyle)

	if result == nil || len(result) != 2 {
		t.Fatalf("searching should return 2 lines, got %v", result)
	}
	if !strings.Contains(result[0], "Search: foo") {
		t.Errorf("search bar should show query, got %q", result[0])
	}
}

func TestRenderScrollbar_NoScroll(t *testing.T) {
	// When total lines <= visibleH, scrollbar should be plain track.
	m := New(80, 30, testutil.TestStyles())
	for i := 0; i < 5; i++ {
		m, _ = m.Update(OutputLineMsg{Text: fmt.Sprintf("line-%d", i)})
	}

	p := scrollbarParams{
		contentH:   10,
		yOffset:    0,
		visibleH:   5, // exactly equal to total lines
		trackStyle: lipgloss.NewStyle(),
		thumbStyle: lipgloss.NewStyle(),
	}
	bar := m.renderScrollbar(p)
	lines := strings.Split(bar, "\n")
	if len(lines) != 10 {
		t.Errorf("scrollbar should have %d lines, got %d", p.contentH, len(lines))
	}
}

func TestRenderScrollbar_WithScroll(t *testing.T) {
	// When total lines > visibleH, scrollbar should show thumb.
	m := New(80, 30, testutil.TestStyles())
	for i := 0; i < 50; i++ {
		m, _ = m.Update(OutputLineMsg{Text: fmt.Sprintf("line-%d", i)})
	}

	p := scrollbarParams{
		contentH:   20,
		yOffset:    10,
		visibleH:   20,
		trackStyle: lipgloss.NewStyle(),
		thumbStyle: lipgloss.NewStyle().Background(lipgloss.Color("7")),
	}
	bar := m.renderScrollbar(p)
	if bar == "" {
		t.Error("scrollbar should not be empty with scrollable content")
	}
}

func TestRenderFooter_WithLogPath(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.logPath = "/tmp/dops/test.log"

	c := m.resolveColors()
	footer := m.renderFooterSection(60, c)
	if !strings.Contains(footer, "test.log") {
		t.Error("footer should contain log path")
	}
}

func TestRenderFooter_TruncatesLongPath(t *testing.T) {
	m := New(80, 30, testutil.TestStyles())
	m.logPath = "/very/long/path/to/some/deeply/nested/directory/structure/logs/dops/default/hello-world/20260331-120000.log"

	c := m.resolveColors()
	footer := m.renderFooterSection(40, c)
	// Should truncate with ellipsis rather than exceed width.
	w := lipgloss.Width(footer)
	if w > 42 { // some tolerance for ANSI
		t.Errorf("footer width %d exceeds content width 40", w)
	}
}
