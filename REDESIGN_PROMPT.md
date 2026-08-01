# REDESIGN ESTRUTURAL DO SITE PÚBLICO -- Hibrido PLDA + Blooty

O site atual e funcional mas visualmente fraco comparado as referencias (PLDA.com.br, Blooty Agencia, WTour Travel). Precisamos de um redesign estrutural, nao mais polish incremental. O objetivo: site premium B2B para prospeccao imobiliaria comercial, no nivel do PLDA.com.br (nossa referencia direta, mesmo nicho).

## Decisoes do usuario

- **Abordagem**: Hibrido PLDA (estrutura/nicho) + Blooty (polish/visual)
- **Fotos**: Usar stock gratuitas (Unsplash/Pexels) -- baixar e self-hostar em `static/img/`
- **Navegacao**: Nav flutuante glass (estilo Blooty, NAO menu lateral do PLDA)
- **Processo**: Direto no codigo, sem Pencil

## Restricoes (NAO QUEBRAR)

- **Stack**: Go + html/template + HTMX + Alpine.js + Tailwind (build-time, NAO CDN)
- **Nao mudar**: Go handlers, routes, data structures em `institutional.go`, auth, middleware
- **Mudar APENAS**: `input.css`, templates HTML, `static/js/`, `static/img/`
- **Self-host**: todas as fotos e icones em `static/img/` (nunca CDN externo)
- Rodar `npx tailwindcss -i input.css -o static/css/app.css --minify` apos mudancas no CSS
- Rodar `go build ./cmd/prospeccao` para verificar que templates compilam
- Atualizar testes em `institutional_test.go` se asserts de string mudarem

---

## FASE 1: Stock Photos -- baixar e self-hostar em static/img/

Baixar estas fotos do Unsplash (use `curl` para baixar, sao gratuitas para uso comercial):

1. **hero-comercial.jpg** -- foto principal do hero
   - Buscar: "commercial real estate building" ou "shopping street retail" ou "modern office building exterior"
   - Tamanho: pelo menos 1920x1080, otimizar para < 300KB
   - Sugestao: `https://images.unsplash.com/photo-1486406146926-c627a92ad1ab` (building exterior) ou similar
   - Salvar em: `static/img/hero-comercial.jpg`

2. **service-expansao.jpg** -- expansao de redes
   - Buscar: "retail store front" ou "shopping mall interior"

3. **service-bts.jpg** -- built to suit
   - Buscar: "construction site commercial" ou "modern building development"

4. **service-strip-mall.jpg** -- strip mall
   - Buscar: "strip mall shopping center" ou "small shopping street"

5. **service-lajes.jpg** -- lajes comerciais
   - Buscar: "modern office interior" ou "corporate office space"

6. **service-prospeccao.jpg** -- prospeccao de ponto
   - Buscar: "city street commercial" ou "urban retail district"

7. **about-founder.jpg** -- escritorio corporativo (para quem-somos)
   - Buscar: "modern office desk" ou "corporate meeting room"

8. **cta-bg.jpg** -- fundo para secoes CTA
   - Buscar: "city skyline night" ou "urban aerial view"

**Para cada foto**: baixar com `curl`, redimensionar/comprimir se necessario (usar `sips` no macOS ou instalar `imagemagick`), garantir que cada uma seja < 300KB. Se `sips` nao funcionar, baixar versoes otimizadas do Unsplash com parametro `&w=1920&q=80&fm=jpg`.

---

## FASE 2: Nav Glass Flutuante (estilo Blooty)

Reescrever `internal/template/partials/nav.html` completamente. A nav atual e sticky retangular simples. A nova deve ser:

### Estrutura

- **Floating pill**: `position: fixed`, `top: 12px`, `left: 12px`, `right: 12px`, `border-radius: 9999px`
- **Glassmorphism**: `backdrop-filter: blur(20px) saturate(180%)`, `background: rgba(252,249,248,0.85)`
- **Shadow**: `0 4px 24px rgba(26,43,76,0.08)`
- **Scroll detection**: adicionar classe `.scrolled` apos 20px de scroll (background fica mais opaco)
- **Logo + nome da empresa** a esquerda (usar `logo-symbol.png`)
- **Links a direita**: Inicio, Quem somos, Servicos, Nossos clientes, Fale Conosco (btn)
- **Mobile**: hamburger que abre menu pill abaixo da nav (nao fullscreen drawer)
- **Altura**: 56px (menor que atual 64px, mais elegante)

### CSS a adicionar no input.css

