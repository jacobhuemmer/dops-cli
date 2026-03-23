package executor

import (
	"bufio"
	"context"
	"fmt"
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		close(lines)
		errs <- fmt.Errorf("stdout pipe: %w", err)
		return lines, errs
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		close(lines)
		errs <- fmt.Errorf("stderr pipe: %w", err)
		return lines, errs
	}

	if err := cmd.Start(); err != nil {
		close(lines)
		errs <- fmt.Errorf("start script: %w", err)
		return lines, errs
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- OutputLine{Text: scanner.Text(), IsStderr: false}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			lines <- OutputLine{Text: scanner.Text(), IsStderr: true}
		}
	}()

	go func() {
		wg.Wait()
		close(lines)
		errs <- cmd.Wait()
	}()

	return lines, errs
}

var _ Runner = (*ScriptRunner)(nil)
