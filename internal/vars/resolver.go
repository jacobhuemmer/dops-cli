package vars

import (
	"fmt"

	"dops/internal/domain"
)

type VarResolver interface {
	Resolve(cfg *domain.Config, catalogName, runbookName string, params []domain.Parameter) map[string]string
}

type DefaultVarResolver struct{}

func NewDefaultResolver() *DefaultVarResolver {
	return &DefaultVarResolver{}
}

func (r *DefaultVarResolver) Resolve(cfg *domain.Config, catalogName, runbookName string, params []domain.Parameter) map[string]string {
	result := make(map[string]string)

	// Layer 1: global vars
	for k, v := range cfg.Vars.Global {
		result[k] = toString(v)
	}

	// Layer 2: catalog vars (overrides global)
	if cat, ok := cfg.Vars.Catalog[catalogName]; ok {
		for k, v := range cat.Vars {
			result[k] = toString(v)
		}

		// Layer 3: runbook vars (overrides catalog)
		if rb, ok := cat.Runbooks[runbookName]; ok {
			for k, v := range rb {
				result[k] = toString(v)
			}
		}
	}

	// Filter to only requested parameter names
	filtered := make(map[string]string)
	for _, p := range params {
		if v, ok := result[p.Name]; ok {
			filtered[p.Name] = v
		}
	}

	return filtered
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

var _ VarResolver = (*DefaultVarResolver)(nil)
