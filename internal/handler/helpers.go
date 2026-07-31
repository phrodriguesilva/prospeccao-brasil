package handler

import (
	"math/big"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// toPgText converts a string to *string (for nullable text columns).
func toPgText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// fromPgText converts *string to a string.
func fromPgText(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// toPgInt32 converts a string to *int32 (for nullable int columns).
func toPgInt32(s string) *int32 {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	i := int32(n)
	return &i
}

// fromPgInt32 converts *int32 to a string.
func fromPgInt32(i *int32) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(int(*i))
}

// toPgNumeric converts a string to pgtype.Numeric.
func toPgNumeric(s string) pgtype.Numeric {
	if s == "" {
		return pgtype.Numeric{Valid: false}
	}
	n, _, err := big.ParseFloat(s, 10, 200, big.ToNearestEven)
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	num := pgtype.Numeric{Valid: true}
	_ = num.Scan(n.Text('f', -1))
	return num
}

// fromPgNumeric converts pgtype.Numeric to a string.
func fromPgNumeric(n pgtype.Numeric) string {
	if !n.Valid {
		return ""
	}
	f, err := n.Float64Value()
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(f.Float64, 'f', 2, 64)
}

// parseDecimal converts a string price to pgtype.Numeric.
func parseDecimal(s string) pgtype.Numeric {
	return toPgNumeric(s)
}

// formatDecimal converts pgtype.Numeric to a display string.
func formatDecimal(n pgtype.Numeric) string {
	return fromPgNumeric(n)
}

// formatBRL formats a numeric string as Brazilian Real.
func formatBRL(s string) string {
	if s == "" {
		return "R$ 0,00"
	}
	return "R$ " + s
}
