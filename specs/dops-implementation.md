# dops — Developer Operations TUI
### Implementation Spec · Handoff document for AI-assisted implementation

---

## 1. Overview

`dops` is a terminal user interface (TUI) built in Go using the Bubble Tea framework. It provides a browsable catalog of automation scripts — called **runbooks** — that operators can select, parameterize, and execute directly from the terminal.

The tool is designed for DevOps and platform engineering workflows where teams maintain collections of operational scripts that need to be discoverable and safely invocable without editing raw shell scripts by hand.

---

## 2. Diagrams

The following diagrams are included alongside this spec to aid the design and implementation process. Additional diagrams will be added as the project evolves.

| File | Description |
|---|---|
| `tui-layout.png` | Wireframe of the main TUI view showing the sidebar, metadata panel, output pane, and footer regions |
| `tui-wizard.png` | Wireframe of the wizard overlay showing the parameter input form and its relationship to the main view |
| `tui-form-output-example.png` | Wireframe showing the shared header + body layout pattern used by both the wizard form and the output pane |

---

## 3. Directory Structure

All dops data lives under `~/.dops/`. The root contains a single `config.json` and a `catalogs/` subdirectory. A `themes/` directory holds user-defined custom themes.

```
~/.dops/
├── config.json
├── themes/
│   ├── dracula.json
│   └── solarized.json
├── keys/
│   └── keys.txt
└── catalogs/
    ├── default/
    │   └── hello-world/
    │       ├── runbook.yaml
    │       └── script.sh
    ├── local/
    │   └── hello-world/
    │       ├── runbook.yaml
    │       └── script.sh
    └── public-catalog.git/
        └── hello-world/
            ├── runbook.yaml
            └── script.sh
```

Each catalog is a subdirectory under `catalogs/`. Each runbook is a subdirectory inside a catalog, named after the runbook. Every runbook directory contains exactly two files: `runbook.yaml` (the manifest) and `script.sh` (the entrypoint).

---

## 4. Configuration — `config.json`

`config.json` is the single source of truth for tool behavior, catalog registry, theme, and all saved inputs. It lives at `~/.dops/config.json`.

### 4.1 Schema

```json
{
  "theme": "tokyonight",
  "defaults": {
    "max_risk_level": "medium"
  },
  "catalogs": [
    {
      "name": "default",
      "path": "~/.dops/catalogs/default",
      "active": true,
      "policy": {
        "max_risk_level": "medium"
      }
    },
    {
      "name": "local",
      "path": "~/.dops/catalogs/local",
      "active": true,
      "policy": {
        "max_risk_level": "critical"
      }
    },
    {
      "name": "public-catalog",
      "path": "~/.dops/catalogs/public-catalog.git",
      "active": false,
      "policy": {
        "max_risk_level": "low"
      }
    }
  ],
  "vars": {
    "global": {
      "region": "us-east-1",
      "environment": "production"
    },
    "catalog": {
      "default": {
        "namespace": "platform",
        "cert_manager_token": "age1qyqszqgpqyqszqgpqyqszqgp...",
        "runbooks": {
          "rotate-tls-certificates": {
            "dry_run": false
          }
        }
      }
    }
  }
}
```

### 4.2 Field Reference

| Field | Description |
|---|---|
| `theme` | Name of the active theme. dops checks `~/.dops/themes/` first, then falls back to bundled themes |
| `defaults.max_risk_level` | Fallback policy applied to catalogs that do not specify their own |
| `catalogs[].name` | Display name of the catalog, used as the group header in the TUI sidebar |
| `catalogs[].path` | Absolute path to the catalog directory on disk |
| `catalogs[].active` | Boolean. Only active catalogs are loaded and displayed at startup |
| `catalogs[].policy.max_risk_level` | Ceiling on which runbooks are surfaced. Runbooks exceeding this level are filtered at load time |
| `vars.global.<key>` | Key/value pairs available to every runbook across all catalogs |
| `vars.catalog.<name>.<key>` | Key/value pairs shared across all runbooks in the named catalog |
| `vars.catalog.<name>.runbooks.<name>.<key>` | Key/value pairs specific to a single runbook |

### 4.3 Risk Level Order

Risk levels form an ordered scale used for policy enforcement:

