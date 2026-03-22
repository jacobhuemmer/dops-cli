---
name: clean-code-boundaries
description: Clean Code boundary and encapsulation principles for Go. Use when integrating third-party libraries, calling external APIs, wrapping SDK clients, defining package visibility, or managing dependencies at the edges of the system.
user-invocable: false
---

# Clean Code Boundaries for Go

Apply these principles when code interacts with external dependencies or defines visibility boundaries.

## Wrap External Dependencies

- Never let third-party types leak into your domain:
  ```go
  // Bad: domain code depends on AWS SDK types
  func Deploy(ctx context.Context, client *s3.Client, bucket string) error

  // Good: domain code depends on your interface
  type ObjectStore interface {
      Put(ctx context.Context, key string, data []byte) error
  }

  func Deploy(ctx context.Context, store ObjectStore, target string) error
  ```
- Create thin adapters that implement your interface using the third-party library
- Adapter packages live at the edge: `internal/adapters/awsstore/`, not mixed into domain code
- When the library changes (or you switch libraries), only the adapter changes

## Define Interfaces at the Consumer

- The package that *uses* the dependency defines the interface it needs
- The package that *implements* it satisfies the interface implicitly
- This is Go's superpower — implicit interface satisfaction means zero coupling between consumer and provider:
  ```
  internal/deploy/        → defines Pusher interface (consumer)
  internal/adapters/aws/  → implements Pusher (provider, doesn't import deploy)
  cmd/root.go             → wires aws.Client as deploy.Pusher
  ```

## Encapsulation via Package Visibility

- Default to unexported (lowercase) — export only what external packages need
- Use `internal/` to enforce boundaries at the compiler level:
  ```
  internal/deploy/   — importable only by this module
  internal/config/   — importable only by this module
  pkg/               — only if you intentionally want external consumers
  ```
- A package's exported API is its contract — keep it minimal
- If you export something, you maintain it. Unexport aggressively.

## Principle of Least Privilege

- Give each component access to only what it needs
- Pass specific dependencies, not the whole application context:
  ```go
  // Bad: function has access to everything
  func Deploy(app *Application) error

  // Good: function has access to exactly what it needs
  func Deploy(pusher Pusher, target string) error
  ```
- Narrow interfaces are a form of least privilege — `io.Reader` over `*os.File`

## Anti-Corruption Layer

- When external systems use different domain language or data models, translate at the boundary
- Your domain types should not be shaped by an external API's response format:
  ```go
  // Adapter translates external model to domain model
  func (a *GitHubAdapter) ListDeployments(ctx context.Context) ([]deploy.Target, error) {
      resp, err := a.client.Repositories.ListDeployments(ctx, a.owner, a.repo, nil)
      if err != nil {
          return nil, fmt.Errorf("list github deployments: %w", err)
      }
      return toDeployTargets(resp), nil  // translate to domain type
  }
  ```

## Learning Tests

- When adopting a new third-party library, write tests that verify YOUR understanding of its behavior
- These tests serve as documentation and catch breaking changes on upgrade:
  ```go
  func TestYAMLMarshal_PreservesOrder(t *testing.T) {
      // Verify our assumption that yaml.Marshal preserves field order
      // If a library upgrade breaks this, we'll know immediately
  }
  ```
- Learning tests live alongside the adapter that uses the library

## Configuration Boundaries

- Environment variables, config files, and CLI flags are external inputs — validate at the boundary
- Parse and validate config into typed structs early; pass typed values inward
- Domain code never reads environment variables or config files directly:
  ```go
  // Bad: buried in domain code
  func Connect() (*DB, error) {
      host := os.Getenv("DB_HOST")  // hidden dependency

  // Good: injected from boundary
  func Connect(host string, port int) (*DB, error) {
  ```

## Spec Traceability

- Specs describe behavior in domain language, not library language
- Boundary wrappers ensure implementation matches domain language
- If a spec says "store the artifact," the domain code calls `store.Put()`, not `s3Client.PutObject()`
