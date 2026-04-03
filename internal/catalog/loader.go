package catalog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dops/internal/adapters"
	"dops/internal/domain"

	"gopkg.in/yaml.v3"
)

type CatalogWithRunbooks struct {
	Catalog  domain.Catalog
	Runbooks []domain.Runbook
	Skills   []domain.Skill
}

type CatalogLoader interface {
	LoadAll(catalogs []domain.Catalog, defaultRisk domain.RiskLevel) ([]CatalogWithRunbooks, error)
	FindByID(id string) (*domain.Runbook, *domain.Catalog, error)
	FindByAlias(alias string) (*domain.Runbook, *domain.Catalog, error)
}

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]os.DirEntry, error)
}

type aliasEntry struct {
	catalogIdx int
	runbookIdx int
}

type DiskCatalogLoader struct {
	fs      FileSystem
	loaded  []CatalogWithRunbooks
	aliases map[string]aliasEntry // alias → location in loaded
}

func NewDiskLoader(fs FileSystem) *DiskCatalogLoader {
	return &DiskCatalogLoader{fs: fs}
}

func (l *DiskCatalogLoader) LoadAll(catalogs []domain.Catalog, defaultRisk domain.RiskLevel) ([]CatalogWithRunbooks, error) {
	var result []CatalogWithRunbooks

	for _, cat := range catalogs {
		if !cat.Active {
			continue
		}

		ceiling := cat.Policy.MaxRiskLevel
		if ceiling == "" {
			ceiling = defaultRisk
		}

		runbooks, skills, err := l.loadCatalog(cat.Name, adapters.ExpandHome(cat.RunbookRoot()), ceiling)
		if err != nil {
			return nil, fmt.Errorf("load catalog %q: %w", cat.Name, err)
		}

		if len(runbooks) > 0 || len(skills) > 0 {
			result = append(result, CatalogWithRunbooks{
				Catalog:  cat,
				Runbooks: runbooks,
				Skills:   skills,
			})
		}
	}

	l.loaded = result
	l.buildAliasIndex()
	return result, nil
}

// buildAliasIndex creates a map from alias → runbook location.
// Duplicate aliases and aliases colliding with IDs are logged and skipped.
func (l *DiskCatalogLoader) buildAliasIndex() {
	l.aliases = make(map[string]aliasEntry)

	// Collect all IDs first to detect collisions.
	ids := make(map[string]bool)
	for _, cwr := range l.loaded {
		for _, rb := range cwr.Runbooks {
			ids[rb.ID] = true
		}
	}

	for ci, cwr := range l.loaded {
		for ri, rb := range cwr.Runbooks {
			l.registerAlias(ids, ci, ri, rb)
		}
	}
}

// registerAlias records each alias for a single runbook, skipping invalid,
// colliding, or duplicate aliases with a logged warning.
func (l *DiskCatalogLoader) registerAlias(ids map[string]bool, ci, ri int, rb domain.Runbook) {
	for _, alias := range rb.Aliases {
		if err := domain.ValidateAlias(alias); err != nil {
			log.Printf("warning: runbook %q: skipping invalid alias %q: %v", rb.ID, alias, err)
			continue
		}
		if ids[alias] {
			log.Printf("warning: runbook %q: alias %q collides with an existing runbook ID, skipping", rb.ID, alias)
			continue
		}
		if existing, ok := l.aliases[alias]; ok {
			existingID := l.loaded[existing.catalogIdx].Runbooks[existing.runbookIdx].ID
			log.Printf("warning: runbook %q: alias %q already claimed by %q, skipping", rb.ID, alias, existingID)
			continue
		}
		l.aliases[alias] = aliasEntry{catalogIdx: ci, runbookIdx: ri}
	}
}

func (l *DiskCatalogLoader) FindByID(id string) (*domain.Runbook, *domain.Catalog, error) {
	for i := range l.loaded {
		for j := range l.loaded[i].Runbooks {
			if l.loaded[i].Runbooks[j].ID == id {
				return &l.loaded[i].Runbooks[j], &l.loaded[i].Catalog, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("runbook %q not found", id)
}

func (l *DiskCatalogLoader) FindByAlias(alias string) (*domain.Runbook, *domain.Catalog, error) {
	entry, ok := l.aliases[alias]
	if !ok {
		return nil, nil, fmt.Errorf("runbook alias %q not found", alias)
	}
	return &l.loaded[entry.catalogIdx].Runbooks[entry.runbookIdx],
		&l.loaded[entry.catalogIdx].Catalog, nil
}

func (l *DiskCatalogLoader) loadCatalog(catalogName, catalogPath string, ceiling domain.RiskLevel) ([]domain.Runbook, []domain.Skill, error) {
	entries, err := l.fs.ReadDir(catalogPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read catalog dir: %w", err)
	}

	var runbooks []domain.Runbook
	var skills []domain.Skill

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		rbPath := filepath.Join(catalogPath, entry.Name(), "runbook.yaml")
		rb, err := l.loadRunbook(rbPath)
		if err != nil {
			return nil, nil, fmt.Errorf("load runbook %q: %w", entry.Name(), err)
		}

		// Generate ID as catalog.runbook if not set in YAML.
		if rb.ID == "" {
			rb.ID = catalogName + "." + entry.Name()
		}

		// Skills: read skill.md, skip adding to runbooks.
		if rb.IsSkill() {
			skillPath := filepath.Join(catalogPath, entry.Name(), "skill.md")
			content, err := l.fs.ReadFile(skillPath)
			if err != nil {
				log.Printf("warning: skill %q missing skill.md, skipping", rb.ID)
				continue
			}
			skills = append(skills, domain.Skill{
				ID:          rb.ID,
				Name:        rb.Name,
				Description: rb.Description,
				Trigger:     rb.Trigger,
				Content:     string(content),
				Catalog:     catalogName,
			})
			continue
		}

		if rb.RiskLevel.Exceeds(ceiling) {
			continue
		}

		runbooks = append(runbooks, *rb)
	}

	return runbooks, skills, nil
}

func (l *DiskCatalogLoader) loadRunbook(path string) (*domain.Runbook, error) {
	data, err := l.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read runbook: %w", err)
	}

	var rb domain.Runbook
	if err := yaml.Unmarshal(data, &rb); err != nil {
		return nil, fmt.Errorf("parse runbook: %w", err)
	}

	return &rb, nil
}

var _ CatalogLoader = (*DiskCatalogLoader)(nil)
