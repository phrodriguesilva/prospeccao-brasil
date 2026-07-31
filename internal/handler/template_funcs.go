package handler

import (
	"encoding/json"
	"html/template"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// TemplateFuncs returns a FuncMap for html/template with helpers for pgtype and UUID display.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"uuidToString":     uuidToString,
		"pgText":           pgTextDisplay,
		"pgInt32":          pgInt32Display,
		"pgNumeric":        pgNumericDisplay,
		"formatBRLNumeric": formatBRLNumeric,
		"sub":              subInt,
		"add":              addInt,
		"joinPhotos":       joinPhotos,
	}
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

func pgTextDisplay(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func pgInt32Display(i *int32) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(int(*i))
}

func pgNumericDisplay(n pgtype.Numeric) string {
	return fromPgNumeric(n)
}

func formatBRLNumeric(n pgtype.Numeric) string {
	s := fromPgNumeric(n)
	if s == "" {
		return "R$ 0,00"
	}
	return "R$ " + strings.ReplaceAll(s, ".", ",")
}

func subInt(a, b int) int {
	return a - b
}

func addInt(a, b int) int {
	return a + b
}

func joinPhotos(data []byte) string {
	var urls []string
	_ = json.Unmarshal(data, &urls)
	return strings.Join(urls, "\n")
}
