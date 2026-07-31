package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// DashboardHandler handles the admin dashboard.
type DashboardHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *DashboardHandler {
	return &DashboardHandler{queries: queries, tmpl: tmpl, log: log}
}

// dashboardData is the template data for the dashboard.
type dashboardData struct {
	Title           string
	ActivePage      string
	UserEmail       string
	UserRole        string
	PropertyCount   int64
	ClientCount     int64
	ProspectCount   int64
	StatusCounts    []statusCount
	RecentProspects []db.ListRecentProspectsWithDetailsRow
	IsEmpty         bool
}

type statusCount struct {
	Status string
	Count  int64
}

// Index renders the admin dashboard at GET /admin.
func (h *DashboardHandler) Index(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctx := r.Context()
	tenantID := user.TenantID

	propCount, err := h.queries.CountPropertiesByTenant(ctx, tenantID)
	if err != nil {
		h.log.ErrorContext(ctx, "dashboard: count properties", "error", err)
		propCount = 0
	}

	clientCount, err := h.queries.CountClientsByTenant(ctx, tenantID)
	if err != nil {
		h.log.ErrorContext(ctx, "dashboard: count clients", "error", err)
		clientCount = 0
	}

	prospectCount, err := h.queries.CountProspectsByTenant(ctx, tenantID)
	if err != nil {
		h.log.ErrorContext(ctx, "dashboard: count prospects", "error", err)
		prospectCount = 0
	}

	statusRows, err := h.queries.CountProspectsByStatus(ctx, tenantID)
	if err != nil {
		h.log.ErrorContext(ctx, "dashboard: count by status", "error", err)
		statusRows = nil
	}

	statusCounts := make([]statusCount, 0, len(statusRows))
	for _, row := range statusRows {
		statusCounts = append(statusCounts, statusCount{Status: row.Status, Count: row.Count})
	}

	recentProspects, err := h.queries.ListRecentProspectsWithDetails(ctx, db.ListRecentProspectsWithDetailsParams{
		TenantID: tenantID,
		Limit:    5,
	})
	if err != nil {
		h.log.ErrorContext(ctx, "dashboard: recent prospects", "error", err)
		recentProspects = nil
	}

	data := dashboardData{
		Title:           "Dashboard",
		ActivePage:      "dashboard",
		UserEmail:       user.Email,
		UserRole:        user.Role,
		PropertyCount:   propCount,
		ClientCount:     clientCount,
		ProspectCount:   prospectCount,
		StatusCounts:    statusCounts,
		RecentProspects: recentProspects,
		IsEmpty:         propCount == 0 && clientCount == 0 && prospectCount == 0,
	}

	if err := h.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		h.log.ErrorContext(ctx, "dashboard: render", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
