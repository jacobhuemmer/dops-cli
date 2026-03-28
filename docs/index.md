---
layout: default
title: Home
nav_order: 1
---

<p align="center">
  <img src="https://raw.githubusercontent.com/jacobhuemmer/dops-cli/main/assets/logo.png" alt="dops logo" width="400" />
</p>

# do(ops) cli
{: .fs-9 }

a runbook toolkit for operators and AI agents.
{: .fs-6 .fw-300 }

[Get Started](/dops-cli/guides/getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/jacobhuemmer/dops-cli){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## Terminal UI

<img src="https://raw.githubusercontent.com/jacobhuemmer/dops-cli/main/assets/demo.gif" alt="dops TUI demo" width="900" />

## Web UI

Also available in the browser with `dops open`:

<img src="https://raw.githubusercontent.com/jacobhuemmer/dops-cli/main/assets/web-demo.gif" alt="dops web UI demo" width="900" />

---

## What is dops?

**dops** is an open-source toolkit for browsing, executing, and managing operational runbooks. It works three ways:

- **TUI** — a full-screen terminal interface with sidebar navigation, parameter wizards, and live streaming output
- **Web UI** — a browser-based interface via `dops open` with the same capabilities
- **MCP server** — expose runbooks as tools for Claude and other AI agents

Runbooks are simple YAML + shell scripts organized in catalogs. No proprietary DSL, no cloud dependency.

---

## Features

### Interactive TUI
- Sidebar with catalog tree, search, collapse/expand
- Metadata panel with runbook details
- Output pane with live streaming, scrollback, text selection
- Field-by-field wizard with parameter validation and persistence
- Risk confirmation gates (high = y/N, critical = type runbook ID)
- 20 built-in themes

### Web UI
- Searchable catalog sidebar with risk indicators
- Parameter forms with dropdowns, toggles, chip multi-select
- Saved values pre-filled with collapsible review section
- Risk confirmation dialogs for high and critical operations
- Real-time execution log streaming with ANSI color support
- Full theme support — mirrors your configured dops theme

### CLI
- `dops run <id>` — execute runbooks non-interactively with `--param` flags
- `dops catalog install <url>` — install shared catalogs from git repos
- `dops config set/get/list` — manage configuration and saved parameters
- `dops open` — launch the web UI
- Shell completion for bash, zsh, fish, powershell

### MCP Server
- Expose runbooks as tools for AI agents via Model Context Protocol
- Stdio and HTTP transports with gzip
- Sensitive parameters excluded from tool schemas
- Schema and style guide resources for runbook creation

### Catalog System
- Organize runbooks locally or install shared catalogs from git
- Runbook aliases for short names (`dops run deploy`)
- Per-catalog risk policies
- Encrypted vault for saved parameters (age: X25519 + ChaCha20-Poly1305)

---

## Quick Install

```sh
# Homebrew
brew install jacobhuemmer/tap/dops

# Go
go install github.com/jacobhuemmer/dops-cli@latest

# From source
git clone https://github.com/jacobhuemmer/dops-cli.git
cd dops-cli && make install
```

---

## Commands

| Command | Description |
|---------|-------------|
| [`dops`](/dops-cli/reference/cli/dops) | Launch the interactive TUI |
| [`dops run`](/dops-cli/reference/cli/dops-run) | Execute a runbook non-interactively |
| [`dops open`](/dops-cli/reference/cli/dops-open) | Launch the web UI in a browser |
| [`dops init`](/dops-cli/reference/cli/dops-init) | Initialize configuration |
| [`dops catalog`](/dops-cli/reference/cli/dops-catalog) | Manage runbook catalogs |
| [`dops config`](/dops-cli/reference/cli/dops-config) | Read and write configuration |
| [`dops mcp`](/dops-cli/reference/cli/dops-mcp) | MCP server for AI agents |
| [`dops completion`](/dops-cli/reference/cli/dops-completion) | Generate shell completions |
| [`dops version`](/dops-cli/reference/cli/dops-version) | Print the version |

[Full CLI Reference](/dops-cli/reference/cli){: .btn .btn-outline .fs-4 }

---

## Support

If you find dops useful, consider [buying me a coffee](https://buymeacoffee.com/jacobhuemmer).
