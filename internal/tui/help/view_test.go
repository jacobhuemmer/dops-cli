package help

import (
	"dops/internal/testutil"
	"strings"
	"testing"
)

func TestRender_SidebarFocus(t *testing.T) {
	styles := testutil.TestStyles()
	result := Render(FocusSidebar, 60, styles)
	if result == "" {
		t.Error("Render should return non-empty content")
	}
	if !strings.Contains(result, "Sidebar") {
		t.Error("should show Sidebar section when focus is sidebar")
	}
	if strings.Contains(result, "Output") {
		t.Error("should NOT show Output section when focus is sidebar")
	}
}

func TestRender_OutputFocus(t *testing.T) {
	styles := testutil.TestStyles()
	result := Render(FocusOutput, 60, styles)
	if !strings.Contains(result, "Output") {
		t.Error("should show Output section when focus is output")
	}
	if strings.Contains(result, "Sidebar") {
		t.Error("should NOT show Sidebar section when focus is output")
	}
}

func TestRender_NilStyles(t *testing.T) {
	result := Render(FocusSidebar, 60, nil)
	if result == "" {
		t.Error("should render even with nil styles")
	}
}

func TestRender_NarrowWidth(t *testing.T) {
	styles := testutil.TestStyles()
	result := Render(FocusSidebar, 10, styles)
	if result == "" {
		t.Error("should render with narrow width")
	}
}

func TestRender_GlobalBindings(t *testing.T) {
	styles := testutil.TestStyles()
	result := Render(FocusSidebar, 80, styles)
	if !strings.Contains(result, "Global") {
		t.Error("should show Global section")
	}
	if !strings.Contains(result, "quit") {
		t.Error("should show quit binding")
	}
}
