---
name: clean-code-design
description: Clean Code design maxims for Go — DRY, KISS, YAGNI, separation of concerns, composition over inheritance, Law of Demeter. Use when making architectural decisions, creating new packages or modules, refactoring code structure, or reviewing overall code organization in Go projects.
user-invocable: false
---

# Clean Code Design Principles for Go

Apply these principles when making structural or architectural decisions.

## YAGNI — You Aren't Gonna Need It

- Implement only what the current spec requires — nothing more
- Do not add configuration options "in case someone needs them later"
- Do not build plugin architectures, feature flags, or extension points speculatively
- If it's not in the spec, it doesn't get built. Period.
- The cost of building the wrong abstraction is higher than the cost of adding it later

## KISS — Keep It Simple

- The simplest solution that satisfies the spec is the best solution
- Three similar lines of code are better than a premature abstraction
- Choose boring technology over clever technology
- If a junior developer can't understand the code in 5 minutes, it's too complex
- Complexity debt compounds faster than technical debt

## DRY — Don't Repeat Yourself

- Every piece of domain knowledge has a single authoritative representation
- DRY applies to knowledge, not just code — identical-looking code serving different purposes is NOT duplication
- Before extracting, ask: "If one of these changes, must the other change too?" If no, they're not duplicates.
- Apply DRY across: code, config, specs, build scripts
- Do NOT apply DRY to tests — test readability beats test DRYness

## Separation of Concerns

- Each package has one clear purpose
- Do not mix I/O with business logic:
  ```go
  // Bad: business logic mixed with I/O
  func ProcessConfig(path string) error {
      data, err := os.ReadFile(path)  // I/O
      // ... validation logic ...      // business logic
      // ... transformation ...         // business logic
      return os.WriteFile(out, result, 0644)  // I/O
  }

  // Good: separated
  func LoadConfig(path string) ([]byte, error) { ... }     // I/O
  func ValidateConfig(data []byte) error { ... }            // logic
  func TransformConfig(data []byte) ([]byte, error) { ... } // logic
  func SaveConfig(path string, data []byte) error { ... }   // I/O
  ```
- CLI parsing, business logic, and infrastructure are separate layers
- For a Go CLI: `cmd/` (CLI wiring) → `internal/` (business logic) → adapters (I/O, external services)

## Composition Over Inheritance

- Go has no inheritance — this principle is built into the language
- Use struct embedding for code reuse, interfaces for polymorphism
- Prefer small, composable types over large monolithic ones:
  ```go
  // Compose behaviors
  type DeployService struct {
      validator Validator
      pusher    Pusher
      notifier  Notifier
  }
  ```
- Favor function parameters and return values over method receivers when behavior doesn't need state

## Law of Demeter

- A function should only call methods on: its receiver, its parameters, objects it creates, its direct fields
- No train wrecks: `s.config.Database.Connection.Pool.Get()` — each `.` is a coupling point
- If you need data deep inside a structure, expose a method at the right level:
  ```go
  // Bad
  pool := s.config.Database.Connection.Pool
  conn := pool.Get()

  // Good
  conn, err := s.GetDBConnection()
  ```

## Package Design

- Packages are the unit of encapsulation in Go
- Name packages by what they provide, not what they contain
- Avoid circular dependencies — if A imports B and B needs A, extract an interface package
- `internal/` enforces visibility at the compiler level — use it
- Keep `main.go` thin: parse flags, wire dependencies, call `Run()`

## Boy Scout Rule

- When implementing a spec, clean up the code you touch
- Rename a confusing variable, extract a function, remove dead code
- Don't refactor unrelated code — scope cleanup to what you're working on
- Small, continuous improvements prevent decay

## Kent Beck's Four Rules of Simple Design (Priority Order)

1. **Runs all the tests** — all spec criteria pass
2. **Contains no duplication** — shared concepts implemented once
3. **Expresses intent** — code reads like the spec
4. **Minimizes classes/functions** — no unnecessary abstractions

When in conflict, higher-numbered rules yield to lower-numbered ones.
