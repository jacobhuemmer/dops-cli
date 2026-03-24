package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"dops/internal/catalog"
	"dops/internal/domain"
)

func TestCatalogListJSON(t *testing.T) {
	catalogs := []catalog.CatalogWithRunbooks{
		{
			Catalog: domain.Catalog{Name: "default"},
			Runbooks: []domain.Runbook{
				{ID: "default.hello", Name: "hello", Description: "Say hello", RiskLevel: domain.RiskLow},
				{ID: "default.deploy", Name: "deploy", Description: "Deploy app", RiskLevel: domain.RiskHigh},
			},
		},
	}

	result, err := CatalogListJSON(catalogs)
	if err != nil {
		t.Fatal(err)
	}

	var summaries []RunbookSummary
	if err := json.Unmarshal([]byte(result), &summaries); err != nil {
		t.Fatal(err)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != "default.hello" {
		t.Errorf("first ID = %q", summaries[0].ID)
	}
}

func TestRunbookDetailJSON(t *testing.T) {
	rb := domain.Runbook{
		ID:          "default.hello",
		Name:        "hello",
		Description: "Say hello",
		Version:     "1.0.0",
	}
	cat := domain.Catalog{Name: "default"}

	result, err := RunbookDetailJSON(rb, cat)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "default.hello") {
		t.Error("should contain runbook ID")
	}
	if !strings.Contains(result, "1.0.0") {
		t.Error("should contain version")
	}
}
