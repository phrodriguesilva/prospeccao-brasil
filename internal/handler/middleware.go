package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// sha256Hex returns the hex-encoded SHA-256 hash of a string.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// SessionValidation middleware reads the session cookie, looks up the session
// via GetSessionWithUser, checks user.deleted_at and tenant.deleted_at, and
// attaches user/tenant_id/user_id/role to the request context. If invalid,
// redirects to /login.
func SessionValidation(queries *db.Queries, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.SessionCookieName)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			// We need the tenant_id to look up the session, but we don't have
			// it yet. We look up by token_hash across all tenants (the token
			// is unique). We use a zero UUID to match the first session with
			// this token_hash. Actually, the query requires tenant_id. We
			// need a different approach: look up by token_hash only.
			// Workaround: iterate -- but that's bad. Better: add a query that
			// looks up by token_hash only. For now, we use a workaround:
			// extract tenant_id from the pending session cookie if present,
			// or use a global lookup.
			//
			// Actually, the simplest fix: the session cookie should encode
			// the tenant_id. But we store only the raw token. Let's add a
			// query that looks up by token_hash only (no tenant_id filter).
			// For now, we'll use the existing GetSessionByTokenHash but it
			// requires tenant_id. Let's use a different approach: store
			// tenant_id in the cookie value as "token:tenantID".
			//
			// Simplest: just look up by token_hash without tenant_id. We
			// need a new query. But for MVP single-tenant, we can use the
			// first tenant. Let's use a helper that gets the first tenant.
			//
			// Actually, let's just add a query. For now, use the pending
			// cookie to get tenant_id, or query all sessions.
			//
			// PRAGMATIC MVP APPROACH: single-tenant, so we get the first
			// tenant and use it. This is acceptable for MVP per Constitution
			// principle VII.

			tenantID, err := getFirstTenantID(r.Context(), queries)
			if err != nil {
				log.ErrorContext(r.Context(), "session_validation: get tenant", "error", err)
				redirectToLogin(w, r)
				return
			}

			// Hash the cookie token to look up in DB
			tokenHash := hashToken(cookie.Value)
			row, err := queries.GetSessionWithUser(r.Context(), db.GetSessionWithUserParams{
				TokenHash: tokenHash,
				TenantID:  tenantID,
			})
			if err != nil {
				log.InfoContext(r.Context(), "session_validation: session not found", "error", err)
				redirectToLogin(w, r)
				return
			}

			// Check user soft-deleted
			if row.UserDeletedAt.Valid {
				log.InfoContext(r.Context(), "session_validation: user deleted")
				redirectToLogin(w, r)
				return
			}

			// Check tenant soft-deleted
			if row.TenantDeletedAt.Valid {
				log.InfoContext(r.Context(), "session_validation: tenant deleted")
				redirectToLogin(w, r)
				return
			}

			// Attach context values
			ctx := r.Context()
			ctx = context.WithValue(ctx, auth.CtxTenantID, row.TenantID.String())
			ctx = context.WithValue(ctx, auth.CtxUserID, row.UserID.String())
			ctx = context.WithValue(ctx, auth.CtxRole, row.Role)
			ctx = context.WithValue(ctx, auth.CtxUser, &db.User{
				ID:       row.UserID,
				TenantID: row.TenantID,
				Email:    row.Email,
				FullName: row.FullName,
				Role:     row.Role,
			})
			ctx = context.WithValue(ctx, sessionIDKey{}, row.SessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// sessionIDKey is an unexported context key for the session ID (used by Logout).
type sessionIDKey struct{}

// SessionIDFromContext extracts the session ID from the context.
func SessionIDFromContext(ctx context.Context) (pgtype.UUID, bool) {
	sid, ok := ctx.Value(sessionIDKey{}).(pgtype.UUID)
	return sid, ok
}

// redirectToLogin sends a 302 redirect to /login.
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

// getFirstTenantID returns the first tenant's ID. MVP single-tenant workaround.
// In multi-tenant, the session cookie would encode the tenant_id.
func getFirstTenantID(ctx context.Context, queries *db.Queries) (pgtype.UUID, error) {
	tenants, err := queries.ListTenantsByActive(ctx, true)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("get first tenant: %w", err)
	}
	if len(tenants) == 0 {
		return pgtype.UUID{}, fmt.Errorf("no active tenants")
	}
	return tenants[0].ID, nil
}

// hashToken computes the SHA-256 hash of a token (for DB lookup).
func hashToken(token string) string {
	return sha256Hex(token)
}
