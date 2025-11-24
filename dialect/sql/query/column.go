package query

// ColumnBuilder is a builder for column definition in table creation.
type ColumnBuilder struct {
	Builder
	typ  string // column type.
	name string // column name.
}

// Column returns a new ColumnBuilder with the given name.
//
//	sql.Column("group_id").Type("int").Attr("UNIQUE")
func Column(name string) *ColumnBuilder { return &ColumnBuilder{name: name} }

// Type sets the column type.
func (c *ColumnBuilder) Type(t string) *ColumnBuilder {
	c.typ = t
	return c
}

// Query returns query representation of a Column.
func (c *ColumnBuilder) Query() (string, []any) {
	c.Ident(c.name)
	if c.typ != "" {
		c.Pad().WriteString(c.typ)
	}
	return c.String(), c.args
}
