package wizard

import (
	"fmt"
	"strings"

	"dops/internal/domain"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type Model struct {
	runbook  domain.Runbook
	catalog  domain.Catalog
	resolved map[string]string
	values   map[string]*string
	form     *huh.Form
	width    int
	height   int
}

func New(rb domain.Runbook, cat domain.Catalog, resolved map[string]string) Model {
	m := Model{
		runbook:  rb,
		catalog:  cat,
		resolved: resolved,
		values:   make(map[string]*string),
	}

	missing := MissingParams(rb.Parameters, resolved)
	if len(missing) == 0 {
		return m
	}

	var fields []huh.Field
	for _, p := range missing {
		val := ""
		if v, ok := resolved[p.Name]; ok {
			val = v
		}
		m.values[p.Name] = &val

		switch p.Type {
		case domain.ParamBoolean:
			boolVal := val == "true"
			boolPtr := &boolVal
			fields = append(fields, huh.NewConfirm().
				Title(p.Name).
				Description(p.Description).
				Value(boolPtr))
		case domain.ParamSelect:
			opts := make([]huh.Option[string], len(p.Options))
			for i, o := range p.Options {
				opts[i] = huh.NewOption(o, o)
			}
			fields = append(fields, huh.NewSelect[string]().
				Title(p.Name).
				Description(p.Description).
				Options(opts...).
				Value(m.values[p.Name]))
		default: // string, integer
			input := huh.NewInput().
				Title(p.Name).
				Description(p.Description).
				Value(m.values[p.Name])
			if p.Secret {
				input = input.EchoMode(huh.EchoModePassword)
			}
			fields = append(fields, input)
		}
	}

	m.form = huh.NewForm(huh.NewGroup(fields...))
	return m
}

func (m Model) Init() tea.Cmd {
	if m.form == nil {
		return nil
	}
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEscape {
			return m, func() tea.Msg { return WizardCancelMsg{} }
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		return m, func() tea.Msg {
			return WizardSubmitMsg{
				Runbook: m.runbook,
				Catalog: m.catalog,
				Params:  m.collectParams(),
			}
		}
	}

	if m.form.State == huh.StateAborted {
		return m, func() tea.Msg { return WizardCancelMsg{} }
	}

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	cmd := BuildCommand(m.runbook, m.mergedParams())
	b.WriteString("  $ " + cmd + "\n")
	b.WriteString(strings.Repeat("─", 60) + "\n")

	if m.form != nil {
		b.WriteString(m.form.View())
	}

	return b.String()
}

func (m Model) collectParams() map[string]string {
	result := make(map[string]string)
	for k, v := range m.resolved {
		result[k] = v
	}
	for k, v := range m.values {
		if v != nil && *v != "" {
			result[k] = *v
		}
	}
	return result
}

func (m Model) mergedParams() map[string]string {
	return m.collectParams()
}

// ShouldSkip returns true when all required parameters are already resolved.
func ShouldSkip(params []domain.Parameter, resolved map[string]string) bool {
	for _, p := range params {
		if !p.Required {
			continue
		}
		if _, ok := resolved[p.Name]; !ok {
			return false
		}
	}
	return true
}

// MissingParams returns parameters that are not yet resolved.
// Includes required params that have no value, and optional params that have no value.
func MissingParams(params []domain.Parameter, resolved map[string]string) []domain.Parameter {
	var missing []domain.Parameter
	for _, p := range params {
		if _, ok := resolved[p.Name]; !ok {
			missing = append(missing, p)
		}
	}
	return missing
}

// BuildCommand formats the dops run command for display.
func BuildCommand(rb domain.Runbook, params map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "dops run %s", rb.ID)
	for _, p := range rb.Parameters {
		if v, ok := params[p.Name]; ok {
			if p.Secret {
				fmt.Fprintf(&b, " --param %s=****", p.Name)
			} else {
				fmt.Fprintf(&b, " --param %s=%s", p.Name, v)
			}
		}
	}
	return b.String()
}
