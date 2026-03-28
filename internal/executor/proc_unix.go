//go:build !windows

package executor

import (
	"os/exec"
	"syscall"
)

// configurePlatformCancel sets up Unix process-group kill on context cancellation.
func configurePlatformCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
}
