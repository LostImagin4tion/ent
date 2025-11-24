package query

// ViewBuilder is a query builder for `CREATE VIEW` statement.
type ViewBuilder struct {
	Builder
	schema  string    // view schema.
	name    string    // view name.
	exists  bool      // check existence.
	columns []Querier // table columns.
	as      Querier   // view query.
}

// CreateView returns a query builder for the `CREATE VIEW` statement.
//
//	t := Table("users")
//	CreateView("clean_users").
//		Columns(
//			Column("id").Type("int").Attr("auto_increment"),
//			Column("name").Type("varchar(255)"),
//		).
//		As(Select(t.C("id"), t.C("name")).From(t))
func CreateView(name string) *ViewBuilder { return &ViewBuilder{name: name} }

// Schema sets the database name for the view.
func (v *ViewBuilder) Schema(name string) *ViewBuilder {
	v.schema = name
	return v
}

// IfNotExists appends the `IF NOT EXISTS` clause to the `CREATE VIEW` statement.
func (v *ViewBuilder) IfNotExists() *ViewBuilder {
	v.exists = true
	return v
}

// Column appends the given column to the `CREATE VIEW` statement.
func (v *ViewBuilder) Column(c *ColumnBuilder) *ViewBuilder {
	v.columns = append(v.columns, c)
	return v
}

// Columns appends a list of columns to the builder.
func (v *ViewBuilder) Columns(columns ...*ColumnBuilder) *ViewBuilder {
	v.columns = make([]Querier, 0, len(columns))
	for i := range columns {
		v.columns = append(v.columns, columns[i])
	}
	return v
}

// As sets the view definition to the builder.
func (v *ViewBuilder) As(as Querier) *ViewBuilder {
	v.as = as
	return v
}

// Query returns query representation of a `CREATE VIEW` statement.
//
// CREATE VIEW [IF NOT EXISTS] name AS
//
//	(view definition)
func (v *ViewBuilder) Query() (string, []any) {
	v.WriteString("CREATE VIEW ")
	if v.exists {
		v.WriteString("IF NOT EXISTS ")
	}
	v.writeSchema(v.schema)
	v.Ident(v.name)
	if len(v.columns) > 0 {
		v.Pad().Wrap(func(b *Builder) { b.JoinComma(v.columns...) })
	}
	v.WriteString(" AS ")
	v.Join(v.as)
	return v.String(), v.args
}
