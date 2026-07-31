# Prospecção Brasil -- Documentation

This is the documentation hub for the Prospecção Brasil project. It is
organized for Obsidian navigation and covers topics NOT already handled by
the spec-driven development artifacts.

## Map of Contents

### Planning

- [PRD](planning/PRD.md) -- Product Requirements Document (vision, problem,
  ICP, MVP scope, roadmap, stack summary).

### Architecture

- [DECISIONS](architecture/DECISIONS.md) -- Architecture Decision Records
  (ADRs). Key technical decisions with context and rationale.

### Operations

- [RUNBOOK](operations/RUNBOOK.md) -- Production operations runbook (health
  checks, DB recovery, incident response). Stub until deployment is a goal.

## What Lives Where

| Topic | Location | Notes |
|-------|----------|-------|
| Feature specs (spec.md, plan.md, tasks.md) | `specs/` | Spec Kit lifecycle artifacts |
| AI agent rules, code conventions, quality gates | `AGENTS.md` | Read by all AI agents |
| Project principles (security, testing, migrations) | `.specify/memory/constitution.md` | Non-negotiable constraints |
| Database schema, query contracts | `specs/002-database-schema/` | SPEC-02 artifacts |
| Product vision, ICP, roadmap | `docs/planning/PRD.md` | This folder |
| Architecture decisions (ADRs) | `docs/architecture/DECISIONS.md` | This folder |
| Production ops runbook | `docs/operations/RUNBOOK.md` | This folder |

Do NOT duplicate spec content, AGENTS.md rules, or constitution principles
in this docs/ folder. Link to them instead.
