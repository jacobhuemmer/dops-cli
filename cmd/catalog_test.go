package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dops/internal/domain"
)

func TestIsValidGitRef(t *testing.T) {
	tests := []struct {
		ref  string
		want bool
	}{
		{"main", true},
		{"v1.0.0", true},
		{"feature/branch-name", true},
		{"abc123", true},
		{"release-1.2.3", true},
		{"refs/tags/v1", true},
		{"my_branch", true},
		{"", false},
		{"ref with spaces", false},
		{"ref;injection", false},
		{"ref&cmd", false},
		{"ref|pipe", false},
		{"ref$(cmd)", false},
		{"ref`cmd`", false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := isValidGitRef(tt.ref)
			if got != tt.want {
				t.Errorf("isValidGitRef(%q) = %v, want %v", tt.ref, got, tt.want)
			}
		})
	}
}

func TestValidateSubPath(t *testing.T) {
	t.Run("valid subdir", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "runbooks")
		os.MkdirAll(sub, 0o755)

		got, err := validateSubPath(root, "runbooks")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "runbooks" {
			t.Errorf("got %q, want %q", got, "runbooks")
		}
	})

	t.Run("nested subdir", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "a", "b")
		os.MkdirAll(sub, 0o755)

		got, err := validateSubPath(root, "a/b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != filepath.Clean("a/b") {
			t.Errorf("got %q, want %q", got, filepath.Clean("a/b"))
		}
	})

	t.Run("absolute path rejected", func(t *testing.T) {
		root := t.TempDir()
		_, err := validateSubPath(root, "/etc/passwd")
		if err == nil {
			t.Fatal("expected error for absolute path")
		}
		if !strings.Contains(err.Error(), "relative path") {
			t.Errorf("error %q should mention relative path", err.Error())
		}
	})

	t.Run("parent traversal rejected", func(t *testing.T) {
		root := t.TempDir()
		_, err := validateSubPath(root, "../escape")
		if err == nil {
			t.Fatal("expected error for parent traversal")
		}
	})

	t.Run("nonexistent subdir rejected", func(t *testing.T) {
		root := t.TempDir()
		_, err := validateSubPath(root, "noexist")
		if err == nil {
			t.Fatal("expected error for nonexistent subdir")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error %q should mention does not exist", err.Error())
		}
	})

	t.Run("file not directory rejected", func(t *testing.T) {
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, "afile"), []byte("x"), 0o644)

		_, err := validateSubPath(root, "afile")
		if err == nil {
			t.Fatal("expected error for file (not directory)")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("error %q should mention not a directory", err.Error())
		}
	})
}

// initDopsDir creates a minimal dopsDir with config for testing catalog commands.
func initDopsDir(t *testing.T, catalogs ...domain.Catalog) string {
	t.Helper()
	dir := t.TempDir()
	dopsDir := filepath.Join(dir, ".dops")
	os.MkdirAll(dopsDir, 0o755)

	cfg := &domain.Config{
		Theme:    "tokyonight",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskMedium},
		Catalogs: catalogs,
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(dopsDir, "config.json"), data, 0o644)

	return dopsDir
}

func readCatalogConfig(t *testing.T, dopsDir string) *domain.Config {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dopsDir, "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	return &cfg
}

func TestCatalogList(t *testing.T) {
	t.Run("empty catalog list", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "list"}, dopsDir)
		if err != nil {
			t.Fatalf("catalog list: %v", err)
		}
	})

	t.Run("lists existing catalogs", func(t *testing.T) {
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "mycat",
			Path:   "/some/path",
			Active: true,
		})
		_, err := executeCmd([]string{"catalog", "list"}, dopsDir)
		if err != nil {
			t.Fatalf("catalog list: %v", err)
		}
	})
}

