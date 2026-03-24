package clipboard

import (
	"encoding/base64"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Copy returns a tea.Cmd that copies text to the clipboard.
// Uses BubbleTea's built-in SetClipboard (which uses OSC 52 in most cases).
// Falls back to a direct OSC 52 write if available.
func Copy(text string) tea.Cmd {
	return tea.SetClipboard(text)
}

// WriteOSC52 writes text to the clipboard using the OSC 52 escape sequence.
// This works in terminals that support it, including over SSH sessions.
func WriteOSC52(text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := fmt.Fprintf(os.Stdout, "\x1b]52;c;%s\x07", encoded)
	return err
}
