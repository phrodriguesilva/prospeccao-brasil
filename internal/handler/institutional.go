package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"prospeccaobrasil/internal/db"
)

// InstitutionalHandler handles public institutional site pages.
// All pages are public (no auth required).
type InstitutionalHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewInstitutionalHandler creates a new InstitutionalHandler.
func NewInstitutionalHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *InstitutionalHandler {
	return &InstitutionalHandler{
		queries: queries,
		tmpl:    tmpl,
		log:     log,
	}
}

// pageData is the base data passed to all institutional templates.
type pageData struct {
	ActivePage       string
	Success          bool
	Form             contactForm
	Errors           contactErrors
	Testimonials     []testimonial
	Metrics          []metric
	Services         []serviceDetail
	Service          *serviceDetail
	Segments         []segmentDetail
	TeamMembers      []teamMember
	Partners         []partnerCategory
	ClientLogos      []clientLogo
	Results          []workResult
	SocialCauses     []socialCause
	InvestorServices []serviceDetail
}

type contactForm struct {
	Company string
	Name    string
	Email   string
	Phone   string
	Subject string
	Message string
}

type contactErrors struct {
	Name    string
	Email   string
	Subject string
	Message string
	Generic string
}

type testimonial struct {
	Name    string
	Company string
	Quote   string
	Metric  string
}

type metric struct {
	Label string
	Value string
}

type serviceDetail struct {
	Slug        string
	Title       string
	Summary     string
	Description string
	Methodology []string
	CTA         string
}

type segmentDetail struct {
	Slug        string
	Title       string
	Icon        string
	Description string
}

type teamMember struct {
	Name     string
	Role     string
	Bio      string
	Initials string
}

type partnerCategory struct {
	Title    string
	Partners []string
}

type clientLogo struct {
	Name string
	Path string
}

type workResult struct {
	Name string
	Path string
}

type socialCause struct {
	Name        string
	Description string
	URL         string
}

// defaultTestimonials contains the 3 testimonials from the legacy site.
var defaultTestimonials = []testimonial{
	{
		Name:    "Larissa Mello",
		Company: "Rede de Farmácias",
		Quote:   "A Prospecção Brasil selecionou os melhores pontos comerciais. Eles foram extremamente úteis para ajudar-nos a identificar novos mercados e pontos de venda.",
	},
	{
		Name:    "Roberto Andrade",
		Company: "Expansão de Redes",
		Quote:   "A empresa tem nos ajudado a encontrar novos parceiros de negócios que se alinharam com nossas metas de crescimento.",
	},
	{
		Name:    "João Viana",
		Company: "Investidor Imobiliário",
		Quote:   "A Prospecção Brasil tem fornecido informações preciosas sobre mercados-alvo que ajudaram a acelerar nosso crescimento.",
	},
}

// defaultMetrics contains the 4 market metrics for the metrics strip.
var defaultMetrics = []metric{
	{Label: "Pontos Comercializados", Value: "500+"},
	{Label: "Clientes Atendidos", Value: "50+"},
	{Label: "Cidades Atendidas", Value: "100+"},
	{Label: "Anos de Mercado", Value: "15+"},
}

