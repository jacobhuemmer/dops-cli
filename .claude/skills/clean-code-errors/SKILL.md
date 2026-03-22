---
name: clean-code-errors
description: Clean Code error handling principles for Go. Use when writing, editing, or reviewing error handling in Go code — error returns, error wrapping, sentinel errors, custom error types, error checking patterns, and panic/recover usage.
user-invocable: false
---

# Clean Code Error Handling for Go

Apply these principles when writing or modifying error handling in Go.

## Fail Fast

- Validate inputs at system boundaries (CLI args, config, API inputs) immediately
- Return errors at the first point of failure — don't continue in an inconsistent state
- Precondition checks go at the top of the function:
  ```go
  func Deploy(ctx context.Context, target string) error {
      if target == "" {
          return errors.New("target must not be empty")
      }
      // ... proceed with valid input
  }
  ```

## Error Wrapping

- Always add context when propagating errors with `fmt.Errorf` and `%w`:
  ```go
  result, err := s.store.Get(ctx, id)
  if err != nil {
      return nil, fmt.Errorf("get deployment %s: %w", id, err)
  }
  ```
- The wrap message should describe what *this function* was doing, not restate the underlying error
- This creates a chain: `"deploy to prod: get deployment abc123: connection refused"`
- Use `%w` (not `%v`) so callers can use `errors.Is` and `errors.As`

## Sentinel Errors

- Define package-level sentinel errors for conditions callers need to check:
  ```go
  var (
      ErrNotFound    = errors.New("not found")
      ErrConflict    = errors.New("conflict")
  )
  ```
- Keep sentinels minimal — only for conditions that change caller behavior
- Callers check with `errors.Is(err, ErrNotFound)`, which works through wrap chains

## Custom Error Types

- Use custom types when the caller needs structured information from the error:
  ```go
  type ValidationError struct {
      Field   string
      Message string
  }

  func (e *ValidationError) Error() string {
      return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
  }
  ```
- Callers extract with `errors.As`:
  ```go
  var ve *ValidationError
  if errors.As(err, &ve) {
      // handle validation error with access to ve.Field
  }
  ```
- Do not overuse — simple `fmt.Errorf` wrapping is sufficient for most cases

## Error Handling Patterns

- Handle the error OR return it — never both (don't log AND return):
  ```go
  // Bad: error is handled twice
  if err != nil {
      log.Error("failed to deploy", "err", err)
      return err  // caller will also log it
  }

  // Good: return with context, let the caller decide
  if err != nil {
      return fmt.Errorf("deploy: %w", err)
  }
  ```
- Log errors at the top of the call stack (usually `main` or the HTTP/CLI handler)
- Use early returns to keep the happy path unindented

## Never Ignore Errors

- Every error return must be checked
- If you genuinely don't need the error, document why:
  ```go
  _ = w.Close() // best-effort cleanup; write already succeeded
  ```
- Never use blank identifier for errors silently

## Panic and Recover

- `panic` is for programmer errors (bugs), never for expected runtime conditions
- Do not use `panic` for input validation, missing files, network errors, etc.
- `recover` belongs only in top-level middleware (HTTP handler, CLI root) — never in library code
- If you're writing `panic` followed by a string, you almost certainly want `return error` instead

## Do Not Return Nil for Errors

- If a function returns `(T, error)`, a nil error means T is valid — always
- Do not return `(nil, nil)` to mean "not found" — use a sentinel error or a boolean:
  ```go
  // Bad
  func Find(id string) (*Thing, error) {
      // returns (nil, nil) if not found — caller must check both

  // Good
  func Find(id string) (*Thing, error) {
      // returns (nil, ErrNotFound) if not found
  ```

## Spec Traceability

- Specs should define error scenarios explicitly
- Each spec error case becomes a test case
- Error messages should be user-meaningful when they surface in CLI output
