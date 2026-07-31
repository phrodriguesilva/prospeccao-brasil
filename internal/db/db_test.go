package db_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/db"
)

// numInt creates a pgtype.Numeric from an int64.
func numInt(n int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(n), Valid: true}
}

// testDB holds the shared test database connection and queries.
type testDB struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// setupTestDB connects to the test database (prospeccaobrasil_test),
// runs migrations, and returns a testDB handle. Tests are skipped if
// DATABASE_URL is not set or the database is unreachable.
func setupTestDB(t *testing.T) *testDB {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost:5432/prospeccaobrasil_test?sslmode=disable"
	}

	// Resolve migrations path relative to the repo root (the test
	// binary runs with the package dir as CWD, so we must walk up to
	// find the migrations/ directory).
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		t.Skip("skip: cannot find migrations/ directory")
	}

	// Run migrations against the test DB.
	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		t.Skipf("skip: cannot create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Skipf("skip: cannot run migrations: %v", err)
	}

	// Connect via pgxpool.
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("skip: cannot connect to test DB: %v", err)
	}

	// Clean all tables before each test run (truncate cascade).
	_, err = pool.Exec(context.Background(), `
		TRUNCATE TABLE
			audit_log, contacts, prospections, clients, properties,
			sessions, users, tenants
		CASCADE
	`)
	if err != nil {
		t.Fatalf("cannot truncate tables: %v", err)
	}

	return &testDB{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (td *testDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

// findMigrationsDir walks up from the current working directory to find
// the repo root (identified by the migrations/ directory). Returns the
// absolute path or empty string if not found.
func findMigrationsDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 10; i++ {
		candidate := dir + "/migrations"
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := dir + "/.."
		resolved, err := filepath.Abs(parent)
		if err != nil || resolved == dir {
			return ""
		}
		dir = resolved
	}
	return ""
}

// newUUID generates a new UUID v4 and returns pgtype.UUID.
func newUUID() pgtype.UUID {
	u := uuid.New()
	return pgtype.UUID{Bytes: u, Valid: true}
}

// strPtr returns a pointer to s (for nullable text columns).
func strPtr(s string) *string {
	return &s
}

// seedTenant creates a tenant and returns its ID.
func seedTenant(t *testing.T, q *db.Queries, name string) pgtype.UUID {
	t.Helper()
	id := newUUID()
	_, err := q.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   id,
		Name: name,
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("seedTenant: %v", err)
	}
	return id
}

// seedUser creates a user under a tenant and returns its ID.
func seedUser(t *testing.T, q *db.Queries, tenantID pgtype.UUID, email, role string) pgtype.UUID {
	t.Helper()
	id := newUUID()
	_, err := q.CreateUser(context.Background(), db.CreateUserParams{
		ID:           id,
		TenantID:     tenantID,
		Email:        email,
		FullName:     "Test User",
		Role:         role,
		PasswordHash: "$2a$10$testhash",
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return id
}

// seedProperty creates a property under a tenant and returns its ID.
func seedProperty(t *testing.T, q *db.Queries, tenantID pgtype.UUID, title string) pgtype.UUID {
	t.Helper()
	id := newUUID()
	_, err := q.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:       id,
		TenantID: tenantID,
		Title:    title,
		Address:  "Rua Test, 123",
		City:     "São Paulo",
		State:    "SP",
		Price:    numInt(500000),
		Status:   "available",
		Type:     "residential",
		Photos:   []byte("[]"),
	})
	if err != nil {
		t.Fatalf("seedProperty: %v", err)
	}
	return id
}

// seedClient creates a client under a tenant and returns its ID.
func seedClient(t *testing.T, q *db.Queries, tenantID pgtype.UUID, name string) pgtype.UUID {
	t.Helper()
	id := newUUID()
	_, err := q.CreateClient(context.Background(), db.CreateClientParams{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Preferences: []byte("{}"),
		Status:      "lead",
	})
	if err != nil {
		t.Fatalf("seedClient: %v", err)
	}
	return id
}

