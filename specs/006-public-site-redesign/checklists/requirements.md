# Requirements Checklist: Public Site Redesign

**Feature**: SPEC-06
**Created**: 2026-07-31
**Purpose**: Validate that spec.md requirements are complete, clear, consistent, measurable, and cover all scenarios -- BEFORE implementation begins.

---

## Requirement Completeness

- [ ] CHK001: Todas as 5 paginas publicas (Home, Servicos, Quem Somos, Nossos Clientes, Fale Conosco) tem user stories com Gherkin acceptance scenarios?
- [ ] CHK002: Cada user story tem prioridade atribuida (P1-P6) e justificativa de prioridade?
- [ ] CHK003: Cada user story tem "Independent Test" descrito (como testar isoladamente)?
- [ ] CHK004: A rota nova `/servicos/{slug}` esta explicitamente listada como requisito funcional (FR-026)?
- [ ] CHK005: O requisito de Pencil designs antes de HTML/CSS esta documentado (FR-027)?
- [ ] CHK006: A correcao CSS dos 3 templates de auth (login, totp_setup, totp_verify) tem user story dedicada (US6)?
- [ ] CHK007: A proibicao de copy de software ("carga cognitiva", "pipeline", "plataforma") esta como requisito funcional (FR-025)?
- [ ] CHK008: O fallback de imagens (gradiente/Deep Navy) esta documentado como edge case?
- [ ] CHK009: O comportamento com depoimentos insuficientes esta documentado como edge case?
- [ ] CHK010: O comportamento com logos ausentes esta documentado como edge case?
- [ ] CHK011: O fallback no-JS do formulario de contato esta documentado como edge case (nao quebrar SPEC-04)?
- [ ] CHK012: O requisito de mobile (375px sem scroll horizontal) esta documentado (FR-023, SC-008)?
- [ ] CHK013: As metricas com valor zero ou indefinido tem edge case documentado?
- [ ] CHK014: A separacao de host (public vs sistema.*) esta mencionada como assumption?

## Requirement Clarity

- [ ] CHK015: "Hero full-bleed com fotografia" esta definido sem ambiguidade (imagem cobre largura total, overlay para legibilidade do texto)?
- [ ] CHK016: "Primeiro viewport premium" (SC-003) e mensuravel? O spec define o que constitui "premium" (hero com foto, headline de mercado, metricas, CTA)?
- [ ] CHK017: "Paginas profundas de servicos" (US2) esta definido: o que constitui "profundo" (metodologia, etapas, CTA -- FR-007)?
- [ ] CHK018: "Apresentacao visual premium" do Fale Conosco (US5) e distinguivel do estado atual? O spec diz o que muda (info de contato em destaque na pagina, nao so footer)?
- [ ] CHK019: "Classes do design system" para auth (FR-018) lista exemplos concretos (form-input, btn, btn-primary, card, alert)?
- [ ] CHK020: O slug dos 5 servicos minimos esta listado no spec (FR-006) para evitar ambiguidade na implementacao?

## Requirement Consistency

- [ ] CHK021: FR-024 (copy pt-BR de consultoria) e consistente com FR-025 (proibicao de copy de software) e Non-Goals (nao vender software no dominio publico)?
- [ ] CHK022: FR-022 (manter brand tokens) e consistente com Non-Goals (nao reescrever paleta) e Assumptions (design system reutilizado)?
- [ ] CHK023: FR-015 (manter comportamento do formulario) e consistente com FR-016 (adicionar info de contato em destaque) -- um nao contradiz o outro?
- [ ] CHK024: FR-019 (funcionalidade auth identica) e consistente com FR-017/FR-018 (apenas CSS muda)?
- [ ] CHK025: US5 (Fale Conosco visual) e US6 (auth CSS) nao tem dependencias circulares nem conflito de arquivos?

## Requirement Measurability

- [ ] CHK026: SC-001 (tempo de carga < 1s) e mensuravel com ferramenta especifica (Lighthouse ou DevTools)?
- [ ] CHK027: SC-002 (Lighthouse score >= 90) e mensuravel e tem categorias definidas (Best Practices, Accessibility)?
- [ ] CHK028: SC-004 (ausencia de palavras proibidas) e mensuravel via grep nos templates?
- [ ] CHK029: SC-006 (make check passa) e mensuravel com comando concreto?
- [ ] CHK030: SC-008 (375px sem scroll horizontal) e mensuravel com DevTools ou teste automatizado?

## Requirement Coverage

- [ ] CHK031: O spec cobre o que acontece quando uma imagem de hero nao carrega (edge case)?
- [ ] CHK032: O spec cobre o que acontece quando o visitante acessa `/servicos/{slug}` com slug inexistente (FR-008, edge case)?
- [ ] CHK033: O spec cobre o comportamento do formulario de contato com JavaScript desabilitado (edge case, FR-015)?
- [ ] CHK034: O spec cobre o comportamento das paginas de auth em mobile (FR-023 se aplica a auth)?
- [ ] CHK035: O spec cobre a interacao entre o redesign publico e o sistema interno (assumption: nao afeta)?
- [ ] CHK036: O spec menciona que o footer e nav serao atualizados (FR-028) sem reescrever a estrutura?
- [ ] CHK037: O spec menciona que a newsletter permanece funcional (FR-029)?
- [ ] CHK038: O spec define onde as imagens sao armazenadas (FR-030, `static/img/`)?

## Boundary & Scope Validation

- [ ] CHK039: Non-Goals inclui explicitamente "nao automatizar deploy no VPS"?
- [ ] CHK040: Non-Goals inclui explicitamente "nao vender software no dominio publico"?
- [ ] CHK041: Non-Goals inclui explicitamente "nao reescrever paleta de marca"?
- [ ] CHK042: Non-Goals inclui explicitamente "nao criar SPA ou React"?
- [ ] CHK043: Non-Goals inclui explicitamente "nao adicionar depoimentos dinamicos (banco)"?
- [ ] CHK044: Non-Goals inclui explicitamente "nao mudar funcionalidade do formulario de contato"?
- [ ] CHK045: Non-Goals inclui explicitamente "nao mudar funcionalidade de autenticacao"?
