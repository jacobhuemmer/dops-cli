<p align="center">
  <img src="assets/logo.png" alt="dops logo" width="500" />
</p>

# dops — the do(ops) cli

`dops` provides a browsable catalog of automation scripts that operators can select, parameterize, and execute directly from the terminal. Built for DevOps and platform engineering workflows.

<p align="center">
  <img src="assets/demo.gif" alt="dops demo" width="900" />
</p>

## Features

### Interactive TUI

- **Sidebar** — collapsible catalog tree with fuzzy search
- **Metadata panel** — runbook details, risk level, click-to-copy path
- **Output pane** — live streaming output with scroll, search, and text selection
- **Wizard** — field-by-field parameter input with per-field save control
- **Help overlay** — context-aware keybinding display (`?` key)

### Execution

- **Live streaming** — stdout/stderr streamed in real-time
- **Log persistence** — execution output saved to timestamped log files
- **Process control** — `ctrl+x` to stop running execution
- **Risk gates** — confirmation required for high/critical risk runbooks
- **Dry-run mode** — preview resolved command without executing

### MCP Server

AI agents can discover and execute runbooks via the [Model Context Protocol](https://modelcontextprotocol.io):

- **Tools** — each runbook exposed as an MCP tool with JSON Schema
- **Resources** — catalog listing and runbook details
- **Transports** — stdio (for Claude Code) and HTTP with gzip
- **Security** — sensitive params excluded from schema, loaded from local config
- **Progress** — real-time output streaming via MCP notifications

### Configuration & Security

- **Local config** — user-editable settings in `~/.dops/config.json`
- **Encrypted vault** — saved parameter values in `~/.dops/vault.json`, encrypted with [age](https://github.com/FiloSottile/age) (X25519 + ChaCha20-Poly1305)
- **Tamper detection** — age AEAD authenticates the entire vault; any modification causes decryption failure
- **File permissions** — `vault.json` and `keys.txt` locked to `0600`

### CLI

- `dops` — launch the TUI
- `dops run <id>` — execute a runbook by ID
- `dops config set/get/unset/list` — manage configuration
- `dops catalog list/add/remove/install/update` — manage catalogs
- `dops mcp serve` — start MCP server
- `dops mcp tools` — list available MCP tools

## Install

### Homebrew

```bash
brew tap jacobhuemmer/tap
brew install dops
```

### Go

```bash
go install github.com/jacobhuemmer/dops-cli@latest
```

### Docker (MCP server)

```bash
# Mount your local catalogs and config into the container
docker run -i -v ~/.dops:/data/dops ghcr.io/jacobhuemmer/dops-cli:latest
```

### From source

```bash
git clone https://github.com/jacobhuemmer/dops-cli.git
cd dops-cli
make build
./bin/dops
```

## Quick Start

1. **Create a catalog** with runbook scripts:

```
~/.dops/catalogs/default/
├── hello-world/
│   ├── runbook.yaml
│   └── script.sh
└── check-health/
    ├── runbook.yaml
    └── script.sh
```

2. **Define a runbook** (`runbook.yaml`):

```yaml
name: check-health
version: 1.0.0
description: Runs health checks against a service endpoint
risk_level: medium
script: script.sh
parameters:
  - name: endpoint
    type: string
    required: true
    description: The endpoint to check
    scope: global             # local | global | catalog | runbook
```

3. **Launch dops**:

```bash
dops
```

4. **Navigate** with arrow keys, **run** with Enter, **scroll** output with Up/Down, **search** with `/`.

## Parameter Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Free text input | endpoints, names, paths |
| `boolean` | Yes/No toggle | dry_run, verbose |
| `integer` | Whole number (negative ok) | offsets, deltas |
| `number` | Non-negative whole number (0+) | ports, replicas, days, timeout |
| `float` | Decimal number | percentages, thresholds |
| `select` | Single selection from options | environment, region |
| `multi_select` | Multiple selections from options | features, policies |
| `file_path` | File path input | config files |
| `resource_id` | Resource identifier | ARNs, URIs |

## Vault

dops stores all saved parameter values in an encrypted vault (`~/.dops/vault.json`) rather than in plaintext config.

```
~/.dops/
├── config.json    # User-editable (theme, catalogs, defaults)
├── vault.json     # Encrypted parameter store (0600)
└── keys/
    └── keys.txt   # age X25519 identity (0600, auto-generated)
```

### How It Works

The vault encrypts its entire contents as a single [age](https://age-encryption.org/) blob. When saved, all parameter values are serialized to JSON, encrypted, and wrapped in a versioned envelope:

```json
{
  "version": 1,
  "data": "age1..."
}
```

Values inside the vault are stored as plaintext — there is no per-value encryption. The vault provides encryption at the file level.

### Tamper Detection

The vault uses age's authenticated encryption (ChaCha20-Poly1305 AEAD). The Poly1305 authentication tag covers the entire ciphertext, so any modification — even a single bit flip — causes decryption to fail. This is similar to how [sops](https://github.com/getsops/sops) uses a MAC to detect tampering, but with a key difference:

| | sops | dops vault |
|---|------|------------|
| **Encryption** | Per-value (keys visible, values encrypted) | Entire file (fully opaque) |
| **Tamper detection** | Explicit HMAC over all encrypted values | Inherent from AEAD authentication tag |
| **Diffability** | Human-readable diffs (keys in plaintext) | Opaque blob — no meaningful diffs |
| **Selective edits** | Edit individual values in place | Must decrypt, modify, re-encrypt entire file |

sops needs a separate MAC because it encrypts values individually — without it, someone could swap or modify encrypted values undetected. The dops vault encrypts everything as one blob, so AEAD covers the entire payload inherently. There is nothing to tamper with selectively.

If `vault.json` is modified outside dops, the CLI prints a clear error:

```
vault.json is corrupted or was modified outside dops
```

Recovery: delete `vault.json` and re-enter saved values.

### Key Management

- Identity auto-generated at `~/.dops/keys/keys.txt` on first use
- Single key per dops installation
- If `keys.txt` is lost, the vault cannot be decrypted — delete and re-enter values

### Migration

Users upgrading from v0.2.0 get automatic one-time migration: vars are moved from `config.json` to `vault.json` on first startup. No user action required.

## MCP Integration

### Claude Code

Add to `.claude/settings.json`:

```json
{
  "mcpServers": {
    "dops": {
      "command": "dops",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Docker

```bash
# stdio transport — mount your catalogs/config
docker run -i -v ~/.dops:/data/dops ghcr.io/jacobhuemmer/dops-cli:latest

# HTTP transport with gzip
docker run -p 8080:8080 -v ~/.dops:/data/dops ghcr.io/jacobhuemmer/dops-cli:latest --transport http --port 8080
```

> **Note:** The container uses `DOPS_HOME=/data/dops`. Mount your local `~/.dops` directory to `/data/dops` to provide catalogs, config, and themes. You can also set `DOPS_HOME` to any path when running dops outside Docker.

## Keyboard Shortcuts

### Sidebar
| Key | Action |
|-----|--------|
| `↑↓` | Navigate runbooks |
| `←→` | Collapse/expand catalog |
| `Enter` | Run selected runbook |
| `/` | Search runbooks |

### Output
| Key | Action |
|-----|--------|
| `↑↓ j/k` | Scroll one line |
| `PgUp/PgDn` | Scroll one page |
| `h/l` | Scroll left/right |
| `/` | Search output |
| `n/N` | Next/prev match |
| `y` | Copy selection |

### Global
| Key | Action |
|-----|--------|
| `Tab` | Switch pane focus |
| `?` | Help overlay |
| `ctrl+x` | Stop execution |
| `ctrl+shift+p` | Command palette |
| `q` | Quit |

## Themes

dops ships with 6 bundled themes. Default: `tokyomidnight`.

| Theme | Style |
|-------|-------|
| `tokyonight` | Dark — cool blue accents |
| `tokyomidnight` | Dark — deeper background (default) |
| `catppuccin-mocha` | Dark — warm pastels |
| `catppuccin-latte` | Light — warm pastels |
| `nord` | Dark — muted blue-gray |
| `rosepine-dawn` | Light — soft lavender |

Each theme includes dark and light variants. dops auto-detects your terminal background and selects the appropriate variant.

```sh
dops config set theme=catppuccin-mocha
```

Custom themes go in `~/.dops/themes/<name>.json`. See the [configuration reference](https://jacobhuemmer.github.io/dops-cli/reference/configuration) for the full schema.

## Shell Completion

```bash
# Bash
dops completion bash > /etc/bash_completion.d/dops

# Zsh
dops completion zsh > "${fpath[1]}/_dops"

# Fish
dops completion fish > ~/.config/fish/completions/dops.fish

# PowerShell
dops completion powershell | Out-String | Invoke-Expression
```

## Development

```bash
make build       # Build binary
make test        # Run tests
make vet         # Go vet
make lint        # golangci-lint
make screenshots # Generate VHS screenshots
make docker      # Build Docker image
make ci          # Run CI checks (vet + test + build)
```

## Support

If you find do(ops) cli useful, consider [buying me a coffee](https://buymeacoffee.com/jacobhuemmer)!

<p align="center">
  <a href="https://buymeacoffee.com/jacobhuemmer">
    <img src="assets/buymeacoffee.png" alt="Buy Me a Coffee" width="200" />
  </a>
</p>

## License

MIT
