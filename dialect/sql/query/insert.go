package query

import (
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
)

// InsertBuilder is a builder for `INSERT INTO` statement.
type InsertBuilder struct {
	Builder
	table     string
	schema    string
	columns   []string
	defaults  bool
	returning []string
	values    [][]any
	conflict  *conflict
}

// Insert creates a builder for the `INSERT INTO` statement.
//
//	Insert("users").
//		Columns("name", "age").
//		Values("a8m", 10).
//		Values("foo", 20)
//
// Note: Insert inserts all values in one batch.
func Insert(table string) *InsertBuilder { return &InsertBuilder{table: table} }

// Schema sets the database name for the insert table.
func (i *InsertBuilder) Schema(name string) *InsertBuilder {
	i.schema = name
	return i
}

// Set is a syntactic sugar API for inserting only one row.
func (i *InsertBuilder) Set(column string, v any) *InsertBuilder {
	i.columns = append(i.columns, column)
	if len(i.values) == 0 {
		i.values = append(i.values, []any{v})
	} else {
		i.values[0] = append(i.values[0], v)
	}
	return i
}

// Columns appends columns to the INSERT statement.
func (i *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	i.columns = append(i.columns, columns...)
	return i
}

// Values append a value tuple for the insert statement.
func (i *InsertBuilder) Values(values ...any) *InsertBuilder {
	i.values = append(i.values, values)
	return i
}

// Default sets the default values clause based on the dialect type.
func (i *InsertBuilder) Default() *InsertBuilder {
	i.defaults = true
	return i
}

// Returning adds the `RETURNING` clause to the insert statement.
// Supported by SQLite and PostgreSQL.
func (i *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	i.returning = columns
	return i
}

type (
	// conflict holds the configuration for the
	// `ON CONFLICT` / `ON DUPLICATE KEY` clause.
	conflict struct {
		target struct {
			constraint string
			columns    []string
			where      *Predicate
		}
		action struct {
			nothing bool
			where   *Predicate
			update  []func(*UpdateSet)
		}
	}

	// ConflictOption allows configuring the
	// conflict config using functional options.
	ConflictOption func(*conflict)
)

// ConflictColumns sets the unique constraints that trigger the conflict
// resolution on insert to perform an upsert operation. The columns must
// have a unique constraint applied to trigger this behaviour.
//
//	sql.Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			sql.ConflictColumns("id"),
//			sql.ResolveWithNewValues(),
//		)
func ConflictColumns(names ...string) ConflictOption {
	return func(c *conflict) {
		c.target.columns = names
	}
}

// ConflictConstraint allows setting the constraint
// name (i.e. `ON CONSTRAINT <name>`) for PostgreSQL.
//
//	sql.Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			sql.ConflictConstraint("users_pkey"),
//			sql.ResolveWithNewValues(),
//		)
func ConflictConstraint(name string) ConflictOption {
	return func(c *conflict) {
		c.target.constraint = name
	}
}

// ConflictWhere allows inference of partial unique indexes. See, PostgreSQL
// doc: https://www.postgresql.org/docs/current/sql-insert.html#SQL-ON-CONFLICT
func ConflictWhere(p *Predicate) ConflictOption {
	return func(c *conflict) {
		c.target.where = p
	}
}

