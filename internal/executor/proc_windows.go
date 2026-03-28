//go:build windows

package executor

import "os/exec"

// configurePlatformCancel is a no-op on Windows.
// exec.CommandContext already defaults Cancel to cmd.Process.Kill(),
// which maps to TerminateProcess on Windows.
func configurePlatformCancel(cmd *exec.Cmd) {}