func TestCatalogAdd(t *testing.T) {
	t.Run("add local directory", func(t *testing.T) {
		dopsDir := initDopsDir(t)

		// Create a directory to add as catalog.
		catalogDir := t.TempDir()

		_, err := executeCmd([]string{"catalog", "add", catalogDir}, dopsDir)
		if err != nil {
			t.Fatalf("catalog add: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if len(cfg.Catalogs) != 1 {
			t.Fatalf("catalogs = %d, want 1", len(cfg.Catalogs))
		}
		if !cfg.Catalogs[0].Active {
			t.Error("catalog should be active")
		}
	})

	t.Run("add nonexistent path", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "add", "/nonexistent/path/xyz"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
		if !strings.Contains(err.Error(), "path does not exist") {
			t.Errorf("error %q should mention path does not exist", err.Error())
		}
	})

	t.Run("add file instead of directory", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		f := filepath.Join(t.TempDir(), "afile")
		os.WriteFile(f, []byte("x"), 0o644)

		_, err := executeCmd([]string{"catalog", "add", f}, dopsDir)
		if err == nil {
			t.Fatal("expected error for file path")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("error %q should mention not a directory", err.Error())
		}
	})

	t.Run("duplicate catalog name rejected", func(t *testing.T) {
		catalogDir := t.TempDir()
		name := filepath.Base(catalogDir)

		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   name,
			Path:   "/original/path",
			Active: true,
		})

		_, err := executeCmd([]string{"catalog", "add", catalogDir}, dopsDir)
		if err == nil {
			t.Fatal("expected error for duplicate catalog")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error %q should mention already exists", err.Error())
		}
	})

	t.Run("add with display name", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		catalogDir := t.TempDir()

		_, err := executeCmd([]string{"catalog", "add", "--display-name", "My Catalog", catalogDir}, dopsDir)
		if err != nil {
			t.Fatalf("catalog add: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if cfg.Catalogs[0].DisplayName != "My Catalog" {
			t.Errorf("display name = %q, want %q", cfg.Catalogs[0].DisplayName, "My Catalog")
		}
	})

	t.Run("add with invalid display name", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		catalogDir := t.TempDir()

		longName := strings.Repeat("a", 51)
		_, err := executeCmd([]string{"catalog", "add", "--display-name", longName, catalogDir}, dopsDir)
		if err == nil {
			t.Fatal("expected error for display name too long")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "add"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for missing arg")
		}
	})
}

func TestCatalogRemove(t *testing.T) {
	t.Run("remove existing catalog", func(t *testing.T) {
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "mycat",
			Path:   "/some/path",
			Active: true,
		})

		_, err := executeCmd([]string{"catalog", "remove", "mycat"}, dopsDir)
		if err != nil {
			t.Fatalf("catalog remove: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if len(cfg.Catalogs) != 0 {
			t.Errorf("catalogs = %d, want 0", len(cfg.Catalogs))
		}
	})

	t.Run("remove nonexistent catalog", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "remove", "nonexistent"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for nonexistent catalog")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error %q should mention not found", err.Error())
		}
	})

	t.Run("remove only the named catalog", func(t *testing.T) {
		dopsDir := initDopsDir(t,
			domain.Catalog{Name: "keep", Path: "/keep", Active: true},
			domain.Catalog{Name: "drop", Path: "/drop", Active: true},
		)

		_, err := executeCmd([]string{"catalog", "remove", "drop"}, dopsDir)
		if err != nil {
			t.Fatalf("catalog remove: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if len(cfg.Catalogs) != 1 {
			t.Fatalf("catalogs = %d, want 1", len(cfg.Catalogs))
		}
		if cfg.Catalogs[0].Name != "keep" {
			t.Errorf("remaining catalog = %q, want keep", cfg.Catalogs[0].Name)
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "remove"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for missing arg")
		}
	})
}

