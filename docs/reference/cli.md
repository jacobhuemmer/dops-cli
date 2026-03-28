---
title: CLI Commands
---

# CLI Commands

dops provides four interfaces: a full-screen **TUI** (default), a **CLI** for scripting, a **Web UI** for browsers, and an **MCP server** for AI agents.

---

## Core commands

| Command | Description |
|---------|-------------|
| [`dops`](cli/dops) | Launch the interactive TUI |
| [`dops run`](cli/dops-run) | Execute a runbook non-interactively |
| [`dops open`](cli/dops-open) | Launch the web UI in a browser |
| [`dops init`](cli/dops-init) | Initialize dops configuration |

## Catalog commands

| Command | Description |
|---------|-------------|
| [`dops catalog list`](cli/dops-catalog) | List configured catalogs |
| [`dops catalog add`](cli/dops-catalog) | Add a local catalog directory |
| [`dops catalog remove`](cli/dops-catalog) | Remove a catalog from config |
| [`dops catalog install`](cli/dops-catalog) | Install a catalog from a git repository |
| [`dops catalog update`](cli/dops-catalog) | Update a git-installed catalog |

## Configuration commands

| Command | Description |
|---------|-------------|
| [`dops config set`](cli/dops-config) | Set a configuration value |
| [`dops config get`](cli/dops-config) | Get a configuration value |
| [`dops config unset`](cli/dops-config) | Remove a saved value |
| [`dops config list`](cli/dops-config) | Display the full configuration |

## AI agent commands

| Command | Description |
|---------|-------------|
| [`dops mcp serve`](cli/dops-mcp) | Start the MCP server |
| [`dops mcp tools`](cli/dops-mcp) | List available MCP tools |

## Other commands

| Command | Description |
|---------|-------------|
| [`dops completion`](cli/dops-completion) | Generate shell completion scripts |
| [`dops version`](cli/dops-version) | Print the version |
