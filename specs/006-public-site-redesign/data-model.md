# Data Model: Public Site Redesign

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

---

## Overview

SPEC-06 e primariamente redesign visual (HTML/CSS/templates). O unico
cambio no schema do banco e adicionar a coluna `company` (opcional) a
tabela `contact_submissions` para o novo campo "Empresa" no formulario
de Fale Conosco.

Nao ha novas tabelas. Nao ha novas entidades de banco. Servicos,
depoimentos e metricas sao estaticos em Go (maps/slices), nao no banco.

---

## Migration: 000003_add_company_to_contact_submissions

```sql
-- SPEC-06: Add optional company field to contact_submissions
-- for the "Empresa" field in the redesigned Fale Conosco form.
ALTER TABLE contact_submissions
    ADD COLUMN company VARCHAR(255) NULL;
```

**Down migration**:
```sql
-- SPEC-06: Remove company column from contact_submissions.
ALTER TABLE contact_submissions
    DROP COLUMN IF EXISTS company;
```

**Rationale**: O formulario do legacy site tem campo "Empresa". O SPEC-04
nao incluiu. O SPEC-06 adiciona como opcional (NULL permitido) para nao
quebrar submissoes existentes.

---

## Entities (Static, Not in Database)

### Service

```go
type serviceDetail struct {
    Slug        string   // URL path: /servicos/{slug}
    Title       string   // Display title
    Summary     string   // Short description for index card
    Description string   // Long description for detail page
    Methodology []string // List of methodology steps
    CTA         string   // CTA text (default: "Fale com um especialista")
}
```

**Instances** (5 minimos, estaticos em Go):
1. `expansao-de-redes` -- Expansao de Redes
2. `built-to-suit` -- Built to Suit (BTS)
3. `strip-mall` -- Strip Mall / Centros de Conveniencia
4. `lajes-comerciais` -- Lajes Comerciais
5. `prospeccao-de-ponto` -- Prospecacao de Ponto Comercial

### Testimonial

```go
type testimonial struct {
    Name    string // Author name
    Company string // Author company/role
    Quote   string // Testimonial text
    Metric  string // Optional metric badge
}
```

**Instances** (3, do legacy site, estaticos em Go):
1. Larissa Mello -- "A Prospecacao Brasil selecionou os melhores pontos comerciais..."
2. Roberto Andrade -- "A empresa tem nos ajudado a encontrar novos parceiros..."
3. Joao Viana -- "A Prospecacao Brasil tem fornecido informacoes preciosas..."

### Metric

```go
type metric struct {
    Label string // "Pontos Comercializados"
    Value string // "100+" or "0"
}
```

**Instances** (4, estaticos em Go):
1. Pontos Comercializados
2. Clientes Atendidos
3. Cidades Atendidas
4. Anos de Mercado

---

## Existing Tables (Unchanged)

### contact_submissions (SPEC-04, +1 column)

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| name | VARCHAR(255) | NOT NULL | |
| email | VARCHAR(255) | NOT NULL | |
| phone | VARCHAR(20) | NULL | |
| **company** | **VARCHAR(255)** | **NULL** | **NEW (SPEC-06)** |
| subject | VARCHAR(255) | NOT NULL | |
| message | TEXT | NOT NULL | |
| status | VARCHAR(20) | NOT NULL DEFAULT 'new', CHECK in ('new','read','archived') | |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() | |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() | |

### newsletter_subscribers (SPEC-04, unchanged)

Nenhuma alteracao.

---

## sqlc Changes

### queries/contacts.sql (adicionar coluna company)

A query `CreateContactSubmission` precisa aceitar o parametro `company`
(opcional, nullable). Regenerar com `make sqlc`.

```sql
-- name: CreateContactSubmission :one
INSERT INTO contact_submissions (name, email, phone, company, subject, message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
```

A query `ListContactSubmissions` ja usa `SELECT *` e automaticamente
inclui a nova coluna apos regenerar o sqlc.