func TestCatalogInstall_FlagValidation(t *testing.T) {
	t.Run("invalid risk level rejected early", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{
			"catalog", "install",
			"--risk", "yolo",
			"https://example.com/org/repo.git",
		}, dopsDir)
		if err == nil {
			t.Fatal("expected error for invalid risk level")
		}
		if !strings.Contains(err.Error(), "invalid risk level") {
			t.Errorf("error %q should mention invalid risk level", err.Error())
		}
	})

	t.Run("invalid display name rejected early", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		longName := strings.Repeat("x", 51)
		_, err := executeCmd([]string{
			"catalog", "install",
			"--display-name", longName,
			"https://example.com/org/repo.git",
		}, dopsDir)
		if err == nil {
			t.Fatal("expected error for invalid display name")
		}
	})

	t.Run("existing directory rejected", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		catalogsDir := filepath.Join(dopsDir, "catalogs")
		os.MkdirAll(filepath.Join(catalogsDir, "repo"), 0o755)

		_, err := executeCmd([]string{
			"catalog", "install",
			"https://example.com/org/repo.git",
		}, dopsDir)
		if err == nil {
			t.Fatal("expected error for existing directory")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error %q should mention already exists", err.Error())
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "install"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for missing arg")
		}
	})
}

func TestCatalogUpdate_Validation(t *testing.T) {
	t.Run("nonexistent catalog rejected", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "update", "nonexistent"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for nonexistent catalog")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error %q should mention not found", err.Error())
		}
	})

	t.Run("local-only catalog cannot update without display-name", func(t *testing.T) {
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "local",
			Path:   "/some/path",
			Active: true,
			// No URL — local only.
		})

		_, err := executeCmd([]string{"catalog", "update", "local"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for local-only catalog")
		}
		if !strings.Contains(err.Error(), "local-only") {
			t.Errorf("error %q should mention local-only", err.Error())
		}
	})

	t.Run("invalid risk level rejected", func(t *testing.T) {
		// Need a catalog with URL and a real path for EvalSymlinks.
		catPath := t.TempDir()
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "mycat",
			Path:   catPath,
			URL:    "https://example.com/org/repo.git",
			Active: true,
		})

		_, err := executeCmd([]string{
			"catalog", "update",
			"--risk", "invalid",
			"--display-name", "skip-git", // set display-name to avoid git pull
			"mycat",
		}, dopsDir)
		if err == nil {
			t.Fatal("expected error for invalid risk level")
		}
		if !strings.Contains(err.Error(), "invalid risk level") {
			t.Errorf("error %q should mention invalid risk level", err.Error())
		}
	})

	t.Run("display-name only update on local catalog succeeds", func(t *testing.T) {
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "local",
			Path:   "/some/path",
			Active: true,
		})

		_, err := executeCmd([]string{
			"catalog", "update",
			"--display-name", "New Name",
			"local",
		}, dopsDir)
		if err != nil {
			t.Fatalf("update display-name: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if cfg.Catalogs[0].DisplayName != "New Name" {
			t.Errorf("display name = %q, want %q", cfg.Catalogs[0].DisplayName, "New Name")
		}
	})

	t.Run("clear display-name", func(t *testing.T) {
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:        "local",
			DisplayName: "Old Name",
			Path:        "/some/path",
			Active:      true,
		})

		_, err := executeCmd([]string{
			"catalog", "update",
			"--display-name", "",
			"local",
		}, dopsDir)
		if err != nil {
			t.Fatalf("update display-name: %v", err)
		}

		cfg := readCatalogConfig(t, dopsDir)
		if cfg.Catalogs[0].DisplayName != "" {
			t.Errorf("display name = %q, want empty", cfg.Catalogs[0].DisplayName)
		}
	})

	t.Run("invalid ref rejected", func(t *testing.T) {
		catPath := t.TempDir()
		dopsDir := initDopsDir(t, domain.Catalog{
			Name:   "mycat",
			Path:   catPath,
			URL:    "https://example.com/org/repo.git",
			Active: true,
		})

		_, err := executeCmd([]string{
			"catalog", "update",
			"--ref", "ref with spaces",
			"--display-name", "skip", // set to avoid default git pull path
			"mycat",
		}, dopsDir)
		if err == nil {
			t.Fatal("expected error for invalid ref")
		}
		if !strings.Contains(err.Error(), "invalid git ref") {
			t.Errorf("error %q should mention invalid git ref", err.Error())
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		dopsDir := initDopsDir(t)
		_, err := executeCmd([]string{"catalog", "update"}, dopsDir)
		if err == nil {
			t.Fatal("expected error for missing arg")
		}
	})
}

