package query

// SelectTable is a table selector.
type SelectTable struct {
	Builder
	as     string
	name   string
	schema string
	quote  bool
}

// Table returns a new table selector.
//
//	t1 := Table("users").As("u")
//	return Select(t1.C("name"))
func Table(name string) *SelectTable {
	return &SelectTable{quote: true, name: name}
}

// Schema sets the schema name of the table.
func (s *SelectTable) Schema(name string) *SelectTable {
	s.schema = name
	return s
}

// As adds the AS clause to the table selector.
func (s *SelectTable) As(alias string) *SelectTable {
	s.as = alias
	return s
}

// C returns a formatted string for the table column.
func (s *SelectTable) C(column string) string {
	name := s.name
	if s.as != "" {
		name = s.as
	}
	b := &Builder{dialect: s.dialect}
	if s.as == "" {
		b.writeSchema(s.schema)
	}
	b.Ident(name).WriteByte('.').Ident(column)
	return b.String()
}

// Columns returns a list of formatted strings for the table columns.
func (s *SelectTable) Columns(columns ...string) []string {
	names := make([]string, 0, len(columns))
	for _, c := range columns {
		names = append(names, s.C(c))
	}
	return names
}

// Unquote makes the table name to be formatted as raw string (unquoted).
// It is useful when you don't want to query tables under the current database.
// For example: "INFORMATION_SCHEMA.TABLE_CONSTRAINTS" in MySQL.
func (s *SelectTable) Unquote() *SelectTable {
	s.quote = false
	return s
}

// ref returns the table reference.
func (s *SelectTable) ref() string {
	if !s.quote {
		return s.name
	}
	b := &Builder{dialect: s.dialect}
	b.writeSchema(s.schema)
	b.Ident(s.name)
	if s.as != "" {
		b.WriteString(" AS ")
		b.Ident(s.as)
	}
	return b.String()
}

// implement the table view.
func (*SelectTable) view() {}
