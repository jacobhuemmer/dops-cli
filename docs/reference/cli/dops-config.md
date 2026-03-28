---
layout: default
title: dops config
nav_order: 5
parent: CLI Commands
grand_parent: Reference
---

# dops config

Read and write dops configuration.

## Synopsis

```
dops config <subcommand> [args]
```

## Description

Manage settings stored in `~/.dops/config.json` and parameter values in the encrypted vault. The CLI automatically routes `vars.*` key paths to the vault.

## Subcommands

| Command | Description |
|---------|-------------|
| `dops config set key=value` | Set a configuration value |
| `dops config get key` | Get a configuration value |
| `dops config unset key` | Remove a saved value |
| `dops config list` | Display the full configuration (secrets masked) |

### dops config set

Set a configuration value or save a parameter to the vault.

```sh
dops config set key=value
```

Configuration keys are written to `config.json`. Variable paths (`vars.*`) are written to the encrypted vault.

### dops config get

Read a configuration value.

```sh
dops config get key
```

### dops config unset

Remove a configuration value or saved parameter.

```sh
dops config unset key
```

### dops config list

Display the full configuration with secrets masked.

```sh
dops config list
```

## Examples

```sh
# Set theme
dops config set theme=dracula

# Save a global parameter (stored in encrypted vault)
dops config set vars.global.region=us-east-1

# Save a catalog-scoped parameter
dops config set vars.catalog.infra.cluster=prod-us

# Save a runbook-scoped parameter
dops config set vars.catalog.infra.runbooks.health-check.endpoint=https://api.example.com

# Read a value
dops config get theme

# List all config
dops config list

# Remove a saved value
dops config unset vars.global.region
```

## See also

- [Configuration Reference](../configuration) — full config schema, themes, vault details
