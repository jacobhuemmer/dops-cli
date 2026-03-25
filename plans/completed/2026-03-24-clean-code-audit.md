# Clean Code Audit

## Date: 2026-03-24

## Context

Uncle Bob style audit of the entire codebase. Identified 21 items across P0 (critical), P1 (maintainability), and P2 (style). Working through sequentially with manual approval on each change.

## P0 — Critical (Completed)

### P0-1: Surface save errors in wizard
- **File**: `internal/tui/wizard/model.go`
- **Issue**: `config.Set` and `store.Save` errors silently swallowed in `saveCurrentField()`
- **Fix**: Capture errors, display as validation message (`m.err`)

### P0-2: Extract shared ExpandHome
- **Files**: `internal/adapters/fs.go` (new), `internal/tui/app.go`, `internal/mcp/tools.go`, `internal/catalog/loader.go`, `cmd/run.go`
- **Issue**: 4 identical copies of `expandTilde`/`expandHome` across packages
- **Fix**: Single `adapters.ExpandHome()`, all callers delegate

### P0-3: Numeric type validation
- **Files**: `internal/domain/runbook.go`, `internal/tui/wizard/model.go`, `internal/mcp/schema.go`
- **Issue**: Hand-rolled char range validation rejected negative numbers, no decimal support
- **Fix**: Three distinct numeric types:
  - `integer` — any whole number (negative ok), `strconv.Atoi`
  - `number` — non-negative (0+), `strconv.Atoi` + `>= 0`
  - `float` — decimal, `strconv.ParseFloat`

### P0-4: Gzip middleware breaks SSE
- **File**: `internal/mcp/server.go`
- **Issue**: Gzip wrapping buffered SSE events, breaking MCP progress streaming
- **Fix**: Skip gzip when `Accept: text/event-stream` header present

## P1 — Maintainability (Pending)

| # | Issue | File |
|---|-------|------|
| 5 | `Update()` god method (120+ lines) | app.go |
| 6 | `viewNormal()` 110 lines | app.go |
| 7 | Output `View()` 186 lines | output/model.go |
| 8 | Duplicated Yes/No toggle rendering | wizard/model.go |
| 9 | Duplicated nil-style guard pattern | wizard/model.go |
| 10 | `startExecution` 97 lines, dual paths | app.go |
| 11 | Duplicated config loading in cmd/ | cmd/*.go |
| 12 | Mouse translation 4 identical cases | app.go |
| 13 | Coordinate extraction 4 identical cases | app.go |
| 14 | Full re-render for click hit testing | app.go |
| 15 | Layout computation duplicated | app.go |

## P2 — Style (Pending)

| # | Issue | File |
|---|-------|------|
| 16 | Magic numbers in layout | app.go |
| 17 | Dead code: `matchLineSet` | output/model.go |
| 18 | Scrollbar glyph mismatch | output/model.go |
| 19 | Hand-rolled ANSI regex | app.go |
| 20 | Deprecated `strings.Title` | cmd/root.go |
| 21 | Exported `MissingParams` only used in tests | wizard/model.go |
