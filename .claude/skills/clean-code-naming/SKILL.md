---
name: clean-code-naming
description: Go naming conventions and Clean Code naming principles. Use when creating, renaming, or reviewing names of functions, variables, types, constants, packages, or interfaces in Go code. Triggers on new type definitions, function signatures, variable declarations, and package naming.
user-invocable: false
---

# Clean Code Naming for Go

Apply these naming principles to all Go code. Go idioms take precedence over generic Clean Code advice.

## Package Names

- Short, lowercase, single-word names: `http`, `fmt`, `user`
- No underscores, no mixedCaps
- Name should describe what the package *provides*, not what it *contains*
- Avoid generic names: `util`, `common`, `helpers`, `misc` — find a more specific name
- The package name is part of the call site: `http.Get` not `httputil.HTTPGet`

## Interface Names

- Single-method interfaces use the method name + `er` suffix: `Reader`, `Writer`, `Closer`, `Formatter`
- Multi-method interfaces describe the capability: `ReadWriter`, `Handler`
- Do NOT prefix with `I` (no `IReader`) — this is not Go idiom
- Define interfaces where they are *consumed*, not where they are *implemented*

## Function and Method Names

- Use MixedCaps (exported) or mixedCaps (unexported)
- Start with a verb: `Get`, `Set`, `New`, `Create`, `Parse`, `Validate`
- Getters omit `Get`: use `Name()` not `GetName()` (Go convention)
- Setters use `Set`: `SetName()`
- Constructors use `New`: `NewServer`, `NewClient`
- Boolean-returning functions read as questions: `IsValid`, `HasPermission`, `CanRetry`
- The name should make the function's behavior obvious — if you need a comment to explain what it does, rename it

## Variable Names

- Short names in small scopes: `i`, `n`, `r`, `w`, `err`, `ctx` are idiomatic Go
- Longer, descriptive names in larger scopes or when the type isn't obvious
- Receivers are 1-2 letters, consistent across methods: `s` for `Server`, `c` for `Client`
- No Hungarian notation, no type prefixes (`strName`, `intCount`)
- Boolean variables read as predicates: `ok`, `found`, `done`, `valid`
- Avoid `data`, `info`, `temp`, `tmp`, `val` — they say nothing

## Constants

- MixedCaps like variables: `MaxRetries`, `DefaultTimeout`
- Not ALL_CAPS (that's not Go idiom, except for generated code)
- Group related constants with `iota` when sequential values make sense

## Type Names

- Nouns that describe what the type *is*: `Server`, `Request`, `Config`
- No `Type` suffix: `User` not `UserType`
- No `Struct` suffix: `Config` not `ConfigStruct`
- Avoid stutter with package name: `user.User` is fine, but `user.UserService` stutters — prefer `user.Service`

## Spec Traceability

- Names in code should mirror the domain language used in specs
- If the spec says "deployment target", the type is `DeploymentTarget`, not `Dest` or `Location`
- Consistent vocabulary between spec and code eliminates translation overhead

## What NOT to Do

- Do not rename well-established Go conventions (`err`, `ctx`, `ok`, `i`)
- Do not add comments to compensate for bad names — fix the name instead
- Do not use abbreviations unless they are universally understood in the domain
- Do not use `var` names that shadow builtin or imported names
