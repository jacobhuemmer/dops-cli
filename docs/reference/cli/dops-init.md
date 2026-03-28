---
title: dops init
---

# dops init

Initialize dops configuration.

## Synopsis

```
dops init
```

## Description

Sets up the `~/.dops` directory with a default configuration. If no catalogs exist, scaffolds a hello-world runbook to get started.

On macOS and Linux, the hello-world runbook uses a POSIX shell script (`script.sh`). On Windows, it uses a PowerShell script (`script.ps1`).

The created directory structure:

```
~/.dops/
├── config.json
├── catalogs/
│   └── default/
│       └── hello-world/
│           ├── runbook.yaml
│           └── script.sh
└── keys/
```

Running `dops init` is safe if `~/.dops` already exists — it will not overwrite existing files.

## Options

None.

## Examples

```sh
# Initialize dops
dops init

# Then launch the TUI
dops
```

## See also

- [dops](dops) — launch the TUI
- [Getting Started](../../guides/getting-started)