// services contains the 5 service detail pages.
var services = map[string]serviceDetail{
	"expansao-de-redes": {
		Slug:        "expansao-de-redes",
		Title:       "Expansão de Redes",
		Summary:     "Planejamento e execução da expansão de redes varejistas e franquias em todo o Brasil.",
		Description: "A dinâmica do mercado exige que estejamos atentos às demandas. Auxiliamos empresas varejistas ou não, desde o planejamento até a implantação do negócio, expansão, definição de mercado alvo, valores e estudo de rentabilidade. Trabalhamos com redes de franquias e varejo, em qualquer tipo de ponto comercial: lojas de rua, calçadão de bairros, shopping centers, etc.",
		Methodology: []string{
			"Plano Diretor para Expansão: níveis geográficos de análise (Brasil, estados, micro e meso regiões, municípios, bairros e ponto)",
			"Ranqueamento do potencial de mercado e identificação de vácuos de atuação",
			"Desenvolvimento de cronograma de abertura e definição de territórios",
			"Análise de macrolocalização: população, renda, polos correlatos, concorrência",
			"Análise de microlocalização: atratividade, acessibilidade, visibilidade, requisitos físicos",
			"Evitar canibalização entre pontos da rede",
		},
		CTA: "Fale com um especialista",
	},
	"built-to-suit": {
		Slug:        "built-to-suit",
		Title:       "Built to Suit (BTS)",
		Summary:     "Projetos sob medida para aquisição ou desenvolvimento de imóveis comerciais.",
		Description: "Atuamos para aquisição ou venda de terrenos, shopping centers, pontos comerciais e empreendimentos imobiliários. Buscamos operações de permuta, desenvolvimento ou outras opções que possibilitem a melhor decisão a ser tomada por empresários, executivos, empreendedores e lojistas.",
		Methodology: []string{
			"Identificação de terrenos e imóveis com potencial para desenvolvimento BTS",
			"Análise de viabilidade técnica e comercial do empreendimento",
			"Estruturação de operações de permuta e desenvolvimento",
			"Negociação com proprietários, fundos de investimento e incorporadores",
			"Acompanhamento da implantação até a entrega do ponto",
		},
		CTA: "Fale com um especialista",
	},
	"strip-mall": {
		Slug:        "strip-mall",
		Title:       "Strip Mall / Centros de Conveniência",
		Summary:     "Desenvolvimento e gestão de centros comerciais de bairro e strip malls.",
		Description: "O entendimento que a chave do sucesso para qualquer ponto comercial, seja loja de rua ou shopping center, é um planejamento da comercialização que leve em conta a situação mercadológica, os hábitos dos consumidores locais e que atenda ao mix adequado. Aplicamos essa visão ao desenvolvimento e gestão de strip malls e centros de conveniência.",
		Methodology: []string{
			"Planejamento de mix de lojas âncoras e satélites",
			"Análise de potencial de consumo da região de influência",
			"Definição de estratégia de comercialização e posicionamento",
			"Seleção e negociação com operadores de redes varejistas",
			"Gestão de occupancy e otimização do fluxo de clientes",
		},
		CTA: "Fale com um especialista",
	},
	"lajes-comerciais": {
		Slug:        "lajes-comerciais",
		Title:       "Lajes Comerciais",
		Summary:     "Locação e comercialização de lajes corporativas para empresas e investidores.",
		Description: "Atuamos na locação e comercialização de lajes corporativas, conectando proprietários de imóveis a empresas que buscam espaços de escritório e lajes comerciais de alto padrão. Nossa expertise em real estate comercial garante a melhor decisão para ambas as partes.",
		Methodology: []string{
			"Levantamento de lajes disponíveis e análise de potencial comercial",
			"Estudo de viabilidade e precificação de mercado",
			"Abordagem a empresas-alvo e investidores",
			"Negociação de valores locatícios e condições contratuais",
			"Suporte documental e acompanhamento da contratação",
		},
		CTA: "Fale com um especialista",
	},
	"prospeccao-de-ponto": {
		Slug:        "prospeccao-de-ponto",
		Title:       "Prospecção de Ponto Comercial",
		Summary:     "Identificação e negociação do ponto comercial ideal para o seu negócio.",
		Description: "Definimos em conjunto os requisitos físicos e mercadológicos ideais para o seu negócio. Realizamos o levantamento de valores comerciais praticados na região alvo, abordagem em imóveis de interesse (vagos e ocupados), avaliação dos imóveis alvo, e contato e negociação com proprietários e imobiliárias.",
		Methodology: []string{
			"Definição dos requisitos físicos e mercadológicos ideais em conjunto com o cliente",
			"Levantamento de valores comerciais praticados na região alvo",
			"Abordagem em imóveis de interesse (vagos e ocupados)",
			"Avaliação dos imóveis alvo com base em fluxo, acessibilidade e visibilidade",
			"Contato e negociação com proprietários e imobiliárias",
			"Suporte documental para viabilizar as normas contratuais",
		},
		CTA: "Fale com um especialista",
	},
	"conselho-consultivo": {
		Slug:        "conselho-consultivo",
		Title:       "Conselho Consultivo para Expansão",
		Summary:     "Assessoria estratégica de especialistas em expansão de lojas e franquias.",
		Description: "Para empresas que estão iniciando seu processo de expansão ou que precisam profissionalizar seu time de expansão. Compartilhamos know-how de mais de 15 anos em real estate comercial, ajudando a responder perguntas-chave: por onde começar? Crescer conforme oportunidades ou com plano estratégico? Foco em seeding strategy ou acelerar capilaridade? Shopping ou rua? Em qual corredor, perto de quais lojas?",
		Methodology: []string{
			"Diagnóstico do estágio atual de expansão da empresa",
			"Definição de estratégia de crescimento (seeding vs. aceleração)",
			"Profissionalização do time de expansão e criação de comitê",
			"Aplicação de inteligência de geomarketing na tomada de decisão",
			"Avaliação de novas praças e mercados-alvo",
			"Acompanhamento contínuo como conselheiros estratégicos",
		},
		CTA: "Fale com um especialista",
	},
	"sale-leaseback": {
		Slug:        "sale-leaseback",
		Title:       "Operações de Sale & Leaseback",
		Summary:     "Venda do imóvel com locação de volta ao proprietário, liberando capital para a operação.",
		Description: "Sale & Leaseback é uma operação em que o imóvel é vendido e simultaneamente locado de volta ao antigo proprietário por meio de contrato de longo prazo. Esta estrutura permite que a empresa reinvesta o capital da venda em sua operação principal, expandindo e modernizando sua estrutura com o capital liberado.",
		Methodology: []string{
			"Avaliação do ativo imobiliário e estruturação da operação",
			"Identificação de investidores e fundos interessados",
			"Negociação de condições de venda e contratos de locação",
			"Suporte jurídico e tributário durante todo o processo",
			"Acompanhamento pós-operação para garantir a continuidade",
		},
		CTA: "Fale com um especialista",
	},
	"transferencia-de-pontos": {
		Slug:        "transferencia-de-pontos",
		Title:       "Transferência de Pontos Comerciais",
		Summary:     "Intermediação na transferência e repasse de pontos entre redes varejistas.",
		Description: "Auxiliamos na transferência de pontos comerciais entre redes, seja para desmobilização de ativos ou para realocação estratégica. Cuidamos de toda a negociação entre as partes, garantindo que o ponto encontre o operador certo e que o transition seja suave e sem perdas para nenhuma das partes envolvidas.",
		Methodology: []string{
			"Levantamento de pontos disponíveis para transferência",
			"Abordagem a redes interessadas no perfil do ponto",
			"Negociação de valores de repasse e condições contratuais",
			"Coordenação da transição entre operadores",
			"Suporte documental para formalização da transferência",
		},
		CTA: "Fale com um especialista",
	},
	"ingresso-em-mercados": {
		Slug:        "ingresso-em-mercados",
		Title:       "Ingresso em Novos Mercados",
		Summary:     "Estratégia e suporte para redes que buscam entrar em novas praças e regiões.",
		Description: "Entrar em um novo mercado vai muito além de encontrar o ponto comercial indicado por software de geomarketing. O Brasil possui aspectos culturais relevantes em suas microrregiões que influenciam diretamente no sucesso ou fracasso de uma operação. Nossa experiência em múltiplas praças permite identificar não apenas o ponto certo, mas também os ajustes de formato, mix e posicionamento necessários para cada mercado.",
		Methodology: []string{
			"Análise cultural e comportamental da praça-alvo",
			"Estudo de concorrência local e vácuos de mercado",
			"Adaptação de formato e mix para a realidade regional",
			"Identificação e negociação do ponto comercial ideal",
			"Suporte na implantação e primeiros meses de operação",
		},
		CTA: "Fale com um especialista",
	},
	// Investidores e Family Offices
	"administracao-de-portfolios": {
		Slug:        "administracao-de-portfolios",
		Title:       "Administração de Portfólios Imobiliários",
		Summary:     "Gestão profissional de portfólios imobiliários para investidores e family offices.",
		Description: "Oferecemos gestão profissional de portfólios imobiliários para investidores e family offices que buscam maximizar a rentabilidade de seus ativos comerciais. Cuidamos da locação, renovação, reajuste e desmobilização de imóveis, garantindo a melhor relação risco-retorno para o portfólio.",
		Methodology: []string{
			"Diagnóstico do portfólio atual e identificação de oportunidades",
			"Plano de otimização: locação, renovação, reajuste ou desmobilização",
			"Abordagem a operadores e redes para locação de imóveis vagos",
			"Gestão de contratos e acompanhamento de occupancy",
			"Relatórios periódicos de performance e recomendações estratégicas",
		},
		CTA: "Fale com um especialista",
	},
	"estruturacao-de-contratos": {
		Slug:        "estruturacao-de-contratos",
		Title:       "Estruturação de Contratos Atípicos",
		Summary:     "Estruturação de contratos complexos e atípicos para operações imobiliárias.",
		Description: "Estruturamos contratos atípicos para operações imobiliárias que exigem soluções sob medida: permuta, sale & leaseback, built to suit, joint ventures, dentre outras. Trabalhamos em conjunto com assessoria jurídica para garantir segurança jurídica e tributária em operações complexas.",
		Methodology: []string{
			"Identificação da estrutura contratual mais adequada à operação",
			"Coordenação com assessoria jurídica para redação dos contratos",
			"Negociação de cláusulas e condições com as partes envolvidas",
			"Suporte tributário para otimização da operação",
			"Acompanhamento até a formalização e registro",
		},
		CTA: "Fale com um especialista",
	},
	"avaliacao-estrategica": {
		Slug:        "avaliacao-estrategica",
		Title:       "Avaliação Estratégica de Ativo Imobiliário",
		Summary:     "Avaliação estratégica de ativos imobiliários comerciais para tomada de decisão.",
		Description: "Realizamos avaliações estratégicas de ativos imobiliários comerciais para investidores, proprietários e empresas que precisam de uma visão técnica e de mercado para tomar decisões de manter, vender, locar ou desenvolver. Nossa avaliação combina análise de mercado, comparáveis e potencial de rentabilidade.",
		Methodology: []string{
			"Vistoria técnica e levantamento de características do imóvel",
			"Análise de mercado: comparáveis, tendências e potencial da região",
			"Avaliação de potencial de rentabilidade e retorno sobre investimento",
			"Cenários: manter, vender, locar ou desenvolver",
			"Relatório de avaliação com recomendações estratégicas",
		},
		CTA: "Fale com um especialista",
	},
}

