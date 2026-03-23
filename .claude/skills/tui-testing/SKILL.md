---
name: tui-testing
description: Visual testing of TUI and CLI output using VHS and Freeze. Use when making changes to TUI views, CLI output formatting, theme styling, layout, or any visual component. Must be used after any change to View() functions, lipgloss styles, footer, sidebar, metadata, output pane, wizard overlay, or CLI error formatting.
user-invocable: false
---

# Visual Testing with VHS and Freeze

After any change to visual output (TUI or CLI), verify the result using VHS or Freeze before considering the change complete.

## When to Use Which

| Tool | Use For |
|---|---|
| **VHS** | TUI testing — interactive flows, navigation, execution, wizard, multi-step sequences |
| **Freeze** | CLI output testing — static command output, error formatting, config list |

## VHS — TUI Testing

VHS records terminal sessions from `.tape` scripts and produces screenshots/GIFs.

**Directory structure:**
- `tapes/` — tape scripts (`.tape` files)
- `tapes/screenshots/` — output screenshots (`.png`) and GIFs (`.gif`)

**Write tapes in `tapes/`, output to `tapes/screenshots/`:**

```tape
Output tapes/screenshots/test-name.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

Screenshot tapes/screenshots/test-name.png

Type "q"
Sleep 500ms
```

**Run:** `vhs tapes/test-name.tape`

**Then read the screenshot** to verify the visual result matches expectations.

**Key VHS commands:**
- `Type "text"` — type characters
- `Enter`, `Up`, `Down`, `Left`, `Right`, `Escape`, `Tab` — key presses
- `Sleep 500ms` / `Sleep 2s` — wait for rendering/execution
- `Screenshot tapes/name.png` — capture current state
- `Output tapes/name.gif` — set GIF output

**IMPORTANT:** Always prefix with `DOPS_NO_ALT_SCREEN=1` in VHS tapes. The real app uses alt screen (full terminal takeover), but VHS can't capture BubbleTea v2 declarative alt screen.
```tape
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
```

**Best practices:**
- Always `Sleep 2s` after launching the TUI to let it render
- `Sleep 500ms` after navigation for state to update
- `Sleep 2s` or more after `Enter` that triggers script execution
- Take screenshots at each meaningful state change
- Name screenshots descriptively: `tapes/sidebar-collapse.png`, `tapes/wizard-form.png`
- End tapes with `Type "q"` and `Sleep 500ms` to cleanly exit

**Testing flow for TUI changes:**
1. Make the code change
2. `go build -o ./dops .`
3. Write or update a `.tape` file targeting the changed behavior
4. `vhs tapes/the-tape.tape`
5. Read the screenshot(s) to verify
6. If wrong, fix and repeat from step 2

## Freeze — CLI Output Testing

Freeze captures static terminal output as styled screenshots.

**Usage (output to `tapes/screenshots/`):**
```bash
./dops run unknown.runbook 2>&1 | freeze -o tapes/screenshots/cli-error.png
./dops config list 2>&1 | freeze -o tapes/screenshots/cli-config.png
./dops version 2>&1 | freeze -o tapes/screenshots/cli-version.png
```

**With custom styling:**
```bash
./dops config list 2>&1 | freeze -o tapes/screenshots/output.png --theme "Catppuccin Mocha"
```

**Testing flow for CLI changes:**
1. Make the code change
2. `go build -o ./dops .`
3. Run the CLI command piped to freeze
4. Read the screenshot to verify
5. If wrong, fix and repeat

## Limitations

**VHS does not support mouse events.** Mouse click behavior must be tested via unit tests using synthetic `tea.MouseClickMsg` messages. VHS can only test keyboard-driven interactions.

For mouse testing, write unit tests like:
```go
m, cmd := m.Update(tea.MouseClickMsg{X: 5, Y: 2, Button: tea.MouseLeft})
```

Run mouse tests with: `go test ./internal/tui/sidebar/ -run "Mouse" -v`

## What to Verify

**After sidebar changes:** collapse/expand indicators, risk badges, selection cursor, tree connectors, scrollbar, search filtering

**After metadata changes:** runbook name, description, risk badge color, version, ID, border alignment

**After output pane changes:** header background fill, command text visibility, stderr coloring, log path readability, placeholder text, scrollbar

**After footer changes:** keybind key colors, description text, background fill, all states (normal, wizard, running, palette)

**After theme changes:** border visibility, color contrast, dark/light variant selection

**After layout changes:** panel proportions, no dead space, responsive to window size, borders aligned

## Do NOT Skip Visual Testing

Even if unit tests pass, visual bugs (invisible borders, wrong colors, layout overflow, text wrapping) are only caught by looking at the rendered output. Always verify visually after changes to view code.
