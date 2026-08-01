package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
	"prospeccaobrasil/internal/handler"
)

// healthResponse is the body returned by GET /healthz.
type healthResponse struct {
	Status string `json:"status"`
}

// healthHandler returns 200 {"status":"ok"} for liveness probes.
// It is public (no auth) per Constitution principle V.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

// loadEncryptionKey reads ENCRYPTION_KEY from env (base64-encoded 32 bytes)
// and returns the raw 32-byte key. Falls back to deriving from the raw string
// if not base64 (for dev convenience).
func loadEncryptionKey() ([]byte, error) {
	enc := os.Getenv("ENCRYPTION_KEY")
	if enc == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY not set")
	}
	// Try base64 decode first
	if key, err := base64.StdEncoding.DecodeString(enc); err == nil && len(key) == 32 {
		return key, nil
	}
	// Fall back: if the raw string is 32 bytes, use it directly
	if len(enc) == 32 {
		return []byte(enc), nil
	}
	// Last resort: pad/truncate to 32 bytes (dev only)
	if len(enc) < 32 {
		padded := make([]byte, 32)
		copy(padded, enc)
		return padded, nil
	}
	return []byte(enc[:32]), nil
}

// loadTemplates parses all HTML templates from internal/template/ and subdirectories.
func loadTemplates() (*template.Template, error) {
	templateDir := filepath.Join("internal", "template")
	// Parse all templates in one pass using a glob pattern that matches all HTML files
	// in the template directory and its subdirectories.
	tmpl, err := template.New("").Funcs(handler.TemplateFuncs()).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	// Parse partials
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "partials", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	// Parse fragments
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "fragments", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse fragments: %w", err)
	}
	// Parse admin templates (recursive)
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin templates: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "properties", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/properties templates: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "clients", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/clients templates: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "prospections", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/prospections templates: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "contacts", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/contacts templates: %w", err)
	}
	return tmpl, nil
}

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	// Structured logging via slog (Constitution principle V).
	var logHandler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	log := slog.New(logHandler)
	slog.SetDefault(log)

	// Load encryption key for TOTP secret encryption
	encKey, err := loadEncryptionKey()
	if err != nil {
		log.Error("startup: encryption key", "error", err)
		os.Exit(1)
	}

	// Load HMAC key for pending session cookies (reuse ENCRYPTION_KEY)
	hmacKey := encKey

	// Determine cookie Secure flag from APP_BASE_URL
	appBaseURL := os.Getenv("APP_BASE_URL")
	secure := auth.IsSecure(appBaseURL)

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Error("startup: DATABASE_URL not set")
		os.Exit(1)
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Error("startup: pgxpool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)

	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		log.Error("startup: load templates", "error", err)
		os.Exit(1)
	}

	// Initialize auth service + handler
	limiter := auth.NewRateLimiter()
	svc := auth.NewService(queries, pool, encKey, limiter, log)
	authHandler := handler.NewAuthHandler(svc, queries, tmpl, log, secure, hmacKey)
	institutionalHandler := handler.NewInstitutionalHandler(queries, tmpl, log)
	contactHandler := handler.NewContactHandler(queries, tmpl, log, limiter)
	newsletterHandler := handler.NewNewsletterHandler(queries, tmpl, log, limiter)
	dashboardHandler := handler.NewDashboardHandler(queries, tmpl, log)
	propertyHandler := handler.NewPropertyHandler(queries, tmpl, log)
	clientHandler := handler.NewClientHandler(queries, tmpl, log)
	prospectionHandler := handler.NewProspectionHandler(queries, tmpl, log)
	pdfHandler := handler.NewPDFHandler(queries, tmpl, log)

	// Build host-based router dispatcher.
	// sistema.prospeccaobrasil.com -> internal system only
	// prospeccaobrasil.com / .com.br -> public institutional site only
	// localhost / unknown -> dev mode (serves everything, for local dev + tests)
	publicRouter := buildPublicRouter(institutionalHandler, contactHandler, newsletterHandler)
	internalRouter := buildInternalRouter(authHandler, dashboardHandler, propertyHandler, clientHandler, prospectionHandler, pdfHandler, queries, log)
	devRouter := buildDevRouter(institutionalHandler, contactHandler, newsletterHandler, authHandler, dashboardHandler, propertyHandler, clientHandler, prospectionHandler, pdfHandler, queries, log)

	topHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(strings.SplitN(r.Host, ":", 2)[0])
		switch {
		case strings.HasPrefix(host, "sistema."):
			internalRouter.ServeHTTP(w, r)
		case isPublicDomain(host):
			publicRouter.ServeHTTP(w, r)
		default:
			// localhost, 127.0.0.1, empty host (tests), unknown -- dev mode
			devRouter.ServeHTTP(w, r)
		}
	})

	// HTTP server with graceful shutdown
	server := &http.Server{
		Addr:              addr,
		Handler:           topHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("starting server", "addr", addr, "secure", secure)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-done
	log.Info("shutdown signal received", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown", "error", err)
	}
	log.Info("server stopped")
}

