# PROMPT DE EXECUCAO -- Redesign Completo do Prospeccao Brasil

> Este arquivo e um prompt autocontido para um agente de IA executar. Contém
> toda a informacao necessaria sobre o estado atual, as decisoes de design, e
> as tarefas concretas. **Leia por completo antes de executar qualquer tarefa.**

## Contexto do Projeto

Prospeccao Brasil e uma plataforma de prospeccao imobiliaria comercial (Go +
HTMX + Alpine.js + Tailwind + Postgres). Tem duas faces:

1. **Site publico institucional** -- vitrine B2B para varejistas, investidores e
   proprietarios de imoveis comerciais.
2. **Sistema interno (admin)** -- gestao de imoveis, clientes e prospeccoes
   para o administrador (Luiz Claudio).

Uma auditoria de design foi realizada comparando o projeto contra 5 sites de
referencia do mercado (PLDA, RR Negocios, Ocupantes, Amplitude RE, Grupo Sinop).
O relatorio completo esta em `docs/auditoria_design_completa.md`.

**Nota geral atual: 6.6/10.** O site nao e brego, mas falta consistencia,
profundidade de conteudo e autoridade institucional.

---

## Restricoes Absolutas (NAO QUEBRAR)

- **Stack**: Go + html/template + HTMX + Alpine.js + Tailwind (build-time, NAO CDN)
- **NAO mudar**: Go handlers, routes, data structures, middleware, auth
- **Mudar APENAS**: `input.css`, `tailwind.config.js`, templates HTML em
  `internal/template/`, `static/js/`, `static/img/`
- **Self-host**: todas as fotos e assets em `static/` (nunca CDN externo)
- **Sem emojis**: em nenhum lugar -- codigo, UI, comentarios, commits
- **Apos mudancas CSS**: rodar `npx tailwindcss -i input.css -o static/css/app.css --minify`
- **Apos mudancas templates**: rodar `go build ./cmd/prospeccao`
- **Validacao final**: rodar `make check` (lint + test + build-css + build + ast-grep)
- **Nao fazer deploy** -- apenas codigo. O usuario cuida do deploy.

---

## Decisoes de Design Ja Tomadas

### Identidade Visual (MANTER)
- **Paleta principal**: Deep Navy `#031636` + Sobrio Gold `#765a1a` / `#d4af6a`
- **Surface**: Warm off-white `#fcf9f8` (fundo do site publico)
- **Tipografia**: Montserrat (display/headlines) + Inter (body/UI)
- **Formato botoes publico**: Pill shape `rounded-full`
- **Nav publica**: Glass floating pill (`.glass-nav`) -- ja implementada e SUPERIOR aos concorrentes
- **Hero**: Full-bleed Ken Burns com foto background -- manter e refinar
- **WhatsApp float**: ja implementado -- manter
- **Scroll reveal**: ja implementado via IntersectionObserver -- manter
- **Back-to-top + Scroll progress**: ja implementados -- manter

