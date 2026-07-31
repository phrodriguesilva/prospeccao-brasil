package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	ct := res.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var body healthResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("expected status ok, got %q", body.Status)
	}
}
