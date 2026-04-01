package mcp

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dops/internal/catalog"
	"dops/internal/domain"
	"dops/internal/executor"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeRunner struct{}

func (r *fakeRunner) Run(ctx context.Context, scriptPath string, env map[string]string) (<-chan executor.OutputLine, <-chan error) {
	lines := make(chan executor.OutputLine, 10)
	errs := make(chan error, 1)
	go func() {
		lines <- executor.OutputLine{Text: "hello from MCP"}
		lines <- executor.OutputLine{Text: "execution complete"}
		close(lines)
		errs <- nil
	}()
	return lines, errs
}

func testCatalogs() []catalog.CatalogWithRunbooks {
	return []catalog.CatalogWithRunbooks{
		{
			Catalog: domain.Catalog{Name: "default", Path: filepath.Join(os.TempDir(), "test")},
			Runbooks: []domain.Runbook{
				{
					ID:          "default.echo",
					Name:        "echo",
					Description: "Echoes a message",
					RiskLevel:   domain.RiskLow,
					Script:      "script.sh",
					Parameters: []domain.Parameter{
						{Name: "message", Type: domain.ParamString, Required: true},
					},
				},
				{
					ID:          "default.sensitive",
					Name:        "sensitive",
					Description: "Has secrets",
					RiskLevel:   domain.RiskMedium,
					Script:      "script.sh",
					Parameters: []domain.Parameter{
						{Name: "endpoint", Type: domain.ParamString, Required: true},
						{Name: "api_key", Type: domain.ParamString, Secret: true},
					},
				},
			},
		},
	}
}

func TestServer_ToolsList(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()

	if _, err := srv.srv.Connect(ctx, t1, nil); err != nil {
		t.Fatal(err)
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// List tools.
	var tools []*mcpsdk.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	// Verify sensitive tool has warning in description.
	var sensitiveDesc string
	for _, tool := range tools {
		if tool.Name == "default.sensitive" {
			sensitiveDesc = tool.Description
		}
	}
	if !strings.Contains(sensitiveDesc, "api_key") {
		t.Error("sensitive tool description should mention api_key")
	}

	// Verify sensitive param NOT in schema.
	for _, tool := range tools {
		if tool.Name == "default.sensitive" {
			schema, _ := json.Marshal(tool.InputSchema)
			if strings.Contains(string(schema), "api_key") {
				t.Error("api_key should not be in schema")
			}
		}
	}
}

func TestServer_ResourcesList(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// Read catalog resource.
	result, err := session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "dops://catalog"})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected resource contents")
	}

	text := result.Contents[0].Text
	if !strings.Contains(text, "default.echo") {
		t.Error("catalog should contain echo runbook")
	}
}

func TestServer_ToolCall(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// Call echo tool.
	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "default.echo",
		Arguments: map[string]any{"message": "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(*mcpsdk.TextContent).Text
	if !strings.Contains(text, "hello from MCP") {
		t.Errorf("output should contain 'hello from MCP', got: %s", text)
	}
	if !strings.Contains(text, "Exit code: 0") {
		t.Error("should contain exit code")
	}
}

func TestServer_MaxRiskFilters(t *testing.T) {
	cats := []catalog.CatalogWithRunbooks{
		{
			Catalog: domain.Catalog{Name: "test", Path: filepath.Join(os.TempDir(), "test")},
			Runbooks: []domain.Runbook{
				{ID: "test.low", Name: "low", RiskLevel: domain.RiskLow, Script: "s.sh"},
				{ID: "test.high", Name: "high", RiskLevel: domain.RiskHigh, Script: "s.sh"},
				{ID: "test.critical", Name: "critical", RiskLevel: domain.RiskCritical, Script: "s.sh"},
			},
		},
	}

	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: cats,
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
		MaxRisk:  domain.RiskLow,
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	var tools []*mcpsdk.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		tools = append(tools, tool)
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool (low only), got %d", len(tools))
	}
	if tools[0].Name != "test.low" {
		t.Errorf("expected test.low, got %s", tools[0].Name)
	}
}

func TestServer_SchemaResources(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	tests := []struct {
		uri      string
		contains string
	}{
		{"dops://schema/runbook", "runbook.yaml"},
		{"dops://schema/shell-style", "Shell Script Style Guide"},
	}

	for _, tt := range tests {
		result, err := session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: tt.uri})
		if err != nil {
			t.Fatalf("ReadResource(%s): %v", tt.uri, err)
		}
		if len(result.Contents) == 0 {
			t.Fatalf("ReadResource(%s): no contents", tt.uri)
		}
		if !strings.Contains(result.Contents[0].Text, tt.contains) {
			t.Errorf("ReadResource(%s) should contain %q", tt.uri, tt.contains)
		}
	}
}

func TestServer_Prompts(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		DopsHome: "/tmp/test-dops",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	result, err := session.GetPrompt(ctx, &mcpsdk.GetPromptParams{
		Name: "create-runbook",
		Arguments: map[string]string{
			"catalog":     "infra",
			"name":        "check-health",
			"description": "Check service health",
			"risk_level":  "medium",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Messages) == 0 {
		t.Fatal("expected prompt messages")
	}

	text := result.Messages[0].Content.(*mcpsdk.TextContent).Text
	if !strings.Contains(text, "check-health") {
		t.Error("prompt should contain runbook name")
	}
	if !strings.Contains(text, "/tmp/test-dops") {
		t.Error("prompt should contain dops home path")
	}
	if !strings.Contains(text, "medium") {
		t.Error("prompt should contain risk level")
	}
}

func TestToolError(t *testing.T) {
	result := toolError("something broke")
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	text := result.Content[0].(*mcpsdk.TextContent).Text
	if text != "something broke" {
		t.Errorf("got %q, want %q", text, "something broke")
	}
}

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	handler := gzipMiddleware(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("expected Content-Encoding: gzip")
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal("failed to create gzip reader:", err)
	}
	defer gr.Close()

	body, _ := io.ReadAll(gr)
	if string(body) != "hello world" {
		t.Errorf("got %q, want %q", string(body), "hello world")
	}
}

func TestGzipMiddleware_SkipsWithoutAcceptEncoding(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	handler := gzipMiddleware(inner)
	req := httptest.NewRequest("GET", "/", nil)
	// No Accept-Encoding header.
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not gzip without Accept-Encoding")
	}
	if rec.Body.String() != "hello world" {
		t.Errorf("got %q", rec.Body.String())
	}
}

func TestGzipMiddleware_SkipsSSE(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("event: data\n"))
	})

	handler := gzipMiddleware(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "text/event-stream")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not gzip SSE responses")
	}
}

func TestServer_RunbookDetailResource(t *testing.T) {
	srv := NewServer(ServerConfig{
		Version:  "test",
		Catalogs: testCatalogs(),
		Runner:   &fakeRunner{},
		Config:   &domain.Config{},
	})

	ctx := context.Background()
	t1, t2 := mcpsdk.NewInMemoryTransports()
	srv.srv.Connect(ctx, t1, nil)

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	result, err := session.ReadResource(ctx, &mcpsdk.ReadResourceParams{URI: "dops://catalog/default.echo"})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Contents) == 0 {
		t.Fatal("expected resource contents")
	}

	text := result.Contents[0].Text
	if !strings.Contains(text, "default.echo") {
		t.Error("detail should contain runbook ID")
	}
}