```css
.glass-nav {
  position: fixed; top: 12px; left: 12px; right: 12px;
  border-radius: 9999px;
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  background: rgba(252, 249, 248, 0.80);
  box-shadow: 0 4px 24px rgba(26, 43, 76, 0.08), inset 0 1px 0 rgba(255,255,255,0.5);
  border: 1px solid rgba(26, 43, 76, 0.06);
  z-index: 100;
  transition: all 0.3s ease;
}
.glass-nav.scrolled {
  background: rgba(252, 249, 248, 0.95);
  box-shadow: 0 8px 32px rgba(26, 43, 76, 0.12), inset 0 1px 0 rgba(255,255,255,0.6);
}
```

### JS para toggle .scrolled (adicionar no reveal.js ou inline Alpine)

```js
window.addEventListener('scroll', function() {
  var nav = document.querySelector('.glass-nav');
  if (nav) nav.classList.toggle('scrolled', window.scrollY > 20);
});
```

**IMPORTANTE**: como a nav agora e `fixed`, adicionar `padding-top` no `<body>` de todas as paginas publicas (pelo menos 80px) para compensar.

---

## FASE 3: Hero Redesign (espelhar PLDA + polish Blooty)

Reescrever o hero do `home.html` completamente. O hero atual e navy solido sem foto. O novo deve ser:

### Estrutura

- **Full-bleed**: `min-height: 100vh` (ou pelo menos 90vh), nao 600px
- **Background**: foto `hero-comercial.jpg` com overlay gradiente navy (left-to-right, 90% -> 40% opacity)
- **Ken Burns lento**: `animation: ken-burns 20s ease-in-out infinite alternate` (zoom de 1.0 -> 1.08)
- **Gradiente adicional**: radial glow no canto (como Blooty hero-glow)

### Conteudo (centralizado ou alinhado a esquerda)

- **Eyebrow**: "RETAIL SERVICE PARA GRANDES EMPRESAS" (gold, uppercase, letter-spacing 0.2em)
- **Headline GRANDE**: "Alugamos seu imovel para grandes redes varejistas" (nao "Encontramos o ponto comercial ideal" -- copiar a promessa do PLDA, que e mais direta)
  - Tamanho: `text-5xl md:text-6xl lg:text-7xl` (MUITO maior que atual)
  - Font-weight: 800
  - Line-height: 1.1
  - Cor: white
  - Destacar "grandes redes varejistas" com `text-gradient` gold
- **Subheadline**: "Inteligencia em Real Estate Comercial. Prospeccao, estruturacao e comercializacao de pontos comerciais em todo o Brasil."
  - Tamanho: `text-lg md:text-xl`
  - Cor: `white/80`
  - Max-width: 600px
- **2 CTAs**:
  - Primario: "Solicite uma apresentacao" (btn gold, pill shape, shadow glow)
  - Secundario: "Conheca nossos servicos" (outline branco, pill shape)
- **Metricas inline** abaixo dos CTAs: 500+ Pontos | 50+ Clientes | 100+ Cidades | 15+ Anos (nao em secao separada, integrado no hero)

### CSS do hero

```css
.hero-fullbleed {
  position: relative; min-height: 90vh; display: flex; align-items: center;
  overflow: hidden;
}
.hero-fullbleed .hero-img {
  position: absolute; inset: 0; background-size: cover; background-position: center;
  animation: ken-burns 20s ease-in-out infinite alternate;
}
@keyframes ken-burns {
  0% { transform: scale(1) translate(0, 0); }
  100% { transform: scale(1.08) translate(-1%, -1%); }
}
@media (prefers-reduced-motion: reduce) {
  .hero-fullbleed .hero-img { animation: none; }
}
```

---

## FASE 4: Tipografia e Espacamento (premium feel)

O site atual usa tamanhos pequenos e espacamento apertado. Para parecer premium:

### Tipografia

- **Headlines de secao**: `text-4xl md:text-5xl` (atual e `text-3xl`, muito pequeno)
- **Display do hero**: `text-5xl md:text-6xl lg:text-7xl`
- **Body text**: `text-lg` para paragrafos importantes (atual e `text-base`)
- **Eyebrow**: manter `0.875rem` mas aumentar `letter-spacing` para `0.2em`
- Adicionar `font-weight: 800` para headlines (atual e 700)

### Espacamento

- **Secoes**: `py-24 md:py-32` (atual e `py-20`, muito apertado)
- **Container**: `max-w-6xl` (atual e `max-w-7xl`, muito largo para texto)
- **Gap entre cards**: `gap-8 md:gap-10`
- **Margem entre secoes**: mais whitespace, menos conteudo por secao
- **Padding interno dos cards**: `p-8` (atual e `p-6`)

---

## FASE 5: Services com Fotos (nao cards genericos)

Reescrever a secao de servicos no `home.html` e `servicos.html`. Em vez de cards texto-only, usar cards com foto:

### HOME (3 servicos destaque + link "ver todos")

- Grid de 3 cards (expansao-de-redes, built-to-suit, prospeccao-de-ponto)
- Cada card: foto no topo (`h-48`, `object-cover`), titulo, summary, "Saiba mais ->"
- Hover: foto zoom (`scale 1.05`), card lift, border glow
- Usar `premium-card` mas com imagem

