package domain

import (
	"testing"
	"time"
)

func TestNewExecutionRecord(t *testing.T) {
	r := NewExecutionRecord("infra.deploy", "deploy", "infra", ExecCLI)

	if r.ID == "" {
		t.Error("ID should be set")
	}
	if r.RunbookID != "infra.deploy" {
		t.Errorf("RunbookID = %q", r.RunbookID)
	}
	if r.RunbookName != "deploy" {
		t.Errorf("RunbookName = %q", r.RunbookName)
	}
	if r.CatalogName != "infra" {
		t.Errorf("CatalogName = %q", r.CatalogName)
	}
	if r.Status != ExecRunning {
		t.Errorf("Status = %q, want running", r.Status)
	}
	if r.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
	if r.Interface != ExecCLI {
		t.Errorf("Interface = %q, want cli", r.Interface)
	}
}

func TestExecutionRecord_Complete_Success(t *testing.T) {
	r := NewExecutionRecord("default.hello", "hello", "default", ExecTUI)
	time.Sleep(time.Millisecond) // ensure nonzero duration
	r.Complete(0, 10, "Done")

	if r.Status != ExecSuccess {
		t.Errorf("Status = %q, want success", r.Status)
	}
	if r.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", r.ExitCode)
	}
	if r.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
	if r.Duration == "" {
		t.Error("Duration should be set")
	}
	if r.OutputLines != 10 {
		t.Errorf("OutputLines = %d, want 10", r.OutputLines)
	}
	if r.OutputSummary != "Done" {
		t.Errorf("OutputSummary = %q", r.OutputSummary)
	}
}

func TestExecutionRecord_Complete_Failure(t *testing.T) {
	r := NewExecutionRecord("default.hello", "hello", "default", ExecMCP)
	r.Complete(1, 5, "error: timeout")

	if r.Status != ExecFailed {
		t.Errorf("Status = %q, want failed", r.Status)
	}
	if r.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", r.ExitCode)
	}
}

func TestExecutionRecord_Cancel(t *testing.T) {
	r := NewExecutionRecord("default.hello", "hello", "default", ExecWeb)
	r.Cancel()

	if r.Status != ExecCancelled {
		t.Errorf("Status = %q, want cancelled", r.Status)
	}
	if r.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1", r.ExitCode)
	}
	if r.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
}

func TestExecutionRecord_MaskSecrets(t *testing.T) {
	r := NewExecutionRecord("default.hello", "hello", "default", ExecCLI)
	r.Parameters = map[string]string{
		"endpoint": "https://api.example.com",
		"api_key":  "sk-12345",
		"region":   "us-east-1",
	}

	r.MaskSecrets([]string{"api_key"})

	if r.Parameters["api_key"] != "****" {
		t.Errorf("api_key = %q, want ****", r.Parameters["api_key"])
	}
	if r.Parameters["endpoint"] != "https://api.example.com" {
		t.Error("non-secret param should not be masked")
	}
	if r.Parameters["region"] != "us-east-1" {
		t.Error("non-secret param should not be masked")
	}
}

func TestExecutionRecord_MaskSecrets_NoParams(t *testing.T) {
	r := NewExecutionRecord("default.hello", "hello", "default", ExecCLI)
	// No panic with nil params
	r.MaskSecrets([]string{"api_key"})
}

func TestExecutionRecord_UniqueIDs(t *testing.T) {
	r1 := NewExecutionRecord("a.b", "b", "a", ExecCLI)
	r2 := NewExecutionRecord("a.b", "b", "a", ExecCLI)
	if r1.ID == r2.ID {
		t.Error("IDs should be unique")
	}
}
