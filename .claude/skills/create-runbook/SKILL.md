---
name: create-runbook
description: "Create a new dops runbook with correct YAML schema and POSIX shell script. Use when the user asks to create a new automation, runbook, or script for dops. Triggers on: 'create a runbook', 'add automation', 'new script', 'scaffold runbook'."
user-invocable: true
---

# Create a dops Runbook

## Directory Structure

Runbooks live under `DOPS_HOME/catalogs/<catalog>/<runbook-name>/`:

```
~/.dops/catalogs/<catalog>/<runbook-name>/
├── runbook.yaml    # Runbook definition
└── script.sh       # Automation script (POSIX sh)
```

## runbook.yaml Schema

```yaml
name: <runbook-name>          # Must match directory name
version: 1.0.0
description: Short description of what this runbook does
risk_level: low               # low | medium | high | critical
script: script.sh
parameters:
  - name: endpoint
    type: string              # See parameter types below
    required: true
    description: What this parameter does
    scope: global             # global | catalog | runbook
    default: ""               # Optional default value
    secret: false             # If true, masked in UI, excluded from MCP
    options: []               # Required for select and multi_select
```

## Parameter Types

| Type | Description | Validation |
|------|-------------|------------|
| `string` | Free text | — |
| `boolean` | Yes/No toggle | — |
| `integer` | Whole number (negative ok) | `strconv.Atoi` |
| `number` | Non-negative whole number (0+) | Must be >= 0 |
| `float` | Decimal number | `strconv.ParseFloat` |
| `select` | Single choice from options | Requires `options` list |
| `multi_select` | Multiple choices from options | Requires `options` list |
| `file_path` | File system path | — |
| `resource_id` | Resource identifier (ARN, URI) | — |

## Risk Levels

| Level | TUI Confirmation | MCP Confirmation |
|-------|-----------------|------------------|
| `low` | None | None |
| `medium` | None | None |
| `high` | y/N prompt | `_confirm_id` param must match runbook ID |
| `critical` | Type runbook ID | `_confirm_word` param must be "CONFIRM" |

## Scopes

| Scope | Where saved | Use when |
|-------|------------|----------|
| `global` | `vars.global.<name>` | Shared across all runbooks (API tokens, regions) |
| `catalog` | `vars.catalog.<cat>.<name>` | Shared within a catalog |
| `runbook` | `vars.catalog.<cat>.runbooks.<rb>.<name>` | Specific to this runbook |

## Script Template (POSIX sh)

```sh
#!/bin/sh
set -eu

# dops passes parameters as UPPERCASE environment variables.
# Parameter "endpoint" → $ENDPOINT
# Parameter "dry_run" → $DRY_RUN
ENDPOINT="${ENDPOINT:?endpoint is required}"
DRY_RUN="${DRY_RUN:-false}"

main() {
  echo "==> Stage 1/2: Validate"
  echo "    Checking ${ENDPOINT}..."
  # TODO: implement

  echo ""
  echo "==> Stage 2/2: Execute"
  echo "    Running operation..."
  # TODO: implement

  echo ""
  echo "✓ Done"
}

main "$@"
```

## Shell Style Rules (POSIX-compatible)

1. **Use `#!/bin/sh`** — not `#!/bin/bash` (POSIX compatibility for Linux/macOS)
2. **Use `set -eu`** — not `set -euo pipefail` (pipefail is not POSIX)
3. **Quote all variables**: `"${var}"` not `$var`
4. **Use `[ ]` not `[[ ]]`** — POSIX test
5. **Use `$(command)`** not backticks
6. **Stderr for errors**: `echo "error" >&2`
7. **Indent with 2 spaces**, no tabs
8. **Put `main()` at the bottom** of the script
9. **Use `command -v`** not `which`
10. **Use `printf`** over `echo -e` for portability

## Output Conventions

```sh
# Stage headers
echo "==> Stage 1/3: Build"

# Indented details
echo "    Compiling source..."

# Success
echo "✓ Build complete"

# Failure
echo "✗ Build failed" >&2

# Summary block
echo "========================================="
echo "  Summary"
echo "========================================="
echo "  Status: SUCCESS"
echo "========================================="
```

## Workflow

1. Determine the catalog name and runbook name
2. Create the directory: `mkdir -p ~/.dops/catalogs/<catalog>/<name>`
3. Write `runbook.yaml` with the schema above
4. Write `script.sh` following the POSIX template
5. `chmod +x script.sh`
6. If the catalog is new, register it: `dops catalog add ~/.dops/catalogs/<catalog>`
7. Test: launch `dops`, navigate to the runbook, execute it
