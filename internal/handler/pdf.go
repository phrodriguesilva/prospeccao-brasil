package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// PDFHandler handles PDF generation via chromedp.
type PDFHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewPDFHandler creates a new PDFHandler.
func NewPDFHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *PDFHandler {
	return &PDFHandler{queries: queries, tmpl: tmpl, log: log}
}

// pdfData is the template data for the PDF HTML.
type pdfData struct {
	Title       string
	Address     string
	Price       string
	Type        string
	AreaSqm     string
	Bedrooms    string
	Bathrooms   string
	Description string
	Photos      []string
}

// GeneratePropertyPDF generates a PDF presentation for a property.
func (h *PDFHandler) GeneratePropertyPDF(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	prop, err := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Check if Chrome/Chromium is available
	chromePath := findChrome()
	if chromePath == "" {
		h.log.ErrorContext(r.Context(), "pdf: chrome not found")
		http.Error(w, "Geração de PDF não disponível. Chrome não instalado no servidor.", http.StatusInternalServerError)
		return
	}

	// Parse photos from JSONB
	var photos []string
	if len(prop.Photos) > 0 {
		_ = json.Unmarshal(prop.Photos, &photos)
	}

	data := pdfData{
		Title:       prop.Title,
		Address:     prop.Address + ", " + prop.City + " - " + prop.State,
		Price:       formatBRL(fromPgNumeric(prop.Price)),
		Type:        prop.Type,
		AreaSqm:     fromPgNumeric(prop.AreaSqm),
		Bedrooms:    fromPgInt32(prop.Bedrooms),
		Bathrooms:   fromPgInt32(prop.Bathrooms),
		Description: fromPgText(prop.Description),
		Photos:      photos,
	}

	// Render HTML template
	var htmlBuf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&htmlBuf, "properties_pdf.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "pdf: render template", "error", err)
		http.Error(w, "Erro ao gerar PDF.", http.StatusInternalServerError)
		return
	}

	// Write HTML to temp file
	tmpDir, err := os.MkdirTemp("", "prospeccao-pdf-*")
	if err != nil {
		h.log.ErrorContext(r.Context(), "pdf: temp dir", "error", err)
		http.Error(w, "Erro ao gerar PDF.", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	htmlPath := filepath.Join(tmpDir, "property.html")
	if err := os.WriteFile(htmlPath, htmlBuf.Bytes(), 0644); err != nil {
		h.log.ErrorContext(r.Context(), "pdf: write html", "error", err)
		http.Error(w, "Erro ao gerar PDF.", http.StatusInternalServerError)
		return
	}

	// Generate PDF via chromedp
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	var pdfBuf []byte
	err = chromedp.Run(chromeCtx,
		chromedp.Navigate("file://"+htmlPath),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().Do(ctx)
			if err != nil {
				return fmt.Errorf("print to pdf: %w", err)
			}
			pdfBuf = buf
			return nil
		}),
	)

	if err != nil {
		h.log.ErrorContext(r.Context(), "pdf: chromedp run", "error", err)
		http.Error(w, "Erro ao gerar PDF. Tente novamente.", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "pdf_generated", "property_id", uuid.UUID(prop.ID.Bytes).String(), "size", len(pdfBuf))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="imovel-%s.pdf"`, uuid.UUID(prop.ID.Bytes).String()))
	_, _ = w.Write(pdfBuf)
}

// findChrome finds the Chrome/Chromium binary path.
func findChrome() string {
	candidates := []string{
		"chromium-browser",
		"chromium",
		"google-chrome",
		"google-chrome-stable",
	}
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path
		}
	}
	// Check common macOS paths
	macPaths := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium-browser",
		"/snap/bin/chromium",
	}
	for _, p := range macPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