// servicesList returns all regular services as a slice for index pages.
func servicesList() []serviceDetail {
	return []serviceDetail{
		services["expansao-de-redes"],
		services["prospeccao-de-ponto"],
		services["built-to-suit"],
		services["strip-mall"],
		services["lajes-comerciais"],
		services["conselho-consultivo"],
		services["sale-leaseback"],
		services["transferencia-de-pontos"],
		services["ingresso-em-mercados"],
	}
}

// investorServicesList returns the 3 investor-focused services.
func investorServicesList() []serviceDetail {
	return []serviceDetail{
		services["administracao-de-portfolios"],
		services["estruturacao-de-contratos"],
		services["avaliacao-estrategica"],
	}
}

// segmentsList returns all segments of activity.
var segmentsList = []segmentDetail{
	{Slug: "fast-food", Title: "Fast Food", Icon: "burger", Description: "Redes de alimentação rápida e QSR, com foco em fluxo de pessoas e acessibilidade."},
	{Slug: "supermercados", Title: "Supermercados e Atacarejos", Icon: "cart", Description: "Grandes superfícies varejistas de alimentação, com requisitos de estacionamento e logística."},
	{Slug: "farmacias", Title: "Farmácias", Icon: "health", Description: "Redes de farmácias e drogarias, com foco em densidade demográfica e competição local."},
	{Slug: "eletro-moveis", Title: "Eletroeletrônicos e Móveis", Icon: "tv", Description: "Redes de eletrônicos, móveis e eletrodomésticos, com requisitos de área e visibilidade."},
	{Slug: "petshops", Title: "Petshops", Icon: "pet", Description: "Redes de pet shops e veterinárias, em expansão acelerada no varejo brasileiro."},
	{Slug: "bazar", Title: "Bazar e Varejo Geral", Icon: "store", Description: "Redes de bazar, variedades e varejo geral, com mix diversificado de produtos."},
	{Slug: "servicos-financeiros", Title: "Serviços Financeiros", Icon: "bank", Description: "Bancos, financeiras e correspondentes, com foco em segurança e fluxo de clientes."},
	{Slug: "cafeterias", Title: "Cafeterias", Icon: "coffee", Description: "Redes de cafeterias e coffee shops, com foco em localização premium e público qualificado."},
	{Slug: "outros", Title: "Outros Segmentos", Icon: "grid", Description: "Moda, beleza, serviços, academias e demais segmentos do varejo e serviços."},
}

