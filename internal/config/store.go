package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"dops/internal/domain"
)

type ConfigStore interface {
	Load() (*domain.Config, error)
	Save(cfg *domain.Config) error
	EnsureDefaults() (*domain.Config, error)
}

// FileSystem is the subset of adapters.FileSystem that ConfigStore needs.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}

type FileConfigStore struct {
	fs   FileSystem
	path string
}

func NewFileStore(fs FileSystem, path string) *FileConfigStore {
	return &FileConfigStore{fs: fs, path: path}
}

func (s *FileConfigStore) Load() (*domain.Config, error) {
	data, err := s.fs.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func (s *FileConfigStore) Save(cfg *domain.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := s.fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := s.fs.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (s *FileConfigStore) EnsureDefaults() (*domain.Config, error) {
	cfg, err := s.Load()
	if err == nil {
		return cfg, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cfg = defaultConfig()
	if err := s.Save(cfg); err != nil {
		return nil, fmt.Errorf("write default config: %w", err)
	}

	return cfg, nil
}

func defaultConfig() *domain.Config {
	return &domain.Config{
		Theme:    "github",
		Defaults: domain.Defaults{MaxRiskLevel: domain.RiskMedium},
		Catalogs: []domain.Catalog{},
	}
}
