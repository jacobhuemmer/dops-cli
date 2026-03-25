# Vault — Encrypted Parameter Storage
### Spec · v0.3.0

---

## 1. Overview

`vault.json` is an encrypted file that stores all saved parameter values for dops runbooks. It replaces the `vars` section previously stored as plaintext inside `config.json`.

The vault encrypts its entire contents as a single age blob, providing confidentiality and tamper detection via authenticated encryption (AEAD). The file is locked to `0600` permissions and is only readable/writable by the dops CLI.

---

## 2. Directory Layout

```
~/.dops/
├── config.json    # User-editable settings (theme, catalogs, defaults)
├── vault.json     # CLI-managed encrypted parameter store (0600)
└── keys/
    └── keys.txt   # age X25519 identity (0600, auto-generated)
```

---

## 3. File Format

```json
{
  "version": 1,
  "data": "age1..."
}
```

| Field | Type | Description |
|-------|------|-------------|
| `version` | int | Schema version. Currently `1`. |
| `data` | string | age-encrypted blob containing the serialized Vars JSON. |

### Decrypted Payload

When decrypted, `data` yields a JSON object matching the `domain.Vars` structure:

```json
{
  "global": {
    "region": "us-east-1",
    "api_token": "sk-..."
  },
  "catalog": {
    "infra": {
      "namespace": "production",
      "runbooks": {
        "scale-deployment": {
          "replicas": "3"
        }
      }
    }
  }
}
```

All values are stored as plaintext inside the encrypted blob. There is no per-value encryption — the vault provides encryption at the file level.

---

## 4. Encryption

### Algorithm
- **age** with X25519 key exchange
- AEAD: ChaCha20-Poly1305 (provided by age internally)
- Tamper detection is inherent — any modification to the ciphertext causes decryption failure

### Key Management
- Identity stored at `~/.dops/keys/keys.txt`
- Auto-generated on first use via `crypto.NewAgeEncrypter(keysDir)`
- Same key used for vault encryption and any legacy per-value encryption

### File Permissions
- `vault.json`: `0600` (owner read/write only)
- `keys/keys.txt`: `0600` (owner read/write only)

---

## 5. Operations

### Load

1. Read `vault.json` from disk
2. Parse JSON envelope, extract `data` field
3. Decrypt `data` with age identity from `keys/keys.txt`
4. Unmarshal decrypted JSON into `domain.Vars`
5. If `vault.json` does not exist, return empty `Vars` (first run)
6. If decryption fails, return error (tampered or wrong key)

### Save

1. Marshal `domain.Vars` to JSON
2. Encrypt JSON with age recipient from `keys/keys.txt`
3. Wrap in envelope: `{"version": 1, "data": "<ciphertext>"}`
4. Write to `vault.json` with `0600` permissions
5. Atomic write (write to temp file, then rename) to prevent corruption

### Migration (one-time)

On startup, if `config.json` contains a non-empty `vars` section and `vault.json` does not exist:

1. Extract `vars` from loaded config
2. Save to `vault.json` (encrypted)
3. Remove `vars` from config
4. Save cleaned `config.json`

After migration, `config.json` no longer contains a `vars` key.

---

## 6. API

### Vault struct

```go
type Vault struct {
    path    string
    keysDir string
}

func New(path, keysDir string) *Vault
func (v *Vault) Load() (*domain.Vars, error)
func (v *Vault) Save(vars *domain.Vars) error
func (v *Vault) Exists() bool
```

### Integration Points

| Component | Operation | How it uses the vault |
|-----------|-----------|----------------------|
| `cmd/root.go` | Load | Loads vault on startup, populates `cfg.Vars` |
| `cmd/root.go` | Migrate | Moves vars from config.json to vault.json |
| `cmd/config.go` | Read/Write | `vars.*` get/set/unset/list commands |
| `cmd/run.go` | Write | Saves parameter inputs after execution |
| TUI wizard | Write | Saves field values on "Save for future runs? Yes" |
| TUI app | Read | Resolves saved vars for wizard pre-fill |
| MCP server | Read | Resolves saved vars for tool execution |
| Var resolvers | Read | Reads `cfg.Vars` (populated from vault at load) |

---

## 7. Behavior

### First Run (no vault.json)
- Vault load returns empty `Vars`
- No error, no migration needed
- Vault created on first parameter save

### Existing Install (vars in config.json)
- Migration runs automatically on startup
- Vars moved to vault.json, removed from config.json
- User sees no change in behavior

