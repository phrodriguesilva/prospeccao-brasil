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
	ActivePage   string
	Success      bool
	Form         contactForm
	Errors       contactErrors
	Testimonials []testimonial
	Metrics      []metric
	Services     []serviceDetail
	Service      *serviceDetail
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
}

// servicesList returns all services as a slice for index pages.
func servicesList() []serviceDetail {
	return []serviceDetail{
		services["expansao-de-redes"],
		services["built-to-suit"],
		services["strip-mall"],
		services["lajes-comerciais"],
		services["prospeccao-de-ponto"],
	}
}

// Home renders the home page at GET /.
func (h *InstitutionalHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "home.html", pageData{
		ActivePage:   "home",
		Testimonials: defaultTestimonials,
		Metrics:      defaultMetrics,
		Services:     servicesList(),
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

// renderPage renders a page using the template.
func (h *InstitutionalHandler) renderPage(w http.ResponseWriter, name string, data pageData) {
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		h.log.Error("render page", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
