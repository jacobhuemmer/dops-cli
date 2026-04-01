package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCatalog_Label(t *testing.T) {
	tests := []struct {
		name        string
		catalog     Catalog
		wantLabel   string
	}{
		{
			name:      "returns display name when set",
			catalog:   Catalog{Name: "my-repo", DisplayName: "Production Ops"},
			wantLabel: "Production Ops",
		},
		{
			name:      "falls back to name when display name empty",
			catalog:   Catalog{Name: "my-repo"},
			wantLabel: "my-repo",
		},
		{
			name:      "falls back to name when display name blank",
			catalog:   Catalog{Name: "my-repo", DisplayName: ""},
			wantLabel: "my-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.catalog.Label(); got != tt.wantLabel {
				t.Errorf("Label() = %q, want %q", got, tt.wantLabel)
			}
		})
	}
}

func TestCatalog_RunbookRoot(t *testing.T) {
	tests := []struct {
		name    string
		catalog Catalog
		want    string
	}{
		{
			name:    "no subpath",
			catalog: Catalog{Path: "/home/user/.dops/catalogs/repo"},
			want:    "/home/user/.dops/catalogs/repo",
		},
		{
			name:    "with subpath",
			catalog: Catalog{Path: "/home/user/.dops/catalogs/repo", SubPath: "src/runbooks"},
			want:    "/home/user/.dops/catalogs/repo/src/runbooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.catalog.RunbookRoot(); got != tt.want {
				t.Errorf("RunbookRoot() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid short name",
			input: "Prod Ops",
		},
		{
			name:  "valid at limit",
			input: strings.Repeat("a", 50),
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 51),
			wantErr: true,
			errMsg:  "50 characters or fewer",
		},
		{
			name:    "non-printable character",
			input:   "hello\x00world",
			wantErr: true,
			errMsg:  "non-printable",
		},
		{
			name:  "empty is valid",
			input: "",
		},
		{
			name:  "unicode is valid",
			input: "Producción Ops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDisplayName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCatalogVars_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		cv   CatalogVars
		want map[string]any
	}{
		{
			name: "vars only",
			cv: CatalogVars{
				Vars:     map[string]any{"region": "us-east-1", "env": "prod"},
				Runbooks: nil,
			},
			want: map[string]any{"region": "us-east-1", "env": "prod"},
		},
		{
			name: "vars with runbooks",
			cv: CatalogVars{
				Vars: map[string]any{"region": "us-east-1"},
				Runbooks: map[string]map[string]any{
					"deploy": {"version": "1.0"},
				},
			},
			want: map[string]any{
				"region": "us-east-1",
				"runbooks": map[string]any{
					"deploy": map[string]any{"version": "1.0"},
				},
			},
		},
		{
			name: "empty vars no runbooks",
			cv: CatalogVars{
				Vars:     map[string]any{},
				Runbooks: map[string]map[string]any{},
			},
			want: map[string]any{},
		},
		{
			name: "nil vars nil runbooks",
			cv:   CatalogVars{},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.cv)
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal result: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCatalogVars_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantVars     map[string]any
		wantRunbooks map[string]map[string]any
		wantErr      bool
	}{
		{
			name:         "flat vars only",
			input:        `{"region":"us-east-1","env":"prod"}`,
			wantVars:     map[string]any{"region": "us-east-1", "env": "prod"},
			wantRunbooks: map[string]map[string]any{},
		},
		{
			name:  "vars with runbooks",
			input: `{"region":"us-east-1","runbooks":{"deploy":{"version":"1.0"}}}`,
			wantVars: map[string]any{"region": "us-east-1"},
			wantRunbooks: map[string]map[string]any{
				"deploy": {"version": "1.0"},
			},
		},
		{
			name:         "empty object",
			input:        `{}`,
			wantVars:     map[string]any{},
			wantRunbooks: map[string]map[string]any{},
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cv CatalogVars
			err := json.Unmarshal([]byte(tt.input), &cv)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}
			if !reflect.DeepEqual(cv.Vars, tt.wantVars) {
				t.Errorf("Vars = %v, want %v", cv.Vars, tt.wantVars)
			}
			if !reflect.DeepEqual(cv.Runbooks, tt.wantRunbooks) {
				t.Errorf("Runbooks = %v, want %v", cv.Runbooks, tt.wantRunbooks)
			}
		})
	}
}

func TestCatalogVars_RoundTrip(t *testing.T) {
	original := CatalogVars{
		Vars: map[string]any{"key": "value", "count": float64(42)},
		Runbooks: map[string]map[string]any{
			"deploy": {"target": "prod"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CatalogVars
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(original.Vars, decoded.Vars) {
		t.Errorf("Vars mismatch after round-trip: got %v, want %v", decoded.Vars, original.Vars)
	}
	if !reflect.DeepEqual(original.Runbooks, decoded.Runbooks) {
		t.Errorf("Runbooks mismatch after round-trip: got %v, want %v", decoded.Runbooks, original.Runbooks)
	}
}
