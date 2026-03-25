package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"dops/internal/crypto"
	"dops/internal/domain"
)

const currentVersion = 1

// envelope is the on-disk JSON format for vault.json.
type envelope struct {
	Version int    `json:"version"`
	Data    string `json:"data"`
}

// Vault manages encrypted parameter storage in vault.json.
type Vault struct {
	path    string
	keysDir string
}

// New creates a Vault backed by the given file path and age keys directory.
func New(path, keysDir string) *Vault {
	return &Vault{path: path, keysDir: keysDir}
}

// Exists returns true if vault.json exists on disk.
func (v *Vault) Exists() bool {
	_, err := os.Stat(v.path)
	return err == nil
}

// Load reads and decrypts vault.json, returning the stored Vars.
// Returns empty Vars if the file does not exist (first run).
func (v *Vault) Load() (*domain.Vars, error) {
	data, err := os.ReadFile(v.path)
	if os.IsNotExist(err) {
		return &domain.Vars{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse vault: %w", err)
	}

	if env.Version != currentVersion {
		return nil, fmt.Errorf("unsupported vault version: %d", env.Version)
	}

	enc, err := crypto.NewAgeEncrypter(v.keysDir)
	if err != nil {
		return nil, fmt.Errorf("init decryption: %w", err)
	}

	plaintext, err := enc.Decrypt(env.Data)
	if err != nil {
		return nil, fmt.Errorf("vault.json is corrupted or was modified outside dops: %w", err)
	}

	var vars domain.Vars
	if err := json.Unmarshal([]byte(plaintext), &vars); err != nil {
		return nil, fmt.Errorf("parse vault data: %w", err)
	}

	return &vars, nil
}

// Save encrypts and writes Vars to vault.json with 0600 permissions.
// Uses atomic write (temp file + rename) to prevent corruption.
func (v *Vault) Save(vars *domain.Vars) error {
	plaintext, err := json.Marshal(vars)
	if err != nil {
		return fmt.Errorf("marshal vault data: %w", err)
	}

	enc, err := crypto.NewAgeEncrypter(v.keysDir)
	if err != nil {
		return fmt.Errorf("init encryption: %w", err)
	}

	ciphertext, err := enc.Encrypt(string(plaintext))
	if err != nil {
		return fmt.Errorf("encrypt vault: %w", err)
	}

	env := envelope{
		Version: currentVersion,
		Data:    ciphertext,
	}

	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vault envelope: %w", err)
	}

	// Atomic write: write to temp file, then rename.
	dir := filepath.Dir(v.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".vault-*.tmp")
	if err != nil {
		return fmt.Errorf("create vault temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write vault temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close vault temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("set vault permissions: %w", err)
	}

	if err := os.Rename(tmpPath, v.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename vault: %w", err)
	}

	return nil
}
