---
title: dops open
---

# dops open

Launch the web UI in a browser.

## Synopsis

```
dops open [flags]
```

## Description

Starts a local web server and opens the dops web UI in your default browser. The web UI provides the same capabilities as the TUI in a browser-based interface:

- Searchable catalog sidebar with risk indicators
- Parameter forms with saved values pre-filled
- Segmented toggles, chip multi-select, and dropdown inputs
- Risk confirmation dialogs for high and critical operations
- Real-time execution log streaming with ANSI color support
- Full theme support — mirrors your configured dops theme

The SPA is embedded in the Go binary — no Node.js required at runtime.

Press `Ctrl+C` to shut down the server.

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `3000` | HTTP server port |
| `--no-browser` | `false` | Start server without opening browser |

## Examples

```sh
# Launch web UI (opens browser automatically)
dops open

# Use a custom port
dops open --port 8080

# Start server without opening browser
dops open --no-browser
```

## See also

- [dops](dops) — launch the TUI
- [Web UI Guide](../../guides/web-ui)
