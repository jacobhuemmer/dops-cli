package output

import "testing"

func TestSelection_Bounds(t *testing.T) {
	s := TextSelection{Active: true, AnchorX: 5, AnchorY: 2, FocusX: 10, FocusY: 4}
	sx, sy, ex, ey := s.Bounds()
	if sx != 5 || sy != 2 || ex != 10 || ey != 4 {
		t.Errorf("bounds = (%d,%d)-(%d,%d), want (5,2)-(10,4)", sx, sy, ex, ey)
	}

	// Reversed
	s = TextSelection{Active: true, AnchorX: 10, AnchorY: 4, FocusX: 5, FocusY: 2}
	sx, sy, ex, ey = s.Bounds()
	if sx != 5 || sy != 2 || ex != 10 || ey != 4 {
		t.Errorf("reversed bounds = (%d,%d)-(%d,%d), want (5,2)-(10,4)", sx, sy, ex, ey)
	}
}

func TestSelection_IsEmpty(t *testing.T) {
	s := TextSelection{Active: true, AnchorX: 3, AnchorY: 1, FocusX: 3, FocusY: 1}
	if !s.IsEmpty() {
		t.Error("same anchor and focus should be empty")
	}

	s.FocusX = 5
	if s.IsEmpty() {
		t.Error("different focus should not be empty")
	}
}

func TestSelection_ExtractText_SingleLine(t *testing.T) {
	lines := []string{"hello world", "second line"}
	s := TextSelection{Active: true, AnchorX: 0, AnchorY: 0, FocusX: 4, FocusY: 0}
	text := s.ExtractText(lines)
	if text != "hello" {
		t.Errorf("extracted = %q, want %q", text, "hello")
	}
}

func TestSelection_ExtractText_MultiLine(t *testing.T) {
	lines := []string{"first", "second", "third"}
	s := TextSelection{Active: true, AnchorX: 2, AnchorY: 0, FocusX: 2, FocusY: 2}
	text := s.ExtractText(lines)
	if text != "rst\nsecond\nthi" {
		t.Errorf("extracted = %q, want %q", text, "rst\nsecond\nthi")
	}
}

func TestSelection_ExtractText_Empty(t *testing.T) {
	lines := []string{"hello"}
	s := TextSelection{Active: true, AnchorX: 3, AnchorY: 0, FocusX: 3, FocusY: 0}
	text := s.ExtractText(lines)
	if text != "" {
		t.Errorf("empty selection should return empty, got %q", text)
	}
}