```
low  <  medium  <  high  <  critical
```

A catalog with `max_risk_level: medium` will surface runbooks marked `low` or `medium`, and silently exclude `high` and `critical` runbooks. The runbooks are not deleted — they exist on disk but are not loaded into the TUI.

---

## 5. Runbook Manifest — `runbook.yaml`

`runbook.yaml` is the contract between the catalog and the script. It describes what the runbook does, what parameters it accepts, and which script to invoke. It does not describe execution infrastructure — that is owned by `config.json`.

### 5.1 Schema

```yaml
id: "default.hello-world"
name: "hello-world"
description: "Prints a hello world message to stdout"
version: "1.0.0"
risk_level: "low"
script: "./script.sh"

parameters:
  - name: "namespace"
    type: "string"
    required: true
    scope: "catalog"
    secret: false
    description: "Target Kubernetes namespace"
  - name: "cert_manager_token"
    type: "string"
    required: true
    scope: "catalog"
    secret: true
    description: "Cert manager API token"
  - name: "dry_run"
    type: "boolean"
    required: false
    scope: "runbook"
    secret: false
    default: false
    description: "Preview changes without applying"
```

### 5.2 Field Reference

| Field | Description |
|---|---|
| `id` | Globally unique identifier in `<catalog>.<runbook>` format. Used as the CLI invocation key: `dops run <id>`. Must be unique across all catalogs. |
| `name` | Human-friendly display name. Should match the parent directory name |
| `description` | Human-readable description, displayed in the TUI metadata panel |
| `version` | Semver string for the runbook |
| `risk_level` | One of: `low`, `medium`, `high`, `critical`. Compared against catalog policy at load time |
| `script` | Relative path to the entrypoint script, relative to the runbook directory |
| `parameters` | List of input definitions. Each parameter is collected by the wizard before execution |
| `parameters[].scope` | Where to save this input in `config.json`. One of: `global`, `catalog`, `runbook` |
| `parameters[].secret` | Boolean. If `true`, the value is encrypted with age before being written to `config.json` |

### 5.3 Parameter Types

| Type | Behavior |
|---|---|
| `string` | Free text input |
| `boolean` | True/false toggle |
| `integer` | Numeric input |
| `select` | List selection (requires an `options` field listing valid values) |

---

## 6. TUI Layout

The TUI has two views: the **main view** and the **wizard overlay**. Both are built with Bubble Tea. Mouse support is enabled globally — all interactive elements across both views support mouse click interaction. This is enabled at the Bubble Tea program level via `tea.WithMouseCellMotion()`.

### 6.1 Main View

The main view is always visible and has four regions:

| Region | Description |
|---|---|
| **Sidebar** | Displays all active catalogs and their runbooks in an always-expanded folder tree, similar to a directory listing. Catalogs are the top-level nodes and runbooks are their leaves. At startup, the first runbook in the first catalog is automatically highlighted and selected, populating the metadata panel. When a runbook is highlighted, both the catalog header and the runbook leaf are brightened to indicate the active selection. Supports a scrollbar and fuzzy search. Navigate with arrow keys or mouse click. |
| **Metadata** | Displays the `name`, `description`, `version`, and `risk_level` of the currently selected runbook, parsed from its `runbook.yaml`. |
| **Output** | Three-region pane: a **header** showing the command that was run, a **body** streaming stdout/stderr live, and a **footer** showing the path to the saved log file. Clicking the header copies the full command to the clipboard. Clicking the footer copies the log file path to the clipboard. Supports a scrollbar and in-pane search with match highlighting and vim-style navigation. Cleared between executions. |
| **Footer** | Status bar showing current state and available keybindings (e.g. `enter` to run, `q` to quit). |

### 6.2 Sidebar Tree Layout

The sidebar renders catalogs and runbooks as a collapsible folder tree. All catalogs are expanded by default. Each catalog header shows an expand/collapse indicator (`▾` expanded, `▸` collapsed).

```
▾ default/
  ├── hello-world [low]
  └── rotate-tls-certificates [medium]
▸ local/
▾ public-catalog.git/
  └── drain-node [high]
```

