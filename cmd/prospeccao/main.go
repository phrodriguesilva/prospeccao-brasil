package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
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

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	// Structured logging via slog (Constitution principle V).
	// JSON handler in production, text handler in development.
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)

	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
