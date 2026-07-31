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
	tmpl, err := template.New("").ParseGlob(filepath.Join(templateDir, "*.html"))
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

	// Build router
	r := chi.NewRouter()

	// Static files (self-hosted JS/CSS)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public group (no auth) -- institutional site
	r.Group(func(r chi.Router) {
		r.Get("/", institutionalHandler.Home)
		r.Get("/quem-somos", institutionalHandler.QuemSomos)
		r.Get("/servicos", institutionalHandler.Servicos)
		r.Get("/nossos-clientes", institutionalHandler.NossosClientes)
		r.Get("/fale-conosco", institutionalHandler.FaleConosco)
		r.Post("/fale-conosco", contactHandler.Submit)
		r.Post("/newsletter", newsletterHandler.Subscribe)

		r.Get("/healthz", healthHandler)
		r.Get("/login", authHandler.LoginGET)
		r.Post("/login", authHandler.LoginPOST)
		r.Get("/2fa/setup", authHandler.TotpSetupGET)
		r.Post("/2fa/setup", authHandler.TotpSetupPOST)
		r.Get("/2fa/verify", authHandler.TotpVerifyGET)
		r.Post("/2fa/verify", authHandler.TotpVerifyPOST)
	})

	// 404 handler (uses institutional layout)
	r.NotFound(institutionalHandler.NotFound)

	// Protected group (auth required)
	r.Group(func(r chi.Router) {
		r.Use(handler.SessionValidation(queries, log))
		r.Post("/logout", authHandler.LogoutPOST)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(auth.RoleAdmin))
			r.Get("/admin", authHandler.AdminGET)
		})
	})

	// HTTP server with graceful shutdown
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
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
