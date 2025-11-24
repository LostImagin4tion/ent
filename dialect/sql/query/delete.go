package query

// DeleteBuilder is a builder for `DELETE` statement.
type DeleteBuilder struct {
	Builder
	table  string
	schema string
	where  *Predicate
}

// Delete creates a builder for the `DELETE` statement.
//
//	Delete("users").
//		Where(
//			Or(
//				EQ("name", "foo").And().EQ("age", 10),
//				EQ("name", "bar").And().EQ("age", 20),
//				And(
//					EQ("name", "qux"),
//					EQ("age", 1).Or().EQ("age", 2),
//				),
//			),
//		)
func Delete(table string) *DeleteBuilder { return &DeleteBuilder{table: table} }

// Schema sets the database name for the table whose row will be deleted.
func (d *DeleteBuilder) Schema(name string) *DeleteBuilder {
	d.schema = name
	return d
}

// Where appends a where predicate to the `DELETE` statement.
func (d *DeleteBuilder) Where(p *Predicate) *DeleteBuilder {
	if d.where != nil {
		d.where = And(d.where, p)
	} else {
		d.where = p
	}
	return d
}

// FromSelect makes it possible to delete a sub query.
func (d *DeleteBuilder) FromSelect(s *Selector) *DeleteBuilder {
	d.Where(s.where)
	if t := s.Table(); t != nil {
		d.table = t.name
	}
	return d
}

// Query returns query representation of a `DELETE` statement.
func (d *DeleteBuilder) Query() (string, []any) {
	d.WriteString("DELETE FROM ")
	d.writeSchema(d.schema)
	d.Ident(d.table)
	if d.where != nil {
		d.WriteString(" WHERE ")
		d.Join(d.where)
	}
	return d.String(), d.args
}
