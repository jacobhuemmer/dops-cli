package update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.2.0", "0.1.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.1", "0.1.0", true},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", false},
		{"v0.2.0", "0.1.0", true},  // v-prefix tolerance
		{"0.2.0", "v0.1.0", true},  // v-prefix tolerance
		{"1.0.0-rc1", "0.9.0", true}, // pre-release stripped
		{"", "0.1.0", false},
		{"0.1.0", "", false},
		{"bad", "0.1.0", false},
	}

	for _, tt := range tests {
		got := isNewer(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  []int
	}{
		{"1.2.3", []int{1, 2, 3}},
		{"v0.1.0", []int{0, 1, 0}},
		{"0.2.0-rc1", []int{0, 2, 0}},
		{"bad", nil},
		{"1.2", nil},
		{"", nil},
	}

	for _, tt := range tests {
		got := parseSemver(tt.input)
		if tt.want == nil {
			if got != nil {
				t.Errorf("parseSemver(%q) = %v, want nil", tt.input, got)
			}
			continue
		}
		if got == nil || got[0] != tt.want[0] || got[1] != tt.want[1] || got[2] != tt.want[2] {
			t.Errorf("parseSemver(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"", true},
		{"dev", true},
		{"0.1.0-dev", true},
		{"0.1.0", false},
		{"1.0.0", false},
	}

	for _, tt := range tests {
		got := shouldSkip(tt.version)
		if got != tt.want {
			t.Errorf("shouldSkip(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestReadCache_Fresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cacheFile)

	c := cache{
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Latest:    "0.2.0",
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0o644)

	result, ok := readCache(path, "0.1.0")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !result.Available || result.Latest != "0.2.0" {
		t.Errorf("got %+v, want Available=true Latest=0.2.0", result)
	}
}

func TestReadCache_Stale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cacheFile)

	c := cache{
		CheckedAt: time.Now().Add(-25 * time.Hour).UTC().Format(time.RFC3339),
		Latest:    "0.2.0",
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0o644)

	_, ok := readCache(path, "0.1.0")
	if ok {
		t.Fatal("expected cache miss for stale entry")
	}
}

func TestReadCache_NoUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cacheFile)

	c := cache{
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Latest:    "0.1.0",
	}
	data, _ := json.Marshal(c)
	os.WriteFile(path, data, 0o644)

	result, ok := readCache(path, "0.1.0")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if result.Available {
		t.Error("expected Available=false when versions match")
	}
}

// withMockAPI starts an httptest server, points apiBaseURL at it, and restores
// the original URL when the test finishes.
func withMockAPI(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	orig := apiBaseURL
	apiBaseURL = srv.URL
	t.Cleanup(func() {
		apiBaseURL = orig
		srv.Close()
	})
	return srv
}

func TestFetchLatest_Success(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.3.0"})
	}))

	got, err := fetchLatest()
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.3.0" {
		t.Errorf("got %q, want %q", got, "0.3.0")
	}
}

func TestFetchLatest_NonOK(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	_, err := fetchLatest()
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestFetchLatest_BadJSON(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))

	_, err := fetchLatest()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", cacheFile) // sub-dir to test MkdirAll

	writeCache(path, "1.2.3")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("cache file not written:", err)
	}

	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal("invalid JSON in cache:", err)
	}
	if c.Latest != "1.2.3" {
		t.Errorf("cached latest = %q, want %q", c.Latest, "1.2.3")
	}
	if _, err := time.Parse(time.RFC3339, c.CheckedAt); err != nil {
		t.Errorf("cached checked_at not valid RFC3339: %v", err)
	}
}

func TestCheck_UpdateAvailable(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v1.0.0"})
	}))

	dir := t.TempDir()
	result := Check("0.1.0", dir)
	if !result.Available {
		t.Error("expected Available=true")
	}
	if result.Latest != "1.0.0" {
		t.Errorf("Latest = %q, want %q", result.Latest, "1.0.0")
	}

	// Verify cache was written.
	data, err := os.ReadFile(filepath.Join(dir, cacheFile))
	if err != nil {
		t.Fatal("cache not written after Check")
	}
	var c cache
	json.Unmarshal(data, &c)
	if c.Latest != "1.0.0" {
		t.Errorf("cached latest = %q", c.Latest)
	}
}

func TestCheck_NoUpdate(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v0.1.0"})
	}))

	dir := t.TempDir()
	result := Check("0.1.0", dir)
	if result.Available {
		t.Error("expected Available=false when versions match")
	}
}

func TestCheck_SkipsDevVersion(t *testing.T) {
	result := Check("dev", t.TempDir())
	if result.Available {
		t.Error("expected Available=false for dev version")
	}
}

func TestCheck_UsesCache(t *testing.T) {
	// Pre-populate a fresh cache.
	dir := t.TempDir()
	c := cache{
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Latest:    "2.0.0",
	}
	data, _ := json.Marshal(c)
	os.WriteFile(filepath.Join(dir, cacheFile), data, 0o644)

	// Point API at a server that would return a different version.
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(githubRelease{TagName: "v3.0.0"})
	}))

	result := Check("0.1.0", dir)
	if result.Latest != "2.0.0" {
		t.Errorf("should use cached version 2.0.0, got %q", result.Latest)
	}
}

func TestCheck_APIError(t *testing.T) {
	withMockAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	result := Check("0.1.0", t.TempDir())
	if result.Available {
		t.Error("expected Available=false on API error")
	}
}
