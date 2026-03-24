package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ScriptRunner struct{}

func NewScriptRunner() *ScriptRunner {
	return &ScriptRunner{}
}

func (r *ScriptRunner) Run(ctx context.Context, scriptPath string, env map[string]string) (<-chan OutputLine, <-chan error) {
	lines := make(chan OutputLine, 100)
	errs := make(chan error, 1)

	cmd := exec.CommandContext(ctx, "sh", scriptPath)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	// Use io.Pipe instead of cmd.StdoutPipe/StderrPipe for immediate
	// line delivery without OS pipe buffering.
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	if err := cmd.Start(); err != nil {
		close(lines)
		errs <- fmt.Errorf("start script: %w", err)
		return lines, errs
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutR)
		for scanner.Scan() {
			lines <- OutputLine{Text: scanner.Text(), IsStderr: false}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrR)
		for scanner.Scan() {
			lines <- OutputLine{Text: scanner.Text(), IsStderr: true}
		}
	}()

	go func() {
		err := cmd.Wait()
		// Close pipe writers so scanners finish.
		stdoutW.Close()
		stderrW.Close()
		wg.Wait()
		close(lines)
		errs <- err
	}()

	return lines, errs
}

var _ Runner = (*ScriptRunner)(nil)