**Collapse/expand:** catalog folders can be toggled:
- `←` on a catalog header collapses it, hiding its runbooks
- `→` on a collapsed catalog header expands it, revealing its runbooks
- `Enter` or `Space` on a catalog header also toggles collapse/expand
- `←` on a runbook jumps the cursor to its parent catalog header
- **Mouse click** on a catalog header toggles collapse/expand
- **Mouse click** on a runbook selects it and updates the metadata panel
- **Mouse hover** underlines the item under the cursor. The hover highlight clears when the user switches to keyboard navigation.

**Highlight behavior:** when a runbook is highlighted, both the catalog name and the runbook leaf brighten together — the catalog header adopts the `primary` accent color and the runbook leaf uses the full `text` foreground. Unselected items render in `textMuted`. Each runbook name is followed by a colored risk badge (`[low]`, `[medium]`, `[high]`, `[critical]`).

**Startup:** on launch, all catalogs are expanded and the cursor is on the first catalog header. The first runbook is auto-selected for the metadata panel — the user never sees an empty state unless there are no runbooks at all.

**Navigation:** `↑`/`↓` arrow keys move the cursor through all visible items — both catalog headers and runbook leaves. When the cursor lands on a runbook, the metadata panel updates. When the cursor lands on a catalog header, the metadata panel retains the last selected runbook. `Enter` on a runbook triggers execution.

**Scrollbar:** a vertical scrollbar is rendered on the right edge of the sidebar when the total number of runbook entries exceeds the visible height. Scrolls one line at a time via arrow keys or mouse wheel.

**Fuzzy search:** activated by typing `/` while the sidebar has focus. A filter input appears pinned at the bottom of the sidebar panel, displayed as `Filter: <query>█` with the label in `textMuted` and the query in `text` color. As the user types, non-matching runbooks are filtered out of the tree in real time — only runbooks whose name fuzzy-matches the query remain visible. Catalog headers are hidden if all of their runbooks are filtered out. Pressing `Escape` or clearing the input restores the full tree. The highlight follows the first visible match automatically.

### 6.3 Wizard Overlay

