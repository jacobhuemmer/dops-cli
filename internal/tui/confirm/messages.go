package confirm

import "dops/internal/domain"

// ConfirmAcceptMsg is sent when the user confirms execution.
type ConfirmAcceptMsg struct {
	Runbook domain.Runbook
	Catalog domain.Catalog
	Params  map[string]string
}

// ConfirmCancelMsg is sent when the user cancels confirmation.
type ConfirmCancelMsg struct{}