// isPublicDomain returns true for domains that should serve the institutional site.
func isPublicDomain(host string) bool {
	publicDomains := []string{
		"prospeccaobrasil.com",
		"www.prospeccaobrasil.com",
		"prospeccaobrasil.com.br",
		"www.prospeccaobrasil.com.br",
	}
	for _, d := range publicDomains {
		if host == d {
			return true
		}
	}
	return false
}

// staticFileServer returns a handler that serves static files from the static/ directory.
func staticFileServer() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
}

// buildPublicRouter builds the router for the public institutional site.
// Only institutional pages, contact form, newsletter, and healthz are served.
func buildPublicRouter(
	instHandler *handler.InstitutionalHandler,
	contactHandler *handler.ContactHandler,
	newsletterHandler *handler.NewsletterHandler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Handle("/static/*", staticFileServer())
	r.Get("/healthz", healthHandler)
	r.Get("/", instHandler.Home)
	r.Get("/quem-somos", instHandler.QuemSomos)
	r.Get("/servicos", instHandler.Servicos)
	r.Get("/servicos/{slug}", instHandler.ServicoDetalhe)
	r.Get("/nossos-clientes", instHandler.NossosClientes)
	r.Get("/fale-conosco", instHandler.FaleConosco)
	r.Post("/fale-conosco", contactHandler.Submit)
	r.Post("/newsletter", newsletterHandler.Subscribe)
	r.NotFound(instHandler.NotFound)
	return r
}

// buildInternalRouter builds the router for the internal system (sistema.* subdomain).
// Only auth, admin, CRUD, and healthz are served. Institutional pages return 404.
func buildInternalRouter(
	authHandler *handler.AuthHandler,
	dashboardHandler *handler.DashboardHandler,
	propertyHandler *handler.PropertyHandler,
	clientHandler *handler.ClientHandler,
	prospectionHandler *handler.ProspectionHandler,
	pdfHandler *handler.PDFHandler,
	queries *db.Queries,
	log *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Handle("/static/*", staticFileServer())
	r.Get("/healthz", healthHandler)
	// Root redirect: / -> /admin (which redirects to /login if not authenticated)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})
	// Auth routes (public, no session required)
	r.Get("/login", authHandler.LoginGET)
	r.Post("/login", authHandler.LoginPOST)
	r.Get("/2fa/setup", authHandler.TotpSetupGET)
	r.Post("/2fa/setup", authHandler.TotpSetupPOST)
	r.Get("/2fa/verify", authHandler.TotpVerifyGET)
	r.Post("/2fa/verify", authHandler.TotpVerifyPOST)
	// Protected routes (session required + admin role)
	r.Group(func(r chi.Router) {
		r.Use(handler.SessionValidation(queries, log))
		r.Post("/logout", authHandler.LogoutPOST)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(auth.RoleAdmin))
			// Dashboard
			r.Get("/admin", dashboardHandler.Index)
			// Properties CRUD
			r.Get("/properties", propertyHandler.List)
			r.Get("/properties/new", propertyHandler.New)
			r.Post("/properties", propertyHandler.Create)
			r.Get("/properties/{id}", propertyHandler.Detail)
			r.Get("/properties/{id}/edit", propertyHandler.Edit)
			r.Post("/properties/{id}", propertyHandler.Update)
			r.Post("/properties/{id}/delete", propertyHandler.Delete)
			r.Get("/properties/{id}/pdf", pdfHandler.GeneratePropertyPDF)
			// Clients CRUD
			r.Get("/clients", clientHandler.List)
			r.Get("/clients/new", clientHandler.New)
			r.Post("/clients", clientHandler.Create)
			r.Get("/clients/{id}", clientHandler.Detail)
			r.Get("/clients/{id}/edit", clientHandler.Edit)
			r.Post("/clients/{id}", clientHandler.Update)
			r.Post("/clients/{id}/delete", clientHandler.Delete)
			r.Post("/clients/{id}/contacts", clientHandler.CreateContact)
			// Prospections CRUD
			r.Get("/prospections", prospectionHandler.List)
			r.Get("/prospections/new", prospectionHandler.New)
			r.Post("/prospections", prospectionHandler.Create)
			r.Get("/prospections/{id}", prospectionHandler.Detail)
			r.Get("/prospections/{id}/edit", prospectionHandler.Edit)
			r.Post("/prospections/{id}", prospectionHandler.Update)
			r.Post("/prospections/{id}/delete", prospectionHandler.Delete)
			r.Post("/prospections/{id}/contacts", prospectionHandler.CreateContact)
		})
	})
	// 404 for any non-internal route (e.g., institutional pages on sistema.*)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return r
}

