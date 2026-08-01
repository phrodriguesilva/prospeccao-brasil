# Feature Specification: Public Site Redesign

**Feature Branch**: `006-public-site-redesign`

**Created**: 2026-07-31

**Status**: Draft

**Input**: User description: "SPEC-06: Public Site Redesign. Redesignar o site institucional publico para vender o SERVICO de prospeccao imobiliaria comercial -- nao o software. Hero com fotografia, paginas profundas de servicos, prova social, copy de consultoria de varejo. Manter tokens de marca, HTMX, Go html/template. Pencil designs primeiro."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Home Institucional Premium (Priority: P1)

Um executivo de expansao de uma rede varejista chega ao site publico
`prospeccaobrasil.com` pela primeira vez, vindo de uma busca no Google
ou indicacao. Ele ve um hero full-bleed com fotografia de ponto
comercial impactante, um headline de mercado ("Encontramos o ponto
comercial ideal para a sua rede"), um subtitulo que posiciona a
Prospecção Brasil como consultoria de real estate comercial (nao como
software), uma faixa de metricas (pontos comercializados, cidades,
anos de mercado), uma secao de servicos com links para paginas
profundas, prova social (depoimentos de clientes), e um CTA forte
"Solicite uma apresentacao" que leva a /fale-conosco. O primeiro
viewport transmite credibilidade institucional premium -- nao um
generico Tailwind card stack.

**Why this priority**: A Home e a porta de entrada e a primeira
impressao de credibilidade. Se ela nao transmite autoridade de
consultoria imobiliaria comercial, o visitante sai antes de ver
servicos ou contato. E a unica pagina que 100% dos visitantes veem.

**Independent Test**: Abrir `http://localhost:8080/` e verificar:
hero com fotografia, headline de mercado (nao "reduzimos carga
cognitiva"), faixa de metricas, secao de servicos com links, pelo
menos um depoimento, CTA "Solicite uma apresentacao". Navegacao
funciona para todas as paginas.

**Acceptance Scenarios**:

1. **Given** que o visitante acessa `/`, **When** a pagina carrega, **Then** o hero exibe fotografia de imovel comercial em full-bleed com overlay, headline "Encontramos o ponto comercial ideal" (ou equivalente de mercado), e CTA visivel acima da dobra
2. **Given** que o visitante acessa `/`, **When** a pagina carrega, **Then** a faixa de metricas exibe 4 numeros (Pontos Comercializados, Clientes Atendidos, Cidades Atendidas, Anos de Mercado)
3. **Given** que o visitante acessa `/`, **When** ele rola a pagina, **Then** a secao de servicos exibe no minimo 3 servicos com links para `/servicos/{slug}`
4. **Given** que o visitante acessa `/`, **When** ele rola ate a secao de prova social, **Then** exibe pelo menos 2 depoimentos com nome e cargo/empresa do autor
5. **Given** que o visitante acessa `/` em um dispositivo mobile, **When** a pagina carrega, **Then** o hero, metricas, servicos e CTA se reorganizam em coluna unica sem quebra de layout
6. **Given** que o visitante acessa `/`, **When** ele clica em "Solicite uma apresentacao", **Then** e redirecionado para `/fale-conosco`

---

### User Story 2 - Paginas Profundas de Servicos (Priority: P2)

Um diretor de expansao de uma rede de farmacias pesquisa "expansao de
redes varejistas" e chega a `/servicos`. Ele ve um indice de servicos
com cards descritivos (nao icones genericos) e clica em "Expansao de
Redes". A pagina detalhada `/servicos/expansao-de-redes` explica a
metodologia (plano diretor, analise de macro/microlocalizacao,
prospeccao de ponto), mostra como a Prospecção Brasil executa cada
etapa, e tem um CTA "Fale com um especialista". O mesmo padrao se
repete para os outros servicos: Built to Suit (BTS), Strip Mall /
Centros de Conveniencia, Lajes Comerciais, e Prospeccao de Ponto
Comercial. Cada pagina profunda tem conteudo real do legacy site
adaptado, nao placeholder.

**Why this priority**: Servicos e onde o visitante decide se a
empresa tem a profundidade tecnica que ele precisa. Cards genericos
do SPEC-04 ("Relatorios PDF", "Gestao de Pipeline") vendem software,
nao servico. Paginas profundas com metodologia real sao o que
converte um executivo de expansao.

**Independent Test**: Abrir `http://localhost:8080/servicos` e
verificar o indice com 5+ servicos. Clicar em cada um e verificar
que a pagina detalhada tem: titulo, descricao da metodologia, pelo
menos uma secao de "Como fazemos", e CTA.

**Acceptance Scenarios**:

1. **Given** que o visitante acessa `/servicos`, **When** a pagina carrega, **Then** exibe um indice com no minimo 5 servicos: Expansao de Redes, Built to Suit (BTS), Strip Mall / Centros de Conveniencia, Lajes Comerciais, Prospeccao de Ponto Comercial
2. **Given** que o visitante acessa `/servicos`, **When** ele clica em "Expansao de Redes", **Then** e redirecionado para `/servicos/expansao-de-redes` que exibe conteudo detalhado (metodologia, etapas, CTA)
3. **Given** que o visitante acessa `/servicos/built-to-suit`, **When** a pagina carrega, **Then** exibe descricao do servico BTS, como funciona, e CTA "Fale com um especialista"
4. **Given** que o visitante acessa `/servicos/servico-inexistente`, **When** a pagina carrega, **Then** retorna 404 com a pagina de erro institucional
5. **Given** que o visitante acessa `/servicos/expansao-de-redes` em mobile, **When** a pagina carrega, **Then** o conteudo se reorganiza em coluna unica sem quebra

---

### User Story 3 - Quem Somos com Historia e Autoridade (Priority: P3)

Um investidor ou proprietario de imovel pesquisa a empresa antes de
fechar. Ele acessa `/quem-somos` e le a historia do fundador Luiz
Claudio (15 anos na Shell Brasil, especialista em redes de franquia
e varejo), a missao/visao/valores da empresa, e sente que esta lidando
com uma consultoria com autoridade de mercado -- nao com uma startup
de software. A pagina tem fotografia do fundador (ou placeholder
institucional se foto nao disponivel), bloco de missao/visao/valores,
e mencao da licenca CRECI.

**Why this priority**: Quem somos fecha a confianca. O visitante ja
viu o hero e os servicos; agora ele quer saber quem esta por tras.
A historia do Luiz Claudio e da Shell e o diferencial de autoridade.

**Independent Test**: Abrir `http://localhost:8080/quem-somos` e
verificar: nome do fundador, mencao a Shell Brasil, missao, visao,
valores, e mencao CRECI. Nenhum texto sobre "plataforma" ou
"carga cognitiva".

**Acceptance Scenarios**:

1. **Given** que o visitante acessa `/quem-somos`, **When** a pagina carrega, **Then** exibe a historia do fundador Luiz Claudio com mencao a "15 anos Shell Brasil" e especialidade em "redes de franquias e varejo"
2. **Given** que o visitante acessa `/quem-somos`, **When** ele rola a pagina, **Then** encontra blocos de Missao, Visao e Valores (Transparencia, Profissionalismo, Etica, Comprometimento, Agilidade)
3. **Given** que o visitante acessa `/quem-somos`, **When** ele rola ate o final da pagina, **Then** encontra mencao a licenca CRECI (Conselho Federal/Regional de Corretores de Imoveis)
4. **Given** que o visitante acessa `/quem-somos`, **When** ele busca na pagina, **Then** nao encontra mencao a "software", "plataforma", "pipeline", ou "carga cognitiva"

---

### User Story 4 - Nossos Clientes com Prova Social (Priority: P4)

Um prospect acessa `/nossos-clientes` e ve depoimentos reais (Larissa
Mello, Roberto Andrade, Joao Viana -- do legacy site), logos de
empresas parceiras (ou placeholders institucionais se logos nao
disponiveis), e uma secao de numeros que reforca a autoridade. A
pagina substitui o estado vazio atual do SPEC-04. Se nao houver
logos ou depoimentos suficientes, a pagina exibe o que existe com
elegancia -- sem "em breve" ou "placeholder" visivel.

**Why this priority**: Prova social e o ultimo empurrao antes do
contato. O estado vazio atual prejudica a credibilidade. Mesmo com
poucos depoimentos, apresenta-los bem e melhor que nada.

**Independent Test**: Abrir `http://localhost:8080/nossos-clientes`
e verificar: pelo menos 2 depoimentos com nome, faixa de metricas,
e nenhum texto de "vazio" ou "placeholder".

**Acceptance Scenarios**:

1. **Given** que o visitante acessa `/nossos-clientes`, **When** a pagina carrega, **Then** exibe pelo menos 2 depoimentos com nome do autor e texto (Larissa Mello, Roberto Andrade, e/ou Joao Viana)
2. **Given** que o visitante acessa `/nossos-clientes`, **When** a pagina carrega, **Then** exibe uma faixa de metricas (pontos comercializados, clientes atendidos, cidades, anos)
3. **Given** que o visitante acessa `/nossos-clientes`, **When** nao ha logos de empresas cadastradas, **Then** a secao de logos e omitida ou substituida por depoimentos sem exibir "vazio" ou "em breve"
4. **Given** que o visitante acessa `/nossos-clientes`, **When** ele rola ate o final, **Then** encontra um CTA "Solicite uma apresentacao" ou "Fale Conosco"

---

### User Story 5 - Fale Conosco com Apresentacao Visual Melhorada (Priority: P5)

Um prospect decide entrar em contato. Ele acessa `/fale-conosco` e ve
um formulario com apresentacao visual premium (nao o formulario
generico atual), campos para Empresa, Nome, Telefone, Email, e
Mensagem. O formulario continua usando HTMX para validacao async
e persistencia no banco (ja implementado no SPEC-04). Ao submeter
com sucesso, recebe uma mensagem de confirmacao. A pagina tambem
exibe informacoes de contato (endereco, telefones, email, WhatsApp)
de forma proeminente -- nao apenas no footer.

**Why this priority**: O formulario ja funciona (SPEC-04). Esta
historia e apenas melhoraria visual para alinhar com o resto do
redesign. Funcionalidade permanece identica.

**Independent Test**: Abrir `http://localhost:8080/fale-conosco`,
preencher o formulario, submeter, e verificar mensagem de sucesso.
Verificar que endereco e telefones aparecem na pagina (nao so no
footer).

**Acceptance Scenarios**:

1. **Given** que o visitante acessa `/fale-conosco`, **When** a pagina carrega, **Then** exibe o formulario com campos Empresa, Nome, Telefone, Email, Mensagem e botao "Enviar"
2. **Given** que o visitante acessa `/fale-conosco`, **When** a pagina carrega, **Then** exibe informacoes de contato (endereco Botafogo RJ, telefones, email, WhatsApp) em destaque na pagina
3. **Given** que o visitante preenche o formulario com dados validos, **When** ele clica em "Enviar", **Then** recebe mensagem de sucesso via HTMX sem reload da pagina
4. **Given** que o visitante submete o formulario com email invalido, **When** o HTMX processa, **Then** recebe mensagem de erro de validacao inline
5. **Given** que o visitante submete o formulario sem JavaScript (no-JS fallback), **When** o formulario e processado, **Then** a pagina recarrega com mensagem de sucesso ou erro

---

### User Story 6 - Correcao CSS dos Templates de Autenticacao (Priority: P6)

O administrador acessa `sistema.prospeccaobrasil.com/login` e ve uma
pagina de login com visual alinhado ao design system da Prospecção
Brasil (Deep Navy, Montserrat, Inter, classes Tailwind do app.css) --
nao a pagina sem CSS atual com `<style="color:red">`. O mesmo se
aplica a `/2fa/setup` e `/2fa/verify`. A funcionalidade de login,
2FA TOTP, e cookies de sessao permanece identica (SPEC-03). Apenas
a apresentacao visual e corrigida.

**Why this priority**: As paginas de auth sao a unica parte do
sistema que um humano ve hoje (o admin). Sem CSS, parecem
amadoras. Como o redesign do site publico ja esta em andamento,
aproveitar para corrigir as paginas de auth e baixo custo e alto
valor visual.

**Independent Test**: Abrir `http://localhost:8080/login` e
verificar que a pagina carrega com `app.css`, usa classes do
design system (btn, card, form-input), e nao tem estilos inline
`style="color:red"`.

**Acceptance Scenarios**:

1. **Given** que o administrador acessa `/login`, **When** a pagina carrega, **Then** o `<head>` inclui `<link rel="stylesheet" href="/static/css/app.css">` e a pagina usa classes do design system (form-input, btn, btn-primary)
2. **Given** que o administrador acessa `/login` com credenciais invalidas, **When** o formulario retorna erro, **Then** a mensagem de erro usa classes do design system (alert alert-error) em vez de `style="color:red"`
3. **Given** que o administrador acessa `/2fa/setup`, **When** a pagina carrega, **Then** o QR code e o formulario TOTP usam classes do design system com layout centrado em um card
4. **Given** que o administrador acessa `/2fa/verify`, **When** a pagina carrega, **Then** o formulario de codigo TOTP usa classes do design system com layout centrado em um card
5. **Given** que o administrador faz login com credenciais validas, **When** o fluxo 2FA completa, **Then** o comportamento de redirect e cookie de sessao permanece identico ao SPEC-03 (sem mudanca funcional)

---

### Edge Cases

- **Imagens nao disponiveis**: Se a fotografia do hero ou do fundador nao estiver disponivel, o template deve exibir um gradiente ou cor de marca (Deep Navy) como fallback -- nao um `<img>` quebrado. As imagens sao referenciadas em `static/img/` e podem ser placeholders iniciais.
- **Depoimentos insuficientes**: Se houver menos de 3 depoimentos, a secao de prova social exibe os disponiveis sem indicar que falta algo. Nao usar "em breve" ou "placeholder".
- **Logos de clientes ausentes**: Se nao houver logos cadastrados, a secao de logos e omitida da pagina Nossos Clientes. Nao exibir grid vazio.
- **Pagina de servico inexistente**: Se o visitante acessar `/servicos/{slug}` com um slug nao cadastrado, retorna 404 com a pagina de erro institucional (ja existe `404.html`).
- **Formulario sem JavaScript**: O formulario de Fale Conosco ja tem fallback no-JS (SPEC-04). O redesign nao deve quebrar esse comportamento.
- **Metricas com valor zero**: Se as metricas (pontos comercializados, etc.) tiverem valor 0 ou nao estiverem definidas, exibir o numero como "0" ou omitir a faixa -- nao exibir "null" ou quebrar o layout.
- **Mobile em todas as paginas**: Todas as paginas devem funcionar em viewport 375px (iPhone SE) sem scroll horizontal.
- **Auth templates em subdominio diferente**: As paginas de auth sao servidas em `sistema.prospeccaobrasil.com` (host interno). O CSS e servido do mesmo dominio. Nao ha CORS ou cross-domain.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Home (`/`) MUST exibir um hero full-bleed com fotografia de imovel comercial (ou gradiente fallback), headline de mercado em pt-BR, e CTA "Solicite uma apresentacao" acima da dobra
- **FR-002**: Home MUST exibir uma faixa de metricas com 4 indicadores: Pontos Comercializados, Clientes Atendidos, Cidades Atendidas, Anos de Mercado
- **FR-003**: Home MUST exibir uma secao de servicos com no minimo 3 cards, cada um linkando para `/servicos/{slug}`
- **FR-004**: Home MUST exibir pelo menos 2 depoimentos com nome do autor e texto
- **FR-005**: Home MUST exibir um CTA final "Solicite uma apresentacao" ou "Fale Conosco" que linka para `/fale-conosco`
- **FR-006**: Servicos (`/servicos`) MUST exibir um indice com no minimo 5 servicos: Expansao de Redes, Built to Suit (BTS), Strip Mall / Centros de Conveniencia, Lajes Comerciais, Prospeccao de Ponto Comercial
- **FR-007**: Cada servico MUST ter uma pagina detalhada em `/servicos/{slug}` com: titulo, descricao da metodologia, pelo menos uma secao de "Como fazemos" com etapas, e CTA "Fale com um especialista"
- **FR-008**: Paginas de servico com slug inexistente MUST retornar HTTP 404 com a pagina de erro institucional
- **FR-009**: Quem Somos (`/quem-somos`) MUST exibir a historia do fundador Luiz Claudio com mencao a "15 anos Shell Brasil" e especialidade em redes de franquias e varejo
- **FR-010**: Quem Somos MUST exibir blocos de Missao, Visao e Valores (Transparencia, Profissionalismo, Etica, Comprometimento, Agilidade)
- **FR-011**: Quem Somos MUST mencionar a licenca CRECI (Conselho Federal/Regional de Corretores de Imoveis)
- **FR-012**: Nossos Clientes (`/nossos-clientes`) MUST exibir pelo menos 2 depoimentos com nome do autor e texto
- **FR-013**: Nossos Clientes MUST exibir uma faixa de metricas (pontos, clientes, cidades, anos)
- **FR-014**: Nossos Clientes MUST omitir a secao de logos se nao houver logos cadastrados (nao exibir grid vazio ou "em breve")
- **FR-015**: Fale Conosco (`/fale-conosco`) MUST manter o comportamento existente de validacao HTMX, persistencia no banco, e fallback no-JS (SPEC-04)
- **FR-016**: Fale Conosco MUST exibir informacoes de contato (endereco, telefones, email, WhatsApp) em destaque na pagina, nao apenas no footer
- **FR-017**: Templates de autenticacao (`login.html`, `totp_setup.html`, `totp_verify.html`) MUST incluir `<link rel="stylesheet" href="/static/css/app.css">` no `<head>`
- **FR-018**: Templates de autenticacao MUST usar classes do design system (form-input, btn, btn-primary, card, alert) em vez de estilos inline
- **FR-019**: A funcionalidade de login, 2FA TOTP, cookies de sessao, e redirects MUST permanecer identica ao SPEC-03 (apenas apresentacao visual muda)
- **FR-020**: Todas as paginas publicas MUST ser server-rendered com Go `html/template` (sem SPA, sem React)
- **FR-021**: Todas as paginas publicas MUST usar HTMX para interatividade e Alpine.js para micro-estado (ja configurado no SPEC-04)
- **FR-022**: Todas as paginas publicas MUST usar os brand tokens existentes (Deep Navy primary, Sóbrio Gold secondary, Montserrat display, Inter body, soft radius, ambient shadows) -- nao inventar nova paleta
- **FR-023**: Todas as paginas publicas MUST funcionar em viewport 375px (mobile) sem scroll horizontal
- **FR-024**: O copy de todas as paginas publicas MUST ser em pt-BR e posicionar a Prospecção Brasil como consultoria de prospeccao imobiliaria comercial (nao como software)
- **FR-025**: Nenhuma pagina publica MUST conter mencao a "software", "plataforma", "pipeline", "carga cognitiva", "gestao de pipeline", ou "relatorios PDF" como servico
- **FR-026**: O router publico (`buildPublicRouter` em `main.go`) MUST adicionar rotas `GET /servicos/{slug}` para paginas detalhadas de servicos
- **FR-027**: Designs visuais (Pencil) MUST ser criados em `designs/prospeccao.pen` antes da implementacao HTML/CSS, com frames para Home, Servicos (indice + 1 detalhe), Quem Somos, Nossos Clientes, Fale Conosco (desktop + mobile)
- **FR-028**: O nav e footer institucionais MUST ser atualizados com copy de mercado e tratamento visual alinhado ao redesign (mantendo estrutura onde sensato)
- **FR-029**: A newsletter no footer MUST permanecer funcional (SPEC-04)
- **FR-030**: Imagens de hero e conteudo MUST ser armazenadas em `static/img/` e servidas pelo static file server existente

### Key Entities *(include if feature involves data)*

- **Service**: Representa um servico de consultoria (expansao de redes, BTS, strip mall, lajes, prospeccao de ponto). Atributos: slug (URL), titulo, descricao curta, descricao longa (metodologia), etapas (lista), icone/illustracao. Nao e uma entidade de banco -- e estatica em Go (map ou struct) ou um arquivo de conteudo. Referenciado em `/servicos/{slug}`.
- **Testimonial**: Depoimento de cliente. Atributos: autor, cargo/empresa, texto. Pode ser estatico (hardcoded no template) ou dinamico (tabela `testimonials` se o escopo crescer). MVP: estatico no template.
- **Metric**: Indicador de mercado para a faixa de metricas. Atributos: rotulo, valor. Estatico no template (Pontos Comercializados, Clientes Atendidos, Cidades, Anos). Pode virar dinamico no futuro.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Todas as 5 paginas publicas (Home, Servicos, Quem Somos, Nossos Clientes, Fale Conosco) carregam em menos de 1 segundo no localhost (server-rendered HTML, sem JavaScript bloqueante)
- **SC-002**: Todas as 5 paginas publicas passam no Lighthouse "Best Practices" e "Accessibility" com score >= 90 (HTML semantico, contraste de cor, alt em imagens)
- **SC-003**: O primeiro viewport da Home transmite "consultoria imobiliaria comercial premium" em teste qualitativo: hero com fotografia, headline de mercado, metricas, CTA -- sem cards genericos de software
- **SC-004**: Nenhuma pagina publica contem as palavras "software", "plataforma", "pipeline", "carga cognitiva", ou "relatorios PDF" como servico (verificavel via grep nos templates)
- **SC-005**: As 3 paginas de autenticacao (login, 2fa setup, 2fa verify) carregam com `app.css` e usam classes do design system (sem estilos inline `style=`)
- **SC-006**: `make check` passa com 0 issues de lint, todos os testes passam, coverage >= 70% (app) / 85% (auth), build sucede, ast-grep limpo
- **SC-007**: CI no GitHub Actions passa verde apos o merge do SPEC-06
- **SC-008**: Todas as paginas funcionam em viewport 375px sem scroll horizontal (verificavel via DevTools ou teste automatizado)

## Assumptions

- O conteudo (copy, depoimentos, metricas) e adaptado do legacy site `prospeccaobrasil.com.br` e do briefing do usuario. Nao e necessario criar conteudo de marketing do zero.
- As imagens de hero e servicos podem ser placeholders iniciais (gradiente Deep Navy ou foto generica de imovel comercial). O usuario substituira por fotos reais apos o deploy. O template deve suportar a troca facil (path em `static/img/`).
- Os depoimentos (Larissa Mello, Roberto Andrade, Joao Viana) sao do legacy site e podem ser usados como estaticos no template. Se o usuario quiser depoimentos dinamicos no futuro, sera outra spec.
- As metricas (pontos comercializados, etc.) podem ter valores placeholder (ex: 0 ou "100+") ate que o usuario forneça os reais. O template exibe o valor como recebido.
- A estrutura de rotas publicas (`/`, `/quem-somos`, `/servicos`, `/nossos-clientes`, `/fale-conosco`) ja existe no SPEC-04. A nova rota `/servicos/{slug}` sera adicionada.
- O design system (tokens Tailwind, classes `.btn`, `.card`, etc.) ja existe do SPEC-04 e sera reutilizado/estendido. Nao ha necessidade de reescrever o `tailwind.config.js` ou `input.css` do zero.
- O formulario de Fale Conosco (HTMX, validacao, persistencia) ja funciona do SPEC-04. Apenas a apresentacao visual muda.
- As paginas de autenticacao (login, 2fa) foram criadas no SPEC-03 sem CSS. A correcao e apenas adicionar `app.css` e trocar estilos inline por classes do design system.
- O deploy no VPS (Hostinger) e feito pelo usuario apos o SPEC-06 estar completo. O agente nao gerencia VPS, Nginx, DNS, ou SSL.
- O Pencil MCP server esta disponivel para criar os designs em `designs/prospeccao.pen` antes da implementacao.
- A separacao de host (public vs `sistema.*`) ja esta implementada no SPEC-05. O redesign do site publico nao afeta o sistema interno.

## Non-Goals

- **Nao automatizar deploy no VPS**: O deploy (binary, Nginx, DNS, SSL) e responsabilidade do usuario. O agente garante que `make build` e `make check` passam.
- **Nao vender o software interno no dominio publico**: O site publico vende o SERVICO de prospeccao imobiliaria. O software interno (CRUD, PDF, dashboard) vive em `sistema.prospeccaobrasil.com` e nao e mencionado no site publico.
- **Nao reescrever a paleta de marca**: Deep Navy, Sóbrio Gold, Montserrat, Inter, soft radius, ambient shadows permanecem. Mudamos conteudo, layout, e tratamento visual -- nao a identidade visual.
- **Nao criar SPA ou React**: O site continua server-rendered com Go `html/template` + HTMX + Alpine.js.
- **Nao adicionar depoimentos dinamicos (tabela no banco)**: MVP usa depoimentos estaticos no template. Depoimentos dinamicos sera outra spec se necessario.
- **Nao adicionar blog ou content management**: Fora do escopo. O site e institucional, nao um blog.
- **Nao adicionar multi-idioma**: O site e pt-BR apenas.
- **Nao mudar a funcionalidade do formulario de contato**: HTMX, validacao, persistencia, fallback no-JS permanecem do SPEC-04. Apenas visual muda.
- **Nao mudar a funcionalidade de autenticacao**: Login, 2FA, cookies, redirects permanecem do SPEC-03. Apenas CSS muda.
- **Nao adicionar analytics ou tracking**: Fora do escapo do redesign visual. Pode ser adicionado em outra spec.
- **Nao criar paginas de servico dinamicas (CRUD de servicos)**: Os servicos sao estaticos (conteudo definido em Go ou template). CRUD de servicos e desnecessario para uma consultoria com 5-10 servicos fixos.

## Design References

- **Pencil file**: `designs/prospeccao.pen` -- visual source of truth para todas as paginas
- **Frames obrigatórios** (criar antes de implementar):
  - "Home - Desktop" (1440px)
  - "Home - Mobile" (375px)
  - "Servicos Index - Desktop" (1440px)
  - "Servicos Detalhe - Desktop" (1440px)
  - "Quem Somos - Desktop" (1440px)
  - "Nossos Clientes - Desktop" (1440px)
  - "Fale Conosco - Desktop" (1440px)
  - "Login - Desktop" (1440px) -- auth template com design system
- **Brand tokens**: `tailwind.config.js` e `input.css` (Deep Navy `#0D1B2A`, Sóbrio Gold `#C9A961`, Montserrat display, Inter body, soft radius `rounded-lg`, ambient shadows)
- **Layout reference**: https://www.plda.com.br/ (estrutura, hero treatment, deep service pages, social proof) -- com identidade Prospecção Brasil, nao clone literal
- **Content source**: https://prospeccaobrasil.com.br/ (copy, depoimentos, metricas, historia do fundador)
