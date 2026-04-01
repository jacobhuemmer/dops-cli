package vars

import (
	"dops/internal/domain"
	"testing"
)

func TestDefaultVarResolver_Resolve(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global: map[string]any{
				"region":      "us-east-1",
				"environment": "production",
			},
			Catalog: map[string]domain.CatalogVars{
				"default": {
					Vars: map[string]any{
						"namespace": "platform",
						"region":    "eu-west-1", // overrides global
					},
					Runbooks: map[string]map[string]any{
						"hello-world": {
							"dry_run": true,
							"region":  "ap-south-1", // overrides catalog
						},
					},
				},
			},
		},
	}

	params := []domain.Parameter{
		{Name: "region", Scope: "global"},
		{Name: "environment", Scope: "global"},
		{Name: "namespace", Scope: "catalog"},
		{Name: "dry_run", Scope: "runbook"},
	}

	resolver := NewDefaultResolver()
	result := resolver.Resolve(cfg, "default", "hello-world", params)

	tests := []struct {
		key  string
		want string
	}{
		{"region", "ap-south-1"},
		{"environment", "production"},
		{"namespace", "platform"},
		{"dry_run", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := result[tt.key]
			if !ok {
				t.Fatalf("key %q not found in result", tt.key)
			}
			if got != tt.want {
				t.Errorf("result[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestDefaultVarResolver_EmptyScopes(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global:  map[string]any{},
			Catalog: map[string]domain.CatalogVars{},
		},
	}

	params := []domain.Parameter{
		{Name: "region", Scope: "global"},
	}

	resolver := NewDefaultResolver()
	result := resolver.Resolve(cfg, "default", "hello-world", params)

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string", "hello", "hello"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"float64 integer", float64(42), "42"},
		{"float64 decimal", 3.14, "3.14"},
		{"float64 zero", float64(0), "0"},
		{"int via default", 99, "99"},
		{"nil via default", nil, "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.input)
			if got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultVarResolver_FloatValues(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global: map[string]any{
				"count":   float64(5),
				"ratio":   3.14,
			},
		},
	}

	params := []domain.Parameter{
		{Name: "count", Scope: "global"},
		{Name: "ratio", Scope: "global"},
	}

	resolver := NewDefaultResolver()
	result := resolver.Resolve(cfg, "default", "test", params)

	if result["count"] != "5" {
		t.Errorf("count = %q, want 5", result["count"])
	}
	if result["ratio"] != "3.14" {
		t.Errorf("ratio = %q, want 3.14", result["ratio"])
	}
}

func TestDefaultVarResolver_MissingCatalog(t *testing.T) {
	cfg := &domain.Config{
		Vars: domain.Vars{
			Global:  map[string]any{"region": "us-east-1"},
			Catalog: map[string]domain.CatalogVars{},
		},
	}

	params := []domain.Parameter{
		{Name: "region", Scope: "global"},
		{Name: "namespace", Scope: "catalog"},
	}

	resolver := NewDefaultResolver()
	result := resolver.Resolve(cfg, "nonexistent", "hello-world", params)

	if result["region"] != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", result["region"])
	}
	if _, ok := result["namespace"]; ok {
		t.Error("namespace should not be resolved from missing catalog")
	}
}