### Referencia Principal de Design
- **PLDA** (https://www.plda.com.br/) -- estrutura de conteudo, segmentacao por stakeholder, vocabulario tecnico B2B
- **Ocupantes** (https://www.ocupantes.com.br/) -- tom corporativo multinacional, "10 razoes para contratar"
- **Amplitude RE** (https://amplitudere.com.br/) -- estetica ultra-premium, boutique financeira

---

## INVENTARIO COMPLETO DE ARQUIVOS

### Templates Publicos (7 paginas)
```
internal/template/home.html                    -- Homepage
internal/template/quem-somos.html              -- Sobre / Institucional
internal/template/servicos.html                -- Indice de servicos
internal/template/servico-detalhe.html         -- Detalhe de cada servico
internal/template/nossos-clientes.html         -- Social proof / clientes
internal/template/fale-conosco.html            -- Contato
internal/template/404.html                     -- Erro 404
```

### Partials Publicos (6 arquivos)
```
internal/template/partials/nav.html            -- Glass nav flutuante
internal/template/partials/footer.html         -- Footer 4 colunas navy
internal/template/partials/site_gate.html      -- Modal area restrita
internal/template/partials/brand_icons.html    -- Favicons e meta tags
internal/template/partials/floating_elements.html -- Scroll progress, back-to-top, WhatsApp
internal/template/partials/metrics.html        -- Strip de metricas
```

### Fragments HTMX (4 arquivos)
```
internal/template/fragments/contact_success.html
internal/template/fragments/contact_error.html
internal/template/fragments/newsletter_success.html
internal/template/fragments/newsletter_error.html
```

### Auth (3 templates)
```
internal/template/login.html
internal/template/totp_setup.html
internal/template/totp_verify.html
```

### Admin System (15 templates)
```
internal/template/admin/_layout.html
internal/template/admin/dashboard.html
internal/template/partials/internal_nav.html
internal/template/admin/properties/list.html
internal/template/admin/properties/form.html
internal/template/admin/properties/detail.html
internal/template/admin/properties/pdf.html
internal/template/admin/clients/list.html
internal/template/admin/clients/form.html
internal/template/admin/clients/detail.html
internal/template/admin/prospections/list.html
internal/template/admin/prospections/form.html
internal/template/admin/prospections/detail.html
internal/template/admin/contacts/_form.html
internal/template/admin/contacts/_log.html
```

### CSS e Config
```
input.css                                      -- 754 linhas, source CSS
tailwind.config.js                             -- 158 linhas, design tokens
static/css/app.css                             -- output compilado
```

### Assets
```
static/img/logo-symbol.png                    -- Marca simbolo (nav, login)
static/img/logo-full.png                      -- Logo completo com texto
static/img/whatsapp-icon.svg                  -- Icone WhatsApp
static/img/hero-comercial.jpg                 -- Hero principal (235KB)
static/img/about-founder.jpg                  -- Background quem somos (109KB)
static/img/cta-bg.jpg                         -- Skyline para CTAs (274KB)
static/img/service-expansao.jpg               -- Foto servico (98KB)
static/img/service-bts.jpg                    -- Foto servico (52KB)
static/img/service-strip-mall.jpg             -- Foto servico (96KB)
static/img/service-lajes.jpg                  -- Foto servico (57KB)
static/img/service-prospeccao.jpg             -- Foto servico (94KB)
static/img/favicon-32.png
static/img/apple-touch-icon.png
static/img/icon-192.png
```

---

## FASE 1: UNIFICACAO DO DESIGN SYSTEM (Fundacao -- Fazer Primeiro)

**Objetivo:** Eliminar a fragmentacao entre site publico (warm) e admin (cool).
Criar um design system unico e coerente que ambos os lados compartilhem.

### 1.1 Migrar hex hardcoded para tokens Tailwind

Todos os valores hardcoded abaixo DEVEM ser substituidos por classes Tailwind
que referenciam os tokens em `tailwind.config.js`:

| Hex hardcoded | Token Tailwind equivalente | Onde aparece |
|:-------------|:--------------------------|:-------------|
| `#765a1a` | `text-secondary` ou `bg-secondary` | `input.css` (linhas 259, 267, 279, 295, 383, 621, 680, 693, 701, 751), templates inline styles |
| `#d4af6a` | Criar token `secondary.light: "#d4af6a"` no tailwind.config.js | `input.css` (linhas 259, 283, 530, 621, 645, 723, 737, 751) |
| `#f5e6c8` | Criar token `secondary.lightest: "#f5e6c8"` | `input.css` (linha 283) |
| `#031636` | `bg-primary` ou `text-primary` | `input.css` (linhas 295, 307, 364, 637), templates inline styles |
| `#051f47` | `primary.container` ja existe como `#1a2b4c` -- usar esse ou criar | `input.css` (linhas 326, 364) |
| `#25d366` | `bg-whatsapp-green` | `input.css` (linha 340) |
| `#cbd5e1` | `border-outline-variant` ou criar `neutral.border` | `input.css` (linha 693) |
| `#1e293b` | `text-on-surface` ou criar alias | `input.css` (linha 696) |
| `#94a3b8` | `text-outline` | `input.css` (linha 704) |
| `#e2e8f0` | `border-surface-variant` | `input.css` (linhas 250, 547, 580) |
| `#475569` | `text-on-surface-variant` | `input.css` (linha 470) |

**Acao concreta:**
1. Adicionar em `tailwind.config.js` > `colors` > `secondary`:
   ```js
   light: "#d4af6a",
   lightest: "#f5e6c8",
   ```
2. Adicionar em `tailwind.config.js` > `colors`:
   ```js
   neutral: {
     border: "#e2e8f0",
     "border-hover": "#cbd5e1",
     muted: "#94a3b8",
     text: "#1e293b",
   },
   ```
3. Substituir TODOS os `style="color: #xxx"` e `style="background: #xxx"` nos
   templates por classes Tailwind (`text-secondary`, `bg-primary`, etc.)
4. Em `input.css`, substituir hex hardcoded por `theme('colors.secondary.DEFAULT')`,
   `theme('colors.primary.DEFAULT')`, etc. usando a funcao `theme()` do Tailwind.

### 1.2 Unificar a escala tipografica

O token `display-lg` e `48px` no tailwind.config.js, mas os heroes usam
`text-5xl md:text-6xl lg:text-7xl` (48px -> 60px -> 72px). Resolver assim:

**Acao:** Atualizar `tailwind.config.js` com escala responsiva:
```js
fontSize: {
  // Display responsivo (hero headlines)
  "display-lg": ["clamp(2.5rem, 5vw, 4.5rem)", {
    lineHeight: "1.1",
    fontWeight: "800",
    letterSpacing: "-0.02em"
  }],
  // ... manter o resto
}
```

E nos templates, usar `text-display-lg font-display` em vez de
`text-5xl md:text-6xl lg:text-7xl font-extrabold`.

### 1.3 Unificar componentes (publico + admin compartilham a mesma base)

Refatorar `input.css` para que cada componente tenha UMA classe base com
variantes via modificadores. Estrutura alvo:

```css
/* === CARDS === */
.card {
  /* Base compartilhada */
  @apply bg-white rounded-xl overflow-hidden border transition-all duration-300;
  border-color: theme('colors.neutral.border');
}
.card--elevated {
  /* Admin: sombra sutil, sem hover lift */
  @apply shadow-sm;
}
.card--premium {
  /* Publico: hover lift + gold top bar animada */
  @apply shadow-sm;
  /* animacao gold bar no ::before */
}
.card--glass {
  /* Publico: formulario glass */
  backdrop-filter: blur(12px);
  background: rgba(255, 255, 255, 0.9);
}

/* === BOTOES === */
.btn {
  /* Base compartilhada: inline-flex, center, gap, font, transition, focus ring */
  @apply inline-flex items-center justify-center gap-2 font-sans font-semibold
         transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2
         disabled:opacity-50 disabled:cursor-not-allowed;
}
.btn--pill { @apply rounded-full; }     /* Publico */
.btn--rounded { @apply rounded-lg; }    /* Admin */
.btn--primary { @apply bg-primary text-on-primary hover:bg-primary-container focus:ring-primary; }
.btn--secondary { @apply bg-secondary text-on-secondary hover:bg-secondary-container focus:ring-secondary; }
.btn--ghost { @apply bg-transparent text-primary hover:bg-surface-container focus:ring-primary; }
.btn--outline { @apply border border-primary text-primary bg-transparent hover:bg-primary hover:text-on-primary; }
.btn--error { @apply bg-error text-on-error hover:bg-error-container focus:ring-error; }
.btn--glow {
  box-shadow: 0 4px 20px theme('colors.secondary.DEFAULT' / 0.3);
}
.btn--sm { @apply px-3 py-1.5 text-sm; }
.btn--md { @apply px-5 py-2.5 text-body-md; }
.btn--lg { @apply px-7 py-3 text-body-lg; }

/* === INPUTS === */
.input {
  /* Base compartilhada */
  @apply w-full text-body-md text-on-surface
         focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed
         placeholder:text-outline transition-all duration-200;
}
.input--bordered {
  /* Admin: full border box */
  @apply rounded-lg border px-3.5 py-2.5 bg-white
         focus:ring-2 focus:ring-primary focus:border-primary;
  border-color: theme('colors.neutral.border');
}
.input--premium {
  /* Publico: bottom border only */
  @apply bg-transparent border-b-2 px-1 py-3;
  border-color: theme('colors.neutral.border-hover');
}
.input--premium:focus {
  border-bottom-color: theme('colors.secondary.DEFAULT');
}
```

**IMPORTANTE:** Apos essa refatoracao, atualizar TODOS os templates que
usam as classes antigas (`.premium-card`, `.btn-primary`, `.btn-filled`,
`.btn-tonal`, `.form-input`, `.form-input-premium`, `.form-glass`).
Fazer busca global por cada classe e substituir.

### 1.4 Alinhar visual do admin com a marca

O admin atualmente usa `bg-slate-50` e `border-slate-200` que sao frios e
genericos. Migrar para os tokens warm da marca:

| Admin atual | Substituir por |
|:-----------|:---------------|
| `bg-slate-50` | `bg-surface` (`#fcf9f8`) |
| `bg-white` (cards) | `bg-surface-container-lowest` |
| `border-slate-200` | `border-outline-variant` |
| `border-slate-300` (inputs) | `border-outline` |
| `text-slate-600` (headers tabela) | `text-on-surface-variant` |
| `text-slate-700` (labels) | `text-on-surface-variant` |
| `text-slate-800` (body) | `text-on-surface` |
| `text-slate-900` (titles) | `text-on-surface` |
| `bg-slate-100` (btn tonal) | `bg-surface-container-low` |
| `hover:bg-slate-50` (table rows) | `hover:bg-surface-container-low` |

**Atualizar em:**
- `internal/template/admin/_layout.html` -- `.admin-shell` background
- `internal/template/partials/internal_nav.html` -- ja esta navy, manter
- Todos os templates admin (`dashboard.html`, `properties/*.html`, `clients/*.html`, `prospections/*.html`)
- `input.css`: classes `.admin-shell`, `.filter-bar`, `.data-table`, `.form-input`, `.form-label`, `.card`

**Adicionar accent gold no admin:**
- CTAs principais do admin (criar imovel, criar cliente) devem usar
  `btn--secondary` (gold) em vez de `btn--primary` (navy) para destaque
- Active state na sidebar: adicionar borda esquerda gold `border-l-2 border-secondary-light`

### 1.5 Corrigir bug do contact_error.html

O fragment `contact_error.html` renderiza `.card` + `.form-input` quando o
formulario original e `.form-glass` + `.form-input-premium`. Apos o swap HTMX
de erro, o visual quebra.

**Acao:** Abrir `internal/template/fragments/contact_error.html` e substituir
as classes do form re-renderizado para manter `.form-glass` e os inputs com
a variante premium (`.input--premium` na nomenclatura nova).

---

## FASE 2: CONTEUDO E AUTORIDADE DO SITE PUBLICO

**Objetivo:** Alinhar o conteudo com o padrao do mercado (PLDA/Ocupantes).
Nao e redesign visual -- e enriquecimento de conteudo nas estruturas existentes.

### 2.1 Top Contact Bar (barra superior de contato)

Adicionar uma barra fina ACIMA da glass nav com:
- Telefone: `(XX) XXXXX-XXXX` (pegar do fale-conosco.html existente)
- Email: `contato@prospeccaobrasil.com` (pegar do fale-conosco.html)
- CRECI: `CRECI XX XXXXX` (pegar do footer ou quem-somos.html)
- Links sociais: Instagram, LinkedIn (icones SVG pequenos)

**Arquivo:** `internal/template/partials/nav.html`
**Posicao:** Antes do `<nav class="glass-nav">`, fixado no topo
**Estilo:** Background navy `bg-primary`, texto branco `text-on-primary/80`,
altura `28px`, font-size `12px`, z-index acima da glass nav.
**Responsivo:** Ocultar em mobile (`hidden md:flex`)
**Ajustar:** `body-nav-offset` padding de 80px para ~110px para compensar

### 2.2 Hero -- Proposta de Valor Direta

Atualizar o headline do hero em `internal/template/home.html`:

**Eyebrow:** "RETAIL & REAL ESTATE" (manter estilo atual)
**Headline:** Reescrever para algo direto e especifico como:
- "Conectamos redes varejistas aos melhores pontos comerciais do Brasil"
- ou "Inteligencia imobiliaria para expansao de redes varejistas"

O headline atual pode ser generico demais. Verificar o texto existente e
ajustar para que responda "O que voce faz e para quem?" em uma frase.

**Subheadline:** Adicionar credenciais inline:
"Prospeccao, estruturacao e comercializacao de pontos comerciais. XX anos
no mercado, XX+ cidades, XX+ clientes."

### 2.3 Reestruturar Navegacao

Mudar os links de navegacao de generico para segmentado por stakeholder:

**Atual:** Inicio | Quem Somos | Servicos | Nossos Clientes | Fale Conosco
**Novo:** Inicio | Servicos (dropdown) | Para Investidores | Sobre Nos | Contato

O dropdown de Servicos deve listar:
- Expansao de Redes Varejistas
- Built to Suit
- Prospeccao de Ponto Comercial
- Strip Mall
- Lajes Corporativas

**Arquivo:** `internal/template/partials/nav.html`
**Nota:** O dropdown precisa funcionar com Alpine.js (x-data, x-show, @click.away)
dentro da glass nav pill. Em mobile, expandir inline.

### 2.4 Grid de Logos de Clientes

Adicionar na homepage (apos servicos, antes do CTA) uma secao com logos
de clientes/marcas atendidas em grid horizontal.

Se nao houver logos reais disponiveis, criar a secao com placeholder de
nomes de empresas estilizados em texto (Montserrat, font-weight: 700,
color: `text-on-surface-variant`, opacity 0.5) em grid 3x2 ou 4x2.

**Titulo da secao:** "Marcas que Confiam em Nos" ou "Nossos Parceiros"
**Estilo:** Fundo `bg-surface-container-low`, logos/nomes em grayscale,
hover para cor original (se forem imagens).

### 2.5 Secao "Por que nos contratar?"

Adicionar em `quem-somos.html` ou como secao na homepage:
- 4-6 diferenciais numerados em cards premium
- Cada card com numero gold grande + titulo + descricao curta
- Exemplos de diferenciais:
  1. "Conhecimento profundo do mercado varejista brasileiro"
  2. "Rede de contatos com as principais redes do pais"
  3. "Estruturacao juridica e financeira de contratos atipicos"
  4. "Atuacao em mais de 100 cidades"
  5. "Avaliacao estrategica com foco em rentabilidade"
  6. "Acompanhamento completo do processo"

### 2.6 Pagina LGPD / Politica de Privacidade

Criar novo template `internal/template/privacidade.html` com:
- Texto padrao de politica de privacidade
- Coleta de dados via formulario de contato e newsletter
- Cookies utilizados
- Direitos do titular (LGPD)
- Contato do encarregado

**Nota:** Sera necessario adicionar uma rota no Go router. Verificar em
`internal/server/` ou `cmd/prospeccao/main.go` como as rotas publicas sao
registradas e adicionar `/privacidade` com handler renderizando o template.
Adicionar link no footer.

### 2.7 Endereco Fisico em Destaque

Adicionar na pagina de contato (`fale-conosco.html`) e no footer um destaque
do endereco fisico da empresa. Se ja existir no footer, dar mais prominencia.

Seguir o padrao PLDA: "Platinum Tower, Av. Carlos Gomes, 700 / 1006,
Porto Alegre | RS" -- endereco premium em destaque.

### 2.8 Vocabulario Tecnico B2B

Revisar o texto de TODOS os templates publicos e enriquecer com vocabulario
tecnico do mercado imobiliario corporativo:

| Termo generico | Substituir por |
|:--------------|:---------------|
| "Aluguel" | "Locacao Comercial" |
| "Imovel" | "Ativo Imobiliario" ou "Ponto Comercial" |
| "Construir sob medida" | "Built to Suit (BTS)" |
| "Vender e alugar de volta" | "Sale & Leaseback (SLB)" |
| "Escritorio" | "Laje Corporativa" |
| "Shopping pequeno" | "Strip Mall" |
| "Contrato" | "Contrato Atipico" (quando aplicavel) |
| "Investimento" | "Structuring de Investimento" |

---

## FASE 3: MELHORIA DO SISTEMA ADMIN

**Objetivo:** Elevar a qualidade visual do admin sem alterar funcionalidade.

### 3.1 Warm Surface no Admin

Conforme FASE 1.4, substituir todas as referencias slate por tokens warm.
O admin deve usar `bg-surface` como fundo principal, nao `bg-slate-50`.

### 3.2 Dashboard com Mais Impacto

Em `internal/template/admin/dashboard.html`:

- **KPI cards**: Adicionar icone SVG no canto superior direito de cada card
  (casa para imoveis, pessoa para clientes, grafico para prospeccoes)
- **Numero principal**: Usar `font-display text-3xl font-bold text-secondary-light`
  (gold) em vez do estilo atual
- **Tendencia**: Abaixo do numero, adicionar texto contextual:
  "X novos esta semana" ou "X em negociacao"
- **Animacao**: Adicionar `.reveal` nos cards para entrada suave ao carregar

### 3.3 Tabelas de Dados Refinadas

Em `input.css`, melhorar `.data-table`:

```css
.data-table tbody tr {
  transition: background-color 0.15s ease;
}
.data-table tbody tr:nth-child(even) {
  background-color: theme('colors.surface.container-lowest');
}
.data-table tbody tr:hover {
  background-color: theme('colors.surface.container-low');
}
/* Acoes inline na ultima coluna */
.data-table .actions {
  @apply flex items-center gap-1;
}
```

### 3.4 Formularios com Secoes Visuais

Em templates de formulario (`properties/form.html`, `clients/form.html`,
`prospections/form.html`):

Agrupar campos em fieldsets visuais com headers:

```html
<div class="card--elevated mb-6">
  <h3 class="text-headline-md font-display text-on-surface mb-4 pb-3 border-b border-outline-variant">
    Informacoes Basicas
  </h3>
  <!-- campos: titulo, tipo, status -->
</div>

<div class="card--elevated mb-6">
  <h3 class="text-headline-md font-display text-on-surface mb-4 pb-3 border-b border-outline-variant">
    Localizacao
  </h3>
  <!-- campos: endereco, cidade, estado, CEP -->
</div>
```

### 3.5 Empty States com Orientacao

Quando nao ha dados nas listagens, em vez de texto simples, usar:

```html
<div class="text-center py-16">
  <svg class="w-16 h-16 mx-auto text-outline-variant mb-4"><!-- icone contextual --></svg>
  <h3 class="text-headline-md font-display text-on-surface mb-2">Nenhum imovel cadastrado</h3>
  <p class="text-on-surface-variant mb-6">Comece cadastrando seu primeiro imovel para iniciar as prospeccoes.</p>
  <a href="/admin/imoveis/novo" class="btn btn--rounded btn--secondary btn--md">
    Cadastrar Primeiro Imovel
  </a>
</div>
```

---

## FASE 4: POLISH E DIFERENCIACAO

### 4.1 Micro-animacoes nos Status Badges (Admin)

Adicionar pulse sutil no badge "new":
```css
.badge-new {
  animation: badge-pulse 2s ease-in-out infinite;
}
@keyframes badge-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
```

### 4.2 Hover States nas Rows de Tabela

Adicionar cursor pointer e indicador visual nas rows clicaveis:
```css
.data-table tbody tr[onclick],
.data-table tbody tr a {
  cursor: pointer;
}
.data-table tbody tr[onclick]:hover {
  background-color: theme('colors.surface.container');
}
```

### 4.3 Transicoes de Pagina Suaves

Se o browser suportar View Transitions API, adicionar:
```html
<meta name="view-transition" content="same-origin">
```
Em `<head>` de todos os templates. Isso e progressivo -- nao quebra browsers
antigos.

### 4.4 Print CSS para PDF de Imovel

Verificar `internal/template/admin/properties/pdf.html` e garantir que:
- Nao usa glass effects ou backdrop-filter (nao renderiza em PDF)
- Cores sao solidas
- Logo aparece com caminho absoluto
- Margens adequadas para impressao A4

---

## ORDEM DE EXECUCAO

```
FASE 1 (Fundacao -- fazer primeiro, tudo depende disso)
  1.1  Adicionar tokens no tailwind.config.js
  1.2  Atualizar escala tipografica
  1.3  Refatorar componentes em input.css (card, btn, input)
  1.4  Migrar admin de slate para tokens warm
  1.5  Corrigir contact_error.html
  -> Compilar CSS: npx tailwindcss -i input.css -o static/css/app.css --minify
  -> Build Go: go build ./cmd/prospeccao
  -> Rodar: make check

FASE 2 (Conteudo -- site publico)
  2.1  Adicionar top contact bar
  2.2  Refinar hero headline
  2.3  Reestruturar navegacao com dropdown
  2.4  Adicionar grid de logos
  2.5  Adicionar secao "Por que nos contratar"
  2.6  Criar pagina LGPD
  2.7  Destacar endereco fisico
  2.8  Enriquecer vocabulario tecnico
  -> Compilar CSS
  -> Build Go
  -> make check

FASE 3 (Admin)
  3.1  Aplicar warm surface
  3.2  Melhorar dashboard
  3.3  Refinar tabelas
  3.4  Formularios com secoes
  3.5  Empty states
  -> Compilar CSS
  -> Build Go
  -> make check

FASE 4 (Polish)
  4.1  Badge animations
  4.2  Table hover states
  4.3  View Transitions meta
  4.4  PDF print check
  -> Compilar CSS
  -> Build Go
  -> make check FINAL
```

## COMO VERIFICAR SE FICOU BOM

Apos completar todas as fases, o agente deve validar:

1. **Consistencia visual**: Abrir site publico e admin lado a lado. Ambos devem
   compartilhar a mesma paleta (Navy + Gold + Warm Surface), tipografia
   (Montserrat/Inter), e "sensacao" de marca unica.

2. **Zero hex hardcoded**: Rodar `grep -rn "style=\"" internal/template/` e
   verificar que nao ha mais inline styles com cores hardcoded.

3. **Nomenclatura unificada**: Rodar `grep -rn "btn-primary\|btn-filled\|premium-card\|form-input " input.css`
   -- as classes antigas devem ter sido substituidas pelas novas com `--`.

4. **Conteudo institucional**: O hero deve ter proposta de valor direta. A nav
   deve ter segmentacao. O footer deve ter CRECI e endereco. Deve existir secao
   de logos de clientes.

5. **Build limpo**: `make check` deve passar sem erros.

6. **Nao ter emojis**: `grep -rn "[\x{1F600}-\x{1F9FF}]" internal/template/`
   deve retornar vazio.

---

## DOCUMENTOS DE REFERENCIA NO PROJETO

- `AGENTS.md` -- Regras de convencao do projeto (LER ANTES DE COMECAR)
- `docs/auditoria_design_completa.md` -- Relatorio completo da auditoria
- `tailwind.config.js` -- Tokens do design system
- `input.css` -- CSS fonte (754 linhas)

---

*Fim do prompt. Todas as informacoes necessarias estao neste arquivo.
O agente deve ler `AGENTS.md` primeiro, depois este prompt, e executar
na ordem das fases.*
