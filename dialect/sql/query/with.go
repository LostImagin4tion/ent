package query

// WithBuilder is the builder for the `WITH` statement.
type WithBuilder struct {
	Builder
	recursive bool
	ctes      []struct {
		name    string
		columns []string
		s       *Selector
	}
}

// With returns a new builder for the `WITH` statement.
//
//	n := Queries{
//		With("users_view").As(Select().From(Table("users"))),
//		Select().From(Table("users_view")),
//	}
//	return n.Query()
func With(name string, columns ...string) *WithBuilder {
	return &WithBuilder{
		ctes: []struct {
			name    string
			columns []string
			s       *Selector
		}{
			{name: name, columns: columns},
		},
	}
}

// WithRecursive returns a new builder for the `WITH RECURSIVE` statement.
//
//	n := Queries{
//		WithRecursive("users_view").As(Select().From(Table("users"))),
//		Select().From(Table("users_view")),
//	}
//	return n.Query()
func WithRecursive(name string, columns ...string) *WithBuilder {
	w := With(name, columns...)
	w.recursive = true
	return w
}

// Name returns the name of the view.
func (w *WithBuilder) Name() string {
	return w.ctes[0].name
}

// As sets the view sub query.
func (w *WithBuilder) As(s *Selector) *WithBuilder {
	w.ctes[len(w.ctes)-1].s = s
	return w
}

// With appends another named CTE to the statement.
func (w *WithBuilder) With(name string, columns ...string) *WithBuilder {
	w.ctes = append(w.ctes, With(name, columns...).ctes...)
	return w
}

// C returns a formatted string for the WITH column.
func (w *WithBuilder) C(column string) string {
	b := &Builder{dialect: w.dialect}
	b.Ident(w.Name()).WriteByte('.').Ident(column)
	return b.String()
}

// Query returns query representation of a `WITH` clause.
func (w *WithBuilder) Query() (string, []any) {
	w.WriteString("WITH ")
	if w.recursive {
		w.WriteString("RECURSIVE ")
	}
	for i, cte := range w.ctes {
		if i > 0 {
			w.Comma()
		}
		w.Ident(cte.name)
		if len(cte.columns) > 0 {
			w.WriteByte('(')
			w.IdentComma(cte.columns...)
			w.WriteByte(')')
		}
		w.WriteString(" AS ")
		w.Wrap(func(b *Builder) {
			b.Join(cte.s)
		})
	}
	return w.String(), w.args
}

// implement the table view interface.
func (*WithBuilder) view() {}
