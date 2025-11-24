package query

// WindowBuilder represents a builder for a window clause.
// Note that window functions support is limited and used
// only to query rows-limited edges in pagination.
type WindowBuilder struct {
	Builder
	fn        func(*Builder) // e.g. ROW_NUMBER(), RANK()
	partition func(*Builder)
	order     []any
}

// RowNumber returns a new window clause with the ROW_NUMBER() as a function.
// Using this function will assign each row a number, from 1 to N, in the
// order defined by the ORDER BY clause in the window spec.
func RowNumber() *WindowBuilder {
	return Window(func(b *Builder) {
		b.WriteString("ROW_NUMBER()")
	})
}

// Window returns a new window clause with a custom selector allowing
// for custom window functions.
//
//	Window(func(b *Builder) {
//		b.WriteString(Sum(posts.C("duration")))
//	}).PartitionBy("author_id").OrderBy("id"), "duration").
func Window(fn func(*Builder)) *WindowBuilder {
	return &WindowBuilder{fn: fn}
}

// PartitionBy indicates to divide the query rows into groups by the given columns.
// Note that, standard SQL spec allows partition only by columns, and in order to
// use the "expression" version, use the PartitionByExpr.
func (w *WindowBuilder) PartitionBy(columns ...string) *WindowBuilder {
	w.partition = func(b *Builder) {
		b.IdentComma(columns...)
	}
	return w
}

// PartitionExpr indicates to divide the query rows into groups by the given expression.
func (w *WindowBuilder) PartitionExpr(x Querier) *WindowBuilder {
	w.partition = func(b *Builder) {
		b.Join(x)
	}
	return w
}

// OrderBy indicates how to sort rows in each partition.
func (w *WindowBuilder) OrderBy(columns ...string) *WindowBuilder {
	for i := range columns {
		w.order = append(w.order, columns[i])
	}
	return w
}

// OrderExpr appends the `ORDER BY` clause to the window
// partition with custom list of expressions.
func (w *WindowBuilder) OrderExpr(exprs ...Querier) *WindowBuilder {
	for i := range exprs {
		w.order = append(w.order, exprs[i])
	}
	return w
}

// Query returns query representation of the window function.
func (w *WindowBuilder) Query() (string, []any) {
	w.fn(&w.Builder)
	w.WriteString(" OVER ")
	w.Wrap(func(b *Builder) {
		if w.partition != nil {
			b.WriteString("PARTITION BY ")
			w.partition(b)
		}
		joinOrder(w.order, b)
	})
	return w.Builder.String(), w.args
}
