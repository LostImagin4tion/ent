package query

// Querier wraps the basic Query method that is implemented
// by the different builders in this file.
type Querier interface {
	// Query returns the query representation of the element
	// and its arguments (if any).
	Query() (string, []any)
}

// querierErr allowed propagate Querier's inner error
type querierErr interface {
	Err() error
}

// Queries are list of queries join with space between them.
type Queries []Querier

// Query returns query representation of Queriers.
func (n Queries) Query() (string, []any) {
	b := &Builder{}
	for i := range n {
		if i > 0 {
			b.Pad()
		}
		query, args := n[i].Query()
		b.WriteString(query)
		b.args = append(b.args, args...)
	}
	return b.String(), b.args
}

// Raw returns a raw SQL query that is placed as-is in the query.
func Raw(s string) Querier { return &raw{s} }

type raw struct{ s string }

func (r *raw) Query() (string, []any) { return r.s, nil }
