---
title: dops catalog
---

# dops catalog

Manage runbook catalogs.

## Synopsis

```
dops catalog <subcommand> [args] [flags]
```

## Description

Catalogs are directories of runbooks. They can be local directories or cloned from git repositories. Use `dops catalog` subcommands to add, remove, install, and update catalogs.

## Subcommands

| Command | Description |
|---------|-------------|
| `dops catalog list` | List configured catalogs |
| `dops catalog add <path>` | Add a local catalog directory |
| `dops catalog remove <name>` | Remove a catalog from config |
| `dops catalog install <url>` | Install a catalog from a git repository |
| `dops catalog update <name>` | Update a git-installed catalog |

### dops catalog list

List all configured catalogs with their paths and status.

```sh
dops catalog list
```

### dops catalog add

Add a local directory as a catalog.

```sh
dops catalog add <path> [flags]
```

| Flag | Description |
|------|-------------|
| `--display-name` | Friendly display name for the sidebar |

### dops catalog remove

Remove a catalog from the configuration. Does not delete files.

```sh
dops catalog remove <name>
```

### dops catalog install

Clone a catalog from a git repository.

```sh
dops catalog install <url> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | repo basename | Catalog name |
| `--ref` | default branch | Git ref to checkout (tag, branch, or commit) |
| `--path` | repo root | Subdirectory within the repo containing runbooks |
| `--risk` | `critical` | Max risk level policy (`low`, `medium`, `high`, `critical`) |
| `--display-name` | — | Friendly display name for the sidebar |

### dops catalog update

Pull latest changes for a git-installed catalog.

```sh
dops catalog update <name> [flags]
```

| Flag | Description |
|------|-------------|
| `--ref` | Git ref to checkout (tag, branch, or commit) |
| `--risk` | Update max risk level policy |
| `--display-name` | Set display name (empty to clear) |

## Examples

```sh
# List catalogs
dops catalog list

# Add a local catalog
dops catalog add ~/runbooks --display-name "My Team"

# Install from git
dops catalog install https://github.com/org/runbooks.git

# Install a specific tag with a risk policy
dops catalog install https://github.com/org/runbooks.git \
  --ref v2.0 --risk medium --name prod-runbooks

# Install from a subdirectory
dops catalog install https://github.com/org/monorepo.git \
  --path ops/runbooks --name ops

# Update a git catalog
dops catalog update prod-runbooks --ref v2.1

# Remove a catalog
dops catalog remove old-runbooks
```

## See also

- [Creating Runbooks](../../guides/runbooks) — runbook YAML schema
- [Getting Started](../../guides/getting-started) — first catalog setup
