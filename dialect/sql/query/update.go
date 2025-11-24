package query

import (
	"errors"
	"strconv"
)

// UpdateBuilder is a builder for `UPDATE` statement.
type UpdateBuilder struct {
	Builder
	table     string
	schema    string
	where     *Predicate
	nulls     []string
	columns   []string
	returning []string
	values    []any
	order     []any
	limit     *int
	prefix    Queries
}

// Update creates a builder for the `UPDATE` statement.
//
//	Update("users").Set("name", "foo").Set("age", 10)
func Update(table string) *UpdateBuilder { return &UpdateBuilder{table: table} }

// Schema sets the database name for the updated table.
func (u *UpdateBuilder) Schema(name string) *UpdateBuilder {
	u.schema = name
	return u
}

// Set sets a column to a given value. If `Set` was called before with
// the same column name, it overrides the value of the previous call.
func (u *UpdateBuilder) Set(column string, v any) *UpdateBuilder {
	for i := range u.columns {
		if column == u.columns[i] {
			u.values[i] = v
			return u
		}
	}
	u.columns = append(u.columns, column)
	u.values = append(u.values, v)
	return u
}

// Add adds a numeric value to the given column. Note that, calling Set(c)
// after Add(c) will erase previous calls with c from the builder.
func (u *UpdateBuilder) Add(column string, v any) *UpdateBuilder {
	u.columns = append(u.columns, column)
	u.values = append(u.values, ExprFunc(func(b *Builder) {
		b.WriteString("COALESCE")
		b.Wrap(func(b *Builder) {
			b.Ident(Table(u.table).C(column)).Comma().WriteByte('0')
		})
		b.WriteString(" + ")
		b.Arg(v)
	}))
	return u
}

// SetNull sets a column as null value.
func (u *UpdateBuilder) SetNull(column string) *UpdateBuilder {
	u.nulls = append(u.nulls, column)
	return u
}

// Where adds a where predicate for update statement.
func (u *UpdateBuilder) Where(p *Predicate) *UpdateBuilder {
	if u.where != nil {
		u.where = And(u.where, p)
	} else {
		u.where = p
	}
	return u
}

// FromSelect makes it possible to update entities that match the sub-query.
func (u *UpdateBuilder) FromSelect(s *Selector) *UpdateBuilder {
	u.Where(s.where)
	if t := s.Table(); t != nil {
		u.table = t.name
	}
	return u
}

// Empty reports whether this builder does not contain update changes.
func (u *UpdateBuilder) Empty() bool {
	return len(u.columns) == 0 && len(u.nulls) == 0
}

// OrderBy appends the `ORDER BY` clause to the `UPDATE` statement.
// Supported by SQLite and MySQL.
func (u *UpdateBuilder) OrderBy(columns ...string) *UpdateBuilder {
	if u.postgres() {
		u.AddError(errors.New("ORDER BY is not supported by PostgreSQL"))
		return u
	}
	for i := range columns {
		u.order = append(u.order, columns[i])
	}
	return u
}

// Limit appends the `LIMIT` clause to the `UPDATE` statement.
// Supported by SQLite and MySQL.
func (u *UpdateBuilder) Limit(limit int) *UpdateBuilder {
	if u.postgres() {
		u.AddError(errors.New("LIMIT is not supported by PostgreSQL"))
		return u
	}
	u.limit = &limit
	return u
}

// Prefix prefixes the UPDATE statement with list of statements.
func (u *UpdateBuilder) Prefix(stmts ...Querier) *UpdateBuilder {
	u.prefix = append(u.prefix, stmts...)
	return u
}

// Returning adds the `RETURNING` clause to the insert statement.
// Supported by SQLite and PostgreSQL.
func (u *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	u.returning = columns
	return u
}

// Query returns query representation of an `UPDATE` statement.
func (u *UpdateBuilder) Query() (string, []any) {
	b := u.Builder.clone()
	if len(u.prefix) > 0 {
		b.join(u.prefix, " ")
		b.Pad()
	}
	b.WriteString("UPDATE ")
	b.writeSchema(u.schema)
	b.Ident(u.table).WriteString(" SET ")
	u.writeSetter(&b)
	if u.where != nil {
		b.WriteString(" WHERE ")
		b.Join(u.where)
	}
	joinReturning(u.returning, &b)
	joinOrder(u.order, &b)
	if u.limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.Itoa(*u.limit))
	}
	return b.String(), b.args
}

// writeSetter writes the "SET" clause for the UPDATE statement.
func (u *UpdateBuilder) writeSetter(b *Builder) {
	for i, c := range u.nulls {
		if i > 0 {
			b.Comma()
		}
		b.Ident(c).WriteString(" = NULL")
	}
	if len(u.nulls) > 0 && len(u.columns) > 0 {
		b.Comma()
	}
	for i, c := range u.columns {
		if i > 0 {
			b.Comma()
		}
		b.Ident(c).WriteString(" = ")
		switch v := u.values[i].(type) {
		case Querier:
			b.Join(v)
		default:
			b.Arg(v)
		}
	}
}