The wizard overlays on top of the main view when the user confirms they want to execute the selected runbook. It is built using the [Huh](https://github.com/charmbracelet/huh) library — a Charm-native form framework that integrates cleanly with Bubble Tea. It steps through each parameter defined in `runbook.yaml` one at a time, collecting input. On completion, it writes all inputs to `config.json` and invokes the script.

The wizard layout mirrors the output pane — a header above the form body, both contained within the overlay:

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ $ dops run <id>                            │  │  ← header
│  ├───────────────────────────────────────────────────────┤  │
│  │                                                       │  │
│  │                      Form                             │  │  ← Huh form body
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Header** — displays the `dops run` command that will be executed once the form is submitted, including the automation ID. Clicking the header copies the command to the clipboard. Updates live as parameters are filled in to reflect the full resolved command.

**Form body** — renders all parameters defined in `runbook.yaml` as Huh input fields. Supports full mouse interaction — fields can be clicked to focus, toggles can be clicked to flip, and select pickers support mouse selection.

| Property | Behavior |
|---|---|
| **Trigger** | User selects a runbook and presses `enter` |
| **Steps** | One step per parameter in the runbook's `parameters` list |
| **Input types** | `string`: text field · `boolean`: toggle · `select`: list picker |
| **Mouse support** | All form fields support mouse click interaction via Huh's built-in mouse handling |
| **Secrets** | Parameters with `secret: true` are masked during input and encrypted with age before being written to `config.json` |
| **Save behavior** | All inputs are always saved to `config.json` after wizard completion, written to the scope defined by each parameter's `scope` field |
| **Skip behavior** | If all required parameters already have a saved value resolved from `config.json`, the wizard is skipped entirely and the runbook runs immediately |
| **Partial skip** | If some required parameters are missing, the wizard runs for only the missing fields — pre-filling everything already resolved |
| **Cancellation** | `Escape` closes the wizard without executing or saving |

### 6.4 Output Pane

The output pane has three distinct regions, as shown in the wireframe diagram (`tui-layout.png`):

```
┌─────────────────────────────────────────────────────────────┐
│ $ dops run <id> --param hello=world               │  ← header
├─────────────────────────────────────────────────────────────┤
│ hello, world!                                                │
│                                                              │  ← body
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ Saved to /tmp/2026.01.01-010102-default-hello-world.log      │  ← footer
└─────────────────────────────────────────────────────────────┘
```

**Header** — displays the full command that was executed, formatted as a shell invocation. Clicking it copies the command to the clipboard, allowing the user to re-run it outside of dops.

**Body** — streams stdout and stderr live as the script runs. stderr is rendered in the `error` theme color to distinguish it from normal output.

**Footer** — displays the path to the saved log file once execution completes. Clicking the footer copies the log path to the clipboard.

**Log filename format:**

```
/tmp/YYYY.MM.DD-HHmmss-<catalog>-<runbook>.log
```

Example: `/tmp/2026.01.01-010102-default-hello-world.log`

**Scrollbar:** a vertical scrollbar is rendered on the right edge of the output body when the content exceeds the visible height. Scrolls one line at a time via arrow keys or mouse wheel.

**Search:** activated by typing `/` while the output pane has focus. A search input appears at the bottom of the output body. As the user types, all matching occurrences in the stdout/stderr buffer are highlighted in place — the content is not filtered, matches are highlighted inline. Pressing `Enter` confirms the search and enters navigation mode:

- The status line shows the current match position and total count, e.g. `[2/7]`
- `n` moves to the next match downward
- `N` moves to the previous match upward
- The view scrolls automatically to keep the current match visible
- `Escape` exits search mode and clears all highlights

### 6.5 Visual Requirements

All four regions of the main view must be visually distinct bordered panels. Borders must use a color with sufficient contrast against the background — at minimum, the `fgMuted` palette value. Panels must not bleed into each other.

**Panel structure:**
- Each panel (sidebar, metadata, output) is wrapped in a `lipgloss.RoundedBorder()` with the `border` token as the foreground color
- The active/focused panel uses `borderActive` instead
- Panels fill their allocated space — no dead/empty areas at the bottom of the screen

**Sidebar panel requirements:**
- Selection indicator `>` renders in `primary` color (blue), bold
- Tree connectors (`├──`, `└──`) render in `textMuted`
- Each runbook name is followed by a colored risk badge: `[low]` in green, `[medium]` in yellow, `[high]` in orange, `[critical]` in red
- Background fills with `backgroundPanel`

**Metadata panel requirements:**
- Runbook name in `text` color, bold
- Description in `textMuted`
- Risk level as a colored badge (same colors as sidebar badges)
- Panel has its own rounded border, visually separate from the output pane below it

**Output pane requirements:**
- Header sub-region has `backgroundElement` as background fill — must be visibly different from the body
- The `$ dops run ...` command text must be readable (use `text` foreground on `backgroundElement` background)
- Body uses default `background`
- stderr lines in `error` color
- Footer sub-region has `backgroundElement` as background fill
- Log path text in `textMuted` — must be readable
- When no execution has occurred, show a centered placeholder: `"Press enter to run a runbook"`
- Output clears when a different runbook is selected

**Footer bar requirements:**
- Full-width bar with `backgroundPanel` as background
- Keybind keys in `primary`, descriptions in `textMuted`
- Must have consistent padding from the left edge

**Layout proportions:**
- Sidebar: 25% of width, minimum 20, maximum 40 columns
- Right panel fills remaining width
- Metadata: auto-height based on content (approximately 6-8 lines)
- Output: fills remaining vertical space between metadata and footer
- Footer: single line, pinned to bottom

---

## 7. Vars — Saved Inputs

All saved inputs live under the `vars` key in `config.json`. This section is written automatically by the TUI wizard and by `dops config set`. It is never edited manually.

### 7.1 Structure

```
vars.global.<key>                              # available to all runbooks
vars.catalog.<catalog-name>.<key>              # available to all runbooks in a catalog
vars.catalog.<catalog-name>.runbooks.<name>.<key>  # specific to one runbook
```

### 7.2 Resolution Precedence

When a runbook is about to execute, inputs are resolved in this order — each level overrides the previous:

```
vars.global  →  vars.catalog.<n>  →  vars.catalog.<n>.runbooks.<n>
```

### 7.3 Secrets

Any parameter with `secret: true` in `runbook.yaml` is encrypted with age using the key at `~/.dops/keys/keys.txt` before being written to `config.json`. The encrypted value is stored as an `age1...` ciphertext string. At execution time, dops detects any value beginning with `age1` and decrypts it before passing it to the script as an environment variable. Plain values are passed through as-is.

The `keys/` directory is intended to eventually support syncing `keys.txt` to a cloud storage bucket for multi-machine use.

---

## 8. CLI — `dops config`

`dops config` is the command-line interface for reading and writing `config.json` without opening the TUI. It follows the same `key=value` convention as `az config set`.

### 8.1 Commands

```bash
# Set a value
dops config set theme=dracula
dops config set defaults.max_risk_level=high
dops config set vars.global.region=us-east-1
dops config set vars.global.token=abc123 --secret

# Set catalog and runbook-scoped vars
dops config set vars.catalog.default.namespace=platform
dops config set vars.catalog.default.runbooks.rotate-tls.dry_run=false

# Read a value (secrets are masked)
dops config get vars.global.region

# Delete a value
dops config unset vars.global.region

# View entire config (all secrets masked)
dops config list
```

### 8.2 Flags

| Flag | Description |
|---|---|
| `--secret` | Encrypt the value with age before writing to `config.json`. Only valid with `dops config set`. |

### 8.3 Key path convention

Dot-notation maps directly to the `config.json` structure. `vars.catalog.default.namespace` resolves to `config.vars.catalog.default.namespace`. The CLI never writes to `catalogs[]` — that array is managed manually or by a future `dops catalog add` command.

### 8.4 Error Output

CLI errors use styled output with a colored badge, not raw error text or usage dumps:

```
  ERROR  Runbook not found

  runbook "unknown.runbook" not found
```

- **Badge**: bold white text on red/pink background (`ERROR`)
- **Title**: concise error summary on the same line as the badge
- **Detail**: muted/gray text on the next line with the specific error message
- **No usage dump**: errors do not print command usage. Use `--help` for that.

---

## 8b. CLI — `dops run`

`dops run` executes a runbook directly from the command line without opening the TUI. It uses the runbook's `id` field for invocation and accepts parameters as `--param key=value` flags.

### 8b.1 Usage

```bash
# Run a runbook by ID
dops run default.hello-world

# Run with parameter overrides
dops run default.hello-world --param namespace=staging --param dry_run=true

# Run with a secret parameter (prompted interactively if not saved)
dops run default.rotate-tls-certificates --param cert_manager_token=abc123 --secret cert_manager_token
```

### 8b.2 Behavior

1. Look up the runbook by `id` across all active catalogs
2. Resolve saved inputs from `config.json` (global → catalog → runbook precedence)
3. Apply `--param` overrides on top of resolved values
4. If required parameters are still missing, prompt interactively (one field at a time)
5. Save all inputs to `config.json` at the scope defined by each parameter
6. Execute the script with parameters as environment variables
7. Stream stdout/stderr to the terminal (no TUI — plain output)
8. Write log file to `/tmp/YYYY.MM.DD-HHmmss-<catalog>-<runbook>.log`

### 8b.3 Flags

| Flag | Description |
|---|---|
| `--param key=value` | Override a parameter value. Repeatable. |
| `--secret key` | Mark a `--param` value as secret — encrypt before saving. Repeatable. |
| `--no-save` | Execute without saving inputs to `config.json` |
| `--dry-run` | Show the resolved command and parameters without executing |

### 8b.4 Error Cases

- Unknown `id` → exit with error: `runbook "foo.bar" not found`
- Runbook exceeds catalog risk policy → exit with error: `runbook "foo.bar" blocked by risk policy (high > medium)`
- Missing required parameter with no TTY → exit with error: `missing required parameter "namespace" (no TTY for interactive prompt)`

---

## 9. Command Palette

The command palette is a fuzzy-searchable overlay inside the TUI, triggered with `CTRL+SHIFT+P`. It provides quick access to config operations without leaving the app, modeled after VSCode and Ghostty.

### 9.1 Supported Commands

| Command | Description |
|---|---|
| `theme: set` | Pick from available bundled and user themes |
| `config: set` | Set any config value by dot-notation key path |
| `config: view` | Display current `config.json` with all secrets masked |
| `config: delete` | Remove a saved input by key path |
| `secrets: re-encrypt` | Re-encrypt all age-encrypted values with a new key from `~/.dops/keys/keys.txt` |

### 9.2 Behavior

The palette opens as a full-width overlay at the top of the TUI. Typing filters the command list in real time. Selecting a command either executes immediately (e.g. `config: view`) or opens a secondary input prompt for the required value. All writes go through the same code path as `dops config set`.

---

## 10. Theming

dops supports a JSON-based theme system. Themes are loaded in the following priority order, with later sources overriding earlier ones:

1. Built-in themes — embedded in the binary (`dark`, `light`)
2. User themes — `~/.dops/themes/*.json`

The active theme is set in `config.json` via the `"theme"` field. dops checks user themes first, then falls back to bundled themes.

### 10.1 Theme File Structure

Themes use a two-section approach inspired by OpenCode's theme system:

- **`defs`** — a named palette. Define your colors here once and reference them by name throughout the theme. Values can be hex strings (`#RRGGBB`, `#RRGGBBAA`) or `"none"` for terminal transparency.
- **`theme`** — the token map. Each token has a `dark` and `light` variant, so a single file covers both modes. dops detects the terminal background at startup and selects the appropriate variant.

```json
{
  "$schema": "https://dops.sh/theme.json",
  "name": "tokyonight",
  "defs": {
    "bg":        "#1a1b26",
    "bgDark":    "#16161e",
    "bgElem":    "#292e42",
    "bgVisual":  "#283457",
    "fg":        "#c0caf5",
    "fgDark":    "#a9b1d6",
    "fgMuted":   "#565f89",
    "blue":      "#7aa2f7",
    "cyan":      "#7dcfff",
    "green":     "#9ece6a",
    "teal":      "#73daca",
    "orange":    "#ff9e64",
    "red":       "#f7768e",
    "redDark":   "#db4b4b",
    "purple":    "#bb9af7",
    "magenta":   "#bb9af7",
    "yellow":    "#e0af68",
    "dayBg":     "#e1e2e7",
    "dayBgPanel":"#d5d6db",
    "dayBgElem": "#c4c8da",
    "dayFg":     "#3760bf",
    "dayFgMuted":"#848cb5",
    "dayBlue":   "#2e7de9",
    "dayGreen":  "#587539",
    "dayOrange": "#b15c00",
    "dayRed":    "#f52a65",
    "dayPurple": "#9854f1"
  },
  "theme": {
    "background":        { "dark": "bg",       "light": "dayBg"      },
    "backgroundPanel":   { "dark": "bgDark",   "light": "dayBgPanel" },
    "backgroundElement": { "dark": "bgElem",   "light": "dayBgElem"  },
    "text":              { "dark": "fg",        "light": "dayFg"      },
    "textMuted":         { "dark": "fgMuted",  "light": "dayFgMuted" },
    "primary":           { "dark": "blue",     "light": "dayBlue"    },
    "border":            { "dark": "fgMuted",  "light": "dayFgMuted" },
    "borderActive":      { "dark": "blue",     "light": "dayBlue"    },
    "success":           { "dark": "green",    "light": "dayGreen"   },
    "warning":           { "dark": "orange",   "light": "dayOrange"  },
    "error":             { "dark": "red",      "light": "dayRed"     },
    "risk": {
      "low":      { "dark": "green",   "light": "dayGreen"  },
      "medium":   { "dark": "yellow",  "light": "dayOrange" },
      "high":     { "dark": "orange",  "light": "dayOrange" },
      "critical": { "dark": "redDark", "light": "dayRed"    }
    }
  }
}
```

### 10.2 Token Reference

#### Backgrounds

| Token | Description |
|---|---|
| `background` | Base application background |
| `backgroundPanel` | Sidebar and secondary panel backgrounds |
| `backgroundElement` | Input fields, selected items, wizard overlay background |

#### Text

| Token | Description |
|---|---|
| `text` | Primary foreground text |
| `textMuted` | Dimmed text — labels, hints, keybind descriptions |

#### Structure

| Token | Description |
|---|---|
| `primary` | Accent color — active borders, selected catalog header, keybind keys, wizard border, input cursor |
| `border` | Default border color for panels and dividers |
| `borderActive` | Border color for the focused/active panel |

#### Semantics

| Token | Description |
|---|---|
| `success` | Script exit success indicator in output pane |
| `warning` | General warning state |
| `error` | stderr text color, script failure indicator |

#### Risk Levels

| Token | Description |
|---|---|
| `risk.low` | Badge color for `low` risk runbooks |
| `risk.medium` | Badge color for `medium` risk runbooks |
| `risk.high` | Badge color for `high` risk runbooks |
| `risk.critical` | Badge color for `critical` risk runbooks |

### 10.3 Bundled Themes

dops ships with one built-in theme embedded in the binary. Additional themes will be added in future releases.

| Name | Description |
|---|---|
| `tokyonight` | Default theme. Dark variant uses the official Tokyo Night Night palette (`#1a1b26` background). Light variant uses the Tokyo Night Day palette (`#e1e2e7` background). Colors sourced from [folke/tokyonight.nvim](https://github.com/folke/tokyonight.nvim). |

User themes in `~/.dops/themes/` always take precedence over bundled themes. If a user creates `~/.dops/themes/tokyonight.json`, it overrides the bundled version.

### 10.4 Color Format

All color values must be valid hex strings: `#RGB`, `#RGBA`, `#RRGGBB`, or `#RRGGBBAA`. The special value `"none"` is supported for backgrounds to inherit terminal transparency.

---

## 11. Go Implementation Notes

### 11.1 Bubble Tea Model

Mouse support is enabled at program initialization:

```go
p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

The model handles `tea.MouseMsg` events for clickable regions — sidebar runbook selection, output pane header and footer copy actions, and wizard form interactions delegated to Huh.

```go
type model struct {
    catalogs []Catalog      // loaded from config.json, active only
    runbooks []RunbookGroup // grouped by catalog, filtered by policy
    selected *Runbook       // currently highlighted runbook
    wizard   *WizardState  // nil when inactive
    output   string        // accumulated stdout/stderr
    running  bool          // true while script is executing
    theme    Theme         // resolved theme, applied via lipgloss
}
```

### 11.2 Key Packages

| Package | Purpose |
|---|---|
| `github.com/charmbracelet/bubbletea` | Core TUI framework — model, update, view loop |
| `github.com/charmbracelet/bubbles/list` | Sidebar runbook list with grouping |
| `github.com/charmbracelet/huh` | Wizard form — parameter input fields, toggles, and select pickers |
| `github.com/charmbracelet/bubbles/textinput` | String parameter input in wizard |
| `github.com/charmbracelet/lipgloss` | Layout, borders, colors, and styling |
| `encoding/json` | Parsing `config.json` and theme files |
| `gopkg.in/yaml.v3` | Parsing `runbook.yaml` |
| `os/exec` | Script execution with stdout/stderr streaming via `tea.Cmd` |
| `filippo.io/age` | Encrypting and decrypting secret values in `config.json` |

### 11.3 Script Execution

Parameters collected by the wizard are passed to the script as **environment variables**. The script is invoked via `os/exec` and its stdout/stderr piped back to the TUI via a `tea.Cmd` that sends messages as each line is read. This keeps the Bubble Tea event loop non-blocking while streaming live output.

### 11.4 Theme Loading

Theme resolution order at startup:

1. Read `"theme"` field from `config.json`
2. Search `~/.dops/themes/<name>.json`
3. If not found, fall back to bundled theme by name
4. If neither found, fall back to bundled `tokyonight` theme
5. Resolve all `defs` references in `theme` tokens
6. Detect terminal background (dark/light) and select the appropriate variant per token
7. Build `lipgloss.Style` values from resolved colors and store in `model.theme`

---

## 12. Startup Behavior

On launch, dops should:

1. Read and parse `~/.dops/config.json`
2. Resolve and load the active theme (see §11.4)
3. Filter catalogs where `active: true`
4. For each active catalog, walk its subdirectories and parse `runbook.yaml` in each
5. Filter out runbooks whose `risk_level` exceeds the catalog's `max_risk_level`
6. Group remaining runbooks by catalog name
7. Auto-select the first runbook of the first active catalog and populate the metadata panel
8. Render the main TUI view

If `config.json` does not exist, dops should create `~/.dops/` and a default `config.json` with an empty `catalogs` array and the `tokyonight` theme, then display an empty state in the sidebar.

---

*dops implementation spec — generated from design session*
