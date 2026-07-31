package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleLevel(t *testing.T) {
	tests := []struct {
		role string
		want int
	}{
		{RoleAdmin, 4},
		{RoleCorretor, 3},
		{RoleAssistente, 2},
		{RoleFinanceiro, 1},
		{"unknown", 0},
	}
	for _, tt := range tests {
		got := RoleLevel(tt.role)
		if got != tt.want {
			t.Errorf("RoleLevel(%q) = %d, want %d", tt.role, got, tt.want)
		}
	}
}

func TestRequireRoleAdminAccessingAdmin(t *testing.T) {
	mw := RequireRole(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ctx := context.WithValue(context.Background(), CtxRole, RoleAdmin)
	req := httptest.NewRequest("GET", "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("admin accessing admin-only: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireRoleAssistenteAccessingAdmin(t *testing.T) {
	mw := RequireRole(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ctx := context.WithValue(context.Background(), CtxRole, RoleAssistente)
	req := httptest.NewRequest("GET", "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("assistente accessing admin-only: got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireRoleNoRoleInContext(t *testing.T) {
	mw := RequireRole(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("no role in context: got %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRoleFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), CtxRole, RoleAdmin)
	role, ok := RoleFromContext(ctx)
	if !ok || role != RoleAdmin {
		t.Errorf("RoleFromContext = %q, %v, want %q, true", role, ok, RoleAdmin)
	}
	_, ok = RoleFromContext(context.Background())
	if ok {
		t.Error("RoleFromContext with no role should return false")
	}
}

func TestTenantIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), CtxTenantID, "tenant-123")
	tid, ok := TenantIDFromContext(ctx)
	if !ok || tid != "tenant-123" {
		t.Errorf("TenantIDFromContext = %q, %v, want tenant-123, true", tid, ok)
	}
}
