package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dops/internal/crypto"
	"dops/internal/domain"
)

// helper creates a vault backed by a temp directory with a real age key.
func newTestVault(t *testing.T) (*Vault, string) {
	t.Helper()
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	vaultPath := filepath.Join(dir, "vault.json")
	return New(vaultPath, keysDir), dir
}

func TestNew(t *testing.T) {
	v := New("/tmp/vault.json", "/tmp/keys")
	if v.path != "/tmp/vault.json" {
		t.Errorf("path = %q, want /tmp/vault.json", v.path)
	}
	if v.keysDir != "/tmp/keys" {
		t.Errorf("keysDir = %q, want /tmp/keys", v.keysDir)
	}
}

func TestExists_NoFile(t *testing.T) {
	v, _ := newTestVault(t)
	if v.Exists() {
		t.Error("Exists() should return false when vault file does not exist")
	}
}

func TestExists_WithFile(t *testing.T) {
	v, _ := newTestVault(t)
	if err := os.WriteFile(v.path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !v.Exists() {
		t.Error("Exists() should return true when vault file exists")
	}
}

func TestLoad_NonExistentFile_ReturnsEmptyVars(t *testing.T) {
	v, _ := newTestVault(t)
	vars, err := v.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if vars == nil {
		t.Fatal("Load() returned nil vars")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	v, _ := newTestVault(t)
	if err := os.WriteFile(v.path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := v.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid JSON")
	}
}

func TestLoad_UnsupportedVersion(t *testing.T) {
	v, _ := newTestVault(t)
	env := envelope{Version: 999, Data: "irrelevant"}
	data, _ := json.Marshal(env)
	if err := os.WriteFile(v.path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := v.Load()
	if err == nil {
		t.Fatal("Load() expected error for unsupported version")
	}
}

func TestLoad_CorruptedCiphertext(t *testing.T) {
	v, _ := newTestVault(t)
	// Version is correct but data is not valid age ciphertext.
	env := envelope{Version: 1, Data: "not-valid-ciphertext"}
	data, _ := json.Marshal(env)
	if err := os.WriteFile(v.path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := v.Load()
	if err == nil {
		t.Fatal("Load() expected error for corrupted ciphertext")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	v, _ := newTestVault(t)
	vars := &domain.Vars{
		Global: map[string]any{
			"region": "us-west-2",
		},
	}

	if err := v.Save(vars); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if !v.Exists() {
		t.Fatal("vault file should exist after Save")
	}

	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	val, ok := loaded.Global["region"]
	if !ok {
		t.Fatal("expected 'region' in loaded vars")
	}
	if val != "us-west-2" {
		t.Errorf("region = %v, want us-west-2", val)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	v, _ := newTestVault(t)
	vars := &domain.Vars{}

	if err := v.Save(vars); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	info, err := os.Stat(v.path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 600", perm)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	vaultPath := filepath.Join(dir, "nested", "deep", "vault.json")
	v := New(vaultPath, keysDir)

	vars := &domain.Vars{}
	if err := v.Save(vars); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if !v.Exists() {
		t.Error("vault file should exist after Save in nested dir")
	}
}

func TestSave_InvalidDirectory(t *testing.T) {
	// Use /dev/null as parent dir (not a directory) to force MkdirAll failure.
	v := New("/dev/null/impossible/vault.json", t.TempDir())
	vars := &domain.Vars{}
	err := v.Save(vars)
	if err == nil {
		t.Fatal("Save() expected error when directory creation fails")
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	v, _ := newTestVault(t)

	// Save initial data.
	vars1 := &domain.Vars{Global: map[string]any{"key": "value1"}}
	if err := v.Save(vars1); err != nil {
		t.Fatalf("Save() #1: %v", err)
	}

	// Save different data.
	vars2 := &domain.Vars{Global: map[string]any{"key": "value2"}}
	if err := v.Save(vars2); err != nil {
		t.Fatalf("Save() #2: %v", err)
	}

	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	if loaded.Global["key"] != "value2" {
		t.Errorf("key = %v, want value2", loaded.Global["key"])
	}
}

func TestLoad_ValidEnvelopeButBadDecryptedJSON(t *testing.T) {
	v, _ := newTestVault(t)

	// Encrypt something that is NOT valid JSON for domain.Vars.
	enc, err := crypto.NewAgeEncrypter(v.keysDir)
	if err != nil {
		t.Fatalf("NewAgeEncrypter: %v", err)
	}

	ciphertext, err := enc.Encrypt("not valid json for vars")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	env := envelope{Version: 1, Data: ciphertext}
	data, _ := json.Marshal(env)
	if err := os.WriteFile(v.path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = v.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid decrypted JSON")
	}
}

func TestLoad_UnreadableFile(t *testing.T) {
	v, _ := newTestVault(t)
	// Create the file but make it unreadable.
	if err := os.WriteFile(v.path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(v.path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(v.path, 0o600) })

	_, err := v.Load()
	if err == nil {
		t.Fatal("Load() expected error for unreadable file")
	}
}

func TestSave_ReadOnlyTargetDir(t *testing.T) {
	// Create a directory, save once, then make the directory read-only.
	// The second save should fail at either CreateTemp or Rename.
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	targetDir := filepath.Join(dir, "vault-dir")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	vaultPath := filepath.Join(targetDir, "vault.json")
	v := New(vaultPath, keysDir)

	vars := &domain.Vars{}
	// First save should succeed.
	if err := v.Save(vars); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// Make the directory read-only so temp file creation fails.
	if err := os.Chmod(targetDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(targetDir, 0o755) })

	err := v.Save(vars)
	if err == nil {
		t.Fatal("Save() expected error when directory is read-only")
	}
}

func TestSave_EmptyVars(t *testing.T) {
	v, _ := newTestVault(t)
	vars := &domain.Vars{}

	if err := v.Save(vars); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	if loaded.Global != nil && len(loaded.Global) > 0 {
		t.Error("expected empty global vars")
	}
}

func TestSave_WritesValidEnvelope(t *testing.T) {
	v, _ := newTestVault(t)
	vars := &domain.Vars{Global: map[string]any{"x": "y"}}

	if err := v.Save(vars); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	data, err := os.ReadFile(v.path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("envelope JSON parse: %v", err)
	}

	if env.Version != vaultFormatVersion {
		t.Errorf("version = %d, want %d", env.Version, vaultFormatVersion)
	}
	if env.Data == "" {
		t.Error("envelope data should not be empty")
	}
}
