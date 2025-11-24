package query

// DialectBuilder prefixes all root builders with the `Dialect` constructor.
type DialectBuilder struct {
	dialect string
}

// Dialect creates a new DialectBuilder with the given dialect name.
func Dialect(name string) *DialectBuilder {
	return &DialectBuilder{name}
}

// String builds a dialect-aware expression string from the given callback.
func (d *DialectBuilder) String(f func(*Builder)) string {
	b := &Builder{}
	b.SetDialect(d.dialect)
	f(b)
	return b.String()
}

// Expr builds a dialect-aware expression from the given callback.
func (d *DialectBuilder) Expr(f func(*Builder)) Querier {
	return Expr(d.String(f))
}

// CreateView creates a ViewBuilder for the configured dialect.
//
//	t := Table("users")
//	Dialect(dialect.Postgres).
//		CreateView("users").
//		Columns(
//			Column("id").Type("int"),
//			Column("name").Type("varchar(255)"),
//		).
//		As(Select(t.C("id"), t.C("name")).From(t))
func (d *DialectBuilder) CreateView(name string) *ViewBuilder {
	b := CreateView(name)
	b.SetDialect(d.dialect)
	return b
}

// Column creates a ColumnBuilder for the configured dialect.
//
//	Dialect(dialect.Postgres)..
//		Column("group_id").Type("int").Attr("UNIQUE")
func (d *DialectBuilder) Column(name string) *ColumnBuilder {
	b := Column(name)
	b.SetDialect(d.dialect)
	return b
}

// Insert creates a InsertBuilder for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		Insert("users").Columns("age").Values(1)
func (d *DialectBuilder) Insert(table string) *InsertBuilder {
	b := Insert(table)
	b.SetDialect(d.dialect)
	return b
}

// Update creates a UpdateBuilder for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		Update("users").Set("name", "foo")
func (d *DialectBuilder) Update(table string) *UpdateBuilder {
	b := Update(table)
	b.SetDialect(d.dialect)
	return b
}

// Delete creates a DeleteBuilder for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		Delete().From("users")
func (d *DialectBuilder) Delete(table string) *DeleteBuilder {
	b := Delete(table)
	b.SetDialect(d.dialect)
	return b
}

// Select creates a Selector for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		Select().From(Table("users"))
func (d *DialectBuilder) Select(columns ...string) *Selector {
	b := Select(columns...)
	b.SetDialect(d.dialect)
	return b
}

// SelectExpr is like Select, but supports passing arbitrary
// expressions for SELECT clause.
//
//	Dialect(dialect.Postgres).
//		SelectExpr(expr...).
//		From(Table("users"))
func (d *DialectBuilder) SelectExpr(exprs ...Querier) *Selector {
	b := SelectExpr(exprs...)
	b.SetDialect(d.dialect)
	return b
}

// Table creates a SelectTable for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		Table("users").As("u")
func (d *DialectBuilder) Table(name string) *SelectTable {
	b := Table(name)
	b.SetDialect(d.dialect)
	return b
}

// With creates a WithBuilder for the configured dialect.
//
//	Dialect(dialect.Postgres).
//		With("users_view").
//		As(Select().From(Table("users")))
func (d *DialectBuilder) With(name string) *WithBuilder {
	b := With(name)
	b.SetDialect(d.dialect)
	return b
}
