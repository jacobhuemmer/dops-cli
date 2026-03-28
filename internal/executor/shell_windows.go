//go:build windows

package executor

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// ShellFor returns the interpreter and arguments for the given script path.
// On Windows: tries pwsh (PowerShell Core), then powershell.exe.
// For .sh files: tries bash (Git Bash), then pwsh, then powershell.exe.
func ShellFor(scriptPath string) (string, []string) {
	ext := strings.ToLower(filepath.Ext(scriptPath))

	if ext == ".ps1" {
		return findPowerShell(), []string{"-NoProfile", "-NonInteractive", "-File", scriptPath}
	}

	// .sh or no extension: prefer bash (Git Bash), fall back to PowerShell.
	if bash, err := exec.LookPath("bash"); err == nil {
		return bash, []string{scriptPath}
	}
	return findPowerShell(), []string{"-NoProfile", "-NonInteractive", "-File", scriptPath}
}

// findPowerShell returns the best available PowerShell binary.
// Prefers pwsh (PowerShell Core 7+) over powershell.exe (Windows PowerShell 5.1).
func findPowerShell() string {
	if pwsh, err := exec.LookPath("pwsh"); err == nil {
		return pwsh
	}
	return "powershell.exe"
}