### Tampered vault.json
- Age decryption fails with error
- CLI prints clear error: "vault.json is corrupted or was modified outside dops"
- User must delete vault.json and re-enter saved values

### Lost keys.txt
- Vault cannot be decrypted
- CLI prints clear error: "cannot decrypt vault — age key not found"
- User must delete vault.json and re-enter saved values

### Empty vault
- Valid state — vault exists with encrypted empty Vars
- `{"version":1,"data":"age1..."}`  where decrypted payload is `{}`

---

## 8. What Changes from v0.2.0

| Before (v0.2.0) | After (v0.3.0) |
|------------------|----------------|
| Vars stored as plaintext in `config.json` | Vars stored encrypted in `vault.json` |
| Per-value age encryption for `secret: true` params | Entire vault encrypted; no per-value encryption |
| `config.json` contains `vars` key | `config.json` has no `vars` key |
| `crypto.IsEncrypted()` checks `age1` prefix on values | Vault decryption handles all values uniformly |
| File permissions inherited from config | `vault.json` locked to `0600` |

---

## 9. Backward Compatibility

- Automatic one-time migration from config.json vars to vault.json
- `DecryptingVarResolver` kept temporarily to handle legacy per-value encrypted values during migration window
- After migration, all values in vault are plaintext (inside the encrypted blob)
- No breaking changes to runbook.yaml schema or CLI commands

---

## 10. Tamper Detection

### How It Works

The vault uses age's authenticated encryption (ChaCha20-Poly1305 AEAD). The Poly1305 authentication tag covers the entire ciphertext. During decryption, age recomputes the tag and compares it to the stored tag. If any byte has been modified — ciphertext, nonce, or header — the tags won't match and decryption fails.

This means:
- **Bit flips** in the ciphertext → authentication failure
- **Truncation** of the file → parse or authentication failure
- **Replacement** of the `data` field → authentication failure (different key stream)
- **Replay** of an old vault → succeeds (same key, valid ciphertext) — this is acceptable since the vault is a local-only store

### Comparison with sops

[sops](https://github.com/getsops/sops) is a widely-used tool for managing encrypted secrets in config files. It takes a different architectural approach:

| | sops | dops vault |
|---|------|------------|
| **What's encrypted** | Individual values (keys remain plaintext) | Entire file content |
| **File readability** | Human-readable structure, encrypted values | Fully opaque blob |
| **Tamper detection** | Explicit HMAC (MAC field) computed over all encrypted values | Inherent from AEAD authentication tag |
| **Why tamper detection works** | MAC covers all values; changing any value invalidates the MAC | AEAD covers entire ciphertext; changing any byte invalidates the auth tag |
| **Diffability** | Git-friendly — keys visible, values change | Not diff-friendly — entire blob changes on any edit |
| **Selective edits** | Edit individual values with `sops set` | Must decrypt, modify, re-encrypt entire payload |
| **Key rotation** | `sops updatekeys` rotates without decrypting values | Re-encrypt entire vault with new key |

**Why sops needs a separate MAC:** sops encrypts values individually while leaving keys and structure as plaintext. Without the MAC, an attacker could swap encrypted values between keys, remove key-value pairs, or inject values encrypted with a known key — all without detection. The MAC binds the entire tree together.

**Why the dops vault doesn't need one:** The vault encrypts everything as a single blob. There's no structure to manipulate — the only attack is modifying the ciphertext, which AEAD catches inherently.

### Error Messages

| Scenario | Error |
|----------|-------|
| Modified ciphertext | `vault.json is corrupted or was modified outside dops` |
| Missing key file | `init decryption: read key file: ...` |
| Wrong key | `vault.json is corrupted or was modified outside dops` (age can't distinguish wrong key from tampered data) |

---

## 11. Security Considerations

- **At rest**: all saved values encrypted with age AEAD (ChaCha20-Poly1305)
- **In memory**: values decrypted and held in `domain.Vars` during process lifetime
- **On disk**: `vault.json` and `keys.txt` both `0600`
- **No double encryption**: secret values stored as plaintext inside the encrypted vault
- **Tamper detection**: AEAD authentication tag covers entire ciphertext — no silent corruption
- **Atomic writes**: temp file + rename prevents partial writes on crash
- **Key scope**: single age identity per dops installation
- **No remote key management**: keys are local-only; lost keys mean lost vault
