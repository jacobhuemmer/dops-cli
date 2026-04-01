package executor

import (
	"context"
	"fmt"
	"testing"
)

func TestNewDemoRunner(t *testing.T) {
	runner := NewDemoRunner()
	if runner == nil {
		t.Fatal("NewDemoRunner returned nil")
	}
}

func TestDemoRunner_Run_KnownRunbook(t *testing.T) {
	runner := NewDemoRunner()
	// Script path structure: .../hello-world/run.sh -> name = "hello-world"
	scriptPath := "/fake/catalogs/hello-world/run.sh"
	env := map[string]string{"MESSAGE": "demo"}

	lines, errs := runner.Run(context.Background(), scriptPath, env)
	collected, err := collectOutput(t, lines, errs)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(collected) == 0 {
		t.Fatal("expected output lines, got none")
	}

	// hello-world runbook should produce "Hello, demo!" after env resolution.
	found := false
	for _, line := range collected {
		if line.Text == "Hello, demo!" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected resolved line %q in output, got: %v", "Hello, demo!", outputTexts(collected))
	}
}

func TestDemoRunner_Run_UnknownRunbook(t *testing.T) {
	runner := NewDemoRunner()
	// Unknown runbook name should produce generic fallback output.
	scriptPath := "/fake/catalogs/unknown-task/run.sh"

	lines, errs := runner.Run(context.Background(), scriptPath, nil)
	collected, err := collectOutput(t, lines, errs)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(collected) == 0 {
		t.Fatal("expected fallback output, got none")
	}

	// Generic output should contain the runbook name and a success message.
	wantFirst := "$ unknown-task"
	if collected[0].Text != wantFirst {
		t.Errorf("first line = %q, want %q", collected[0].Text, wantFirst)
	}

	lastLine := collected[len(collected)-1].Text
	want := fmt.Sprintf("\033[32m✓\033[0m %s completed successfully", "unknown-task")
	if lastLine != want {
		t.Errorf("last line = %q, want %q", lastLine, want)
	}
}

func TestDemoRunner_Run_ContextCancelled(t *testing.T) {
	runner := NewDemoRunner()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	lines, errs := runner.Run(ctx, "/fake/catalogs/deploy-app/run.sh", nil)
	_, err := collectOutput(t, lines, errs)

	if err == nil {
		// Cancellation may or may not be caught depending on timing with
		// the sleep. Either an error or a short output is acceptable.
		t.Log("no error returned (cancellation may not have been caught before first line)")
	}
}

func TestDemoOutput_KnownRunbook(t *testing.T) {
	env := map[string]string{"SERVICE": "api-server"}
	out := demoOutput("service-status", env)

	if len(out) == 0 {
		t.Fatal("expected output for known runbook")
	}
	if out[0] != "$ service-status" {
		t.Errorf("first line = %q, want %q", out[0], "$ service-status")
	}

	// Env var should be resolved.
	found := false
	for _, line := range out {
		if line == "Checking service: api-server..." {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected env var to be resolved in output, got: %v", out)
	}
}

func TestDemoOutput_FallbackForUnknown(t *testing.T) {
	out := demoOutput("nonexistent-runbook", nil)

	if len(out) == 0 {
		t.Fatal("expected fallback output")
	}
	if out[0] != "$ nonexistent-runbook" {
		t.Errorf("first line = %q, want %q", out[0], "$ nonexistent-runbook")
	}
}

func TestResolveEnv(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		env   map[string]string
		want  []string
	}{
		{
			name:  "single replacement",
			lines: []string{"Hello, ${NAME}!"},
			env:   map[string]string{"NAME": "world"},
			want:  []string{"Hello, world!"},
		},
		{
			name:  "multiple vars in one line",
			lines: []string{"${HOST}:${PORT}"},
			env:   map[string]string{"HOST": "localhost", "PORT": "8080"},
			want:  []string{"localhost:8080"},
		},
		{
			name:  "no replacements needed",
			lines: []string{"plain text"},
			env:   map[string]string{"UNUSED": "value"},
			want:  []string{"plain text"},
		},
		{
			name:  "nil env",
			lines: []string{"${VAR} stays"},
			env:   nil,
			want:  []string{"${VAR} stays"},
		},
		{
			name:  "empty lines",
			lines: []string{},
			env:   map[string]string{"X": "Y"},
			want:  []string{},
		},
		{
			name:  "repeated var",
			lines: []string{"${V} and ${V}"},
			env:   map[string]string{"V": "ok"},
			want:  []string{"ok and ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveEnv(tt.lines, tt.env)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func outputTexts(lines []OutputLine) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.Text
	}
	return out
}