// UpdateWhere allows setting the update condition. Only rows
// for which this expression returns true will be updated.
func UpdateWhere(p *Predicate) ConflictOption {
	return func(c *conflict) {
		c.action.where = p
	}
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported by SQLite and PostgreSQL.
//
//	sql.Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			sql.ConflictColumns("id"),
//			sql.DoNothing()
//		)
func DoNothing() ConflictOption {
	return func(c *conflict) {
		c.action.nothing = true
	}
}

// ResolveWithIgnore sets each column to itself to force an update and return the ID,
// otherwise does not change any data. This may still trigger update hooks in the database.
//
//	sql.Insert("users").
//		Columns("id").
//		Values(1).
//		OnConflict(
//			sql.ConflictColumns("id"),
//			sql.ResolveWithIgnore()
//		)
//
//	// Output:
//	// MySQL: INSERT INTO `users` (`id`) VALUES(1) ON DUPLICATE KEY UPDATE `id` = `users`.`id`
//	// PostgreSQL: INSERT INTO "users" ("id") VALUES(1) ON CONFLICT ("id") DO UPDATE SET "id" = "users"."id
func ResolveWithIgnore() ConflictOption {
	return func(c *conflict) {
		c.action.update = append(c.action.update, func(u *UpdateSet) {
			for _, c := range u.columns {
				u.SetIgnore(c)
			}
		})
	}
}

// ResolveWithNewValues updates columns using the new values proposed
// for insertion using the special EXCLUDED/VALUES table.
//
//	sql.Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			sql.ConflictColumns("id"),
//			sql.ResolveWithNewValues()
//		)
//
//	// Output:
//	// MySQL: INSERT INTO `users` (`id`, `name`) VALUES(1, 'Mashraki) ON DUPLICATE KEY UPDATE `id` = VALUES(`id`), `name` = VALUES(`name`),
//	// PostgreSQL: INSERT INTO "users" ("id") VALUES(1) ON CONFLICT ("id") DO UPDATE SET "id" = "excluded"."id, "name" = "excluded"."name"
func ResolveWithNewValues() ConflictOption {
	return func(c *conflict) {
		c.action.update = append(c.action.update, func(u *UpdateSet) {
			for _, c := range u.columns {
				u.SetExcluded(c)
			}
		})
	}
}

// ResolveWith allows setting a custom function to set the `UPDATE` clause.
//
//	Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			ConflictColumns("name"),
//			ResolveWith(func(u *UpdateSet) {
//				u.SetIgnore("id")
//				u.SetNull("created_at")
//				u.Set("name", Expr(u.Excluded().C("name")))
//			}),
//		)
func ResolveWith(fn func(*UpdateSet)) ConflictOption {
	return func(c *conflict) {
		c.action.update = append(c.action.update, fn)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	sql.Insert("users").
//		Columns("id", "name").
//		Values(1, "Mashraki").
//		OnConflict(
//			sql.ConflictColumns("id"),
//			sql.ResolveWithNewValues()
//		)
func (i *InsertBuilder) OnConflict(opts ...ConflictOption) *InsertBuilder {
	if i.conflict == nil {
		i.conflict = &conflict{}
	}
	for _, opt := range opts {
		opt(i.conflict)
	}
	return i
}

// UpdateSet describes a set of changes of the `DO UPDATE` clause.
type UpdateSet struct {
	*UpdateBuilder
	columns []string
}

// Table returns the table the `UPSERT` statement is executed on.
func (u *UpdateSet) Table() *SelectTable {
	return Dialect(u.UpdateBuilder.dialect).Table(u.UpdateBuilder.table)
}

// Columns returns all columns in the `INSERT` statement.
func (u *UpdateSet) Columns() []string {
	return u.columns
}

// UpdateColumns returns all columns in the `UPDATE` statement.
func (u *UpdateSet) UpdateColumns() []string {
	return append(u.UpdateBuilder.nulls, u.UpdateBuilder.columns...)
}

// Set sets a column to a given value.
func (u *UpdateSet) Set(column string, v any) *UpdateSet {
	u.UpdateBuilder.Set(column, v)
	return u
}

// Add adds a numeric value to the given column.
func (u *UpdateSet) Add(column string, v any) *UpdateSet {
	u.UpdateBuilder.Add(column, v)
	return u
}

// SetNull sets a column as null value.
func (u *UpdateSet) SetNull(column string) *UpdateSet {
	u.UpdateBuilder.SetNull(column)
	return u
}

// SetIgnore sets the column to itself. For example, "id" = "users"."id".
func (u *UpdateSet) SetIgnore(name string) *UpdateSet {
	return u.Set(name, Expr(u.Table().C(name)))
}

// SetExcluded sets the column name to its EXCLUDED/VALUES value.
// For example, "c" = "excluded"."c", or `c` = VALUES(`c`).
func (u *UpdateSet) SetExcluded(name string) *UpdateSet {
	switch u.UpdateBuilder.Dialect() {
	case dialect.MySQL:
		u.UpdateBuilder.Set(name, ExprFunc(func(b *Builder) {
			b.WriteString("VALUES(").Ident(name).WriteByte(')')
		}))
	default:
		t := Dialect(u.UpdateBuilder.dialect).Table("excluded")
		u.UpdateBuilder.Set(name, Expr(t.C(name)))
	}
	return u
}

// Query returns query representation of an `INSERT INTO` statement.
func (i *InsertBuilder) Query() (string, []any) {
	query, args, _ := i.QueryErr()
	return query, args
}

// QueryErr returns query representation of an `INSERT INTO`
// statement and any error occurred in building the statement.
func (i *InsertBuilder) QueryErr() (string, []any, error) {
	b := i.Builder.clone()
	b.WriteString("INSERT INTO ")
	b.writeSchema(i.schema)
	b.Ident(i.table).Pad()
	if i.defaults && len(i.columns) == 0 {
		i.writeDefault(&b)
	} else {
		b.WriteByte('(').IdentComma(i.columns...).WriteByte(')')
		b.WriteString(" VALUES ")
		for j, v := range i.values {
			if j > 0 {
				b.Comma()
			}
			b.WriteByte('(').Args(v...).WriteByte(')')
		}
	}
	if i.conflict != nil {
		i.writeConflict(&b)
	}
	joinReturning(i.returning, &b)
	return b.String(), b.args, b.Err()
}

func (i *InsertBuilder) writeDefault(b *Builder) {
	switch i.Dialect() {
	case dialect.MySQL:
		b.WriteString("VALUES ()")
	case dialect.SQLite, dialect.Postgres:
		b.WriteString("DEFAULT VALUES")
	}
}

func (i *InsertBuilder) writeConflict(b *Builder) {
	switch i.Dialect() {
	case dialect.MySQL:
		b.WriteString(" ON DUPLICATE KEY UPDATE ")
		// Fallback to ResolveWithIgnore() as MySQL
		// does not support the "DO NOTHING" clause.
		if i.conflict.action.nothing {
			i.OnConflict(ResolveWithIgnore())
		}
	case dialect.SQLite, dialect.Postgres:
		b.WriteString(" ON CONFLICT")
		switch t := i.conflict.target; {
		case t.constraint != "" && len(t.columns) != 0:
			b.AddError(fmt.Errorf("duplicate CONFLICT clauses: %q, %q", t.constraint, t.columns))
		case t.constraint != "":
			b.WriteString(" ON CONSTRAINT ").Ident(t.constraint)
		case len(t.columns) != 0:
			b.WriteString(" (").IdentComma(t.columns...).WriteByte(')')
		}
		if p := i.conflict.target.where; p != nil {
			b.WriteString(" WHERE ").Join(p)
		}
		if i.conflict.action.nothing {
			b.WriteString(" DO NOTHING")
			return
		}
		b.WriteString(" DO UPDATE SET ")
	}
	if len(i.conflict.action.update) == 0 {
		b.AddError(errors.New("missing action for 'DO UPDATE SET' clause"))
	}
	u := &UpdateSet{UpdateBuilder: Dialect(i.dialect).Update(i.table), columns: i.columns}
	u.Builder = *b
	for _, f := range i.conflict.action.update {
		f(u)
	}
	u.writeSetter(b)
	if p := i.conflict.action.where; p != nil {
		p.qualifier = i.table
		b.WriteString(" WHERE ").Join(p)
	}
}
