package mcp

import (
	"context"
	"encoding/json"
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
