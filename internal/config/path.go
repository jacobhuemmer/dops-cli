package config

import (
	"fmt"
	"strings"

	"dops/internal/domain"
)

func Get(cfg *domain.Config, keyPath string) (any, error) {
	if keyPath == "" {
		return nil, fmt.Errorf("key path must not be empty")
	}

	parts := strings.Split(keyPath, ".")

	switch parts[0] {
	case "theme":
		if len(parts) != 1 {
			return nil, fmt.Errorf("unknown path: %s", keyPath)
		}
		return cfg.Theme, nil
	case "defaults":
		return getDefaults(cfg, parts[1:])
	case "vars":
		return getVars(cfg, parts[1:])
	default:
		return nil, fmt.Errorf("unknown top-level key: %q", parts[0])
	}
}

func Set(cfg *domain.Config, keyPath string, value any) error {
	if keyPath == "" {
		return fmt.Errorf("key path must not be empty")
	}

	parts := strings.Split(keyPath, ".")

	switch parts[0] {
	case "theme":
		if len(parts) != 1 {
			return fmt.Errorf("unknown path: %s", keyPath)
		}
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("theme must be a string")
		}
		cfg.Theme = s
		return nil
	case "defaults":
		return setDefaults(cfg, parts[1:], value)
	case "vars":
		return setVars(cfg, parts[1:], value)
	default:
		return fmt.Errorf("unknown top-level key: %q", parts[0])
	}
}

func Unset(cfg *domain.Config, keyPath string) error {
	if keyPath == "" {
		return fmt.Errorf("key path must not be empty")
	}

	parts := strings.Split(keyPath, ".")

	if parts[0] != "vars" || len(parts) < 3 {
		return fmt.Errorf("can only unset vars paths, got: %s", keyPath)
	}

	return unsetVars(cfg, parts[1:])
}

func getDefaults(cfg *domain.Config, parts []string) (any, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("incomplete defaults path")
	}
	switch parts[0] {
	case "max_risk_level":
		return cfg.Defaults.MaxRiskLevel, nil
	default:
		return nil, fmt.Errorf("unknown defaults key: %q", parts[0])
	}
}

func setDefaults(cfg *domain.Config, parts []string, value any) error {
	if len(parts) == 0 {
		return fmt.Errorf("incomplete defaults path")
	}
	switch parts[0] {
	case "max_risk_level":
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("max_risk_level must be a string")
		}
		rl, err := domain.ParseRiskLevel(s)
		if err != nil {
			return err
		}
		cfg.Defaults.MaxRiskLevel = rl
		return nil
	default:
		return fmt.Errorf("unknown defaults key: %q", parts[0])
	}
}

// getVars handles: vars.global.<key>, vars.catalog.<cat>.<key>, vars.catalog.<cat>.runbooks.<rb>.<key>
func getVars(cfg *domain.Config, parts []string) (any, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("incomplete vars path")
	}

	switch parts[0] {
	case "global":
		// vars.global.<key>
		key := parts[1]
		if cfg.Vars.Global == nil {
			return nil, fmt.Errorf("global var %q not found", key)
		}
		val, ok := cfg.Vars.Global[key]
		if !ok {
			return nil, fmt.Errorf("global var %q not found", key)
		}
		return val, nil

	case "catalog":
		return getCatalogVars(cfg, parts[1:])

	default:
		return nil, fmt.Errorf("unknown vars scope: %q", parts[0])
	}
}

// getCatalogVars handles: <cat>.<key> or <cat>.runbooks.<rb>.<key>
func getCatalogVars(cfg *domain.Config, parts []string) (any, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("incomplete catalog vars path")
	}

	catName := parts[0]
	cat, ok := cfg.Vars.Catalog[catName]
	if !ok {
		return nil, fmt.Errorf("catalog %q not found in vars", catName)
	}

	if parts[1] == "runbooks" {
		// vars.catalog.<cat>.runbooks.<rb>.<key>
		if len(parts) < 4 {
			return nil, fmt.Errorf("incomplete runbook vars path")
		}
		rbName := parts[2]
		rb, ok := cat.Runbooks[rbName]
		if !ok {
			return nil, fmt.Errorf("runbook %q not found in catalog %q", rbName, catName)
		}
		key := parts[3]
		val, ok := rb[key]
		if !ok {
			return nil, fmt.Errorf("runbook var %q not found", key)
		}
		return val, nil
	}

	// vars.catalog.<cat>.<key>
	key := parts[1]
	val, ok := cat.Vars[key]
	if !ok {
		return nil, fmt.Errorf("catalog var %q not found", key)
	}
	return val, nil
}

