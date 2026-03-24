package output

// TextSelection tracks a click/drag text selection within the output log pane.
type TextSelection struct {
	Active  bool
	AnchorX int // X where mouse was initially pressed
	AnchorY int // Y where mouse was initially pressed (viewport-local row)
	FocusX  int // X of current/final drag position
	FocusY  int // Y of current/final drag position
}

// Bounds returns the selection normalized to reading order (top-left to bottom-right).
func (s *TextSelection) Bounds() (startX, startY, endX, endY int) {
	if s.AnchorY < s.FocusY || (s.AnchorY == s.FocusY && s.AnchorX <= s.FocusX) {
		return s.AnchorX, s.AnchorY, s.FocusX, s.FocusY
	}
	return s.FocusX, s.FocusY, s.AnchorX, s.AnchorY
}

// IsEmpty returns true if the selection covers zero characters (click without drag).
func (s *TextSelection) IsEmpty() bool {
	return s.AnchorX == s.FocusX && s.AnchorY == s.FocusY
}

// Reset clears the selection state.
func (s *TextSelection) Reset() {
	s.Active = false
	s.AnchorX = 0
	s.AnchorY = 0
	s.FocusX = 0
	s.FocusY = 0
}

// ExtractText returns the plain-text content from visible lines covered by the selection.
// visibleLines are the currently rendered log lines (already sliced by yOffset).
func (s *TextSelection) ExtractText(visibleLines []string) string {
	if !s.Active || s.IsEmpty() {
		return ""
	}

	startX, startY, endX, endY := s.Bounds()
	if startY < 0 || startY >= len(visibleLines) {
		return ""
	}
	if endY >= len(visibleLines) {
		endY = len(visibleLines) - 1
	}

	if startY == endY {
		// Single-line selection.
		line := visibleLines[startY]
		runes := []rune(line)
		lx := max(0, startX)
		rx := min(len(runes), endX+1)
		if lx >= rx {
			return ""
		}
		return string(runes[lx:rx])
	}

	// Multi-line selection.
	var result []string
	for y := startY; y <= endY; y++ {
		if y >= len(visibleLines) {
			break
		}
		line := visibleLines[y]
		runes := []rune(line)
		if y == startY {
			lx := max(0, startX)
			if lx < len(runes) {
				result = append(result, string(runes[lx:]))
			}
		} else if y == endY {
			rx := min(len(runes), endX+1)
			result = append(result, string(runes[:rx]))
		} else {
			result = append(result, line)
		}
	}

	return joinLines(result)
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for _, l := range lines[1:] {
		result += "\n" + l
	}
	return result
}
