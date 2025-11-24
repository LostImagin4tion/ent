package query

import "strings"

// As suffixed the given column with an alias (`a` AS `b`).
func As(ident string, as string) string {
	b := &Builder{}
	b.fromIdent(ident)
	b.Ident(ident).Pad().WriteString("AS")
	b.Pad().Ident(as)
	return b.String()
}

// Distinct prefixed the given columns with the `DISTINCT` keyword (DISTINCT `id`).
func Distinct(idents ...string) string {
	b := &Builder{}
	if len(idents) > 0 {
		b.fromIdent(idents[0])
	}
	b.WriteString("DISTINCT")
	b.Pad().IdentComma(idents...)
	return b.String()
}

func IsFunc(s string) bool {
	return strings.Contains(s, "(") && strings.Contains(s, ")")
}

func isAlias(s string) bool {
	return strings.Contains(s, " AS ") || strings.Contains(s, " as ")
}

func isModifier(s string) bool {
	for _, m := range [...]string{"DISTINCT", "ALL", "WITH ROLLUP"} {
		if strings.HasPrefix(s, m) {
			return true
		}
	}
	return false
}
