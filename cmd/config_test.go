package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dops/internal/domain"
	"dops/internal/vault"
)

func setupTestEnv(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	dopsDir := filepath.Join(dir, ".dops")
	os.MkdirAll(dopsDir, 0o755)
	os.MkdirAll(filepath.Join(dopsDir, "keys"), 0o700)
	configPath := filepath.Join(dopsDir, "config.json")

	cfg := &domain.Config{
		Theme:    "tokyonight",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskMedium},
		Catalogs: []domain.Catalog{},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(configPath, data, 0o644)

	// Seed vault with initial vars.
	vaultPath := filepath.Join(dopsDir, "vault.json")
	keysDir := filepath.Join(dopsDir, "keys")
	vlt := vault.New(vaultPath, keysDir)
	vars := &domain.Vars{
		Global:  map[string]any{"region": "us-east-1"},
		Catalog: map[string]domain.CatalogVars{},
	}
	if err := vlt.Save(vars); err != nil {
		t.Fatalf("seed vault: %v", err)
	}

	return dopsDir, configPath
}

func readVault(t *testing.T, dopsDir string) *domain.Vars {
	t.Helper()
	vaultPath := filepath.Join(dopsDir, "vault.json")
	keysDir := filepath.Join(dopsDir, "keys")
	vlt := vault.New(vaultPath, keysDir)
	vars, err := vlt.Load()
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}
	return vars
}

func readConfig(t *testing.T, path string) *domain.Config {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	return &cfg
}

func executeCmd(args []string, dopsDir string) (string, error) {
	cmd := newRootCmd(dopsDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestConfigSet_Theme(t *testing.T) {
	dopsDir, configPath := setupTestEnv(t)

	_, err := executeCmd([]string{"config", "set", "theme=dracula"}, dopsDir)
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	cfg := readConfig(t, configPath)
	if cfg.Theme != "dracula" {
		t.Errorf("theme = %q, want dracula", cfg.Theme)
	}
}

func TestConfigSet_GlobalVar(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	_, err := executeCmd([]string{"config", "set", "vars.global.environment=production"}, dopsDir)
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	vars := readVault(t, dopsDir)
	if vars.Global["environment"] != "production" {
		t.Errorf("environment = %v, want production", vars.Global["environment"])
	}
	if vars.Global["region"] != "us-east-1" {
		t.Errorf("region = %v, want us-east-1", vars.Global["region"])
	}
}

func TestConfigSet_Secret(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	// Secrets are stored as plaintext inside the encrypted vault — no per-value encryption.
	_, err := executeCmd([]string{"config", "set", "vars.global.token=mysecret"}, dopsDir)
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	vars := readVault(t, dopsDir)
	val, ok := vars.Global["token"]
	if !ok {
		t.Fatal("token not found in vault")
	}
	if val != "mysecret" {
		t.Errorf("token = %v, want mysecret (plaintext inside encrypted vault)", val)
	}
}

func TestConfigGet(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	out, err := executeCmd([]string{"config", "get", "vars.global.region"}, dopsDir)
	if err != nil {
		t.Fatalf("config get: %v", err)
	}

	if out != "us-east-1\n" {
		t.Errorf("output = %q, want %q", out, "us-east-1\n")
	}
}

func TestConfigGet_NotFound(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	_, err := executeCmd([]string{"config", "get", "vars.global.nonexistent"}, dopsDir)
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestConfigUnset(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	_, err := executeCmd([]string{"config", "unset", "vars.global.region"}, dopsDir)
	if err != nil {
		t.Fatalf("config unset: %v", err)
	}

	vars := readVault(t, dopsDir)
	if _, ok := vars.Global["region"]; ok {
		t.Error("region should have been removed from vault")
	}
}

func TestConfigList(t *testing.T) {
	dopsDir, _ := setupTestEnv(t)

	out, err := executeCmd([]string{"config", "list"}, dopsDir)
	if err != nil {
		t.Fatalf("config list: %v", err)
	}

	if len(out) == 0 {
		t.Error("config list produced no output")
	}
}