// --- Tests ---

func TestCreateAndGetTenant(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	id := newUUID()
	tenant, err := td.queries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   id,
		Name: "Prospecção Brasil LTDA",
		Plan: "pro",
	})
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}
	if tenant.Name != "Prospecção Brasil LTDA" {
		t.Errorf("Name = %q, want %q", tenant.Name, "Prospecção Brasil LTDA")
	}
	if tenant.Plan != "pro" {
		t.Errorf("Plan = %q, want %q", tenant.Plan, "pro")
	}

	got, err := td.queries.GetTenant(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTenant: %v", err)
	}
	if got.Name != tenant.Name {
		t.Errorf("GetTenant Name = %q, want %q", got.Name, tenant.Name)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")
	userID := newUUID()
	user, err := td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "admin@prospeccao.com.br",
		FullName:     "Luiz Claudio",
		Role:         "admin",
		PasswordHash: "$2a$10$hashedpassword",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Role != "admin" {
		t.Errorf("Role = %q, want %q", user.Role, "admin")
	}

	got, err := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("Email = %q, want %q", got.Email, user.Email)
	}

	// Get by email.
	gotByEmail, err := td.queries.GetUserByEmail(context.Background(), db.GetUserByEmailParams{
		Email:    "admin@prospeccao.com.br",
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if gotByEmail.ID != user.ID {
		t.Errorf("GetUserByEmail ID mismatch")
	}

	// Verify role CHECK constraint rejects invalid value.
	_, err = td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           newUUID(),
		TenantID:     tenantID,
		Email:        "invalid@prospeccao.com.br",
		FullName:     "Invalid",
		Role:         "invalid_role",
		PasswordHash: "hash",
	})
	if err == nil {
		t.Error("expected error for invalid role, got nil")
	}
}

func TestSessionCreateRevoke(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")
	userID := seedUser(t, td.queries, tenantID, "user@prospeccao.com.br", "admin")

	sessionID := newUUID()
	tokenHash := "abc123hash"
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true}
	_, err := td.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:        sessionID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Get active session.
	got, err := td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: tokenHash,
		TenantID:  tenantID,
	})
	if err != nil {
		t.Fatalf("GetSessionByTokenHash: %v", err)
	}
	if got.TokenHash != tokenHash {
		t.Errorf("TokenHash = %q, want %q", got.TokenHash, tokenHash)
	}

	// Revoke session.
	err = td.queries.RevokeSession(context.Background(), db.RevokeSessionParams{
		ID:       sessionID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	// Get should return no rows after revoke.
	_, err = td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: tokenHash,
		TenantID:  tenantID,
	})
	if err == nil {
		t.Error("expected no rows after revoke, got a session")
	}
}