// defaultTeamMembers contains the team members for the Equipe page.
var defaultTeamMembers = []teamMember{
	{Name: "Luiz Claudio P. André", Role: "CEO e Fundador", Bio: "15 anos de experiência na área administrativa, tributária e comercial da Shell Brasil. Especialista em redes de franquias e varejo, com expertise em prospecção de pontos comerciais em todo o Brasil.", Initials: "LC"},
}

// defaultPartners contains partner categories for the Parceiros page.
var defaultPartners = []partnerCategory{
	{Title: "Governança e Gestão", Partners: []string{"Associações e conselhos profissionais", "Escritórios de advocacia imobiliária", "Contabilidade e tributação"}},
	{Title: "Associacoes", Partners: []string{"CRECI-RJ", "Associação do mercado imobiliário", "Fóruns de real estate e varejo"}},
	{Title: "Parceiros Comerciais", Partners: []string{"Arquitetos e engenheiros", "Empresas de geomarketing", "Consultorias de varejo"}},
}

// defaultClientLogos contains the client logos for the Parceiros page.
var defaultClientLogos = []clientLogo{
	{Name: "Burger King", Path: "/static/img/clients/burger-king.png"},
	{Name: "Habib's", Path: "/static/img/clients/habibs.png"},
	{Name: "Ragazzo", Path: "/static/img/clients/ragazzo.png"},
	{Name: "Drogarias Pacheco", Path: "/static/img/clients/drogarias-pacheco.png"},
	{Name: "Casa & Video", Path: "/static/img/clients/casa-video.png"},
	{Name: "American Pet", Path: "/static/img/clients/american-pet.png"},
	{Name: "TIM", Path: "/static/img/clients/tim.png"},
	{Name: "Opticas Carol", Path: "/static/img/clients/oticas-carol.png"},
	{Name: "Rio Bel Fidelidade", Path: "/static/img/clients/rio-bel.png"},
	{Name: "O Amigao", Path: "/static/img/clients/o-amigao.jpg"},
	{Name: "Casa & Lazer", Path: "/static/img/clients/casa-lazer.png"},
}

