---
layout: default
title: Configuration
nav_order: 3
parent: Reference
---

# Configuration

dops stores its configuration in `~/.dops/config.json`. Override the base directory with the `DOPS_HOME` environment variable.

---

## Directory Layout

```
~/.dops/
├── config.json         # Main configuration
├── keys/               # Encryption keys (age)
│   └── dops.key
├── themes/             # Custom theme overrides
│   └── mytheme.json
└── catalogs/           # Runbook catalogs
    ├── default/
    └── infra/
```

---

## Config Keys

| Key | Type | Description |
|-----|------|-------------|
| `theme` | string | Active theme name (default: `tokyomidnight`) |
| `defaults.max_risk_level` | string | Maximum risk level to load (`low`, `medium`, `high`, `critical`) |
| `catalogs` | array | List of catalog entries with `name`, `path`, `active` |
| `vars.global.<name>` | string | Global saved parameter values |
| `vars.catalog.<cat>.<name>` | string | Catalog-scoped saved values |
| `vars.catalog.<cat>.runbooks.<rb>.<name>` | string | Runbook-scoped saved values |

---

## Themes

dops ships with 6 bundled themes:

| Theme | Style |
|-------|-------|
| `tokyonight` | Dark — cool blue accents |
| `tokyomidnight` | Dark — deeper background (default) |
| `catppuccin-mocha` | Dark — warm pastels |
| `catppuccin-latte` | Light — warm pastels |
| `nord` | Dark — muted blue-gray |
| `rosepine-dawn` | Light — soft lavender |

Each theme includes dark and light variants. dops auto-detects your terminal background and selects the appropriate variant.

Switch themes:

```sh
dops config set theme=catppuccin-mocha
```

### Custom Themes

Create a JSON file in `~/.dops/themes/`:

```json
{
  "name": "my-theme",
  "defs": {
    "bg": "#1a1b26",
    "fg": "#c0caf5",
    "blue": "#7aa2f7",
    "green": "#9ece6a",
    "orange": "#ff9e64",
    "red": "#f7768e"
  },
  "theme": {
    "background":        { "dark": "bg",    "light": "#e1e2e7" },
    "backgroundPanel":   { "dark": "#1f2335", "light": "#d5d6db" },
    "backgroundElement": { "dark": "#292e42", "light": "#c4c8da" },
    "text":              { "dark": "fg",    "light": "#3760bf" },
    "textMuted":         { "dark": "#565f89", "light": "#848cb5" },
    "primary":           { "dark": "blue",  "light": "#2e7de9" },
    "border":            { "dark": "#565f89", "light": "#848cb5" },
    "borderActive":      { "dark": "blue",  "light": "#2e7de9" },
    "success":           { "dark": "green", "light": "#587539" },
    "warning":           { "dark": "orange","light": "#b15c00" },
    "error":             { "dark": "red",   "light": "#f52a65" },
    "risk": {
      "low":      { "dark": "green",   "light": "#587539" },
      "medium":   { "dark": "orange",  "light": "#b15c00" },
      "high":     { "dark": "orange",  "light": "#b15c00" },
      "critical": { "dark": "#db4b4b", "light": "#f52a65" }
    }
  }
}
```

Activate it:

```sh
dops config set theme=my-theme
```

---

## Encryption

Secret parameters are encrypted at rest using [age](https://age-encryption.org/) (X25519). The key is auto-generated at `~/.dops/keys/dops.key` on first use.

Encrypted values in `config.json` are prefixed with `age1-`. They are automatically decrypted when passed to scripts and masked with `****` in all display contexts (TUI, MCP, `config list`).

---

## Shell Completion

```sh
# Bash
dops completion bash > /etc/bash_completion.d/dops

# Zsh
dops completion zsh > "${fpath[1]}/_dops"

# Fish
dops completion fish > ~/.config/fish/completions/dops.fish

# PowerShell
dops completion powershell > dops.ps1
```
