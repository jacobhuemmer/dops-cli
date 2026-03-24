# Output Pane Refinement Plan

## Date: 2026-03-23

## Context

The output pane needed several rendering/behavior improvements and missing features. This plan documents all changes made during the session.

## Changes Implemented

### Phase 1 — ANSI Handling & Core Output Fixes

1. **Replaced regex ANSI handling with `charmbracelet/x/ansi`** — `ansi.Strip()`, `ansi.StringWidth()`, `ansi.Cut()`, `ansi.Truncate()`, `ansi.Wrap()` for correct width calculation and truncation on ANSI output.

2. **Fixed scrollbar** — Replaced broken per-line `scrollbarChar()` with proportional thumb positioning: `thumbHeight = (bodyH²)/total`, position derived from actual scroll offset. Thumb touches bottom at max offset.

3. **Added smart auto-scroll (`atBottom` flag)** — Set `false` on scroll up, re-enable when user scrolls to bottom. New output only auto-scrolls when `atBottom` is true.

4. **Strip `\r` carriage returns** — Prevents CI progress bar output from breaking rendering.

5. **Added horizontal scrolling** — `h`/`l` keys scroll 8 columns. `xOffset` field with `ansi.Cut()` for ANSI-aware slicing. Clamped to `[0, maxLineWidth - textWidth]`.

6. **Added command line wrapping** — Header uses `ansi.Wrap()` for long commands (later simplified to truncation when section borders were removed).

7. **Added vim keys** — `j`/`k` for vertical scroll alongside arrow keys.

### Phase 2 — Log File Persistence

8. **Buffered log writer** — Added 64KB `bufio.Writer` to `adapters/log.go` for performant log file writes. Added `os.MkdirAll` for auto-creating log directory and path traversal protection.

### Phase 3 — Live Output Streaming

9. **Live streaming via `tea.Program.Send()`** — Added `ProgramRef` shared pointer type to `AppDeps`. Execution goroutine sends `OutputLineMsg` per-line as they arrive instead of batch collection. Falls back to collect-and-return for tests.

10. **Switched executor to `io.Pipe()`** — Replaced OS-level `cmd.StdoutPipe()/StderrPipe()` with `io.Pipe()` for immediate line delivery without kernel buffering.

### Phase 4 — Three-Section Layout

11. **Rewrote View() as three stacked sections** — Header (command), log (scrollable content with `backgroundElement` fill), footer (log path). Initially with transparent section borders, later simplified to flat content inside the app's outer border.

12. **Removed section borders** — Simplified to flat content (no borders) inside the app's persistent outer border. Eliminates chrome height overhead that caused overflow.

13. **Fixed line background fill** — Each log line rendered with `logContentStyle.Width(logW).Render(lineText)` so `bgElemColor` extends edge-to-edge.

### Phase 5 — Scroll Confinement

14. **Integrated bubbles `viewport.Model`** — Used purely for scroll state tracking (`YOffset`, `GotoBottom`, `TotalLineCount`). Handles Up/Down/j/k/PgUp/PgDown/Home/End/mouse wheel natively. Custom `View()` reads `YOffset()` to render only visible lines.

15. **Fixed viewport height desync** — `viewNormal()` (value receiver) was calling `SetSize()` on a temporary copy. Added `resizeAll()` called from `Update()` for persistent dimensions. `viewNormal()` uses `ViewWithSize()` with authoritative dimensions at render time.

16. **Fixed line width overflow** — Output was rendering lines at `contentW - 1` width, but the app's border content area was `contentW - 2`. Caused line wrapping that exceeded height. Fixed by passing `contentW - borderSize` as the output's width.

### Phase 6 — Focus Management

17. **Added focus cycling (Tab)** — `focusSidebar` / `focusOutput` enum. Tab switches between sidebar and output pane. Key events routed to focused component.

18. **Added visual focus indicator** — Active pane border uses `borderActive` color, inactive uses `border` color.

19. **Hover-to-focus on output pane** — Mouse hover over output area switches focus for immediate scrolling. Sidebar requires click to steal focus.

### Phase 7 — Layout & Spacing

20. **Preserved output on sidebar navigation** — Removed `m.output.Clear()` from `RunbookSelectedMsg`. Last execution stays visible until a new one starts.

21. **Fixed panel height budget** — `panelRows` was subtracting `layoutMarginLeft` as a phantom bottom margin, wasting 3 rows.

22. **Added 4-row bottom margin** — `layoutMarginBottom = 4` pushes panels up from terminal edge.

23. **Fixed `resizeAll()` missing `layoutMarginBottom`** — Caused output overflow below sidebar.

24. **Output confined within persistent outer border** — App always renders the outer `RoundedBorder`. Output `View()` returns flat content that fits inside it.

25. **Added 1-row gaps** — Between header/log and log/footer sections for visual separation.

26. **Added 1-char left/right padding** — Across all sections for breathing room from border edges.

27. **Added 1-row top padding inside log pane** — Pushes log content down from the top of the bgElemColor area.

### Phase 8 — Theme

28. **Set terminal background from theme** — `View.BackgroundColor` set to the theme's `background` color so the entire terminal uses the dops theme background.

29. **Default theme changed to `tokyomidnight`** — New bundled theme with fallback support.

## Files Modified

| File | Changes |
|---|---|
| `internal/tui/output/model.go` | Complete rewrite: viewport integration, flat 3-section layout, ANSI handling, scrollbar, focus |
| `internal/tui/output/model_test.go` | New tests for all features |
| `internal/tui/output/messages.go` | Added `CopiedHeaderFlashMsg`, `CopiedFooterFlashMsg` |
| `internal/tui/output/search_test.go` | Updated for new model |
| `internal/tui/app.go` | Focus management, `resizeAll()`, `ProgramRef`, hover-to-focus, layout fixes |
| `internal/tui/app_test.go` | Updated for new output behavior |
| `internal/executor/script.go` | Switched to `io.Pipe()` for immediate line delivery |
| `internal/adapters/log.go` | Buffered writer, path traversal protection |
| `cmd/root.go` | Wired `ProgramRef` for live streaming |
| `internal/theme/loader.go` | `tokyomidnight` bundled theme support |
| `internal/config/store.go` | Default theme → `tokyomidnight` |
