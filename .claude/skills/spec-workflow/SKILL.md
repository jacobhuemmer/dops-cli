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

## Plans

Plans are implementation proposals that must be approved before work begins.

**Directory structure:**
```
plans/
├── 2026-03-22-deploy-command.md       # Active (in progress or awaiting approval)
├── 2026-03-22-config-loading.md       # Active
└── completed/
    ├── 2026-03-20-project-setup.md    # Approved and done
    └── 2026-03-21-cli-skeleton.md     # Approved and done
```

**Naming:** `YYYY-MM-DD-<short-description>.md` — datetime prefix followed by a kebab-case description.

**Lifecycle:**
1. Create the plan in `plans/` as an active plan
2. Present to user for review and approval
3. Only after explicit approval, begin implementation
4. When implementation is complete and verified, move the plan to `plans/completed/`

**Rules:**
- Only approved plans are moved to `plans/completed/` — rejected or abandoned plans are deleted
- A plan that needs revision stays in `plans/` until re-approved
- Never implement an unapproved plan

## Workflow

1. **Write the spec first** — before any code
2. **Create a plan** — write an implementation plan in `plans/YYYY-MM-DD-<description>.md`
3. **Get approval** — present the plan to the user; do not proceed without explicit approval
4. **Write tests from acceptance criteria** — each criterion becomes one or more test cases
5. **Implement to pass the tests** — the spec and approved plan define the boundary of work
6. **Verify** — all acceptance criteria checked off
7. **Complete the plan** — move the approved plan to `plans/completed/`

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
