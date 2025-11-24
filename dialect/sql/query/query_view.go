package query

// queryView allows using Querier (expressions) in the FROM clause.
type queryView struct{ Querier }

func (*queryView) view() {}

func (q *queryView) C(column string) string {
	if tv, ok := q.Querier.(TableView); ok {
		return tv.C(column)
	}
	return column
}
