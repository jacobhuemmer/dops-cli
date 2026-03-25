package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dops/internal/domain"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.json")
	keysDir := filepath.Join(dir, "keys")

	v := New(vaultPath, keysDir)

	vars := &domain.Vars{
		Global: map[string]any{
			"region":    "us-east-1",
			"api_token": "sk-secret-123",
		},
		Catalog: map[string]domain.CatalogVars{
			"infra": {
				Vars: map[string]any{"namespace": "production"},
				Runbooks: map[string]map[string]any{
					"scale": {"replicas": "3"},
				},
			},
		},
	}

	if err := v.Save(vars); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Global["region"] != "us-east-1" {
		t.Errorf("global.region = %v, want us-east-1", loaded.Global["region"])
	}
	if loaded.Global["api_token"] != "sk-secret-123" {
		t.Errorf("global.api_token = %v, want sk-secret-123", loaded.Global["api_token"])
	}
	cat, ok := loaded.Catalog["infra"]
	if !ok {
		t.Fatal("missing catalog infra")
	}
	if cat.Vars["namespace"] != "production" {
		t.Errorf("catalog.infra.namespace = %v, want production", cat.Vars["namespace"])
	}
	if cat.Runbooks["scale"]["replicas"] != "3" {
		t.Errorf("catalog.infra.runbooks.scale.replicas = %v, want 3", cat.Runbooks["scale"]["replicas"])
	}
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	v := New(filepath.Join(dir, "vault.json"), filepath.Join(dir, "keys"))

	vars, err := v.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if vars.Global != nil || vars.Catalog != nil {
		t.Errorf("expected empty vars, got %+v", vars)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.json")
	keysDir := filepath.Join(dir, "keys")

	v := New(vaultPath, keysDir)

	if v.Exists() {
		t.Error("expected Exists()=false before save")
	}

	if err := v.Save(&domain.Vars{}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !v.Exists() {
		t.Error("expected Exists()=true after save")
	}
}

func TestTamperDetection(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.json")
	keysDir := filepath.Join(dir, "keys")

	v := New(vaultPath, keysDir)

	if err := v.Save(&domain.Vars{Global: map[string]any{"key": "value"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Tamper with the encrypted data.
	data, _ := os.ReadFile(vaultPath)
	var env envelope
	json.Unmarshal(data, &env)

	// Flip a character in the ciphertext.
	runes := []rune(env.Data)
	if len(runes) > 10 {
		runes[10] = 'X'
	}
	env.Data = string(runes)

	tampered, _ := json.Marshal(env)
	os.WriteFile(vaultPath, tampered, 0o600)

	_, err := v.Load()
	if err == nil {
		t.Fatal("expected error on tampered vault, got nil")
	}
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.json")
	keysDir := filepath.Join(dir, "keys")

	v := New(vaultPath, keysDir)

	if err := v.Save(&domain.Vars{}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("permissions = %o, want 600", perm)
	}
}
