package output

import (
	"dops/internal/theme"
	"strings"
	"testing"
)

func outputTestStyles() *theme.Styles {
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

func TestOutput_AppendLines(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m.SetCommand("dops run default.hello-world")

	m, _ = m.Update(OutputLineMsg{Text: "hello world", IsStderr: false})
	m, _ = m.Update(OutputLineMsg{Text: "error happened", IsStderr: true})

	if len(m.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(m.lines))
	}
	if m.lines[0].Text != "hello world" {
		t.Errorf("line 0 = %q", m.lines[0].Text)
	}
	if !m.lines[1].IsStderr {
		t.Error("line 1 should be stderr")
	}
}

func TestOutput_ExecutionDone(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m, _ = m.Update(ExecutionDoneMsg{LogPath: "/tmp/test.log"})

	if m.logPath != "/tmp/test.log" {
		t.Errorf("logPath = %q", m.logPath)
	}
}

func TestOutput_ViewShowsCommand(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m.SetCommand("dops run default.hello-world --param greeting=world")

	view := m.View()
	if !strings.Contains(view, "dops run default.hello-world") {
		t.Error("view should show command")
	}
}

func TestOutput_ViewShowsLogPath(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m, _ = m.Update(ExecutionDoneMsg{LogPath: "/tmp/2026.01.01-010102-default-hello.log"})

	view := m.View()
	if !strings.Contains(view, "/tmp/2026.01.01-010102-default-hello.log") {
		t.Error("view should show log path")
	}
}

func TestOutput_ViewShowsStdout(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m, _ = m.Update(OutputLineMsg{Text: "hello from script", IsStderr: false})

	view := m.View()
	if !strings.Contains(view, "hello from script") {
		t.Error("view should show stdout")
	}
}

func TestOutput_Clear(t *testing.T) {
	m := New(60, 20, outputTestStyles())
	m.SetCommand("old command")
	m, _ = m.Update(OutputLineMsg{Text: "old line"})
	m, _ = m.Update(ExecutionDoneMsg{LogPath: "/tmp/old.log"})

	m.Clear()

	if m.command != "" || len(m.lines) != 0 || m.logPath != "" {
		t.Error("Clear should reset all fields")
	}
}
