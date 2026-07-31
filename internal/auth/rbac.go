package auth

import (
	"context"
	"net/http"
)

// Role constants for the 4 RBAC roles. The hierarchy is:
// admin (4) > corretor (3) > assistente (2) > financeiro (1).
const (
	RoleAdmin      = "admin"
	RoleCorretor   = "corretor"
	RoleAssistente = "assistente"
	RoleFinanceiro = "financeiro"
)

// ContextKey is the type for context values set by middleware.
type ContextKey string

const (
	CtxUser     ContextKey = "user"
	CtxTenantID ContextKey = "tenant_id"
	CtxUserID   ContextKey = "user_id"
	CtxRole     ContextKey = "role"
)

// RoleLevel returns the numeric level for a role (higher = more privileged).
// Unknown roles return 0 (no access).
func RoleLevel(role string) int {
	switch role {
	case RoleAdmin:
		return 4
	case RoleCorretor:
		return 3
	case RoleAssistente:
		return 2
	case RoleFinanceiro:
		return 1
	default:
		return 0
	}
}

// RequireRole returns middleware that checks if the user's role (from context)
// has a level >= the required role's level. If insufficient, returns 403.
func RequireRole(required string) func(http.Handler) http.Handler {
	requiredLevel := RoleLevel(required)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(CtxRole).(string)
			if !ok || RoleLevel(role) < requiredLevel {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RoleFromContext extracts the role string from the request context.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(CtxRole).(string)
	return role, ok
}

// TenantIDFromContext extracts the tenant_id from the request context.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	tid, ok := ctx.Value(CtxTenantID).(string)
	return tid, ok
}
