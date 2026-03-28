---
title: dops
---

# dops

Launch the interactive TUI.

## Synopsis

```
dops [flags]
```

## Description

When run without a subcommand, `dops` launches a full-screen terminal UI for browsing, parameterizing, and executing runbooks.

The TUI provides:
- Sidebar with catalog tree, search, and collapse/expand
- Metadata panel showing runbook details
- Field-by-field parameter wizard with validation and persistence
- Live streaming output with scrollback, search, and text selection
- Risk confirmation gates for high and critical operations
- 20 built-in themes

On first run, `dops` creates `~/.dops/` with a default configuration if it doesn't exist. Override the base directory with the `DOPS_HOME` environment variable.

## Options

None.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DOPS_HOME` | `~/.dops` | Config and catalog directory |
| `DOPS_NO_ALT_SCREEN` | (unset) | Set to `1` to disable alternate screen buffer |

## Examples

```sh
# Launch the TUI
dops

# Launch with a custom config directory
DOPS_HOME=/opt/dops dops
```

## See also

- [dops init](dops-init) — initialize configuration
- [dops run](dops-run) — execute a runbook non-interactively
- [dops open](dops-open) — launch the web UI
- [Keyboard Shortcuts](../keyboard-shortcuts)
