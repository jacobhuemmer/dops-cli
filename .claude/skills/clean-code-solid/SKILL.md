---
name: clean-code-solid
description: SOLID design principles applied to Go. Use when creating or editing Go types, structs, interfaces, or when making architectural decisions about type relationships, dependency injection, or package structure. Triggers on new struct/interface definitions, refactoring type hierarchies, or adding dependencies between packages.
user-invocable: false
---

# SOLID Principles for Go

Apply these principles when designing types, interfaces, and package structure.

## Single Responsibility Principle (SRP)

- A struct should have one reason to change — one responsibility, one owner
- If a struct has methods spanning multiple concerns (e.g., parsing AND persistence), split it
- Packages are also a unit of responsibility — each package has a clear purpose
- Signal a violation: you can't describe what the struct does without using "and"
- Apply to functions too: a function with multiple flag arguments often handles multiple responsibilities

```go
// Bad: mixed concerns
type Report struct { ... }
func (r *Report) Generate() []byte { ... }
func (r *Report) SaveToFile(path string) error { ... }
func (r *Report) SendEmail(to string) error { ... }

// Good: separated concerns
type Report struct { ... }
func (r *Report) Generate() []byte { ... }

type ReportWriter struct { ... }
func (w *ReportWriter) Save(data []byte, path string) error { ... }

type Notifier struct { ... }
func (n *Notifier) Send(to string, data []byte) error { ... }
```

## Open/Closed Principle (OCP)

- Types should be open for extension, closed for modification
- In Go, this means: define behavior via interfaces, add new implementations without changing existing code
- New specs should result in new types implementing existing interfaces, not edits to switch statements
- Signal a violation: adding a new feature requires modifying existing, tested code

```go
// Good: new output formats don't change existing code
type Formatter interface {
    Format(data []byte) ([]byte, error)
}

type JSONFormatter struct{}
func (f *JSONFormatter) Format(data []byte) ([]byte, error) { ... }

// Adding YAML support = new file, no changes to existing code
type YAMLFormatter struct{}
func (f *YAMLFormatter) Format(data []byte) ([]byte, error) { ... }
```

## Liskov Substitution Principle (LSP)

- Any implementation of an interface must be substitutable without changing correctness
- If a function accepts an `io.Writer`, every writer must behave like a writer — no panics, no no-ops that silently discard data
- Do not implement interfaces partially — if you can't fulfill the contract, don't implement it
- Signal a violation: type assertions or type switches on interface values to handle special cases

## Interface Segregation Principle (ISP)

- Keep interfaces small — Go interfaces are implicitly satisfied, so small interfaces are natural
- Prefer `io.Reader` (1 method) over a `FileSystem` interface with 15 methods
- If a consumer only needs `Read`, don't make it depend on `ReadWriteCloser`
- Define interfaces at the call site (consumer), not at the implementation site
- Standard library models this well: `io.Reader`, `io.Writer`, `fmt.Stringer`

```go
// Bad: fat interface forces unnecessary dependencies
type Repository interface {
    Get(id string) (*Entity, error)
    List() ([]*Entity, error)
    Create(e *Entity) error
    Update(e *Entity) error
    Delete(id string) error
    Migrate() error
    Backup() error
}

// Good: segregated by consumer need
type EntityGetter interface {
    Get(id string) (*Entity, error)
}

type EntityLister interface {
    List() ([]*Entity, error)
}

type EntityWriter interface {
    Create(e *Entity) error
    Update(e *Entity) error
    Delete(id string) error
}
```

## Dependency Inversion Principle (DIP)

- High-level packages should not import low-level packages
- Both should depend on interfaces (abstractions)
- In Go: accept interfaces, return structs
- Inject dependencies via constructor parameters, not global state

```go
// Bad: high-level directly depends on low-level
package deploy

import "myapp/internal/awsclient"

func Deploy(target string) error {
    client := awsclient.New()  // hard dependency
    return client.Push(target)
}

// Good: depend on abstraction
package deploy

type Pusher interface {
    Push(target string) error
}

func Deploy(p Pusher, target string) error {
    return p.Push(target)
}
```

- Wire dependencies at the composition root (usually `main.go` or a `wire` function)
- Never import from `main` or `cmd` in library packages

## Go-Specific Guidance

- Go does not have inheritance — use embedding for reuse, interfaces for polymorphism
- Prefer composition: embed types to reuse behavior, don't simulate class hierarchies
- Accept interfaces, return concrete types — this is the Go way to apply DIP
- Keep the dependency graph acyclic — if two packages need each other, extract a shared interface package
