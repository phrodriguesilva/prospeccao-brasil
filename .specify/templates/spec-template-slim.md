# Feature Specification (Slim): [FEATURE NAME]

**Feature Branch**: `[###-feature-name]`

**Created**: [DATE]

**Status**: Draft

**Template**: slim (for infrastructure/tooling specs). See AGENTS.md
"Spec template selection" for when to use this vs the full template.

**Input**: User description: "$ARGUMENTS"

## Overview

[1-2 paragraphs: what this spec is and why it is a prerequisite for other specs.
Infrastructure specs typically do not deliver product value to lawyers directly;
they deliver engineering value (reproducibility, security, performance, deploy).]

## Context

**Canonical sources:**
- [PRD](../../docs/planning/PRD.md) section [X]
- [PROJECT_SETUP](../../docs/PROJECT_SETUP.md) section [X]
- [EXECUTION_PLAN](../../docs/planning/EXECUTION_PLAN.md) [Phase / SPEC-XX]
- [Constitution](../../.specify/memory/constitution.md)

**Dependencies**: [SPEC-XX, ...] or "None (foundation spec)".
**Gate to run**: [what must be true before this spec runs].

## Goals

1. [Goal 1 -- engineering outcome, not user story]
2. [Goal 2]
3. [Goal 3]

## Non-Goals

The following are explicitly deferred to later specs:

- [Item 1] -> SPEC-XX
- [Item 2] -> SPEC-XX

## Requirements

These requirements are the verifiable acceptance criteria. Each is copied from
the EXECUTION_PLAN and is non-negotiable.

- **FR-001**: [verifiable requirement with a command]
- **FR-002**: [verifiable requirement with a command]
- **FR-003**: [verifiable requirement with a command]

## Constraints

[Copy the constraints that apply to this spec from PROJECT_SETUP and the
constitution. Only include constraints relevant to this spec's scope; do not
boilerplate all 12 constraints if only 3 apply.]

1. [Constraint 1]
2. [Constraint 2]
3. [Constraint 3]

## Definition of Done

[Feature] is done when ALL of the following are verifiable:

| # | Acceptance Criterion | Verification Command | Status |
|---|----------------------|----------------------|--------|
| 1 | [criterion] | `command` | [ ] |
| 2 | [criterion] | `command` | [ ] |
| 3 | [criterion] | `command` | [ ] |

**Spec is ready for `/speckit-plan` when all rows are checked.**