// defaultResults contains work result images for the Parceiros page.
var defaultResults = []workResult{
	{Name: "American Pet", Path: "/static/img/results/american-pet-1.png"},
	{Name: "Drogaria Moderna", Path: "/static/img/results/drogaria-moderna.png"},
	{Name: "Pernambucanas", Path: "/static/img/results/pernambucanas.png"},
	{Name: "Della & Delle", Path: "/static/img/results/della-delle.png"},
	{Name: "American Pet", Path: "/static/img/results/american-pet-2.png"},
	{Name: "Drogarias Pacheco", Path: "/static/img/results/drogarias-pacheco.png"},
	{Name: "Monamie Cosmeticos", Path: "/static/img/results/monamie-cosmeticos.png"},
	{Name: "Loja TIM", Path: "/static/img/results/loja-tim.png"},
	{Name: "Casa & Lazer", Path: "/static/img/results/casa-lazer.png"},
	{Name: "RiHappy", Path: "/static/img/results/rihappy.png"},
}

// defaultSocialCauses contains the social causes for the Responsabilidade Social page.
var defaultSocialCauses = []socialCause{
	{Name: "Gerando Falcões", Description: "ONG que transforma favelas em polos de desenvolvimento social e econômico, impactando milhares de famílias em todo o Brasil.", URL: "https://gerandofalcoes.com"},
	{Name: "Fundação o Pão dos Pobres", Description: "Instituição que oferece educação profissionalizante para jovens em situação de vulnerabilidade social.", URL: "https://www.paodospobres.org.br"},
	{Name: "CPCA Lomba do Pinheiro", Description: "Centro profissionalizante que atende crianças e jovens em comunidade vulnerável, preparando para o mercado de trabalho.", URL: "https://pod.rs.gov.br/lomba-do-pinheiro"},
}

