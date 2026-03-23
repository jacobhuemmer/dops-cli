package domain

import "testing"

func TestValidateRunbookID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "valid id", id: "default.hello-world", wantErr: false},
		{name: "valid with numbers", id: "local.rotate-tls-123", wantErr: false},
		{name: "missing dot", id: "hello-world", wantErr: true},
		{name: "empty string", id: "", wantErr: true},
		{name: "empty catalog segment", id: ".hello-world", wantErr: true},
		{name: "empty runbook segment", id: "default.", wantErr: true},
		{name: "multiple dots", id: "a.b.c", wantErr: true},
		{name: "dot only", id: ".", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRunbookID(tt.id)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateRunbookID(%q) expected error, got nil", tt.id)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateRunbookID(%q) unexpected error: %v", tt.id, err)
			}
		})
	}
}