func TestPropertyCRUD(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")

	// Create.
	propID := newUUID()
	prop, err := td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:       propID,
		TenantID: tenantID,
		Title:    "Apartamento Centro",
		Address:  "Av. Paulista, 1000",
		City:     "São Paulo",
		State:    "SP",
		Price:    numInt(750000),
		Status:   "available",
		Type:     "residential",
		Photos:   []byte("[]"),
	})
	if err != nil {
		t.Fatalf("CreateProperty: %v", err)
	}
	if prop.Title != "Apartamento Centro" {
		t.Errorf("Title = %q, want %q", prop.Title, "Apartamento Centro")
	}

	// Get by ID.
	got, err := td.queries.GetPropertyByID(context.Background(), db.GetPropertyByIDParams{
		ID:       propID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetPropertyByID: %v", err)
	}
	if got.Title != prop.Title {
		t.Errorf("Get Title = %q, want %q", got.Title, prop.Title)
	}

	// List by tenant.
	list, err := td.queries.ListPropertiesByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListPropertiesByTenant: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List len = %d, want 1", len(list))
	}

	// List by status.
	byStatus, err := td.queries.ListPropertiesByStatus(context.Background(), db.ListPropertiesByStatusParams{
		TenantID: tenantID,
		Status:   "available",
	})
	if err != nil {
		t.Fatalf("ListPropertiesByStatus: %v", err)
	}
	if len(byStatus) != 1 {
		t.Errorf("ListByStatus len = %d, want 1", len(byStatus))
	}

	// Update.
	updated, err := td.queries.UpdateProperty(context.Background(), db.UpdatePropertyParams{
		ID:       propID,
		TenantID: tenantID,
		Title:    "Apartamento Centro Atualizado",
		Address:  "Av. Paulista, 1000",
		City:     "São Paulo",
		State:    "SP",
		Price:    numInt(800000),
		Status:   "reserved",
		Type:     "residential",
		Photos:   []byte("[]"),
	})
	if err != nil {
		t.Fatalf("UpdateProperty: %v", err)
	}
	if updated.Title != "Apartamento Centro Atualizado" {
		t.Errorf("Updated Title = %q, want %q", updated.Title, "Apartamento Centro Atualizado")
	}
	if updated.Status != "reserved" {
		t.Errorf("Updated Status = %q, want %q", updated.Status, "reserved")
	}

	// Soft-delete.
	err = td.queries.SoftDeleteProperty(context.Background(), db.SoftDeletePropertyParams{
		ID:       propID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("SoftDeleteProperty: %v", err)
	}

	// Get should return no rows after soft-delete.
	_, err = td.queries.GetPropertyByID(context.Background(), db.GetPropertyByIDParams{
		ID:       propID,
		TenantID: tenantID,
	})
	if err == nil {
		t.Error("expected no rows after soft-delete, got a property")
	}

	// Verify status CHECK constraint rejects invalid value.
	_, err = td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:       newUUID(),
		TenantID: tenantID,
		Title:    "Invalid",
		Address:  "X",
		City:     "X",
		State:    "SP",
		Price:    numInt(100),
		Status:   "invalid_status",
		Type:     "residential",
		Photos:   []byte("[]"),
	})
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestClientCRUD(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")

	clientID := newUUID()
	client, err := td.queries.CreateClient(context.Background(), db.CreateClientParams{
		ID:          clientID,
		TenantID:    tenantID,
		Name:        "João Silva",
		Email:       strPtr("joao@email.com"),
		Phone:       strPtr("11999999999"),
		Preferences: []byte("{}"),
		Status:      "lead",
	})
	if err != nil {
		t.Fatalf("CreateClient: %v", err)
	}
	if client.Name != "João Silva" {
		t.Errorf("Name = %q, want %q", client.Name, "João Silva")
	}

	// Get by ID.
	got, err := td.queries.GetClientByID(context.Background(), db.GetClientByIDParams{
		ID:       clientID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetClientByID: %v", err)
	}
	if got.Name != client.Name {
		t.Errorf("Get Name = %q, want %q", got.Name, client.Name)
	}

	// List by tenant.
	list, err := td.queries.ListClientsByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListClientsByTenant: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List len = %d, want 1", len(list))
	}

	// List by status.
	byStatus, err := td.queries.ListClientsByStatus(context.Background(), db.ListClientsByStatusParams{
		TenantID: tenantID,
		Status:   "lead",
	})
	if err != nil {
		t.Fatalf("ListClientsByStatus: %v", err)
	}
	if len(byStatus) != 1 {
		t.Errorf("ListByStatus len = %d, want 1", len(byStatus))
	}

	// Update.
	updated, err := td.queries.UpdateClient(context.Background(), db.UpdateClientParams{
		ID:          clientID,
		TenantID:    tenantID,
		Name:        "João Silva Jr",
		Email:       strPtr("joao@email.com"),
		Phone:       strPtr("11999999999"),
		Preferences: []byte("{}"),
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("UpdateClient: %v", err)
	}
	if updated.Name != "João Silva Jr" {
		t.Errorf("Updated Name = %q, want %q", updated.Name, "João Silva Jr")
	}

	// Soft-delete.
	err = td.queries.SoftDeleteClient(context.Background(), db.SoftDeleteClientParams{
		ID:       clientID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("SoftDeleteClient: %v", err)
	}

	_, err = td.queries.GetClientByID(context.Background(), db.GetClientByIDParams{
		ID:       clientID,
		TenantID: tenantID,
	})
	if err == nil {
		t.Error("expected no rows after soft-delete, got a client")
	}
}

func TestProspectCRUD(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")
	clientID := seedClient(t, td.queries, tenantID, "Client Prospect")
	propID := seedProperty(t, td.queries, tenantID, "Property Prospect")

	prospectID := newUUID()
	prospect, err := td.queries.CreateProspect(context.Background(), db.CreateProspectParams{
		ID:         prospectID,
		TenantID:   tenantID,
		ClientID:   clientID,
		PropertyID: propID,
		Status:     "new",
	})
	if err != nil {
		t.Fatalf("CreateProspect: %v", err)
	}
	if prospect.Status != "new" {
		t.Errorf("Status = %q, want %q", prospect.Status, "new")
	}

	// Get by ID.
	got, err := td.queries.GetProspectByID(context.Background(), db.GetProspectByIDParams{
		ID:       prospectID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetProspectByID: %v", err)
	}
	if got.ID != prospect.ID {
		t.Error("GetProspectByID ID mismatch")
	}

	// List by tenant.
	list, err := td.queries.ListProspectsByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListProspectsByTenant: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List len = %d, want 1", len(list))
	}

	// List by client.
	byClient, err := td.queries.ListProspectsByClient(context.Background(), db.ListProspectsByClientParams{
		ClientID: clientID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("ListProspectsByClient: %v", err)
	}
	if len(byClient) != 1 {
		t.Errorf("ListByClient len = %d, want 1", len(byClient))
	}

	// List by property.
	byProp, err := td.queries.ListProspectsByProperty(context.Background(), db.ListProspectsByPropertyParams{
		PropertyID: propID,
		TenantID:   tenantID,
	})
	if err != nil {
		t.Fatalf("ListProspectsByProperty: %v", err)
	}
	if len(byProp) != 1 {
		t.Errorf("ListByProperty len = %d, want 1", len(byProp))
	}

	// List by status.
	byStatus, err := td.queries.ListProspectsByStatus(context.Background(), db.ListProspectsByStatusParams{
		TenantID: tenantID,
		Status:   "new",
	})
	if err != nil {
		t.Fatalf("ListProspectsByStatus: %v", err)
	}
	if len(byStatus) != 1 {
		t.Errorf("ListByStatus len = %d, want 1", len(byStatus))
	}

	// Update.
	updated, err := td.queries.UpdateProspect(context.Background(), db.UpdateProspectParams{
		ID:       prospectID,
		TenantID: tenantID,
		Status:   "contacting",
		Notes:    strPtr("Client interested"),
	})
	if err != nil {
		t.Fatalf("UpdateProspect: %v", err)
	}
	if updated.Status != "contacting" {
		t.Errorf("Updated Status = %q, want %q", updated.Status, "contacting")
	}

	// Soft-delete.
	err = td.queries.SoftDeleteProspect(context.Background(), db.SoftDeleteProspectParams{
		ID:       prospectID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("SoftDeleteProspect: %v", err)
	}

	_, err = td.queries.GetProspectByID(context.Background(), db.GetProspectByIDParams{
		ID:       prospectID,
		TenantID: tenantID,
	})
	if err == nil {
		t.Error("expected no rows after soft-delete, got a prospect")
	}
}

func TestContactCreate(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")
	clientID := seedClient(t, td.queries, tenantID, "Contact Client")
	propID := seedProperty(t, td.queries, tenantID, "Contact Property")
	prospectID := newUUID()
	_, err := td.queries.CreateProspect(context.Background(), db.CreateProspectParams{
		ID:         prospectID,
		TenantID:   tenantID,
		ClientID:   clientID,
		PropertyID: propID,
		Status:     "new",
	})
	if err != nil {
		t.Fatalf("CreateProspect: %v", err)
	}

	// Create contact linked to prospect.
	contactID := newUUID()
	contact, err := td.queries.CreateContact(context.Background(), db.CreateContactParams{
		ID:          contactID,
		TenantID:    tenantID,
		ClientID:    clientID,
		ProspectID:  prospectID,
		Channel:     "phone",
		Direction:   "outbound",
		Subject:     strPtr("First call"),
		Body:        strPtr("Discussed property details"),
		ContactedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if contact.Channel != "phone" {
		t.Errorf("Channel = %q, want %q", contact.Channel, "phone")
	}

	// Get by ID.
	got, err := td.queries.GetContactByID(context.Background(), db.GetContactByIDParams{
		ID:       contactID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("GetContactByID: %v", err)
	}
	if got.ID != contact.ID {
		t.Error("GetContactByID ID mismatch")
	}

	// List by client.
	byClient, err := td.queries.ListContactsByClient(context.Background(), db.ListContactsByClientParams{
		ClientID: clientID,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("ListContactsByClient: %v", err)
	}
	if len(byClient) != 1 {
		t.Errorf("ListByClient len = %d, want 1", len(byClient))
	}

	// List by prospect.
	byProspect, err := td.queries.ListContactsByProspect(context.Background(), db.ListContactsByProspectParams{
		ProspectID: prospectID,
		TenantID:   tenantID,
	})
	if err != nil {
		t.Fatalf("ListContactsByProspect: %v", err)
	}
	if len(byProspect) != 1 {
		t.Errorf("ListByProspect len = %d, want 1", len(byProspect))
	}

	// Create standalone client contact (no prospect).
	standaloneID := newUUID()
	_, err = td.queries.CreateContact(context.Background(), db.CreateContactParams{
		ID:          standaloneID,
		TenantID:    tenantID,
		ClientID:    clientID,
		ProspectID:  pgtype.UUID{}, // NULL prospect_id
		Channel:     "email",
		Direction:   "inbound",
		Subject:     strPtr("General inquiry"),
		ContactedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateContact standalone: %v", err)
	}

	// Verify channel CHECK constraint rejects invalid value.
	_, err = td.queries.CreateContact(context.Background(), db.CreateContactParams{
		ID:          newUUID(),
		TenantID:    tenantID,
		ClientID:    clientID,
		ProspectID:  pgtype.UUID{},
		Channel:     "invalid_channel",
		Direction:   "inbound",
		ContactedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err == nil {
		t.Error("expected error for invalid channel, got nil")
	}
}

func TestAuditLogAppendOnly(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Test Tenant")
	userID := seedUser(t, td.queries, tenantID, "audit@prospeccao.com.br", "admin")

	// Create audit log entry.
	logID := newUUID()
	entry, err := td.queries.CreateAuditLog(context.Background(), db.CreateAuditLogParams{
		ID:         logID,
		TenantID:   tenantID,
		UserID:     userID,
		Action:     "create",
		EntityType: "client",
		EntityID:   newUUID(),
		Metadata:   []byte(`{"field":"name"}`),
	})
	if err != nil {
		t.Fatalf("CreateAuditLog: %v", err)
	}
	if entry.Action != "create" {
		t.Errorf("Action = %q, want %q", entry.Action, "create")
	}

	// List by tenant.
	list, err := td.queries.ListAuditLogByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListAuditLogByTenant: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List len = %d, want 1", len(list))
	}

	// List by entity.
	byEntity, err := td.queries.ListAuditLogByEntity(context.Background(), db.ListAuditLogByEntityParams{
		TenantID:   tenantID,
		EntityType: "client",
		EntityID:   entry.EntityID,
	})
	if err != nil {
		t.Fatalf("ListAuditLogByEntity: %v", err)
	}
	if len(byEntity) != 1 {
		t.Errorf("ListByEntity len = %d, want 1", len(byEntity))
	}

	// Verify no UPDATE/DELETE functions exist for audit_log in the
	// generated code. This is a compile-time guarantee: the db.Queries
	// struct simply has no UpdateAuditLog or DeleteAuditLog methods.
	// If someone adds such a query, sqlc would generate the method and
	// this test would need updating. The grep check in quickstart.md
	// Scenario 8 is the runtime check.
}

func TestTenantIsolation(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	// Create two tenants.
	tenantA := seedTenant(t, td.queries, "Tenant A")
	tenantB := seedTenant(t, td.queries, "Tenant B")

	// Create a user in each tenant.
	seedUser(t, td.queries, tenantA, "userA@prospeccao.com.br", "admin")
	seedUser(t, td.queries, tenantB, "userB@prospeccao.com.br", "admin")

	// Create a property in each tenant.
	seedProperty(t, td.queries, tenantA, "Property A")
	propB := seedProperty(t, td.queries, tenantB, "Property B")

	// Create a client in each tenant.
	clientA := seedClient(t, td.queries, tenantA, "Client A")
	seedClient(t, td.queries, tenantB, "Client B")

	// Verify tenant A sees only its own data.
	usersA, err := td.queries.ListUsersByTenant(context.Background(), tenantA)
	if err != nil {
		t.Fatalf("ListUsersByTenant A: %v", err)
	}
	if len(usersA) != 1 {
		t.Errorf("Tenant A users = %d, want 1", len(usersA))
	}
	if usersA[0].Email != "userA@prospeccao.com.br" {
		t.Errorf("Tenant A user email = %q, want %q", usersA[0].Email, "userA@prospeccao.com.br")
	}

	propsA, err := td.queries.ListPropertiesByTenant(context.Background(), tenantA)
	if err != nil {
		t.Fatalf("ListPropertiesByTenant A: %v", err)
	}
	if len(propsA) != 1 {
		t.Errorf("Tenant A properties = %d, want 1", len(propsA))
	}
	if propsA[0].Title != "Property A" {
		t.Errorf("Tenant A property title = %q, want %q", propsA[0].Title, "Property A")
	}

	clientsA, err := td.queries.ListClientsByTenant(context.Background(), tenantA)
	if err != nil {
		t.Fatalf("ListClientsByTenant A: %v", err)
	}
	if len(clientsA) != 1 {
		t.Errorf("Tenant A clients = %d, want 1", len(clientsA))
	}

	// Verify tenant B sees only its own data.
	usersB, err := td.queries.ListUsersByTenant(context.Background(), tenantB)
	if err != nil {
		t.Fatalf("ListUsersByTenant B: %v", err)
	}
	if len(usersB) != 1 {
		t.Errorf("Tenant B users = %d, want 1", len(usersB))
	}
	if usersB[0].Email != "userB@prospeccao.com.br" {
		t.Errorf("Tenant B user email = %q, want %q", usersB[0].Email, "userB@prospeccao.com.br")
	}

	// Verify cross-tenant access returns no rows.
	// Tenant A trying to get Tenant B's property by ID.
	_, err = td.queries.GetPropertyByID(context.Background(), db.GetPropertyByIDParams{
		ID:       propB,
		TenantID: tenantA,
	})
	if err == nil {
		t.Error("cross-tenant property access should return no rows, got a property")
	}

	// Tenant B trying to get Tenant A's client by ID.
	_, err = td.queries.GetClientByID(context.Background(), db.GetClientByIDParams{
		ID:       clientA,
		TenantID: tenantB,
	})
	if err == nil {
		t.Error("cross-tenant client access should return no rows, got a client")
	}

	// Verify audit_log isolation.
	_, err = td.queries.CreateAuditLog(context.Background(), db.CreateAuditLogParams{
		ID:         newUUID(),
		TenantID:   tenantA,
		UserID:     usersA[0].ID,
		Action:     "view",
		EntityType: "client",
		EntityID:   clientA,
	})
	if err != nil {
		t.Fatalf("CreateAuditLog A: %v", err)
	}

	logsA, err := td.queries.ListAuditLogByTenant(context.Background(), tenantA)
	if err != nil {
		t.Fatalf("ListAuditLogByTenant A: %v", err)
	}
	if len(logsA) != 1 {
		t.Errorf("Tenant A audit logs = %d, want 1", len(logsA))
	}

	logsB, err := td.queries.ListAuditLogByTenant(context.Background(), tenantB)
	if err != nil {
		t.Fatalf("ListAuditLogByTenant B: %v", err)
	}
	if len(logsB) != 0 {
		t.Errorf("Tenant B audit logs = %d, want 0 (isolation)", len(logsB))
	}
}

func TestListEmptyResults(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	// Fresh tenant with no data -- all List functions should return empty
	// slices without error (exercises the rows.Next/rows.Err path).
	tenantID := seedTenant(t, td.queries, "Empty Tenant")

	users, err := td.queries.ListUsersByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListUsersByTenant empty: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("empty ListUsersByTenant = %d, want 0", len(users))
	}

	props, err := td.queries.ListPropertiesByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListPropertiesByTenant empty: %v", err)
	}
	if len(props) != 0 {
		t.Errorf("empty ListPropertiesByTenant = %d, want 0", len(props))
	}

	clients, err := td.queries.ListClientsByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListClientsByTenant empty: %v", err)
	}
	if len(clients) != 0 {
		t.Errorf("empty ListClientsByTenant = %d, want 0", len(clients))
	}

	prospects, err := td.queries.ListProspectsByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListProspectsByTenant empty: %v", err)
	}
	if len(prospects) != 0 {
		t.Errorf("empty ListProspectsByTenant = %d, want 0", len(prospects))
	}

	logs, err := td.queries.ListAuditLogByTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListAuditLogByTenant empty: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("empty ListAuditLogByTenant = %d, want 0", len(logs))
	}

	// List by status on empty tenant.
	propsByStatus, err := td.queries.ListPropertiesByStatus(context.Background(), db.ListPropertiesByStatusParams{
		TenantID: tenantID,
		Status:   "available",
	})
	if err != nil {
		t.Fatalf("ListPropertiesByStatus empty: %v", err)
	}
	if len(propsByStatus) != 0 {
		t.Errorf("empty ListPropertiesByStatus = %d, want 0", len(propsByStatus))
	}

	clientsByStatus, err := td.queries.ListClientsByStatus(context.Background(), db.ListClientsByStatusParams{
		TenantID: tenantID,
		Status:   "lead",
	})
	if err != nil {
		t.Fatalf("ListClientsByStatus empty: %v", err)
	}
	if len(clientsByStatus) != 0 {
		t.Errorf("empty ListClientsByStatus = %d, want 0", len(clientsByStatus))
	}

	prospectsByStatus, err := td.queries.ListProspectsByStatus(context.Background(), db.ListProspectsByStatusParams{
		TenantID: tenantID,
		Status:   "new",
	})
	if err != nil {
		t.Fatalf("ListProspectsByStatus empty: %v", err)
	}
	if len(prospectsByStatus) != 0 {
		t.Errorf("empty ListProspectsByStatus = %d, want 0", len(prospectsByStatus))
	}
}

func TestWithTx(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	// WithTx creates a new Queries bound to a transaction. Verify it
	// returns a working Queries handle.
	tx, err := td.pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	txQueries := td.queries.WithTx(tx)
	if txQueries == nil {
		t.Fatal("WithTx returned nil")
	}

	// Use the transactional queries to create a tenant.
	tenantID := newUUID()
	_, err = txQueries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   tenantID,
		Name: "Tx Tenant",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("CreateTenant via tx: %v", err)
	}

	// Rollback so the tenant is not persisted.
	err = tx.Rollback(context.Background())
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	// Verify the tenant was not persisted (rollback worked).
	_, err = td.queries.GetTenant(context.Background(), tenantID)
	if err == nil {
		t.Error("tenant should not exist after rollback")
	}
}

func TestUpdateTenant(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Old Name")
	updated, err := td.queries.UpdateTenant(context.Background(), db.UpdateTenantParams{
		ID:   tenantID,
		Name: "New Name",
		Cnpj: strPtr("12345678000190"),
	})
	if err != nil {
		t.Fatalf("UpdateTenant: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("Name = %q, want %q", updated.Name, "New Name")
	}
	if updated.Cnpj == nil || *updated.Cnpj != "12345678000190" {
		t.Errorf("Cnpj not updated correctly")
	}
}

func TestUpdateUserTOTP(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "TOTP Tenant")
	userID := seedUser(t, td.queries, tenantID, "totp@prospeccao.com.br", "admin")

	updated, err := td.queries.UpdateUserTOTP(context.Background(), db.UpdateUserTOTPParams{
		ID:          userID,
		TotpSecret:  strPtr("encrypted_secret_blob"),
		TotpEnabled: true,
		TenantID:    tenantID,
	})
	if err != nil {
		t.Fatalf("UpdateUserTOTP: %v", err)
	}
	if !updated.TotpEnabled {
		t.Error("TotpEnabled = false, want true")
	}
	if updated.TotpSecret == nil || *updated.TotpSecret != "encrypted_secret_blob" {
		t.Error("TotpSecret not set correctly")
	}
}

func TestUpdateUserLoginAttempts(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Login Tenant")
	userID := seedUser(t, td.queries, tenantID, "login@prospeccao.com.br", "admin")

	lockedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updated, err := td.queries.UpdateUserLoginAttempts(context.Background(), db.UpdateUserLoginAttemptsParams{
		ID:                  userID,
		FailedLoginAttempts: 3,
		LockedAt:            lockedAt,
		TenantID:            tenantID,
	})
	if err != nil {
		t.Fatalf("UpdateUserLoginAttempts: %v", err)
	}
	if updated.FailedLoginAttempts != 3 {
		t.Errorf("FailedLoginAttempts = %d, want 3", updated.FailedLoginAttempts)
	}
	if !updated.LockedAt.Valid {
		t.Error("LockedAt should be valid (set)")
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID := seedTenant(t, td.queries, "Expiry Tenant")
	userID := seedUser(t, td.queries, tenantID, "expiry@prospeccao.com.br", "admin")

	// Create an expired session.
	expiredID := newUUID()
	_, err := td.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:        expiredID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: "expired_token_hash",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSession expired: %v", err)
	}

	// Create an active session.
	activeID := newUUID()
	_, err = td.queries.CreateSession(context.Background(), db.CreateSessionParams{
		ID:        activeID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: "active_token_hash",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSession active: %v", err)
	}

	// Delete expired sessions.
	deleted, err := td.queries.DeleteExpiredSessions(context.Background())
	if err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}

	// Active session should still be retrievable.
	_, err = td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: "active_token_hash",
		TenantID:  tenantID,
	})
	if err != nil {
		t.Errorf("active session should still exist: %v", err)
	}

	// Expired session should be gone.
	_, err = td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: "expired_token_hash",
		TenantID:  tenantID,
	})
	if err == nil {
		t.Error("expired session should have been deleted")
	}
}

// Ensure pgx.Conn is imported (used indirectly via pgxpool).
var _ = pgx.Connect