// buildDevRouter builds a router that serves everything (public + internal).
// Used for localhost development and tests.
func buildDevRouter(
	instHandler *handler.InstitutionalHandler,
	contactHandler *handler.ContactHandler,
	newsletterHandler *handler.NewsletterHandler,
	authHandler *handler.AuthHandler,
	dashboardHandler *handler.DashboardHandler,
	propertyHandler *handler.PropertyHandler,
	clientHandler *handler.ClientHandler,
	prospectionHandler *handler.ProspectionHandler,
	pdfHandler *handler.PDFHandler,
	queries *db.Queries,
	log *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Handle("/static/*", staticFileServer())

	// Public group (no auth) -- institutional site + auth
	r.Group(func(r chi.Router) {
		r.Get("/healthz", healthHandler)
		r.Get("/", instHandler.Home)
		r.Get("/quem-somos", instHandler.QuemSomos)
		r.Get("/servicos", instHandler.Servicos)
		r.Get("/servicos/{slug}", instHandler.ServicoDetalhe)
		r.Get("/nossos-clientes", instHandler.NossosClientes)
		r.Get("/fale-conosco", instHandler.FaleConosco)
		r.Post("/fale-conosco", contactHandler.Submit)
		r.Post("/newsletter", newsletterHandler.Subscribe)

		r.Get("/login", authHandler.LoginGET)
		r.Post("/login", authHandler.LoginPOST)
		r.Get("/2fa/setup", authHandler.TotpSetupGET)
		r.Post("/2fa/setup", authHandler.TotpSetupPOST)
		r.Get("/2fa/verify", authHandler.TotpVerifyGET)
		r.Post("/2fa/verify", authHandler.TotpVerifyPOST)
	})

	r.NotFound(instHandler.NotFound)

	// Protected group (auth required + admin role) -- internal system
	r.Group(func(r chi.Router) {
		r.Use(handler.SessionValidation(queries, log))
		r.Post("/logout", authHandler.LogoutPOST)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(auth.RoleAdmin))
			// Dashboard
			r.Get("/admin", dashboardHandler.Index)
			// Properties CRUD
			r.Get("/properties", propertyHandler.List)
			r.Get("/properties/new", propertyHandler.New)
			r.Post("/properties", propertyHandler.Create)
			r.Get("/properties/{id}", propertyHandler.Detail)
			r.Get("/properties/{id}/edit", propertyHandler.Edit)
			r.Post("/properties/{id}", propertyHandler.Update)
			r.Post("/properties/{id}/delete", propertyHandler.Delete)
			r.Get("/properties/{id}/pdf", pdfHandler.GeneratePropertyPDF)
			// Clients CRUD
			r.Get("/clients", clientHandler.List)
			r.Get("/clients/new", clientHandler.New)
			r.Post("/clients", clientHandler.Create)
			r.Get("/clients/{id}", clientHandler.Detail)
			r.Get("/clients/{id}/edit", clientHandler.Edit)
			r.Post("/clients/{id}", clientHandler.Update)
			r.Post("/clients/{id}/delete", clientHandler.Delete)
			r.Post("/clients/{id}/contacts", clientHandler.CreateContact)
			// Prospections CRUD
			r.Get("/prospections", prospectionHandler.List)
			r.Get("/prospections/new", prospectionHandler.New)
			r.Post("/prospections", prospectionHandler.Create)
			r.Get("/prospections/{id}", prospectionHandler.Detail)
			r.Get("/prospections/{id}/edit", prospectionHandler.Edit)
			r.Post("/prospections/{id}", prospectionHandler.Update)
			r.Post("/prospections/{id}/delete", prospectionHandler.Delete)
			r.Post("/prospections/{id}/contacts", prospectionHandler.CreateContact)
		})
	})
	return r
}
