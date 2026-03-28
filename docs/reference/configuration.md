---
title: Configuration
---

# Configuration

dops stores its configuration in `~/.dops/config.json`. Override the base directory with the `DOPS_HOME` environment variable.

---

## Directory Layout

```
~/.dops/
├── config.json         # User-editable settings (theme, catalogs, defaults)
├── vault.json          # Encrypted parameter store (0600, CLI-managed)
├── keys/               # Encryption keys (age)
│   └── keys.txt
├── themes/             # Custom theme overrides
│   └── mytheme.json
└── catalogs/           # Runbook catalogs
    ├── default/
    └── infra/
```

---

## Config Keys

Settings in `config.json` (user-editable):

| Key | Type | Description |
|-----|------|-------------|
| `theme` | string | Active theme name (default: `github`) |
| `defaults.max_risk_level` | string | Maximum risk level to load (`low`, `medium`, `high`, `critical`) |
| `catalogs` | array | List of catalog entries with `name`, `path`, `active` |

Saved parameter values in `vault.json` (encrypted, CLI-managed):

| Key Path | Scope | Description |
|----------|-------|-------------|
| `vars.global.<name>` | global | Shared across all runbooks |
| `vars.catalog.<cat>.<name>` | catalog | Shared within a catalog |
| `vars.catalog.<cat>.runbooks.<rb>.<name>` | runbook | Specific to one runbook |

Use `dops config set/get/unset` to manage both — the CLI routes `vars.*` paths to the vault automatically.

---

## Themes

dops ships with 20 built-in themes. Default: `github`.

```sh
dops config set theme=dracula
```

| Theme | Theme | Theme | Theme |
|-------|-------|-------|-------|
| `github` | `dracula` | `gruvbox` | `nord` |
| `monokai` | `synthwave` | `nightowl` | `one-dark` |
| `kanagawa` | `everforest` | `solarized` | `espresso` |
| `unicorn` | `ayu` | `zenburn` | `catppuccin-mocha` |
| `catppuccin-latte` | `rosepine-dawn` | `doop` | `tokyomidnight` |

Set `theme=rainbow` for a random theme on every launch.

Each theme includes dark and light variants. dops auto-detects your terminal background and selects the appropriate variant. The web UI mirrors your configured theme.

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

## Vault Encryption

All saved parameter values are stored in `vault.json`, encrypted as a single [age](https://age-encryption.org/) blob.

### Algorithm

- **Key exchange**: X25519
- **AEAD**: ChaCha20-Poly1305 (provided by age internally)
- **Identity**: auto-generated at `~/.dops/keys/keys.txt` on first use

### Tamper Detection

The vault uses age's authenticated encryption (AEAD). The Poly1305 authentication tag covers the entire ciphertext — any modification, even a single bit flip, causes decryption to fail. Unlike tools like [sops](https://github.com/getsops/sops) that need a separate HMAC because they encrypt values individually, the dops vault encrypts everything as one blob, so AEAD covers the entire payload inherently.

If `vault.json` is modified outside dops:

```
vault.json is corrupted or was modified outside dops
```

Recovery: delete `vault.json` and re-enter saved values.

### Design

- Values inside the vault are stored as **plaintext** — no per-value encryption
- The entire file is opaque to anyone without the key
- `vault.json`: `0600` (owner read/write only)
- `keys/keys.txt`: `0600` (owner read/write only)
- Atomic writes (temp file + rename) prevent corruption during save

### Migration

Users upgrading from v0.2.0 get automatic one-time migration on first startup. Vars are moved from `config.json` to the encrypted `vault.json`, and the `vars` key is removed from `config.json`. No user action required.

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