func setVars(cfg *domain.Config, parts []string, value any) error {
	if len(parts) < 2 {
		return fmt.Errorf("incomplete vars path")
	}

	switch parts[0] {
	case "global":
		// vars.global.<key>
		if cfg.Vars.Global == nil {
			cfg.Vars.Global = make(map[string]any)
		}
		cfg.Vars.Global[parts[1]] = value
		return nil

	case "catalog":
		return setCatalogVars(cfg, parts[1:], value)

	default:
		return fmt.Errorf("unknown vars scope: %q", parts[0])
	}
}

func setCatalogVars(cfg *domain.Config, parts []string, value any) error {
	if len(parts) < 2 {
		return fmt.Errorf("incomplete catalog vars path")
	}

	catName := parts[0]
	if cfg.Vars.Catalog == nil {
		cfg.Vars.Catalog = make(map[string]domain.CatalogVars)
	}
	cat := cfg.Vars.Catalog[catName]
	if cat.Vars == nil {
		cat.Vars = make(map[string]any)
	}
	if cat.Runbooks == nil {
		cat.Runbooks = make(map[string]map[string]any)
	}

	if parts[1] == "runbooks" {
		// vars.catalog.<cat>.runbooks.<rb>.<key>
		if len(parts) < 4 {
			return fmt.Errorf("incomplete runbook vars path")
		}
		rbName := parts[2]
		rb := cat.Runbooks[rbName]
		if rb == nil {
			rb = make(map[string]any)
		}
		rb[parts[3]] = value
		cat.Runbooks[rbName] = rb
		cfg.Vars.Catalog[catName] = cat
		return nil
	}

	// vars.catalog.<cat>.<key>
	cat.Vars[parts[1]] = value
	cfg.Vars.Catalog[catName] = cat
	return nil
}

func unsetVars(cfg *domain.Config, parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("incomplete vars path")
	}

	switch parts[0] {
	case "global":
		key := parts[1]
		if cfg.Vars.Global == nil {
			return fmt.Errorf("global var %q not found", key)
		}
		if _, ok := cfg.Vars.Global[key]; !ok {
			return fmt.Errorf("global var %q not found", key)
		}
		delete(cfg.Vars.Global, key)
		return nil

	case "catalog":
		return unsetCatalogVars(cfg, parts[1:])

	default:
		return fmt.Errorf("unknown vars scope: %q", parts[0])
	}
}

func unsetCatalogVars(cfg *domain.Config, parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("incomplete catalog vars path")
	}

	catName := parts[0]
	cat, ok := cfg.Vars.Catalog[catName]
	if !ok {
		return fmt.Errorf("catalog %q not found in vars", catName)
	}

	if parts[1] == "runbooks" {
		if len(parts) < 4 {
			return fmt.Errorf("incomplete runbook vars path")
		}
		rbName := parts[2]
		rb, ok := cat.Runbooks[rbName]
		if !ok {
			return fmt.Errorf("runbook %q not found in catalog %q", rbName, catName)
		}
		key := parts[3]
		if _, ok := rb[key]; !ok {
			return fmt.Errorf("runbook var %q not found", key)
		}
		delete(rb, key)
		cat.Runbooks[rbName] = rb
		cfg.Vars.Catalog[catName] = cat
		return nil
	}

	// vars.catalog.<cat>.<key>
	key := parts[1]
	if _, ok := cat.Vars[key]; !ok {
		return fmt.Errorf("catalog var %q not found", key)
	}
	delete(cat.Vars, key)
	cfg.Vars.Catalog[catName] = cat
	return nil
}
