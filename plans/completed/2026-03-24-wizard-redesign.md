# Wizard Overlay Redesign

## Date: 2026-03-24

## Context

The wizard used Huh form library with default styling. Redesigned to match the legacy's polished custom UI with a left accent bar, panel background, styled field progression, context-sensitive footer hints, and per-field save control.

## Changes Implemented

### Visual Redesign
- Replaced Huh form with custom field-by-field wizard
- Left accent bar: thick left border in `primary` color, no other borders
- Panel background: `backgroundPanel` color
- Header: green `$` + bold command text
- Completed fields: `name: value` in muted text, aligned 15-char columns
- Current field label: bold `primary` color with `:`
- Centered on screen with `lipgloss.Place()`, width 2/3 of terminal

### Per-Type Field Rendering
- **String/integer/filepath/resourceid**: text input with `> ` prompt
- **Boolean**: `[Yes] [No]` toggle buttons with primary highlight on selected
- **Select**: `> option` cursor navigation with ↑↓ keys
- **Multi-select**: `[x]/[ ]` checkboxes with Space toggle, `>` cursor
- **Secret/password**: password echo mode with `*` per character
- **Secret with saved value**: shows `> ••••••••••  (enter to keep, type to override)` — password echo mode activates on first keystroke

### Context-Sensitive Footer Hints
- Text input: `enter next  shift+tab back  esc cancel`
- Secret pre-filled: `enter accept  type to override  shift+tab back  esc cancel`
- Select: `↑↓ navigate  enter select  esc cancel`
- Multi-select: `Space toggle  Up/Down navigate  Enter confirm`
- Boolean: `← → toggle  enter confirm  esc cancel`
- Save prompt: `← → toggle  enter confirm`

### Parameter Persistence Redesign
- **Always show wizard** for ALL params (removed `ShouldSkip` bypass)
- **Pre-fill saved values** from config.json — text shows value, sensitive shows `••••••••••`
- **Enter on pre-fill** → accepts value, advances (no save prompt)
- **New/changed input** → after Enter, shows "Save for future runs? [Yes/No]" (default No)
- **Yes** → saves to config.json at the field's scope (global/catalog/runbook), then advances
- **No** → advances without saving (ephemeral)
- **Removed auto-save** from `startExecution()` — saving is per-field during wizard
- Wizard receives `ConfigStore` and `Config` for direct persistence

### New Parameter Types
- `multi_select` → checkbox list with Space toggle
- `file_path` → text input (future: completion/validation)
- `resource_id` → text input (future: format validation)

### Command Header
- Only shows `--param` flags for values that differ from config defaults
- Config values are loaded automatically by CLI, no need to show them

## Files Modified

| File | Changes |
|---|---|
| `internal/tui/wizard/model.go` | Complete rewrite: custom wizard with persistence |
| `internal/tui/wizard/model_test.go` | Updated for new model structure |
| `internal/tui/app.go` | Overlay styling, removed auto-save, pass store to wizard |
| `internal/domain/runbook.go` | Added ParamMultiSelect, ParamFilePath, ParamResourceID |
