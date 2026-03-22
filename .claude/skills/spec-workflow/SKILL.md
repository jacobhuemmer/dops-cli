---
name: spec-workflow
description: Spec-driven development workflow. Use when creating new features, writing specs, planning implementation from specs, mapping specs to tests, or when the user references a spec file. Triggers on spec creation, implementation planning, or any mention of specs, requirements, or acceptance criteria.
user-invocable: false
---

# Spec-Driven Workflow

This project follows a spec-driven development process. Specs are the source of truth for what gets built.

## Spec Format

Specs are plain markdown files stored in `specs/`. Each spec describes one feature or capability.

Structure:

```markdown
# Feature Name

## Overview
One paragraph describing what this feature does and why it exists.

## Requirements
- Concrete, testable requirements
- Each requirement maps to one or more tests
- Use precise language — "must", "should", "may" per RFC 2119

## Acceptance Criteria
- [ ] Criterion 1 — specific, verifiable condition
- [ ] Criterion 2
- [ ] Criterion 3

## Error Cases
- What happens when X fails
- What the user sees when Y is invalid

## Out of Scope
- What this feature explicitly does NOT do (prevents scope creep)
```

## Workflow

1. **Write the spec first** — before any code
2. **Review the spec** — is it clear, complete, testable?
3. **Write tests from acceptance criteria** — each criterion becomes one or more test cases
4. **Implement to pass the tests** — the spec defines the boundary of work
5. **Verify** — all acceptance criteria checked off

## Spec-to-Code Traceability

- **Names** mirror spec language — if the spec says "deployment target", the type is `DeploymentTarget`
- **Tests** reference the spec — test names reflect acceptance criteria
- **Functions** map to requirements — if you can't trace a function to a spec, question it (YAGNI)
- **Error handling** matches spec error cases — every documented error scenario has a test

## Rules

- If it's not in the spec, don't build it
- If the spec is ambiguous, clarify it before coding — don't guess
- If implementation reveals a gap in the spec, update the spec first
- Specs are living documents — update them when requirements change
- Acceptance criteria are the definition of done

## Spec Review Checklist

Before implementing:
- [ ] Every requirement is testable (no vague language like "should be fast")
- [ ] Error cases are documented
- [ ] Out of scope is defined
- [ ] No implicit requirements — everything is explicit
- [ ] Naming is consistent with existing specs and codebase
