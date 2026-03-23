package config

import (
	"dops/internal/domain"
	"testing"
)

func newTestConfig() *domain.Config {
	return &domain.Config{
		Theme: "tokyonight",
		Defaults: domain.Defaults{
			MaxRiskLevel: domain.RiskMedium,
		},
		Vars: domain.Vars{
			Global: map[string]any{
				"region": "us-east-1",
			},
			Catalog: map[string]domain.CatalogVars{
				"default": {
					Vars: map[string]any{
						"namespace": "platform",
					},
					Runbooks: map[string]map[string]any{
						"hello-world": {
							"dry_run": false,
						},
					},
				},
			},
		},
	}
}

func TestGet(t *testing.T) {
	cfg := newTestConfig()

	tests := []struct {
		name    string
		path    string
		want    any
		wantErr bool
	}{
		{name: "top-level theme", path: "theme", want: "tokyonight"},
		{name: "defaults.max_risk_level", path: "defaults.max_risk_level", want: domain.RiskMedium},
		{name: "global var", path: "vars.global.region", want: "us-east-1"},
		{name: "catalog var", path: "vars.catalog.default.namespace", want: "platform"},
		{name: "runbook var", path: "vars.catalog.default.runbooks.hello-world.dry_run", want: false},
		{name: "nonexistent top-level", path: "nonexistent", wantErr: true},
		{name: "nonexistent global var", path: "vars.global.nonexistent", wantErr: true},
		{name: "nonexistent catalog", path: "vars.catalog.missing.x", wantErr: true},
		{name: "empty path", path: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(cfg, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Get(%q) expected error, got %v", tt.path, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) unexpected error: %v", tt.path, err)
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %v (%T), want %v (%T)", tt.path, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		value   any
		verify  func(*domain.Config) bool
		wantErr bool
	}{
		{
			name:  "set theme",
			path:  "theme",
			value: "dracula",
			verify: func(c *domain.Config) bool {
				return c.Theme == "dracula"
			},
		},
		{
			name:  "set global var",
			path:  "vars.global.environment",
			value: "production",
			verify: func(c *domain.Config) bool {
				return c.Vars.Global["environment"] == "production"
			},
		},
		{
			name:  "set catalog var",
			path:  "vars.catalog.default.token",
			value: "abc123",
			verify: func(c *domain.Config) bool {
				return c.Vars.Catalog["default"].Vars["token"] == "abc123"
			},
		},
		{
			name:  "set runbook var",
			path:  "vars.catalog.default.runbooks.hello-world.count",
			value: 42,
			verify: func(c *domain.Config) bool {
				return c.Vars.Catalog["default"].Runbooks["hello-world"]["count"] == 42
			},
		},
		{
			name:  "set new catalog",
			path:  "vars.catalog.newcat.key",
			value: "val",
			verify: func(c *domain.Config) bool {
				return c.Vars.Catalog["newcat"].Vars["key"] == "val"
			},
		},
		{
			name:  "set new runbook in existing catalog",
			path:  "vars.catalog.default.runbooks.new-rb.flag",
			value: true,
			verify: func(c *domain.Config) bool {
				return c.Vars.Catalog["default"].Runbooks["new-rb"]["flag"] == true
			},
		},
		{
			name:    "empty path",
			path:    "",
			value:   "x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()
			err := Set(cfg, tt.path, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Set(%q) expected error", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("Set(%q) unexpected error: %v", tt.path, err)
			}
			if !tt.verify(cfg) {
				t.Errorf("Set(%q, %v) did not produce expected result", tt.path, tt.value)
			}
		})
	}
}

func TestUnset(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		verify  func(*domain.Config) bool
		wantErr bool
	}{
		{
			name: "unset global var",
			path: "vars.global.region",
			verify: func(c *domain.Config) bool {
				_, ok := c.Vars.Global["region"]
				return !ok
			},
		},
		{
			name: "unset catalog var",
			path: "vars.catalog.default.namespace",
			verify: func(c *domain.Config) bool {
				_, ok := c.Vars.Catalog["default"].Vars["namespace"]
				return !ok
			},
		},
		{
			name: "unset runbook var",
			path: "vars.catalog.default.runbooks.hello-world.dry_run",
			verify: func(c *domain.Config) bool {
				_, ok := c.Vars.Catalog["default"].Runbooks["hello-world"]["dry_run"]
				return !ok
			},
		},
		{
			name:    "unset nonexistent",
			path:    "vars.global.nonexistent",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()
			err := Unset(cfg, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Unset(%q) expected error", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unset(%q) unexpected error: %v", tt.path, err)
			}
			if !tt.verify(cfg) {
				t.Errorf("Unset(%q) did not produce expected result", tt.path)
			}
		})
	}
}
