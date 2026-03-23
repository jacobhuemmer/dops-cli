package wizard

import "dops/internal/domain"

type WizardSubmitMsg struct {
	Runbook domain.Runbook
	Catalog domain.Catalog
	Params  map[string]string
}

type WizardCancelMsg struct{}