### SERVICOS INDEX (todos os 5)

- Grid de 2 colunas (`md:grid-cols-2`)
- Cada card: foto lateral (`w-40`) + conteudo (titulo, summary, link)
- Layout horizontal, nao vertical
- Hover: lift + foto zoom

### SERVICO DETALHE

- Hero com foto do servico (`service-{slug}.jpg`) como background
- Descricao em coluna larga (`max-w-3xl`)
- Metodologia em **timeline visual** (nao cards retangulares):
  - Numero em circulo gold
  - Linha conectando os passos
  - Texto ao lado de cada passo
- CTA no final com foto `cta-bg.jpg` como background

---

## FASE 6: Quem Somos Rico (nao texto corrido)

Reescrever `quem-somos.html`:

- **Hero** com foto `about-founder.jpg` (escritorio corporativo) como background
- **Founder**: layout 2 colunas (foto placeholder circular + bio)
  - Usar iniciais "LC" em circulo gold (nao temos foto do Luiz Claudio)
  - Bio em coluna larga com paragrafos bem espacados
- **Missao/Visao/Valores**: 3 cards `premium-card` com icone gold no topo (nao badge)
- **CRECI**: secao com fundo navy e texto branco, centralizado
- **CTA** no final

---

## FASE 7: Nossos Clientes (depoimentos premium)

Reescrever `nossos-clientes.html`:

- **Hero** navy com eyebrow "SOCIAL PROOF"
- **Metricas em destaque** (nao strip fino): 4 cards grandes com numero gold gigante (`text-5xl`)
- **Depoimentos** em carousel scroll-snap (ja temos) mas com:
  - Aspas decorativas grandes no topo do card (`text-6xl` gold `opacity 20%`)
  - Nome + empresa em destaque
  - Card com fundo cream/surface-container-low (nao branco puro)
- **CTA** no final

---

## FASE 8: Fale Conosco (form + info lado a lado premium)

Reescrever `fale-conosco.html`:

- **Hero** navy com eyebrow "CONTATO"
- **Layout 2 colunas**:
  - Esquerda: formulario em card glass (`backdrop-filter: blur`, fundo semi-transparente)
  - Direita: info de contato com icones (endereco, telefone, email, whatsapp)
- **Form inputs** com estilo premium: `border-bottom` apenas (nao border completo), focus state com gold
- **Botao submit**: pill gold com shadow glow

---

## FASE 9: Footer Rico

Atualizar `footer.html`:

- **Fundo** navy (`bg-primary`) com gradiente sutil
- **4 colunas**:
  1. Logo full + descricao + social (Instagram, LinkedIn)
  2. Navegacao (links do site)
  3. Servicos (links para `/servicos/{slug}`)
  4. Contato + Newsletter
- **Divider gold** no topo do footer (linha gradient)
- **Bottom bar**: copyright + CRECI

---

## FASE 10: Polish Final

- **Buttons**: mudar para pill shape (`rounded-full`) em todos os CTAs publicos
- **Shadows**: adicionar `shadow-lg` nos cards principais
- **Transitions**: aumentar duration para `0.3s` em hovers
- **Scroll reveal**: ja temos, garantir que esta aplicado em todas as secoes
- **Custom scrollbar**: ja temos, manter
- **Back-to-top**: ja temos, manter
- **WhatsApp float**: ja temos, manter
- **Scroll progress**: ja temos, manter

---

## Ordem de Execucao

1. **FASE 1**: Baixar todas as stock photos (`curl` + compressao)
2. **FASE 2**: Reescrever `nav.html` + adicionar CSS `glass-nav` + `padding-top` no body
3. **FASE 3**: Reescrever hero do `home.html`
4. **FASE 4**: Atualizar `input.css` (tipografia, espacamento, buttons pill)
5. **FASE 5**: Reescrever services (home + servicos + detalhe)
6. **FASE 6**: Reescrever quem-somos
7. **FASE 7**: Reescrever nossos-clientes
8. **FASE 8**: Reescrever fale-conosco
9. **FASE 9**: Atualizar footer
10. **FASE 10**: Polish final + build CSS + `go build` + testes

**Apos cada fase**, rodar `go build ./cmd/prospeccao` para verificar que templates compilam.
**Apos mudancas no CSS**, rodar `npx tailwindcss -i input.css -o static/css/app.css --minify`.
**No final**, rodar `make check` completo.

## Regras Finais

- **NAO** fazer deploy para a VPS -- o usuario cuida do deploy.
- **NAO** mudar Go handlers, routes, ou data structures.
- **NAO** adicionar CDN dependencies -- tudo self-hosted.
- **NAO** usar emojis em codigo, UI, comentarios, ou commits.
