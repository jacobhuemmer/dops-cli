# Vault Migration Plan

## Date: 2026-03-25

## Context

Saved parameter values (including encrypted secrets) currently live inside `config.json` alongside user-editable settings like theme, catalogs, and defaults. This is a problem because:

1. Users can accidentally corrupt encrypted values while editing config
2. No tamper detection — someone can modify encrypted blobs without detection
3. Sensitive data shares the same file permissions as non-sensitive config
4. No clear boundary between "user owns this" and "CLI owns this"

## Goal

Move all saved parameter values (`vars`) from `config.json` into a new `vault.json` that is:
- Encrypted as a single age blob (entire file)
- Tamper-resistant via age's AEAD (ChaCha20-Poly1305)
- Permissions locked to `0600`
- Only readable/writable by the CLI

## Design

### File Format

```json
{
  "version": 1,
  "data": "age1<base64-encoded encrypted Vars JSON>"
}
```

- `version` — schema version for future migrations
- `data` — the entire `domain.Vars` struct serialized to JSON, then encrypted with age

When decrypted, `data` yields:

```json
{
  "global": { "region": "us-east-1" },
  "catalog": {
    "infra": {
      "api_token": "secret-value",
      "runbooks": {
        "scale-deployment": {
          "namespace": "production"
        }
      }
    }
  }
}
```

### Encryption

- Age AEAD provides authenticated encryption — any tamper causes decryption failure
- Uses the same `~/.dops/keys/keys.txt` identity already used for per-value encryption
- Individual values are stored as **plaintext inside the encrypted blob** (no double encryption)
- The entire file is opaque to anyone without the key

### File Location

```
~/.dops/
├── config.json    # user-editable (theme, catalogs, defaults) — 0644
├── vault.json     # CLI-managed (all saved vars) — 0600
└── keys/
    └── keys.txt   # age identity — 0600
```

### In-Memory Model

`domain.Config.Vars` stays as the in-memory working copy. All existing code (resolvers, path router, mask) continues to read/write `cfg.Vars` unchanged. Only the persistence boundary changes:

- **Load**: read `config.json` → populate config; read `vault.json` → decrypt → populate `cfg.Vars`
- **Save vars**: serialize `cfg.Vars` → encrypt → write `vault.json` (0600)
- **Save config**: serialize config (without Vars) → write `config.json`

### Migration

On startup, if `config.json` contains a non-empty `vars` key and `vault.json` does not exist:

1. Read vars from `config.json`
2. Write them to `vault.json` (encrypted, 0600)
3. Remove `vars` key from `config.json`
4. Save cleaned `config.json`

This is a one-time, automatic migration.

## Implementation Steps

### Step 1: Create vault package
- `internal/vault/vault.go` — `Vault` struct with `Load()`, `Save()` methods
- `internal/vault/vault_test.go` — round-trip, tamper detection, migration tests
- Uses `crypto.AgeEncrypter` for encrypt/decrypt
- `Save()` writes with `os.WriteFile(path, data, 0o600)`

### Step 2: Update domain
- Add `json:"-"` tag to `Vars` field in `domain.Config` so it's excluded from config.json serialization
- Or handle via custom marshal — keep Vars in struct but don't write to config.json

### Step 3: Wire vault into startup
- `cmd/root.go` — create vault, load vars, populate `cfg.Vars`, pass vault to TUI deps
- Add `Vault` to `tui.AppDeps`
- Migration logic runs here (before TUI launch)

### Step 4: Update write paths
- `internal/tui/wizard/model.go` — `saveCurrentField()` saves via vault instead of config store
- `cmd/run.go` — `saveInputs()` saves via vault
- `cmd/config.go` — `vars.*` set/unset commands save via vault
- Individual secret values no longer need per-value encryption (vault encrypts everything)

### Step 5: Update read paths
- `cmd/config.go` — `vars.*` get/list commands read from `cfg.Vars` (populated from vault at load)
- MCP server — receives `cfg.Vars` via config as before (no change needed)
- Resolvers — no change (still read `cfg.Vars`)

### Step 6: Remove per-value encryption
- Wizard `saveCurrentField()` — remove age encrypt logic (vault handles it)
- `cmd/run.go` `saveInputs()` — remove per-value encryption
- `crypto.IsEncrypted()` / `DecryptingVarResolver` — keep for backward compat during migration, remove later
- `crypto.MaskSecrets()` — update to mask secret-flagged params rather than checking `age1` prefix

### Step 7: Update docs
- README.md — new Vault section with encryption, tamper detection, sops comparison, migration
- docs/reference/configuration.md — expanded Vault Encryption section with AEAD details, comparison table, migration
- docs/guides/runbooks.md — scopes table references vault.json
- specs/vault-v0.3.0.md — new §10 Tamper Detection with sops comparison, error messages
- internal/mcp/prompts.go — scope storage references vault.json
- SKILL.md — no changes needed (shows schema only, not storage details)

## Verification

- [x] `go test ./...` passes
- [x] New vault round-trip tests (encrypt → save → load → decrypt → compare)
- [x] Tamper test (modify vault.json byte → load fails)
- [x] Migration test (config.json with vars → vault.json created, config.json cleaned)
- [x] TUI wizard saves to vault (not config.json)
- [x] `dops run` saves to vault
- [x] `dops config set vars.global.x=y` saves to vault
- [x] `dops config get vars.global.x` reads from vault
- [x] `dops config list` shows vars from vault
- [x] MCP tool execution resolves vars from vault
- [x] File permissions on vault.json are 0600
- [x] Fresh install (no vault.json) works correctly
- [x] Existing install with vars in config.json migrates automatically

## Risks

- **Vault corruption**: if vault.json is corrupted, all saved vars are lost. Consider writing a `.vault.json.bak` before overwriting.
- **Key loss**: if `keys.txt` is deleted, vault becomes unreadable. Document recovery (re-enter values).
- **Concurrent access**: dops doesn't run multiple instances typically, but if it did, concurrent vault writes could conflict. Not a concern for v1.
