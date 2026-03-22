---
name: bubbletea-patterns
description: BubbleTea v2 Elm architecture patterns for Go. Use when creating or editing tea.Model implementations, handling messages, composing components, managing state transitions, or working with commands and subscriptions. Triggers on tea.Model, Update, View, Init, tea.Cmd, tea.Msg usage.
user-invocable: false
---

# BubbleTea v2 Patterns

Import: `charm.land/bubbletea/v2`

## Core Model Interface

```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() tea.View  // v2 returns tea.View, NOT string
}
```

- `Init()` returns the initial command (or `nil` for none)
- `Update()` handles messages, returns updated model + next command
- `View()` returns `tea.View` (use `tea.NewView(s)` to wrap a string)

## tea.View (Declarative — v2 Change)

Terminal features are now declared in the View struct, not via commands or program options:

```go
func (m model) View() tea.View {
    v := tea.NewView(m.render())
    v.AltScreen = true                          // was tea.WithAltScreen() option
    v.MouseMode = tea.MouseModeCellMotion        // was tea.WithMouseCellMotion() option
    v.ReportFocus = true                         // was tea.WithReportFocus() option
    v.WindowTitle = "My App"                     // was tea.SetWindowTitle() cmd
    v.Cursor = tea.NewCursor(m.cursorX, m.cursorY) // was tea.ShowCursor/HideCursor cmd
    return v
}
```

These can change dynamically per render — no need for imperative toggle commands.

## Messages and Commands

```go
type Msg = uv.Event    // still effectively any/interface{}
type Cmd func() Msg    // a function that returns a message
```

Key commands:
- `tea.Quit` — returns `QuitMsg{}`
- `tea.Batch(cmds ...Cmd)` — run commands concurrently
- `tea.Sequence(cmds ...Cmd)` — run commands sequentially (was `tea.Sequentially` in v1)
- `tea.Every(d, fn)` — tick synced with system clock
- `tea.Tick(d, fn)` — independent tick

## Key Events (v2 Change)

```go
// v2 uses KeyPressMsg, not KeyMsg
case tea.KeyPressMsg:
    switch msg.String() {
    case "q", "ctrl+c":
        return m, tea.Quit
    case "space":       // was " " in v1
        m.toggle()
    }
```

Key fields: `msg.Code` (rune), `msg.Text` (string), `msg.Mod` (KeyMod modifiers).

## Mouse Events (v2 Change)

v2 splits mouse into specific types:
- `tea.MouseClickMsg` — button press
- `tea.MouseReleaseMsg` — button release
- `tea.MouseWheelMsg` — scroll
- `tea.MouseMotionMsg` — movement

Each has `.X`, `.Y` fields. MouseClickMsg/ReleaseMsg have `.Button` (e.g., `tea.MouseLeft`).

## Startup Messages

These are sent automatically when the program starts:
- `tea.WindowSizeMsg{Width, Height int}` — initial size + every resize
- `tea.ColorProfileMsg` — terminal color capability
- `tea.EnvMsg` — environment variables

## Component Composition

Embed child models as struct fields. Delegate Update and View:

```go
type model struct {
    list   list.Model
    input  textinput.Model
    state  viewState
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Broadcast messages all children need (e.g., WindowSizeMsg)
    // Route input to focused child only
    switch m.state {
    case listView:
        var cmd tea.Cmd
        m.list, cmd = m.list.Update(msg)
        cmds = append(cmds, cmd)
    case inputView:
        var cmd tea.Cmd
        m.input, cmd = m.input.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}
```

Child components (from `charm.land/bubbles/v2`) return `string` from View(), not `tea.View`. Compose them inside the parent's `tea.NewView()`.

## State Machine Pattern

Use an enum to track which view/state is active:

```go
type viewState int

const (
    listView viewState = iota
    formView
    confirmView
)

type model struct {
    state viewState
    // ... child models
}
```

Route messages and render based on state. This keeps Update clean and predictable.

## Program Setup

```go
p := tea.NewProgram(initialModel,
    tea.WithContext(ctx),       // cancellable
    tea.WithFilter(filterFn),  // suppress/transform messages
)
m, err := p.Run()  // blocking
```

Useful options:
- `tea.WithInput(r)` / `tea.WithOutput(w)` — custom I/O (testing)
- `tea.WithWindowSize(w, h)` — set initial size (testing)
- `tea.WithColorProfile(p)` — force color profile (testing)
- `tea.WithFPS(fps)` — max render FPS (default 60)

External interaction:
- `p.Send(msg)` — inject messages from outside
- `p.Quit()` — graceful shutdown
- `p.Println(args...)` — print above the TUI

## Common Mistakes to Avoid

- Do NOT return a pointer to model — return the value: `return m, cmd`
- Do NOT modify model fields outside of Update — the Elm architecture is message-driven
- Do NOT block in Update — return a `Cmd` for async work
- Do NOT use `tea.WithAltScreen()` option — set `view.AltScreen = true` in View()
- Do NOT use `case " ":` for space — use `case "space":` in v2
- Do NOT use `tea.KeyMsg` for key presses — use `tea.KeyPressMsg` in v2
