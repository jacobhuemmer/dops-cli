package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogWriter_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	lw := NewLogWriter(dir)

	now := time.Date(2026, 1, 1, 1, 1, 2, 0, time.UTC)
	path, err := lw.Create("default", "hello-world", now)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	expected := filepath.Join(dir, "2026.01.01-010102-default-hello-world.log")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}

	// File should exist
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

func TestLogWriter_WriteAndClose(t *testing.T) {
	dir := t.TempDir()
	lw := NewLogWriter(dir)

	now := time.Now()
	path, err := lw.Create("default", "hello-world", now)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	lw.WriteLine("stdout line 1")
	lw.WriteLine("stderr: error here")
	lw.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "stdout line 1") {
		t.Error("log missing stdout line")
	}
	if !strings.Contains(content, "stderr: error here") {
		t.Error("log missing stderr line")
	}
}

func TestLogWriter_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	lw := NewLogWriter(dir)

	tests := []struct {
		catalog  string
		runbook  string
		time     time.Time
		contains string
	}{
		{"prod", "deploy", time.Date(2026, 3, 15, 14, 30, 45, 0, time.UTC), "2026.03.15-143045-prod-deploy.log"},
		{"local", "drain-node", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), "2026.12.31-235959-local-drain-node.log"},
	}

	for _, tt := range tests {
		path, err := lw.Create(tt.catalog, tt.runbook, tt.time)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		lw.Close()

		if !strings.HasSuffix(path, tt.contains) {
			t.Errorf("path %q should end with %q", path, tt.contains)
		}
	}
}
