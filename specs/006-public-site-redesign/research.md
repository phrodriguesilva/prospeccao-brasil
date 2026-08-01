# Research: Public Site Redesign

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

---

## R1: Hero full-bleed com fotografia (CSS/Tailwind)

**Problem**: O hero atual do SPEC-04 e um bloco centrado com fundo `bg-surface-container-low` (off-white). Nao tem fotografia. Precisamos de um hero full-bleed com foto de imovel comercial e overlay para legibilidade do texto.

**Solution**: Usar `bg-cover bg-center` do Tailwind com uma imagem em `static/img/hero-comercial.jpg` (placeholder inicial). Overlay com gradiente `bg-gradient-to-r from-primary/80 to-primary/40` para garantir contraste do texto branco sobre a foto. Texto do hero em `text-on-primary` (branco).

```html
<section class="relative min-h-[600px] flex items-center bg-primary">
  <div class="absolute inset-0 bg-cover bg-center"
       style="background-image: url('/static/img/hero-comercial.jpg')"></div>
  <div class="absolute inset-0 bg-gradient-to-r from-primary/90 to-primary/50"></div>
  <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
    <h1 class="font-display text-display-lg text-on-primary">Encontramos o ponto comercial ideal</h1>
    <a href="/fale-conosco" class="btn btn-secondary btn-lg">Solicite uma apresentacao</a>
  </div>
</section>
```

**Fallback**: Se a imagem nao carregar, o `bg-primary` (Deep Navy) no `<section>` garante que o texto branco permaneca legivel. O overlay gradiente tambem funciona sobre cor solida.

**Decision**: Usar `background-image` inline no `<div>` overlay (nao `<img>`) para evitar layout shift e permitir `bg-cover` sem precisar de JS para calcular altura. Imagem servida de `static/img/` pelo static file server existente.

---

## R2: Paginas detalhadas de servicos (rotas dinamicas com chi)

**Problem**: O SPEC-04 tem apenas `GET /servicos` (pagina unica com cards). O SPEC-06 precisa de `GET /servicos/{slug}` com 5+ paginas detalhadas. Os servicos sao estaticos (5-10 fixos), nao precisam de banco.

**Solution**: Definir servicos como um map Go estatico em `internal/handler/institutional.go`:

```go
type serviceDetail struct {
    Slug        string
    Title       string
    Summary     string
    Description string
    Methodology []string
    CTA         string
}

var services = map[string]serviceDetail{
    "expansao-de-redes": { ... },
    "built-to-suit":     { ... },
    "strip-mall":        { ... },
    "lajes-comerciais":  { ... },
    "prospeccao-de-ponto": { ... },
}
```

O handler `ServicoDetalhe` le o `{slug}` do chi URL param, busca no map, renderiza `servico-detalhe.html`. Se nao encontrar, retorna 404.

**Rota**: `r.Get("/servicos/{slug}", instHandler.ServicoDetalhe)` apos `r.Get("/servicos", instHandler.Servicos)`.

**Decision**: Map estatico em Go. Nao criar tabela `services` no banco -- YAGNI para 5-10 servicos fixos de uma consultoria. Conteudo adaptado do legacy site.

---

## R3: Faixa de metricas (componente reutilizavel)

**Problem**: Home e Nossos Clientes precisam de uma faixa de metricas (Pontos Comercializados, Clientes Atendidos, Cidades, Anos). Os valores podem ser placeholders.

**Solution**: Criar um partial `internal/template/partials/metrics.html` que recebe um slice de structs `{Label, Value}`:

```go
type metric struct {
    Label string
    Value string
}
```

Definir as metricas como estaticas no handler e passar para o template. O partial e incluido com `{{template "metrics" .Metrics}}`.

**Decision**: Partial reutilizavel. Valores estaticos no handler (nao banco). O usuario atualizara os valores reais quando tiver.

---

## R4: Depoimentos estaticos no template

**Problem**: O SPEC-04 ja tem a struct `testimonial` no handler, mas a pagina Nossos Clientes exibe estado vazio. O SPEC-06 precisa de pelo menos 2-3 depoimentos visiveis.

**Solution**: Popular o slice `Testimonials` no handler com os 3 depoimentos do legacy site (Larissa Mello, Roberto Andrade, Joao Viana). Os depoimentos sao hardcoded no handler Go, passados para o template via `pageData`.

```go
var defaultTestimonials = []testimonial{
    {Name: "Larissa Mello", Company: "Rede de Farmacias", Quote: "A Prospecção Brasil selecionou os melhores pontos comerciais..."},
    {Name: "Roberto Andrade", Company: "Varejo", Quote: "A empresa tem nos ajudado a encontrar novos parceiros..."},
    {Name: "Joao Viana", Company: "Investidor", Quote: "A Prospecção Brasil tem fornecido informacoes preciosas..."},
}
```

**Decision**: Estatico no handler. Nao criar tabela `testimonials` -- YAGNI para 3 depoimentos. Se crescer, vira outra spec.

