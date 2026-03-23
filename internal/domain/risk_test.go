package domain

import "testing"

func TestRiskLevel_Exceeds(t *testing.T) {
	tests := []struct {
		name     string
		level    RiskLevel
		ceiling  RiskLevel
		exceeds  bool
	}{
		{name: "low does not exceed low", level: RiskLow, ceiling: RiskLow, exceeds: false},
		{name: "low does not exceed medium", level: RiskLow, ceiling: RiskMedium, exceeds: false},
		{name: "medium does not exceed medium", level: RiskMedium, ceiling: RiskMedium, exceeds: false},
		{name: "medium exceeds low", level: RiskMedium, ceiling: RiskLow, exceeds: true},
		{name: "high exceeds medium", level: RiskHigh, ceiling: RiskMedium, exceeds: true},
		{name: "high exceeds low", level: RiskHigh, ceiling: RiskLow, exceeds: true},
		{name: "high does not exceed high", level: RiskHigh, ceiling: RiskHigh, exceeds: false},
		{name: "critical exceeds high", level: RiskCritical, ceiling: RiskHigh, exceeds: true},
		{name: "critical exceeds low", level: RiskCritical, ceiling: RiskLow, exceeds: true},
		{name: "critical does not exceed critical", level: RiskCritical, ceiling: RiskCritical, exceeds: false},
		{name: "low does not exceed critical", level: RiskLow, ceiling: RiskCritical, exceeds: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.Exceeds(tt.ceiling)
			if got != tt.exceeds {
				t.Errorf("RiskLevel(%q).Exceeds(%q) = %v, want %v", tt.level, tt.ceiling, got, tt.exceeds)
			}
		})
	}
}

func TestParseRiskLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RiskLevel
		wantErr bool
	}{
		{name: "low", input: "low", want: RiskLow},
		{name: "medium", input: "medium", want: RiskMedium},
		{name: "high", input: "high", want: RiskHigh},
		{name: "critical", input: "critical", want: RiskCritical},
		{name: "empty string", input: "", wantErr: true},
		{name: "unknown value", input: "extreme", wantErr: true},
		{name: "uppercase", input: "LOW", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRiskLevel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRiskLevel(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRiskLevel(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseRiskLevel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
