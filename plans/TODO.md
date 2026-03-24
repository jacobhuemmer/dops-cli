# dops-cli — Remaining Feature TODO

All features from the parity audit have been implemented.

## Completed (2026-03-23)

- [x] **Risk confirmation gates** — Low=skip, Medium=y/N, High=type ID, Critical=type CONFIRM
- [x] **Process management (ctrl+x stop)** — Cancellable context, SIGKILL with process group, 2s WaitDelay
- [x] **Help overlay (? key)** — Context-aware keybindings based on focused pane
- [x] **Catalog management CLI** — `dops catalog list/add/remove/install/update`
- [x] **Dry-run mode** — Shows resolved command and env without executing
- [x] **Text selection and clipboard** — Click/drag selection, y to copy, OSC 52 fallback
- [x] **OSC 52 clipboard fallback** — For SSH/remote terminal sessions

## Future Enhancements

- [ ] Selection highlighting in the output view (visual feedback during drag)
- [ ] MCP server integration
- [ ] Additional input types (multi_select, file_path, resource_id) in wizard
- [ ] Sidebar folder compaction (single-child chains → "parent / child")
- [ ] Spinner during execution
- [ ] Update check banner
