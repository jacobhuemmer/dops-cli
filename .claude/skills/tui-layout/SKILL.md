---
name: tui-layout
description: Full-screen TUI layout patterns for BubbleTea v2 and Lipgloss v2. Use when building or modifying View() layouts, panel sizing, container alignment, or responsive terminal UI structure. Triggers on viewNormal, JoinVertical, JoinHorizontal, Width, Height, panelRows, WindowSizeMsg handling, or any layout calculation.
user-invocable: false
---

# TUI Layout Pattern (OpenCode Style)

Full-screen BubbleTea v2 layouts follow a root compositor pattern. There is no explicit "outer frame widget" — the root `View()` method IS the outer container.

## Core Pattern

```go
func (m model) View() tea.View {
    // 1. Guard: don't render until terminal size is known
    if m.width == 0 || m.height == 0 {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }

    // 2. Build panels from budget
    content := m.buildLayout()

    // 3. Outer container: enforce exact terminal dimensions
    content = lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Render(content)

    v := tea.NewView(content)
    v.AltScreen = true
    return v
}
```

## Layout Budget

Work top-down from the known terminal size:

```
m.height (total rows)
├── panelRows = m.height - footerH
│   ├── sidebar (left): Height = panelRows
│   └── right panel: Height = panelRows (forced)
│       ├── metadata: auto-height (measured)
│       └── output: fills remaining
└── footer: footerH rows

m.width (total cols)
├── margin (left + right)
├── sidebar: sw + 2 (border)
├── gap: 1
└── right panel: remaining
```

## Rules

### 1. Measure, don't guess
Always render auto-height components first, then measure with `lipgloss.Height()`:

```go
metaView := buildMetadata(...)
metaActualH := lipgloss.Height(metaView) // includes borders, padding, ANSI

outputH := panelRows - metaActualH - 2  // remainder for output
```

Never hardcode heights like `metaH := 8`. Content, borders, and line wrapping can change the actual rendered height.

### 2. Force matching heights with Width/Height
When two columns must align, force them to the same height:

```go
// Right panel forced to exact panelRows
rightPanel = lipgloss.NewStyle().
    Height(panelRows).
    Width(rightW).
    Render(rightPanel)

// Sidebar also panelRows (contentH + 2 border = panelRows)
sidebarView = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Width(sw).
    Height(panelRows - 2).
    Render(sidebarContent)
```

### 3. Outer container is mandatory
Always wrap the final composed content in `Width(m.width).Height(m.height)`:

```go
content = lipgloss.NewStyle().
    Width(m.width).
    Height(m.height).
    Render(content)
```

This prevents overflow, fills remaining space, and ensures consistent rendering in alt screen. Without it, content may clip, shift, or leave gaps.

### 4. Use margins for centering
Don't let panels touch the terminal edge:

```go
margin := 1
innerW := m.width - (margin * 2)
// ... compute panel widths from innerW ...

body = lipgloss.NewStyle().
    MarginLeft(margin).
    Render(body)
```

### 5. Size propagation is manual
Each child component must receive its allocated dimensions explicitly. Use `SetSize()`, `SetHeight()`, or pass dimensions via constructor/method:

```go
m.sidebar.SetHeight(sidebarContentH)
outputView := m.output.ViewWithSize(rightW, outputContentH)
```

### 6. Border accounting
Lipgloss borders add rows/cols to the rendered output:
- `Border(RoundedBorder())` adds 2 rows (top + bottom) and 2 cols (left + right)
- Content width inside border = `Width` value
- Rendered width = `Width` + 2
- Use `GetHorizontalFrameSize()` / `GetVerticalFrameSize()` for precise accounting

### 7. JoinHorizontal aligns at the top
`lipgloss.JoinHorizontal(lipgloss.Top, left, right)` aligns columns at their first row. If one column is shorter, it leaves blank space at the bottom. To align bottoms, both columns must have the same height (see rule 2).

## Anti-Patterns

- **Don't use padding/newlines to push footer down** — use the outer container `Height` instead
- **Don't hardcode panel heights** — derive from `m.height` and measured components
- **Don't rely on `MaxHeight` for clipping** — it clips from the bottom, hiding footers
- **Don't compute layout in Update()** — compute in View() from current width/height
- **Don't skip the outer container** — without `Width.Height`, alt screen layouts break
