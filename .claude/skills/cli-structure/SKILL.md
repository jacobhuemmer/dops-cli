---
name: cli-structure
description: Go CLI project structure with Cobra and BubbleTea. Use when creating new commands, packages, or modules, organizing project layout, wiring dependencies, setting up Cobra commands, or deciding where code should live. Triggers on new file/package creation, cmd/ changes, internal/ organization, or main.go modifications.
user-invocable: false
---

# CLI Project Structure

This project uses Cobra for CLI routing and BubbleTea v2 for interactive TUI.

## Directory Layout

```
dops/
├── main.go              # Minimal — calls cmd.Execute()
├── cmd/
│   ├── root.go          # Root command, persistent flags, initConfig
│   ├── version.go       # Non-TUI commands
│   └── <command>.go     # One file per subcommand
├── internal/
│   ├── tui/             # BubbleTea models, views, styles
│   │   ├── model.go
│   │   ├── styles.go
│   │   └── ...
│   ├── <domain>/        # Business logic packages
│   └── adapters/        # External service wrappers
├── specs/               # Spec markdown files
├── .pre-commit-config.yaml
└── LICENSE
```

## main.go

Keep it minimal:
```go
package main

import "dops/cmd"

func main() {
    cmd.Execute()
}
```

## Cobra Command Pattern

One command per file in `cmd/`:

```go
package cmd

import "github.com/spf13/cobra"

var deployCmd = &cobra.Command{
    Use:   "deploy [target]",
    Short: "Deploy to a target environment",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Wire dependencies, launch TUI or execute
        return nil
    },
}

func init() {
    rootCmd.AddCommand(deployCmd)
    deployCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview without applying")
}
```

Rules:
- Always use `RunE` (returns error), never `Run`
- Use `Args` validators: `NoArgs`, `ExactArgs(n)`, `MinimumNArgs(n)`, `RangeArgs(min, max)`
- Commands without `RunE` act as groups (print help)

## Cobra + BubbleTea Handoff

Cobra parses flags and args, then hands off to BubbleTea:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // 1. Resolve config from flags/viper
    cfg := resolveConfig(cmd)

    // 2. Create TUI model with injected dependencies
    model := tui.NewModel(cfg)

    // 3. Launch BubbleTea
    p := tea.NewProgram(model)
    _, err := p.Run()
    return err
}
```

Not every command needs a TUI. Simple commands (version, config, completion) use plain stdout.

## Flag Handling

Persistent flags (all subcommands):
```go
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "verbosity (-v, -vv, -vvv)")
```

Flag groups:
```go
cmd.MarkFlagsRequiredTogether("username", "password")
cmd.MarkFlagsMutuallyExclusive("json", "yaml")
cmd.MarkFlagsOneRequired("json", "yaml")
```

## PreRun Hooks

Execution order: `PersistentPreRunE` → `PreRunE` → `RunE` → `PostRunE` → `PersistentPostRunE`

```go
var rootCmd = &cobra.Command{
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initLogging()  // runs before ALL commands
    },
}
```

## Dependency Wiring

Wire dependencies in `cmd/` or a dedicated `wire.go`, NOT in business logic:

```go
// cmd/deploy.go — composition root
RunE: func(cmd *cobra.Command, args []string) error {
    store := awsstore.New(cfg.Region)     // concrete adapter
    deployer := deploy.New(store)          // inject interface
    return deployer.Run(cmd.Context())
}
```

Business logic in `internal/` accepts interfaces, never concrete external types.

## Package Boundaries

- `cmd/` — CLI wiring only. Parses flags, creates dependencies, calls business logic.
- `internal/tui/` — BubbleTea models, views, styles. Depends on business logic interfaces.
- `internal/<domain>/` — Pure business logic. No CLI, no TUI, no external imports.
- `internal/adapters/` — Thin wrappers around external services implementing domain interfaces.

## Rules

- `main.go` does one thing: calls `cmd.Execute()`
- Never import `cmd/` from `internal/`
- Never read environment variables or flags in `internal/` — inject values
- Each subcommand file registers itself via `init()` → `rootCmd.AddCommand()`
- Keep `cmd/` files focused on wiring — no business logic
