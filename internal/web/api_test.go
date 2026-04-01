package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dops/internal/domain"
	"dops/internal/executor"
	"dops/internal/theme"

	catpkg "dops/internal/catalog"
)

// --- Mocks ---

type mockLoader struct {
	runbook *domain.Runbook
	catalog *domain.Catalog
	err     error
}

func (m *mockLoader) FindByID(id string) (*domain.Runbook, *domain.Catalog, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	if m.runbook != nil && m.runbook.ID == id {
		return m.runbook, m.catalog, nil
	}
	return nil, nil, fmt.Errorf("not found")
}

func (m *mockLoader) FindByAlias(alias string) (*domain.Runbook, *domain.Catalog, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	if m.runbook != nil {
		for _, a := range m.runbook.Aliases {
			if a == alias {
				return m.runbook, m.catalog, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("not found")
}

type mockRunner struct {
	lines []string
	err   error
}

func (m *mockRunner) Run(_ context.Context, _ string, _ map[string]string) (<-chan executor.OutputLine, <-chan error) {
	linesCh := make(chan executor.OutputLine, len(m.lines))
	errsCh := make(chan error, 1)
	for _, l := range m.lines {
		linesCh <- executor.OutputLine{Text: l}
	}
	close(linesCh)
	errsCh <- m.err
	return linesCh, errsCh
}

type mockThemeLoader struct {
	themes map[string]*domain.ThemeFile
}

func (m *mockThemeLoader) Load(name string) (*domain.ThemeFile, error) {
	if tf, ok := m.themes[name]; ok {
		return tf, nil
	}
	return nil, fmt.Errorf("theme %q not found", name)
}

type mockConfigStore struct {
	saved     *domain.Config
	saveCalls int
}

func (m *mockConfigStore) Load() (*domain.Config, error)            { return m.saved, nil }
func (m *mockConfigStore) Save(cfg *domain.Config) error            { m.saveCalls++; m.saved = cfg; return nil }
func (m *mockConfigStore) EnsureDefaults() (*domain.Config, error)  { return m.saved, nil }

// --- Helpers ---

var testRunbook = domain.Runbook{
	ID:          "default.hello-world",
	Name:        "hello-world",
	Aliases:     []string{"hw"},
	Description: "Says hello",
	Version:     "1.0.0",
	RiskLevel:   domain.RiskLow,
	Script:      "./script.sh",
	Parameters: []domain.Parameter{
		{Name: "greeting", Type: domain.ParamString, Required: true, Description: "The greeting"},
	},
}

var testRunbookWithSecret = domain.Runbook{
	ID:          "default.secret-job",
	Name:        "secret-job",
	Description: "Has secrets",
	Version:     "1.0.0",
	RiskLevel:   domain.RiskLow,
	Script:      "./run.sh",
	Parameters: []domain.Parameter{
		{Name: "token", Type: domain.ParamString, Required: true, Secret: true},
		{Name: "name", Type: domain.ParamString, Required: true},
	},
}

var testCatalog = domain.Catalog{Name: "default", DisplayName: "My Ops", Active: true}

func testDeps() ServerDeps {
	return ServerDeps{
		Config: &domain.Config{
			Theme:    "github",
			Defaults: domain.Defaults{MaxRiskLevel: domain.RiskHigh},
		},
		Catalogs: []catpkg.CatalogWithRunbooks{
			{
				Catalog:  testCatalog,
				Runbooks: []domain.Runbook{testRunbook},
			},
		},
		Loader: &mockLoader{
			runbook: &testRunbook,
			catalog: &testCatalog,
		},
		Runner: &mockRunner{
			lines: []string{"hello", "world"},
		},
		Theme: &theme.ResolvedTheme{
			Name:   "github",
			Colors: map[string]string{"primary": "#58a6ff", "background": "#0d1117"},
		},
		ThemeLoader: &mockThemeLoader{themes: map[string]*domain.ThemeFile{}},
		Port:        0,
	}
}

func setupTestAPI(t *testing.T) *http.ServeMux {
	t.Helper()
	deps := testDeps()
	a := newAPI(deps)
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	return mux
}

func setupTestAPIWithDeps(t *testing.T, deps ServerDeps) *http.ServeMux {
	t.Helper()
	a := newAPI(deps)
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	return mux
}

// --- Catalog & Runbook Tests ---

func TestListCatalogs(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/catalogs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var catalogs []catalogResponse
	if err := json.Unmarshal(w.Body.Bytes(), &catalogs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(catalogs) != 1 {
		t.Fatalf("expected 1 catalog, got %d", len(catalogs))
	}
	if catalogs[0].Name != "default" {
		t.Errorf("catalog name = %q", catalogs[0].Name)
	}
	if catalogs[0].DisplayName != "My Ops" {
		t.Errorf("display name = %q", catalogs[0].DisplayName)
	}
	if len(catalogs[0].Runbooks) != 1 {
		t.Fatalf("expected 1 runbook, got %d", len(catalogs[0].Runbooks))
	}
	if catalogs[0].Runbooks[0].ID != "default.hello-world" {
		t.Errorf("runbook id = %q", catalogs[0].Runbooks[0].ID)
	}
}

func TestListCatalogs_RunbookFields(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/catalogs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var catalogs []catalogResponse
	if err := json.Unmarshal(w.Body.Bytes(), &catalogs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rb := catalogs[0].Runbooks[0]
	if rb.Name != "hello-world" {
		t.Errorf("name = %q", rb.Name)
	}
	if rb.Description != "Says hello" {
		t.Errorf("description = %q", rb.Description)
	}
	if rb.RiskLevel != "low" {
		t.Errorf("risk_level = %q", rb.RiskLevel)
	}
	if rb.ParamCount != 1 {
		t.Errorf("param_count = %d, want 1", rb.ParamCount)
	}
}

func TestGetRunbook(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/runbooks/default.hello-world", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var detail runbookDetail
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if detail.ID != "default.hello-world" {
		t.Errorf("id = %q", detail.ID)
	}
	if detail.Name != "hello-world" {
		t.Errorf("name = %q", detail.Name)
	}
	if detail.Version != "1.0.0" {
		t.Errorf("version = %q", detail.Version)
	}
	if detail.Script != "./script.sh" {
		t.Errorf("script = %q", detail.Script)
	}
	if len(detail.Parameters) != 1 {
		t.Fatalf("expected 1 param, got %d", len(detail.Parameters))
	}
	if detail.Parameters[0].Name != "greeting" {
		t.Errorf("param name = %q", detail.Parameters[0].Name)
	}
}

func TestGetRunbook_ByAlias(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/runbooks/hw", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var detail runbookDetail
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if detail.ID != "default.hello-world" {
		t.Errorf("id = %q, want default.hello-world", detail.ID)
	}
}

func TestGetRunbook_NotFound(t *testing.T) {
	deps := testDeps()
	deps.Loader = &mockLoader{err: fmt.Errorf("not found")}
	mux := setupTestAPIWithDeps(t, deps)

	req := httptest.NewRequest("GET", "/api/runbooks/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestGetRunbook_MasksSecrets(t *testing.T) {
	rb := testRunbookWithSecret
	cat := testCatalog
	deps := testDeps()
	deps.Loader = &mockLoader{runbook: &rb, catalog: &cat}
	deps.Catalogs = []catpkg.CatalogWithRunbooks{
		{Catalog: cat, Runbooks: []domain.Runbook{rb}},
	}
	// Set up saved values that include a secret.
	deps.Config.Vars = domain.Vars{
		Catalog: map[string]domain.CatalogVars{
			"default": {
				Runbooks: map[string]map[string]any{
					"secret-job": {"token": "supersecret", "name": "alice"},
				},
			},
		},
	}
	mux := setupTestAPIWithDeps(t, deps)

	req := httptest.NewRequest("GET", "/api/runbooks/default.secret-job", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var detail runbookDetail
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Secret param should be masked.
	if v, ok := detail.SavedValues["token"]; ok && v != "••••••••" {
		t.Errorf("secret value should be masked, got %q", v)
	}
	// Non-secret param should be visible.
	if v := detail.SavedValues["name"]; v != "alice" {
		t.Errorf("non-secret value = %q, want %q", v, "alice")
	}
}

// --- Execution Tests ---

func TestExecuteRunbook(t *testing.T) {
	mux := setupTestAPI(t)
	body := strings.NewReader(`{"params": {"greeting": "hi"}}`)
	req := httptest.NewRequest("POST", "/api/runbooks/default.hello-world/execute", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", w.Code)
	}

	var resp executeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ExecutionID == "" {
		t.Error("expected non-empty execution ID")
	}
	if !strings.HasPrefix(resp.ExecutionID, "exec-") {
		t.Errorf("execution id = %q, want exec-* prefix", resp.ExecutionID)
	}
}

func TestExecuteRunbook_NotFound(t *testing.T) {
	deps := testDeps()
	deps.Loader = &mockLoader{err: fmt.Errorf("not found")}
	mux := setupTestAPIWithDeps(t, deps)

	body := strings.NewReader(`{"params": {}}`)
	req := httptest.NewRequest("POST", "/api/runbooks/nonexistent/execute", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestExecuteRunbook_InvalidBody(t *testing.T) {
	mux := setupTestAPI(t)
	body := strings.NewReader(`not json`)
	req := httptest.NewRequest("POST", "/api/runbooks/default.hello-world/execute", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "invalid request body" {
		t.Errorf("error = %q", resp["error"])
	}
}

func TestCancelExecution(t *testing.T) {
	deps := testDeps()
	a := newAPI(deps)
	mux := http.NewServeMux()
	a.registerRoutes(mux)

	// Create a fake execution in the store.
	cancelled := false
	exec := &execution{
		id:     "exec-1",
		cancel: func() { cancelled = true },
		notify: make(chan struct{}, 1),
	}
	a.executions.mu.Lock()
	a.executions.execs["exec-1"] = exec
	a.executions.mu.Unlock()

	req := httptest.NewRequest("POST", "/api/executions/exec-1/cancel", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "cancelled" {
		t.Errorf("status = %q, want cancelled", resp["status"])
	}
	if !cancelled {
		t.Error("cancel function was not called")
	}
}

func TestCancelExecution_NotFound(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("POST", "/api/executions/nonexistent/cancel", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestStreamExecution_NotFound(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/executions/nonexistent/stream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestStreamExecution_CompletedExecution(t *testing.T) {
	deps := testDeps()
	a := newAPI(deps)
	mux := http.NewServeMux()
	a.registerRoutes(mux)

	// Create a completed execution with some output lines.
	exec := &execution{
		id:      "exec-1",
		lines:   []string{"line1", "line2"},
		done:    true,
		exitErr: nil,
		cancel:  func() {},
		notify:  make(chan struct{}, 1),
	}
	a.executions.mu.Lock()
	a.executions.execs["exec-1"] = exec
	a.executions.mu.Unlock()

	req := httptest.NewRequest("GET", "/api/executions/exec-1/stream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "data: line1") {
		t.Error("expected line1 in SSE output")
	}
	if !strings.Contains(body, "data: line2") {
		t.Error("expected line2 in SSE output")
	}
	if !strings.Contains(body, "event: done") {
		t.Error("expected done event in SSE output")
	}
	if !strings.Contains(body, "data: success") {
		t.Error("expected success status in done event")
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestStreamExecution_CompletedWithError(t *testing.T) {
	deps := testDeps()
	a := newAPI(deps)
	mux := http.NewServeMux()
	a.registerRoutes(mux)

	exec := &execution{
		id:      "exec-2",
		lines:   []string{"some output"},
		done:    true,
		exitErr: fmt.Errorf("exit status 1"),
		cancel:  func() {},
		notify:  make(chan struct{}, 1),
	}
	a.executions.mu.Lock()
	a.executions.execs["exec-2"] = exec
	a.executions.mu.Unlock()

	req := httptest.NewRequest("GET", "/api/executions/exec-2/stream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "event: done") {
		t.Error("expected done event")
	}
	if !strings.Contains(body, "error: exit status 1") {
		t.Error("expected error status in done event")
	}
}

// --- Theme Tests ---

func TestGetTheme(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/theme", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp themeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Name != "github" {
		t.Errorf("theme name = %q", resp.Name)
	}
	if resp.Colors["primary"] != "#58a6ff" {
		t.Errorf("primary = %q", resp.Colors["primary"])
	}
}

func TestGetTheme_NilTheme(t *testing.T) {
	deps := testDeps()
	deps.Theme = nil
	mux := setupTestAPIWithDeps(t, deps)

	req := httptest.NewRequest("GET", "/api/theme", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp themeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Name != "default" {
		t.Errorf("theme name = %q, want default", resp.Name)
	}
}

func TestListThemes(t *testing.T) {
	mux := setupTestAPI(t)
	req := httptest.NewRequest("GET", "/api/themes", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp["active"] != "github" {
		t.Errorf("active = %v, want github", resp["active"])
	}
	themes, ok := resp["themes"].([]any)
	if !ok {
		t.Fatal("themes is not an array")
	}
	if len(themes) == 0 {
		t.Error("expected at least one bundled theme")
	}
}

func TestListThemes_NilTheme(t *testing.T) {
	deps := testDeps()
	deps.Theme = nil
	mux := setupTestAPIWithDeps(t, deps)

	req := httptest.NewRequest("GET", "/api/themes", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp["active"] != "" {
		t.Errorf("active = %v, want empty string", resp["active"])
	}
}

func TestSetTheme_BadRequest(t *testing.T) {
	mux := setupTestAPI(t)

	tests := []struct {
		name string
		body string
	}{
		{"invalid json", `not json`},
		{"empty name", `{"name": ""}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/theme", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}
		})
	}
}

func TestSetTheme_NotFound(t *testing.T) {
	deps := testDeps()
	deps.ThemeLoader = &mockThemeLoader{themes: map[string]*domain.ThemeFile{}}
	mux := setupTestAPIWithDeps(t, deps)

	body := strings.NewReader(`{"name": "nonexistent"}`)
	req := httptest.NewRequest("PUT", "/api/theme", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestSetTheme_Success(t *testing.T) {
	deps := testDeps()
	deps.ThemeLoader = &mockThemeLoader{
		themes: map[string]*domain.ThemeFile{
			"tokyo-night": {
				Name: "tokyo-night",
				Defs: map[string]string{"bg": "#1a1b26"},
				Theme: map[string]json.RawMessage{
					"background": json.RawMessage(`{"dark":"bg","light":"bg"}`),
				},
			},
		},
	}
	deps.ConfigStore = &mockConfigStore{saved: deps.Config}
	mux := setupTestAPIWithDeps(t, deps)

	body := strings.NewReader(`{"name": "tokyo-night"}`)
	req := httptest.NewRequest("PUT", "/api/theme", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var resp themeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Name != "tokyo-night" {
		t.Errorf("name = %q, want tokyo-night", resp.Name)
	}
}

func TestSetTheme_DemoModeSkipsSave(t *testing.T) {
	deps := testDeps()
	deps.Demo = true
	store := &mockConfigStore{saved: deps.Config}
	deps.ConfigStore = store
	deps.ThemeLoader = &mockThemeLoader{
		themes: map[string]*domain.ThemeFile{
			"tokyo-night": {
				Name: "tokyo-night",
				Defs: map[string]string{"bg": "#1a1b26"},
				Theme: map[string]json.RawMessage{
					"background": json.RawMessage(`{"dark":"bg","light":"bg"}`),
				},
			},
		},
	}
	mux := setupTestAPIWithDeps(t, deps)

	body := strings.NewReader(`{"name": "tokyo-night"}`)
	req := httptest.NewRequest("PUT", "/api/theme", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// In demo mode, Save should not be called on the config store.
	if store.saveCalls != 0 {
		t.Errorf("Save called %d times in demo mode, want 0", store.saveCalls)
	}
}

// --- Execution Store Tests ---

func TestExecutionStoreEviction(t *testing.T) {
	store := newExecutionStore()

	// Seed the store with maxCompleted+10 completed executions.
	for i := 0; i < maxCompleted+10; i++ {
		id := fmt.Sprintf("exec-%d", i+1)
		store.execs[id] = &execution{
			id:          id,
			done:        true,
			completedAt: time.Now().Add(time.Duration(i) * time.Second),
			notify:      make(chan struct{}, 1),
		}
	}
	store.seq = maxCompleted + 10

	if len(store.execs) != maxCompleted+10 {
		t.Fatalf("setup: expected %d execs, got %d", maxCompleted+10, len(store.execs))
	}

	// Trigger eviction by calling evict under lock.
	store.mu.Lock()
	store.evict()
	store.mu.Unlock()

	if len(store.execs) != maxCompleted {
		t.Fatalf("after eviction: expected %d execs, got %d", maxCompleted, len(store.execs))
	}

	// The 10 oldest (exec-1 through exec-10) should have been removed.
	for i := 1; i <= 10; i++ {
		id := fmt.Sprintf("exec-%d", i)
		if _, ok := store.execs[id]; ok {
			t.Errorf("expected %s to be evicted", id)
		}
	}

	// The newest should still be present.
	newest := fmt.Sprintf("exec-%d", maxCompleted+10)
	if _, ok := store.execs[newest]; !ok {
		t.Errorf("expected %s to still be present", newest)
	}
}

func TestExecutionStoreEvictionKeepsRunning(t *testing.T) {
	store := newExecutionStore()

	// Add maxCompleted+5 completed and 3 still-running executions.
	for i := 0; i < maxCompleted+5; i++ {
		id := fmt.Sprintf("exec-done-%d", i)
		store.execs[id] = &execution{
			id:          id,
			done:        true,
			completedAt: time.Now().Add(time.Duration(i) * time.Second),
			notify:      make(chan struct{}, 1),
		}
	}
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("exec-running-%d", i)
		store.execs[id] = &execution{
			id:     id,
			done:   false,
			notify: make(chan struct{}, 1),
		}
	}

	store.mu.Lock()
	store.evict()
	store.mu.Unlock()

	// Running executions must survive eviction.
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("exec-running-%d", i)
		if _, ok := store.execs[id]; !ok {
			t.Errorf("running execution %s was incorrectly evicted", id)
		}
	}

	// Completed count should be exactly maxCompleted.
	completed := 0
	for _, e := range store.execs {
		if e.done {
			completed++
		}
	}
	if completed != maxCompleted {
		t.Errorf("completed count = %d, want %d", completed, maxCompleted)
	}
}

func TestExecutionStoreGet(t *testing.T) {
	store := newExecutionStore()

	// Empty store returns not found.
	if _, ok := store.get("exec-1"); ok {
		t.Error("expected not found for empty store")
	}

	// Add an execution and retrieve it.
	exec := &execution{id: "exec-1", notify: make(chan struct{}, 1)}
	store.execs["exec-1"] = exec

	got, ok := store.get("exec-1")
	if !ok {
		t.Fatal("expected to find exec-1")
	}
	if got.id != "exec-1" {
		t.Errorf("id = %q", got.id)
	}
}
