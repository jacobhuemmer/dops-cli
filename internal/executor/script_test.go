package executor

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func collectOutput(t *testing.T, lines <-chan OutputLine, errs <-chan error) ([]OutputLine, error) {
	t.Helper()
	var collected []OutputLine
	for line := range lines {
		collected = append(collected, line)
	}
	err := <-errs
	return collected, err
}

func TestScriptRunner_BasicExecution(t *testing.T) {
	runner := NewScriptRunner()
	env := map[string]string{
		"GREETING": "world",
		"COUNT":    "3",
	}

	lines, errs := runner.Run(context.Background(), testdataPath("echo.sh"), env)
	collected, err := collectOutput(t, lines, errs)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(collected) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(collected))
	}

	if collected[0].Text != "hello world" {
		t.Errorf("line 0 = %q, want %q", collected[0].Text, "hello world")
	}
	if collected[0].IsStderr {
		t.Error("line 0 should be stdout")
	}
	if collected[1].Text != "count: 3" {
		t.Errorf("line 1 = %q, want %q", collected[1].Text, "count: 3")
	}
}

func TestScriptRunner_StdoutStderrSeparated(t *testing.T) {
	runner := NewScriptRunner()

	lines, errs := runner.Run(context.Background(), testdataPath("mixed.sh"), nil)
	collected, err := collectOutput(t, lines, errs)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var hasStdout, hasStderr bool
	for _, line := range collected {
		if !line.IsStderr && line.Text == "stdout line 1" {
			hasStdout = true
		}
		if line.IsStderr && line.Text == "stderr line 1" {
			hasStderr = true
		}
	}

	if !hasStdout {
		t.Error("missing stdout line")
	}
	if !hasStderr {
		t.Error("missing stderr line")
	}
}

func TestScriptRunner_FailingScript(t *testing.T) {
	runner := NewScriptRunner()

	lines, errs := runner.Run(context.Background(), testdataPath("fail.sh"), nil)
	_, err := collectOutput(t, lines, errs)

	if err == nil {
		t.Error("expected error for failing script")
	}
}

func TestScriptRunner_ContextCancel(t *testing.T) {
	runner := NewScriptRunner()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	lines, errs := runner.Run(ctx, testdataPath("echo.sh"), map[string]string{"GREETING": "x"})
	_, err := collectOutput(t, lines, errs)

	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
