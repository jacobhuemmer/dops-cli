# dops-cli — Remaining Feature TODO

## High Priority

- [ ] **Risk confirmation gates** — Execution should require confirmation based on risk level:
  - Low: no confirmation
  - Medium: simple y/N prompt
  - High: must type the runbook ID
  - Critical: must type "CONFIRM"

- [ ] **Process management (ctrl+x stop)** — Allow stopping a running execution with ctrl+x. Send SIGKILL with WaitDelay. Update output footer with exit status.

## Medium Priority

- [ ] **Text selection (click/drag to copy)** — Click and drag in the output log to select text. `y` key to copy selection. Section-aware extraction (header/log/footer). Selection highlighting.

- [ ] **Catalog management CLI** — `dops catalog install <url>`, `dops catalog add`, `dops catalog remove`, `dops catalog update`, `dops catalog list`. Git-based catalog installs.

- [ ] **Help overlay (? key)** — Context-aware keybinding help overlay. Shows different bindings per focused pane.

## Low Priority

- [ ] **OSC 52 clipboard fallback** — Fall back to OSC 52 terminal escape for clipboard in SSH/remote sessions when native clipboard is unavailable.

- [ ] **MCP server integration** — MCP server with tools, resources, schema, progress, and watcher support.

- [ ] **Dry-run mode** — `--dry-run` flag that shows what would execute without running the script.
