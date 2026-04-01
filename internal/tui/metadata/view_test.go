package metadata

import (
	"dops/internal/domain"
	"dops/internal/testutil"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	rb := &domain.Runbook{
		ID:          "default.hello-world",
		Name:        "hello-world",
		Description: "Prints a hello world message",
		Version:     "1.0.0",
		RiskLevel:   domain.RiskLow,
	}

	cat := &domain.Catalog{Name: "default", Path: "~/.dops/catalogs/default"}
	out := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 40, Styles: testutil.TestStyles()})

	if !strings.Contains(out, "hello-world") {
		t.Error("output should contain runbook name")
	}
	if !strings.Contains(out, "1.0.0") {
		t.Error("output should contain version")
	}
	if !strings.Contains(out, "low") {
		t.Error("output should contain risk level")
	}
	if !strings.Contains(out, "Prints a hello world message") {
		t.Error("output should contain description")
	}
	if !strings.Contains(out, "catalogs/default") {
		t.Error("output should contain local path")
	}
}

func TestRender_PathTruncation(t *testing.T) {
	rb := &domain.Runbook{
		Name:      "hello-world",
		Version:   "1.0.0",
		RiskLevel: domain.RiskLow,
	}
	cat := &domain.Catalog{Name: "default", Path: "~/.dops/catalogs/default"}

	// Wide enough to show full path.
	wide := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 60, Styles: testutil.TestStyles()})
	if !strings.Contains(wide, "runbook.yaml") {
		t.Error("wide render should show full path")
	}

	// Narrow should truncate with ellipsis.
	narrow := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 30, Styles: testutil.TestStyles()})
	if strings.Contains(narrow, "runbook.yaml") {
		t.Error("narrow render should truncate path")
	}
	if !strings.Contains(narrow, "…") {
		t.Error("narrow render should show ellipsis")
	}
}

func TestRender_GitCatalog(t *testing.T) {
	rb := &domain.Runbook{
		Name:      "drain-node",
		Version:   "2.1.0",
		RiskLevel: domain.RiskHigh,
	}
	cat := &domain.Catalog{Name: "public", URL: "https://github.com/org/public-catalog"}
	out := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 50, Styles: testutil.TestStyles()})

	if !strings.Contains(out, "public-catalog") {
		t.Error("output should contain catalog URL")
	}
}

func TestRender_HyperlinkTarget(t *testing.T) {
	rb := &domain.Runbook{Name: "hello-world", Version: "1.0.0", RiskLevel: domain.RiskLow}

	t.Run("git catalog uses URL hyperlink", func(t *testing.T) {
		cat := &domain.Catalog{Name: "public", URL: "https://github.com/org/repo"}
		out := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 80, Styles: testutil.TestStyles()})
		// OSC8 hyperlink embeds the URL in the escape sequence
		if !strings.Contains(out, "https://github.com/org/repo") {
			t.Error("git catalog should embed URL as hyperlink target")
		}
		if strings.Contains(out, "file://") {
			t.Error("git catalog should NOT use file:// hyperlink")
		}
	})

	t.Run("local catalog uses file hyperlink", func(t *testing.T) {
		cat := &domain.Catalog{Name: "default", Path: "~/.dops/catalogs/default"}
		out := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 80, Styles: testutil.TestStyles()})
		if !strings.Contains(out, "file://") {
			t.Error("local catalog should use file:// hyperlink")
		}
	})
}

func TestRender_CopiedFlash(t *testing.T) {
	rb := &domain.Runbook{
		Name:      "hello-world",
		Version:   "1.0.0",
		RiskLevel: domain.RiskLow,
	}
	cat := &domain.Catalog{Name: "default", Path: "~/.dops/catalogs/default"}
	out := Render(RenderParams{Runbook: rb, Catalog: cat, Width: 60, Copied: true, Styles: testutil.TestStyles()})

	// Path should still be visible (flashed green, not replaced).
	if !strings.Contains(out, "runbook.yaml") {
		t.Error("output should still show path when flash is true")
	}
}

func TestRender_Nil(t *testing.T) {
	out := Render(RenderParams{Width: 40, Styles: testutil.TestStyles()})
	if len(out) == 0 {
		t.Error("nil runbook should still produce output")
	}
}

func TestLocation(t *testing.T) {
	rb := &domain.Runbook{Name: "hello-world"}

	t.Run("local catalog", func(t *testing.T) {
		cat := &domain.Catalog{Path: "~/.dops/catalogs/default"}
		loc := Location(rb, cat)
		if loc != "~/.dops/catalogs/default/hello-world/runbook.yaml" {
			t.Errorf("got %q", loc)
		}
	})

	t.Run("git catalog", func(t *testing.T) {
		cat := &domain.Catalog{URL: "https://github.com/org/repo"}
		loc := Location(rb, cat)
		if loc != "https://github.com/org/repo" {
			t.Errorf("got %q", loc)
		}
	})

	t.Run("nil inputs", func(t *testing.T) {
		if Location(nil, nil) != "" {
			t.Error("nil inputs should return empty")
		}
	})
}