---

## R5: Copy de mercado (proibicao de copy de software)

**Problem**: Os templates atuais do SPEC-04 tem copy de software: "Reduzimos a carga cognitiva do prospector", "Gerenciamos imoveis, clientes e pipeline", "Relatorios PDF", "Gestao de Pipeline". Isso vende o software, nao o servico.

**Solution**: Reescrever todo o copy dos templates publicos com linguagem de consultoria imobiliaria comercial, adaptada do legacy site:

- Hero: "Encontramos o ponto comercial ideal para a sua rede" (legacy: "Provendo as melhores opcoes para seu negocio")
- Servicos: Expansao de Redes, Built to Suit, Strip Mall, Lajes Comerciais, Prospecao de Ponto (legacy + PLDA)
- Quem Somos: Historia do Luiz Claudio, 15 anos Shell, especialista em franquias/varejo (legacy)
- Missao: "Oferecer a melhor qualidade de servicos de prospecao de pontos comerciais" (legacy)
- CTA: "Solicite uma apresentacao" (legacy)

**Validacao**: `grep -ri "carga cognitiva\|pipeline\|plataforma\|software\|relatorios pdf" internal/template/*.html internal/template/partials/*.html` deve retornar vazio apos o redesign.

**Decision**: Reescrever todos os templates publicos. Manter estrutura de arquivos (home.html, servicos.html, etc.) mas trocar conteudo.

---

## R6: Auth templates com design system (CSS fix)

**Problem**: `login.html`, `totp_setup.html`, `totp_verify.html` foram criados no SPEC-03 sem CSS. Usam `<style="color:red">` e nao incluem `app.css`.

**Solution**: Adicionar `<link rel="stylesheet" href="/static/css/app.css">` no `<head>` de cada template. Trocar estilos inline por classes do design system:

- `<p style="color:red">` -> `<div class="alert alert-error">` (adicionar classe `.alert` ao `input.css` se nao existir)
- `<input type="email" name="email">` -> `<input type="email" name="email" class="form-input">`
- `<button type="submit">` -> `<button type="submit" class="btn btn-primary btn-md">`
- Layout centrado em um `<div class="card max-w-md mx-auto mt-20">`

**Nova classe CSS**: Adicionar `.alert` e `.alert-error` ao `input.css`:
```css
.alert { @apply rounded-md p-4 text-body-md; }
.alert-error { @apply bg-error-container text-error border border-error; }
```

**Decision**: Corrigir os 3 templates. Adicionar `.alert` ao design system. Funcionalidade (login, 2FA, cookies) permanece identica.

---

## R7: Imagens estaticas (static/img/)

**Problem**: O SPEC-06 precisa de imagens de hero e possivelmente fotos de servicos/fundador. O static file server atual serve `/static/*` mas o diretorio `static/img/` pode nao existir.

**Solution**: Criar `static/img/` com placeholders:
- `hero-comercial.jpg` -- foto generica de ponto comercial (ou gradiente Deep Navy como fallback no CSS)
- `fundador.jpg` -- foto do Luiz Claudio ou placeholder com iniciais "LC"
- Logos de clientes: omitir se nao disponivel (FR-014)

**Placeholder approach**: Para o hero, usar uma imagem de stock gratuita (Unsplash) de ponto comercial urbano brasileiro. Para o fundador, usar um `<div>` com iniciais "LC" em um circulo Deep Navy (ja existe no template atual).

**Decision**: Criar `static/img/` com hero placeholder. O usuario substituira por fotos reais pos-deploy. Template usa `background-image` com fallback `bg-primary`.

---

## R8: Navegacao e footer (atualizacao visual)

**Problem**: O nav e footer atuais do SPEC-04 sao funcionais mas com copy generico. O footer diz "Inteligencia em prospeccao de imoveis comerciais" (bom) mas nao tem endereco/telefone/WhatsApp (legacy tem).

**Solution**: Atualizar `partials/nav.html` e `partials/footer.html`:
- Nav: manter estrutura (logo + links + CTA). Trocar "Inico" por "Home" (ou manter). Adicionar link WhatsApp no mobile.
- Footer: adicionar bloco com endereco (Praia de Botafogo, 228 - 16 Andar - Botafogo - RJ), telefones (+55 21 99842-3232 / 97034-2617 / 3736-3696), email, links WhatsApp/Instagram. Manter newsletter.

**Decision**: Atualizar footer com info de contato do legacy. Manter newsletter funcional. Adicionar links sociais (WhatsApp, Instagram).

---

## R9: Pencil designs antes de HTML/CSS

**Problem**: AGENTS.md exige Pencil como visual source of truth para UI specs. O SPEC-06 tem FR-027 exigindo frames em `designs/prospeccao.pen` antes da implementacao.

**Solution**: Usar o MCP server `pencil` para criar `designs/prospeccao.pen` com frames:
- Home - Desktop (1440px)
- Home - Mobile (375px)
- Servicos Index - Desktop
- Servicos Detalhe - Desktop
- Quem Somos - Desktop
- Nossos Clientes - Desktop
- Fale Conosco - Desktop
- Login - Desktop (auth com design system)

