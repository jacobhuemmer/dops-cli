package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dops/internal/domain"
)

func setupRunEnv(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	dopsDir := filepath.Join(dir, ".dops")
	os.MkdirAll(filepath.Join(dopsDir, "keys"), 0o700)
	os.MkdirAll(filepath.Join(dopsDir, "catalogs", "default", "hello-world"), 0o755)

	configPath := filepath.Join(dopsDir, "config.json")
	cfg := &domain.Config{
		Theme:    "tokyonight",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskMedium},
		Catalogs: []domain.Catalog{
			{
				Name:   "default",
				Path:   filepath.Join(dopsDir, "catalogs", "default"),
				Active: true,
				Policy: domain.CatalogPolicy{MaxRiskLevel: domain.RiskMedium},
			},
		},
		Vars: domain.Vars{
			Global:  map[string]any{"region": "us-east-1"},
			Catalog: map[string]domain.CatalogVars{},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(configPath, data, 0o644)

	// Write runbook.yaml
	runbookYAML := `id: "default.hello-world"
name: "hello-world"
description: "Test runbook"
version: "1.0.0"
risk_level: "low"
script: "./script.sh"
parameters:
  - name: "greeting"
    type: "string"
    required: true
    scope: "global"
    secret: false
    description: "Greeting message"
  - name: "count"
    type: "integer"
    required: false
    scope: "runbook"
    secret: false
    default: 1
    description: "Number of times"
`
	os.WriteFile(
		filepath.Join(dopsDir, "catalogs", "default", "hello-world", "runbook.yaml"),
		[]byte(runbookYAML), 0o644,
	)

	// Write a simple script
	script := "#!/bin/sh\necho \"Hello $GREETING\"\n"
	scriptPath := filepath.Join(dopsDir, "catalogs", "default", "hello-world", "script.sh")
	os.WriteFile(scriptPath, []byte(script), 0o755)

	return dopsDir, configPath
}

func TestRun_UnknownID(t *testing.T) {
	dopsDir, _ := setupRunEnv(t)

	_, err := executeCmd([]string{"run", "unknown.runbook"}, dopsDir)
	if err == nil {
		t.Error("expected error for unknown runbook ID")
	}
}

func TestRun_DryRun(t *testing.T) {
	dopsDir, _ := setupRunEnv(t)

	out, err := executeCmd([]string{
		"run", "default.hello-world",
		"--param", "greeting=world",
		"--dry-run",
	}, dopsDir)
	if err != nil {
		t.Fatalf("run --dry-run: %v", err)
	}

	if len(out) == 0 {
		t.Error("dry-run produced no output")
	}
}

func TestRun_NoSave(t *testing.T) {
	dopsDir, configPath := setupRunEnv(t)

	before := readConfig(t, configPath)

	_, err := executeCmd([]string{
		"run", "default.hello-world",
		"--param", "greeting=world",
		"--no-save",
	}, dopsDir)
	if err != nil {
		t.Fatalf("run --no-save: %v", err)
	}

	after := readConfig(t, configPath)

	// Config should not have changed
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Error("config was modified despite --no-save")
	}
}

func TestRun_ParamOverride(t *testing.T) {
	dopsDir, configPath := setupRunEnv(t)

	_, err := executeCmd([]string{
		"run", "default.hello-world",
		"--param", "greeting=override-value",
	}, dopsDir)
	if err != nil {
		t.Fatalf("run with --param: %v", err)
	}

	// Verify param was saved to config
	cfg := readConfig(t, configPath)
	val, ok := cfg.Vars.Global["greeting"]
	if !ok {
		t.Fatal("greeting not saved to config")
	}
	if val != "override-value" {
		t.Errorf("greeting = %v, want override-value", val)
	}
}

func TestRun_Executes(t *testing.T) {
	dopsDir, _ := setupRunEnv(t)

	out, err := executeCmd([]string{
		"run", "default.hello-world",
		"--param", "greeting=world",
	}, dopsDir)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(out) == 0 {
		t.Error("run produced no output")
	}
}
