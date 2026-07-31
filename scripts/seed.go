// Command seed creates the initial tenant and admin user for Prospecção Brasil.
// It is idempotent: if the tenant or user already exists, it skips creation.
//
// Usage:
//
//	go run ./scripts/seed.go
//
// Required env vars:
//   - DATABASE_URL: Postgres connection string
//   - ENCRYPTION_KEY: base64-encoded 32-byte key (for TOTP, not used here but required for consistency)
//
// Optional env vars:
//   - ADMIN_EMAIL: admin email (default: admin@prospeccao.com.br)
//   - ADMIN_PASSWORD: admin password (default: changeme)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(log)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Error("DATABASE_URL not set")
		os.Exit(1)
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@prospeccao.com.br"
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "changeme"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Error("pgxpool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)

	// 1. Create tenant (idempotent)
	tenantID, err := getOrCreateTenant(ctx, queries, log)
	if err != nil {
		log.Error("seed tenant", "error", err)
		os.Exit(1)
	}

	// 2. Create admin user (idempotent)
	if err := getOrCreateAdmin(ctx, queries, log, tenantID, adminEmail, adminPassword); err != nil {
		log.Error("seed admin", "error", err)
		os.Exit(1)
	}

	log.Info("seed complete", "tenant_id", tenantID, "admin_email", adminEmail)
}

// getOrCreateTenant creates the default tenant if it does not exist.
// Returns the tenant ID.
func getOrCreateTenant(ctx context.Context, queries *db.Queries, log *slog.Logger) (pgtype.UUID, error) {
	tenants, err := queries.ListTenantsByActive(ctx, true)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("list tenants: %w", err)
	}
	if len(tenants) > 0 {
		log.Info("tenant already exists, skipping", "tenant_id", tenants[0].ID, "name", tenants[0].Name)
		return tenants[0].ID, nil
	}

	tenantID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	tenant, err := queries.CreateTenant(ctx, db.CreateTenantParams{
		ID:   tenantID,
		Name: "Prospeccao Brasil",
		Plan: "free",
	})
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("create tenant: %w", err)
	}
	log.Info("tenant created", "tenant_id", tenant.ID, "name", tenant.Name)
	return tenant.ID, nil
}

// getOrCreateAdmin creates the admin user if it does not exist.
func getOrCreateAdmin(ctx context.Context, queries *db.Queries, log *slog.Logger, tenantID pgtype.UUID, email, password string) error {
	// Check if user already exists
	_, err := queries.GetUserForAuth(ctx, db.GetUserForAuthParams{
		Email:    email,
		TenantID: tenantID,
	})
	if err == nil {
		log.Info("admin user already exists, skipping", "email", email)
		return nil
	}

	// User does not exist -- create
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        email,
		FullName:     "Administrador",
		Role:         auth.RoleAdmin,
		PasswordHash: hash,
	})
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	log.Info("admin user created", "user_id", userID, "email", email, "role", auth.RoleAdmin)
	return nil
}