**Timing**: Pencil designs sao criados no stage de implementacao (apos plan e tasks), nao no stage de plan. O plan define O QUE construir; o Pencil define COMO visualmente. O tasks.md tera uma task explicita "criar Pencil designs" como primeira task de implementacao.

**Decision**: Pencil apos tasks, antes de HTML. Referenciar frames no spec e no tasks.

---

## R10: Mantendo comportamento do formulario (SPEC-04 compatibilidade)

**Problem**: O formulario de Fale Conosco ja funciona (HTMX, validacao, persistencia, fallback no-JS). O SPEC-06 melhora apenas visual. Nao pode quebrar o existente.

**Solution**: Manter o handler `ContactHandler.Submit` e o fragmento `fragments/contact_success.html` / `fragments/contact_error.html` inalterados. Apenas o template `fale-conosco.html` e reescrito com:
- Layout em 2 colunas (formulario + info de contato) no desktop
- Campos atuais (name, email, phone, subject, message) -- adicionar campo "empresa" opcional
- Manter `hx-post="/fale-conosco"` `hx-target="#contact-form-container"` `hx-swap="outerHTML"`
- Manter `action="/fale-conosco" method="POST"` para fallback no-JS

**Novo campo "empresa"**: Adicionar campo opcional "Empresa" ao formulario. O handler `ContactHandler` precisa aceitar o campo extra. Verificar se a tabela `contact_submissions` tem coluna `company` ou se precisa migration. Se precisar, criar migration `000003_add_company_to_contact_submissions.up.sql`.

**Decision**: Verificar schema da tabela. Se nao tiver `company`, adicionar via migration. Handler aceita campo opcional.

---

## R11: Estrutura de templates (manter ou reescrever)

**Problem**: Os templates atuais do SPEC-04 sao standalone (cada um com `<html>`, `<head>`, `<body>` completo). Nao usam um `base.html` shared (embora o SPEC-04 plan mencione).

**Solution**: Manter a estrutura standalone (cada template e uma pagina HTML completa). Isso e mais simples para server-rendered Go templates e evita complexidade de heranca. O nav e footer sao partials incluidos via `{{template "nav" .}}` e `{{template "footer" .}}`.

**Novo template**: `servico-detalhe.html` para paginas `/servicos/{slug}`.

**Decision**: Manter estrutura standalone. Adicionar `servico-detalhe.html`. Reescrever conteudo de todos os templates publicos.

---

## R12: Test strategy para redesign visual

**Problem**: O SPEC-06 e principalmente redesign visual (HTML/CSS). Testes de integracao existentes (SPEC-04) verificam status code e conteudo. Precisamos garantir que o redesign nao quebre os testes existentes e adicionar testes para a nova rota `/servicos/{slug}`.

**Solution**:
- Manter testes existentes do `institutional_test.go` (TestHome, TestQuemSomos, TestServicos, TestNossosClientes, TestFaleConosco). Atualizar assertions se o copy mudar (ex: "carga cognitiva" nao estara mais no body).
- Adicionar `TestServicoDetalhe` (200, conteudo do servico), `TestServicoDetalheNotFound` (404 para slug inexistente).
- Adicionar `TestServicoDetalheAll` (loop sobre todos os slugs, verificar 200).
- Auth template tests: adicionar `TestLoginHasCSS` (verificar `<link rel="stylesheet" href="/static/css/app.css">` no body), `TestTotpSetupHasCSS`, `TestTotpVerifyHasCSS`.
- Coverage: handler ja esta em 85%+. Redesign nao deve reduzir.

**Decision**: Atualizar assertions dos testes existentes. Adicionar testes para nova rota e auth CSS.

---

## R13: PLDA como referencia de estrutura (nao clone)

**Problem**: O usuario citou PLDA (plda.com.br) como referencia de layout/estrutura. Precisamos entender o que aproveitar sem copiar literalmente.

**PLDA patterns aproveitaveis**:
- Hero full-bleed com headline de mercado ("Alugamos seu imovel para grandes redes varejistas")
- Servicos como paginas profundas dedicadas (cada servico tem URL propria)
- Separar Depoimentos e Empresas Parceiras em paginas distintas (nosso caso: juntar em Nossos Clientes)
- Footer com CRECI, endereco, WhatsApp, email
- Tom institucional premium, fotografia-led

**PLDA patterns NAO aproveitar**:
- Menu dropdown complexo com 11+ servicos (nosso caso: 5 servicos, menu simples)
- Wix-style layout (PLDA usa Wix). Nosso caso: Go html/template + Tailwind
- Cores da PLDA (azul claro, branco). Nosso caso: Deep Navy + Sóbrio Gold

**Decision**: Aproveitar estrutura (hero, servicos profundos, prova social separada) com identidade Prospecção Brasil (Deep Navy, Gold, Montserrat).
