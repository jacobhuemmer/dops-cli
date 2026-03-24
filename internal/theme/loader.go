package theme

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	"dops/internal/domain"
)

//go:embed tokyonight.json
var bundledTokyonight []byte

//go:embed tokyomidnight.json
var bundledTokyomidnight []byte

type ThemeLoader interface {
	Load(name string) (*domain.ThemeFile, error)
}

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type FileThemeLoader struct {
	fs        FileSystem
	themesDir string
}

func NewFileLoader(fs FileSystem, themesDir string) *FileThemeLoader {
	return &FileThemeLoader{fs: fs, themesDir: themesDir}
}

func (l *FileThemeLoader) Load(name string) (*domain.ThemeFile, error) {
	// 1. Try user theme
	userPath := filepath.Join(l.themesDir, name+".json")
	data, err := l.fs.ReadFile(userPath)
	if err == nil {
		return parseTheme(data)
	}

	// 2. Try bundled theme
	tf, err := l.loadBundled(name)
	if err == nil {
		return tf, nil
	}

	// 3. Fall back to tokyomidnight
	if name != "tokyomidnight" {
		return l.loadBundled("tokyomidnight")
	}

	return nil, fmt.Errorf("theme %q not found and fallback failed", name)
}

func (l *FileThemeLoader) loadBundled(name string) (*domain.ThemeFile, error) {
	switch name {
	case "tokyonight":
		return parseTheme(bundledTokyonight)
	case "tokyomidnight":
		return parseTheme(bundledTokyomidnight)
	default:
		return nil, fmt.Errorf("no bundled theme %q", name)
	}
}

func parseTheme(data []byte) (*domain.ThemeFile, error) {
	var tf domain.ThemeFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parse theme: %w", err)
	}
	return &tf, nil
}

var _ ThemeLoader = (*FileThemeLoader)(nil)