func TestCatalogCmd_SubcommandWiring(t *testing.T) {
	cmd := newCatalogCmd("/tmp/fake")
	subNames := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}

	expected := []string{"list", "add", "remove", "install", "update"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestRegisterCatalog(t *testing.T) {
	dopsDir := initDopsDir(t)
	cat := domain.Catalog{Name: "new-cat", Path: "/some/path", Active: true}

	if err := registerCatalog(dopsDir, cat); err != nil {
		t.Fatalf("registerCatalog: %v", err)
	}

	cfg := readCatalogConfig(t, dopsDir)
	if len(cfg.Catalogs) != 1 {
		t.Fatalf("catalogs = %d, want 1", len(cfg.Catalogs))
	}
	if cfg.Catalogs[0].Name != "new-cat" {
		t.Errorf("name = %q, want new-cat", cfg.Catalogs[0].Name)
	}
}

func TestRegisterCatalog_AppendsToPrevious(t *testing.T) {
	dopsDir := initDopsDir(t, domain.Catalog{Name: "existing", Path: "/old", Active: true})
	cat := domain.Catalog{Name: "new", Path: "/new", Active: true}

	if err := registerCatalog(dopsDir, cat); err != nil {
		t.Fatalf("registerCatalog: %v", err)
	}

	cfg := readCatalogConfig(t, dopsDir)
	if len(cfg.Catalogs) != 2 {
		t.Fatalf("catalogs = %d, want 2", len(cfg.Catalogs))
	}
}

func TestCloneRepo_ExistingDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "existing")
	os.MkdirAll(target, 0o755)

	err := cloneRepo("https://example.com/repo.git", target, "")
	if err == nil {
		t.Fatal("expected error for existing directory")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q should mention already exists", err.Error())
	}
}

func TestVersionCmd(t *testing.T) {
	dopsDir := initDopsDir(t)
	out, err := executeCmd([]string{"version"}, dopsDir)
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, "version") {
		t.Errorf("output %q should contain 'version'", out)
	}
}

func TestCheckmark(t *testing.T) {
	if checkmark(true) != "\u2713" {
		t.Error("checkmark(true) should be ✓")
	}
	if checkmark(false) != "\u2717" {
		t.Error("checkmark(false) should be ✗")
	}
}

func TestLoadSaveConfig_RoundTrip(t *testing.T) {
	dopsDir := initDopsDir(t, domain.Catalog{Name: "test", Path: "/test", Active: true})

	cfg, err := loadConfig(dopsDir)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(cfg.Catalogs) != 1 {
		t.Fatalf("catalogs = %d, want 1", len(cfg.Catalogs))
	}

	cfg.Catalogs = append(cfg.Catalogs, domain.Catalog{Name: "new", Path: "/new", Active: true})
	if err := saveConfig(dopsDir, cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	cfg2, err := loadConfig(dopsDir)
	if err != nil {
		t.Fatalf("loadConfig after save: %v", err)
	}
	if len(cfg2.Catalogs) != 2 {
		t.Fatalf("catalogs = %d, want 2", len(cfg2.Catalogs))
	}
}

func TestMCPCmd_SubcommandWiring(t *testing.T) {
	cmd := newMCPCmd("/tmp/fake")
	subNames := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}
	if !subNames["serve"] {
		t.Error("missing serve subcommand")
	}
	if !subNames["tools"] {
		t.Error("missing tools subcommand")
	}
}

func TestCloneRepo_InvalidRef(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "newrepo")

	err := cloneRepo("https://example.com/repo.git", target, "ref with spaces")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
	if !strings.Contains(err.Error(), "invalid git ref") {
		t.Errorf("error %q should mention invalid git ref", err.Error())
	}
}
