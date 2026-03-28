//go:build !windows

package executor

import (
	"path/filepath"
	"strings"
)

// ShellFor returns the interpreter and arguments for the given script path.
// On Unix: .ps1 files use pwsh, everything else uses sh.
func ShellFor(scriptPath string) (string, []string) {
	ext := strings.ToLower(filepath.Ext(scriptPath))
	if ext == ".ps1" {
		return "pwsh", []string{"-NoProfile", "-NonInteractive", "-File", scriptPath}
	}
	return "sh", []string{scriptPath}
}
