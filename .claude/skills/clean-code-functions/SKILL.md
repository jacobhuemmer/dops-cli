---
name: clean-code-functions
description: Clean Code function design principles for Go. Use when creating, editing, or reviewing Go functions and methods. Triggers on new function definitions, refactoring existing functions, adding parameters, or when a function is growing in size or complexity.
user-invocable: false
---

# Clean Code Functions for Go

Apply these principles when writing or modifying Go functions.

## Size

- Functions should be small — aim for under 30 lines of logic
- If a function is hard to name, it's doing too much — split it
- If you need to scroll to see the whole function, it's too long
- Extract sections guarded by comments into their own functions — the comment is the name

## Single Responsibility

- A function does ONE thing at ONE level of abstraction
- "One thing" means: one reason for someone to read or change this function
- If you can extract a meaningful sub-function, the original was doing more than one thing

## Single Level of Abstraction (SLA)

- Don't mix orchestration with implementation detail in the same function
- Top-level functions should read like a high-level description:
  ```go
  func (s *Server) HandleDeploy(ctx context.Context, req DeployRequest) error {
      target, err := s.resolveTarget(req)
      if err != nil {
          return fmt.Errorf("resolve target: %w", err)
      }
      if err := s.validate(target); err != nil {
          return fmt.Errorf("validate: %w", err)
      }
      return s.deploy(ctx, target)
  }
  ```
- Lower-level functions handle the detail

## Arguments

- Fewer is better: zero, one, or two arguments are ideal
- Three or more arguments — consider a config/options struct:
  ```go
  // Instead of:
  func Connect(host string, port int, timeout time.Duration, tls bool) error

  // Use:
  type ConnectConfig struct {
      Host    string
      Port    int
      Timeout time.Duration
      TLS     bool
  }
  func Connect(cfg ConnectConfig) error
  ```
- Use functional options pattern for optional configuration in public APIs
- `context.Context` is always the first parameter and doesn't count toward the limit

## Command-Query Separation

- A function should either DO something (command) or RETURN something (query), not both
- Commands: `Save`, `Delete`, `Send` — return `error` only
- Queries: `Get`, `Find`, `List` — return data + `error`
- Exception: Go's `ok` idiom is fine: `val, ok := m[key]`

## No Side Effects

- If named `Validate`, it should not also modify the input
- If named `Get`, it should not create a record if missing (unless explicitly named `GetOrCreate`)
- Functions should not depend on or modify hidden state
- If side effects are necessary, make them explicit in the name

## Return Values

- Return early on errors — avoid deep nesting:
  ```go
  // Good
  if err != nil {
      return err
  }
  // continue happy path

  // Bad
  if err == nil {
      // nested happy path
  }
  ```
- Use named returns sparingly — only when they clarify what is returned in a complex signature
- Return concrete types from constructors, accept interfaces as parameters

## Spec Traceability

- Each spec acceptance criterion should map to one or a small cluster of functions
- If a function cannot be traced to a spec requirement, question whether it's needed (YAGNI)
- The function name should make the spec connection obvious
