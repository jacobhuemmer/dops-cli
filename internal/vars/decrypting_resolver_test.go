package vars

import (
	"dops/internal/domain"
	"fmt"
	"testing"
)

type fakeEncrypter struct {
	decrypted map[string]string
}

func (f *fakeEncrypter) Encrypt(plaintext string) (string, error) {
	return "age1" + plaintext, nil
}

func (f *fakeEncrypter) Decrypt(ciphertext string) (string, error) {
	if val, ok := f.decrypted[ciphertext]; ok {
		return val, nil
	}
	return "", fmt.Errorf("cannot decrypt %q", ciphertext)
}

func TestDecryptingVarResolver_Resolve(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global: map[string]any{
				"region": "us-east-1",
				"token":  "age1encrypted-token",
			},
			Catalog: map[string]domain.CatalogVars{},
		},
	}

	params := []domain.Parameter{
		{Name: "region", Scope: "global"},
		{Name: "token", Scope: "global"},
	}

	enc := &fakeEncrypter{
		decrypted: map[string]string{
			"age1encrypted-token": "decrypted-secret",
		},
	}

	inner := NewDefaultResolver()
	resolver := NewDecryptingResolver(inner, enc)

	result := resolver.Resolve(cfg, "default", "hello-world", params)

	if result["region"] != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", result["region"])
	}

	if result["token"] != "decrypted-secret" {
		t.Errorf("token = %q, want decrypted-secret", result["token"])
	}
}

func TestDecryptingVarResolver_SkipsPlainValues(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global:  map[string]any{"region": "us-east-1"},
			Catalog: map[string]domain.CatalogVars{},
		},
	}

	params := []domain.Parameter{
		{Name: "region", Scope: "global"},
	}

	enc := &fakeEncrypter{decrypted: map[string]string{}}
	inner := NewDefaultResolver()
	resolver := NewDecryptingResolver(inner, enc)

	result := resolver.Resolve(cfg, "default", "hello-world", params)

	if result["region"] != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", result["region"])
	}
}