// Home renders the home page at GET /.
func (h *InstitutionalHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "home.html", pageData{
		ActivePage:   "home",
		Testimonials: defaultTestimonials,
		Metrics:      defaultMetrics,
		Services:     servicesList(),
		Segments:     segmentsList,
	})
}

// QuemSomos renders the "Quem somos" page at GET /quem-somos.
func (h *InstitutionalHandler) QuemSomos(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "quem-somos.html", pageData{ActivePage: "quem-somos"})
}

// Servicos renders the "Servicos" index page at GET /servicos.
func (h *InstitutionalHandler) Servicos(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "servicos.html", pageData{
		ActivePage: "servicos",
		Services:   servicesList(),
	})
}

// ServicoDetalhe renders a service detail page at GET /servicos/{slug}.
func (h *InstitutionalHandler) ServicoDetalhe(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	detail, ok := services[slug]
	if !ok {
		h.NotFound(w, r)
		return
	}
	h.renderPage(w, "servico-detalhe.html", pageData{
		ActivePage: "servicos",
		Service:    &detail,
	})
}

// NossosClientes renders the "Nossos clientes" page at GET /nossos-clientes.
func (h *InstitutionalHandler) NossosClientes(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "nossos-clientes.html", pageData{
		ActivePage:   "nossos-clientes",
		Testimonials: defaultTestimonials,
		Metrics:      defaultMetrics,
	})
}

// FaleConosco renders the "Fale Conosco" page at GET /fale-conosco.
func (h *InstitutionalHandler) FaleConosco(w http.ResponseWriter, r *http.Request) {
	data := pageData{ActivePage: "fale-conosco"}
	if r.URL.Query().Get("success") == "1" {
		data.Success = true
	}
	h.renderPage(w, "fale-conosco.html", data)
}

// NotFound renders the 404 page for unmatched routes.
func (h *InstitutionalHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.renderPage(w, "404.html", pageData{})
}

// Privacidade renders the LGPD privacy policy page at GET /privacidade.
func (h *InstitutionalHandler) Privacidade(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "privacidade.html", pageData{ActivePage: "privacidade"})
}

// Segmentos renders the "Segmentos de Atuacao" page at GET /segmentos.
func (h *InstitutionalHandler) Segmentos(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "segmentos.html", pageData{
		ActivePage: "segmentos",
		Segments:   segmentsList,
	})
}

// Equipe renders the "Equipe" page at GET /equipe.
func (h *InstitutionalHandler) Equipe(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "equipe.html", pageData{
		ActivePage:  "equipe",
		TeamMembers: defaultTeamMembers,
	})
}

// Parceiros renders the "Empresas Parceiras" page at GET /parceiros.
func (h *InstitutionalHandler) Parceiros(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "parceiros.html", pageData{
		ActivePage:  "parceiros",
		Partners:    defaultPartners,
		ClientLogos: defaultClientLogos,
		Results:     defaultResults,
	})
}

// ResponsabilidadeSocial renders the "Responsabilidade Social" page at GET /responsabilidade-social.
func (h *InstitutionalHandler) ResponsabilidadeSocial(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "responsabilidade-social.html", pageData{
		ActivePage:   "responsabilidade-social",
		SocialCauses: defaultSocialCauses,
	})
}

// Investidores renders the "Investidores e Family Offices" page at GET /investidores.
func (h *InstitutionalHandler) Investidores(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "investidores.html", pageData{
		ActivePage:       "investidores",
		InvestorServices: investorServicesList(),
	})
}

// renderPage renders a page using the template.
func (h *InstitutionalHandler) renderPage(w http.ResponseWriter, name string, data pageData) {
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		h.log.Error("render page", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
