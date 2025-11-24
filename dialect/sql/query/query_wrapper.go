package query

import "fmt"

// Wrapper wraps a given Querier with different format.
// Used to prefix/suffix other queries.
type Wrapper struct {
	format  string
	wrapped Querier
}

// Query returns query representation of a wrapped Querier.
func (w *Wrapper) Query() (string, []any) {
	query, args := w.wrapped.Query()
	return fmt.Sprintf(w.format, query), args
}

// SetDialect calls SetDialect on the wrapped query.
func (w *Wrapper) SetDialect(name string) {
	if s, ok := w.wrapped.(state); ok {
		s.SetDialect(name)
	}
}

// Dialect calls Dialect on the wrapped query.
func (w *Wrapper) Dialect() string {
	if s, ok := w.wrapped.(state); ok {
		return s.Dialect()
	}
	return ""
}

// Total returns the total number of arguments so far.
func (w *Wrapper) Total() int {
	if s, ok := w.wrapped.(state); ok {
		return s.Total()
	}
	return 0
}

// SetTotal sets the value of the total arguments.
// Used to pass this information between sub queries/expressions.
func (w *Wrapper) SetTotal(total int) {
	if s, ok := w.wrapped.(state); ok {
		s.SetTotal(total)
	}
}
