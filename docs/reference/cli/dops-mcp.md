---
layout: default
title: dops mcp
nav_order: 7
parent: CLI Commands
grand_parent: Reference
---

# dops mcp

MCP server for AI agent integration.

## Synopsis

```
dops mcp <subcommand> [flags]
```

## Description

Exposes runbooks as tools for AI agents via the [Model Context Protocol](https://modelcontextprotocol.io). Each runbook becomes an MCP tool with a JSON Schema describing its parameters.

## Subcommands

| Command | Description |
|---------|-------------|
| `dops mcp serve` | Start the MCP server |
| `dops mcp tools` | List available MCP tools |

### dops mcp serve

Start the MCP server.

```sh
dops mcp serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--transport` | `stdio` | Transport type: `stdio` or `http` |
| `--port` | `8080` | HTTP port (only for `http` transport) |
| `--allow-risk` | `critical` | Maximum risk level to expose: `low`, `medium`, `high`, `critical` |

### dops mcp tools

List available MCP tools and their schemas.

```sh
dops mcp tools [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--allow-risk` | `critical` | Maximum risk level to show |

## Examples

```sh
# Start stdio server (for Claude Code)
dops mcp serve

# Start HTTP server on port 8080
dops mcp serve --transport http --port 8080

# Only expose low and medium risk runbooks
dops mcp serve --allow-risk medium

# List available tools
dops mcp tools
```

## See also

- [MCP / AI Agents Guide](../../guides/mcp) — setup, Docker, risk controls
