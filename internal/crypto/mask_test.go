package crypto

import (
	"dops/internal/domain"
	"testing"
)

func TestMaskSecrets(t *testing.T) {
	cfg := &domain.Config{
		Theme: "tokyonight",
		Vars: domain.Vars{
			Global: map[string]any{
				"region": "us-east-1",
				"token":  "age1qyqszqgpqyqszqgp",
			},
			Catalog: map[string]domain.CatalogVars{
				"default": {
					Vars: map[string]any{
						"namespace":  "platform",
						"secret_key": "age1abcdefgh",
					},
					Runbooks: map[string]map[string]any{
						"hello": {
							"flag":   true,
							"apikey": "age1zzzzz",
						},
					},
				},
			},
		},
	}

	masked := MaskSecrets(cfg)

	// Original should not be modified
	if cfg.Vars.Global["token"] != "age1qyqszqgpqyqszqgp" {
		t.Error("original config was modified")
	}

	// Plain values should pass through
	if masked.Vars.Global["region"] != "us-east-1" {
		t.Errorf("region = %v, want us-east-1", masked.Vars.Global["region"])
	}
	if masked.Vars.Catalog["default"].Vars["namespace"] != "platform" {
		t.Errorf("namespace = %v, want platform", masked.Vars.Catalog["default"].Vars["namespace"])
	}

	// Encrypted values should be masked
	if masked.Vars.Global["token"] != "****" {
		t.Errorf("token = %v, want ****", masked.Vars.Global["token"])
	}
	if masked.Vars.Catalog["default"].Vars["secret_key"] != "****" {
		t.Errorf("secret_key = %v, want ****", masked.Vars.Catalog["default"].Vars["secret_key"])
	}
	if masked.Vars.Catalog["default"].Runbooks["hello"]["apikey"] != "****" {
		t.Errorf("apikey = %v, want ****", masked.Vars.Catalog["default"].Runbooks["hello"]["apikey"])
	}

	// Non-string values should pass through
	if masked.Vars.Catalog["default"].Runbooks["hello"]["flag"] != true {
		t.Errorf("flag = %v, want true", masked.Vars.Catalog["default"].Runbooks["hello"]["flag"])
	}
}
