package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"dops/internal/domain"
)

func TestRunbookToInputSchema_BasicTypes(t *testing.T) {
	rb := domain.Runbook{
		ID: "default.test",
		Parameters: []domain.Parameter{
			{Name: "name", Type: domain.ParamString, Required: true},
			{Name: "count", Type: domain.ParamInteger},
			{Name: "verbose", Type: domain.ParamBoolean},
			{Name: "env", Type: domain.ParamSelect, Options: []string{"dev", "prod"}},
		},
	}

	schema := RunbookToInputSchema(rb, nil)
	var parsed map[string]any
	if err := json.Unmarshal(schema, &parsed); err != nil {
		t.Fatal(err)
	}

	props := parsed["properties"].(map[string]any)
	if props["name"].(map[string]any)["type"] != "string" {
		t.Error("name should be string type")
	}
	if props["count"].(map[string]any)["type"] != "integer" {
		t.Error("count should be integer type")
	}
	if props["verbose"].(map[string]any)["type"] != "boolean" {
		t.Error("verbose should be boolean type")
	}

	envProp := props["env"].(map[string]any)
	if envProp["type"] != "string" {
		t.Error("env should be string type")
	}
}

func TestRunbookToInputSchema_ExcludesSensitive(t *testing.T) {
	rb := domain.Runbook{
		ID: "default.test",
		Parameters: []domain.Parameter{
			{Name: "endpoint", Type: domain.ParamString, Required: true},
			{Name: "api_key", Type: domain.ParamString, Secret: true},
		},
	}

	schema := RunbookToInputSchema(rb, nil)
	var parsed map[string]any
	json.Unmarshal(schema, &parsed)

	props := parsed["properties"].(map[string]any)
	if _, exists := props["api_key"]; exists {
		t.Error("sensitive param should be excluded from schema")
	}
	if _, exists := props["endpoint"]; !exists {
		t.Error("non-sensitive param should be in schema")
	}
}

func TestRunbookToInputSchema_RiskConfirmation(t *testing.T) {
	rb := domain.Runbook{
		ID:        "infra.dangerous",
		RiskLevel: domain.RiskCritical,
	}

	schema := RunbookToInputSchema(rb, nil)
	var parsed map[string]any
	json.Unmarshal(schema, &parsed)

	props := parsed["properties"].(map[string]any)
	if _, exists := props["_confirm_word"]; !exists {
		t.Error("critical risk should have _confirm_word field")
	}
}

func TestRunbookToDescription_WithSensitive(t *testing.T) {
	rb := domain.Runbook{
		ID:          "infra.deploy",
		Description: "Deploy the app",
		RiskLevel:   domain.RiskMedium,
		Parameters: []domain.Parameter{
			{Name: "api_key", Type: domain.ParamString, Secret: true},
		},
	}

	desc := RunbookToDescription(rb)
	if !strings.Contains(desc, "api_key") {
		t.Error("description should mention sensitive param")
	}
	if !strings.Contains(desc, "dops config set") {
		t.Error("description should mention how to configure")
	}
}
