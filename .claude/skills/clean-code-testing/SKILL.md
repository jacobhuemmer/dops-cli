---
name: clean-code-testing
description: Clean Code testing principles for Go. Use when writing, editing, or reviewing Go test files (*_test.go). Triggers on test function creation, table-driven test setup, test helper design, mock/stub creation, or test refactoring.
user-invocable: false
---

# Clean Code Testing for Go

Apply these principles when writing or modifying Go tests.

## Tests Derive from Specs

- Every spec acceptance criterion becomes one or more test cases
- Name tests after the behavior, not the implementation:
  ```go
  // Bad
  func TestDeployFunction(t *testing.T)

  // Good
  func TestDeploy_FailsWhenTargetUnreachable(t *testing.T)
  ```
- If you can't map a test to a spec requirement, question whether the test (or the code it tests) is needed

## FIRST Principles

- **Fast**: Tests run quickly. No network calls in unit tests. Use interfaces and test doubles.
- **Independent**: Tests don't depend on each other or on execution order. No shared mutable state between tests.
- **Repeatable**: Same result every time, in every environment. No flaky tests. Control time, randomness, and external state.
- **Self-Validating**: Tests produce a boolean result — pass or fail. No manual inspection of output.
- **Timely**: Write tests before or alongside the code, not after. Tests written after tend to rationalize the implementation rather than verify the spec.

## Table-Driven Tests

- Use table-driven tests for related scenarios — this is idiomatic Go:
  ```go
  func TestValidateTarget(t *testing.T) {
      tests := []struct {
          name    string
          target  string
          wantErr bool
      }{
          {name: "valid hostname", target: "prod-01.example.com", wantErr: false},
          {name: "empty target", target: "", wantErr: true},
          {name: "invalid chars", target: "prod 01!", wantErr: true},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              err := ValidateTarget(tt.target)
              if (err != nil) != tt.wantErr {
                  t.Errorf("ValidateTarget(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
              }
          })
      }
  }
  ```
- Each table entry should have a descriptive `name` field
- Keep the test logic in the loop minimal — complexity belongs in the table entries

## One Concept per Test

- A test function verifies one behavioral concept
- Multiple assertions are fine if they all verify the same concept
- If a test name has "and" in it, consider splitting it

## Test Helpers

- Use `t.Helper()` in helper functions so failures report the caller's line number
- Put shared test utilities in a `testutil` package or `_test.go` files
- Test helpers should not use `t.Fatal` unless the failure makes continuing meaningless

## Test Doubles (Go Style)

- Prefer interfaces + simple stub implementations over mocking frameworks
- Fakes are better than mocks for complex behavior:
  ```go
  type fakePusher struct {
      pushed []string
      err    error
  }

  func (f *fakePusher) Push(target string) error {
      f.pushed = append(f.pushed, target)
      return f.err
  }
  ```
- Only mock at boundaries (external services, filesystem, clock) — not internal types
- If you need to mock something deep inside, the design needs refactoring (DIP violation)

## Test Readability

- Tests are documentation — a new developer should understand the feature by reading the test
- Follow Arrange-Act-Assert (Given-When-Then):
  ```go
  func TestDeploy_SucceedsWithValidTarget(t *testing.T) {
      // Arrange
      pusher := &fakePusher{}
      target := "prod-01"

      // Act
      err := Deploy(pusher, target)

      // Assert
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(pusher.pushed) != 1 || pusher.pushed[0] != target {
          t.Errorf("expected push to %q, got %v", target, pusher.pushed)
      }
  }
  ```
- Don't DRY tests at the cost of readability — some duplication in tests is fine
- Use `testdata/` directory for fixture files

## What NOT to Do

- Do not test private functions directly — test through the public API
- Do not write tests that test the Go standard library or third-party libraries
- Do not use `init()` in test files
- Do not skip flaky tests with `t.Skip` — fix the flakiness
- Do not use `time.Sleep` in tests — use channels, waitgroups, or fake clocks
