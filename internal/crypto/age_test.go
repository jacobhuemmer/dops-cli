package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgeEncrypter_EnsureKey(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")

	enc, err := NewAgeEncrypter(keysDir)
	if err != nil {
		t.Fatalf("NewAgeEncrypter: %v", err)
	}

	keysFile := filepath.Join(keysDir, "keys.txt")
	data, err := os.ReadFile(keysFile)
	if err != nil {
		t.Fatalf("read keys file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("keys file is empty")
	}

	// Creating again with same dir should not error or overwrite
	enc2, err := NewAgeEncrypter(keysDir)
	if err != nil {
		t.Fatalf("second NewAgeEncrypter: %v", err)
	}
	_ = enc2
	_ = enc
}

func TestAgeEncrypter_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")

	enc, err := NewAgeEncrypter(keysDir)
	if err != nil {
		t.Fatalf("NewAgeEncrypter: %v", err)
	}

	plaintext := "my-secret-token-abc123"

	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if !IsEncrypted(ciphertext) {
		t.Errorf("ciphertext %q does not start with age1", ciphertext)
	}

	if ciphertext == plaintext {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt = %q, want %q", decrypted, plaintext)
	}
}

func TestAgeEncrypter_DecryptPlaintext(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")

	enc, err := NewAgeEncrypter(keysDir)
	if err != nil {
		t.Fatalf("NewAgeEncrypter: %v", err)
	}

	_, err = enc.Decrypt("not-encrypted-at-all")
	if err == nil {
		t.Error("expected error decrypting plaintext")
	}
}

func TestAgeEncrypter_EmptyPlaintext(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")

	enc, err := NewAgeEncrypter(keysDir)
	if err != nil {
		t.Fatalf("NewAgeEncrypter: %v", err)
	}

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}

	if decrypted != "" {
		t.Errorf("Decrypt = %q, want empty", decrypted)
	}
}
