---
title: dops run
---

# dops run

Execute a runbook non-interactively.

## Synopsis

```
dops run <id> [flags]
```

## Description

Runs a runbook by its full ID (e.g., `infra.health-check`) or alias. Parameters can be passed via `--param` flags. Saved values from the vault are applied automatically — only missing required parameters need to be provided.

If `--dry-run` is specified, dops resolves all parameters and prints the command that would be executed without actually running it.

## Arguments

| Argument | Description |
|----------|-------------|
| `<id>` | Runbook ID (e.g., `infra.health-check`) or alias |

## Options

| Flag | Description |
|------|-------------|
| `--param key=value` | Set a parameter value (repeatable) |
| `--dry-run` | Show resolved command without executing |
| `--no-save` | Execute without saving inputs to vault |

## Examples

```sh
# Run a runbook with parameters
dops run infra.health-check --param endpoint=https://api.example.com

# Run using an alias
dops run deploy --param version=v1.2.3

# Multiple parameters
dops run demo.deploy-app \
  --param environment=staging \
  --param version=v1.2.3 \
  --param features=logging,monitoring

# Dry run — show what would execute
dops run infra.restart-pods --param namespace=prod --dry-run
```

## See also

- [dops](dops) — launch the TUI
- [Creating Runbooks](../../guides/runbooks) — runbook YAML schema and scripting guide
